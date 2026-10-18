package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"konkour-bench/internal/bench"
	"konkour-bench/internal/dataset"
	"konkour-bench/internal/storage"
)

func TestSourceLoadsQuestionsAndSavedSelection(t *testing.T) {
	s, err := storage.Open(filepath.Join(t.TempDir(), "bench.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	source := Source{Store: s, Config: "../../benchmark.sample.json", BaseURL: "https://test.invalid"}
	doc, err := source.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Questions) != 15 || len(doc.Outcomes) != 675 {
		t.Fatalf("questions=%d outcomes=%d", len(doc.Questions), len(doc.Outcomes))
	}
	images := 0
	for _, q := range doc.Questions {
		if q.Image != "" {
			images++
		}
		for _, o := range q.Options {
			if o.Image != "" {
				images++
			}
		}
	}
	if images == 0 {
		t.Fatal("question images missing")
	}
	for _, row := range doc.Rows {
		if row.Accuracy != nil || row.Status != "incomplete" {
			t.Fatal("empty database reported a ranked score")
		}
	}
	snap, err := dataset.Load(source.Config)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.AddRun(storage.Run{ID: "saved", Protocol: bench.Protocol(), Snapshot: snap, BaseURL: source.BaseURL, Selection: storage.Selection{Models: snap.Config.Models[:1], Efforts: []string{"medium"}, Questions: []string{"sample:2"}}}); err != nil {
		t.Fatal(err)
	}
	source.RunID = "latest"
	doc, err = source.Load()
	if err != nil {
		t.Fatal(err)
	}
	if doc.RunID != "saved" || len(doc.Questions) != 1 || doc.Questions[0].ID != "sample:2" || len(doc.Outcomes) != 1 {
		t.Fatalf("saved selection not preserved: %+v", doc.Questions)
	}
}

func TestReportHandlerAndStaticBoundary(t *testing.T) {
	loads := 0
	h := Handler(func() (Document, error) {
		loads++
		return Document{Questions: []Question{{ID: "exam:1", Text: "مقدار $x^2$؟", Image: asset(&dataset.Asset{MIME: "image/png", Bytes: []byte{1, 2}})}}}, nil
	}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("frontend")) }))
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/api/report", nil))
		var doc Document
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &doc) != nil || len(doc.Questions) != 1 || doc.Questions[0].Image != "data:image/png;base64,AQI=" {
			t.Fatalf("invalid response %s", w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("report must not be browser-cached")
		}
	}
	if loads != 1 {
		t.Fatal("concurrent refreshes should share a recent snapshot")
	}
	for path, want := range map[string]int{"/": 200, "/assets/app.js": 200, "/.env": 404, "/data/bench.db": 404, "/api/missing": 404} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != want {
			t.Errorf("%s: %d", path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/api/report", nil))
	if w.Code != 405 {
		t.Fatalf("POST accepted: %d", w.Code)
	}
}

func TestHandlerFailureIsActionableAndDoesNotExposeInternalError(t *testing.T) {
	h := Handler(func() (Document, error) { return Document{}, errors.New("private filesystem detail") }, http.NotFoundHandler())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/report", nil))
	if w.Code != 500 || strings.Contains(w.Body.String(), "private filesystem") || !strings.Contains(w.Body.String(), "server log") {
		t.Fatalf("bad error response %s", w.Body.String())
	}
}
