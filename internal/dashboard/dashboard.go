// Package dashboard serves the React dashboard and locally scored benchmark data.
package dashboard

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"konkour-bench/internal/bench"
	"konkour-bench/internal/dataset"
	"konkour-bench/internal/report"
	"konkour-bench/internal/storage"
)

type Option struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
	Image  string `json:"image,omitempty"`
}
type Question struct {
	ID      string   `json:"id"`
	Number  int      `json:"number"`
	Text    string   `json:"text"`
	Image   string   `json:"image,omitempty"`
	Options []Option `json:"options"`
}
type Document struct {
	report.Document
	Questions []Question `json:"questions"`
}

type Source struct {
	Store                  *storage.Store
	Config, RunID, BaseURL string
}

func (s Source) Load() (Document, error) {
	var snap dataset.Snapshot
	var sel storage.Selection
	var err error
	runID, base := s.RunID, s.BaseURL
	if runID == "" {
		snap, err = dataset.Load(s.Config)
	} else {
		var saved storage.Run
		if runID == "latest" {
			saved, err = s.Store.LatestRun()
		} else {
			saved, err = s.Store.Run(runID)
		}
		if err == nil {
			if saved.Protocol != bench.ProtocolFor(saved.BaseURL) {
				return Document{}, errors.New("saved run uses another protocol")
			}
			snap, sel, runID, base = saved.Snapshot, saved.Selection, saved.ID, saved.BaseURL
		}
	}
	if err != nil {
		return Document{}, err
	}
	doc, err := report.Build(s.Store, snap, runID, base, sel)
	if err != nil {
		return Document{}, err
	}
	result := Document{Document: doc, Questions: []Question{}}
	selected := map[string]bool{}
	for _, o := range doc.Outcomes {
		selected[o.Question] = true
	}
	for _, q := range snap.Questions {
		if !selected[q.ID()] {
			continue
		}
		question := Question{ID: q.ID(), Number: q.Number, Text: q.Text, Image: asset(q.Asset), Options: []Option{}}
		for _, o := range q.Options {
			question.Options = append(question.Options, Option{Number: o.Number, Text: o.Text, Image: asset(o.Asset)})
		}
		result.Questions = append(result.Questions, question)
	}
	return result, nil
}

func asset(a *dataset.Asset) string {
	if a == nil {
		return ""
	}
	return "data:" + a.MIME + ";base64," + base64.StdEncoding.EncodeToString(a.Bytes)
}

// Handler caches a serialized snapshot briefly to bound concurrent database work.
// The only filesystem exposed to clients is the built frontend's directory.
func Handler(load func() (Document, error), assets http.Handler) http.Handler {
	var mu sync.Mutex
	var body []byte
	var expires time.Time
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/report", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if body == nil || time.Now().After(expires) {
			doc, err := load()
			if err == nil {
				body, err = json.Marshal(doc)
			}
			if err != nil {
				slog.Error("dashboard report failed", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"Unable to load benchmark results. Check the server log."}`))
				return
			}
			expires = time.Now().Add(5 * time.Second)
		}
		_, _ = w.Write(body)
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/assets/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/" {
			w.Header().Set("Cache-Control", "no-cache")
		}
		assets.ServeHTTP(w, r)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		mux.ServeHTTP(w, r)
	})
}

func Serve(ctx context.Context, source Source, address, assets string) error {
	root, err := os.OpenRoot(assets)
	if err != nil {
		return fmt.Errorf("open dashboard assets (run make dashboard): %w", err)
	}
	defer root.Close()
	if _, err := root.Stat("index.html"); err != nil {
		return fmt.Errorf("dashboard is not built; run make dashboard: %w", err)
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	server := &http.Server{Handler: Handler(source.Load, http.FileServer(http.FS(root.FS()))), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
	served := make(chan error, 1)
	go func() {
		served <- server.Serve(listener)
	}()
	slog.Info("dashboard listening", "url", "http://"+listener.Addr().String())
	select {
	case err = <-served:
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			if closeErr := server.Close(); closeErr != nil {
				slog.Error("dashboard close failed", "error", closeErr)
			}
			return err
		}
		err = <-served
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
