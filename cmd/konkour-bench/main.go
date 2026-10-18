package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"konkour-bench/internal/bench"
	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/report"
	"konkour-bench/internal/storage"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, os.Args[1:]); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}
func output(v any) error { e := json.NewEncoder(os.Stdout); e.SetIndent("", "  "); return e.Encode(v) }
func list(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			result = append(result, p)
		}
	}
	return result
}
func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: konkour-bench <validate|models|run|report|inspect|reconcile> [flags]; use COMMAND --help")
	}
	command := args[0]
	valid := map[string]bool{"validate": true, "models": true, "run": true, "report": true, "inspect": true, "reconcile": true}
	if command == "--help" || command == "help" {
		fmt.Println("Konkour Bench: validate, models, run, report, inspect, reconcile. Use COMMAND --help for flags.")
		return nil
	}
	if !valid[command] {
		return fmt.Errorf("unknown command %q", command)
	}
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	config := f.String("config", "benchmark.json", "dataset/model configuration JSON")
	dbPath := f.String("db", "data/bench.db", "SQLite storage path")
	base := f.String("base-url", "https://openrouter.ai/api/v1", "OpenRouter API base URL")
	models := f.String("model", "", "comma-separated configured model IDs (default: all)")
	efforts := f.String("effort", "", "comma-separated minimal,medium,max (default: all)")
	questions := f.String("question", "", "comma-separated dataset:number IDs (default: all)")
	concurrency := f.Int("concurrency", 4, "maximum concurrent calls; one per model")
	timeout := f.Duration("timeout", 0, "optional generation timeout; zero means no deadline")
	retryUnknown := f.Bool("retry-unknown", false, "explicitly allow resending uncertain requests; may incur duplicate charges")
	runID := f.String("run", "", "report a saved run ID or 'latest'; default uses current dataset and compatible history")
	dir := f.String("out", "reports/latest", "report output directory")
	hash := f.String("hash", "", "request hash to inspect")
	jsonLogs := f.Bool("log-json", false, "write structured JSON logs to stderr")
	if err := f.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", f.Args())
	}
	if *jsonLogs {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	}
	client := openrouter.New(*base, os.Getenv("OPENROUTER_API_KEY"))
	if command == "run" || command == "reconcile" {
		if client.Key == "" {
			return errors.New("OPENROUTER_API_KEY must be set in the environment (see README for loading .env)")
		}
	}
	if *timeout < 0 {
		return errors.New("timeout cannot be negative")
	}
	if command == "validate" || command == "models" || command == "run" {
		snap, err := dataset.Load(*config)
		if err != nil {
			return err
		}
		for _, warning := range snap.Warnings {
			slog.Warn(warning)
		}
		if command == "validate" {
			images := 0
			categories := map[string]int{}
			for _, q := range snap.Questions {
				if q.HasImage() {
					images++
				}
				categories[q.Category]++
			}
			return output(map[string]any{"questions": len(snap.Questions), "text_only": len(snap.Questions) - images, "with_images": images, "categories": categories, "models": len(snap.Config.Models), "revision": dataset.Hash(snap), "warnings": snap.Warnings})
		}
		if command == "models" {
			selected := list(*models)
			if len(selected) == 0 {
				selected = snap.Config.Models
			}
			cat, err := client.Discover(ctx, selected)
			if err != nil {
				return err
			}
			var rows []map[string]any
			for _, id := range selected {
				m, exists := cat.Models[strings.TrimSuffix(id, ":exacto")]
				tiers := map[string]any{}
				for _, tier := range openrouter.Tiers {
					choices, note := m.Candidates(tier)
					tiers[tier] = map[string]any{"values": choices, "evidence": note}
				}
				rows = append(rows, map[string]any{"model": id, "available": exists, "images": m.Images(), "efforts": tiers})
			}
			return output(rows)
		}
		s, err := storage.Open(*dbPath)
		if err != nil {
			return err
		}
		defer s.Close()
		r := bench.Runner{Store: s, Client: client, Log: slog.Default()}
		summary, runErr := r.Run(ctx, snap, bench.Options{Selection: storage.Selection{Models: list(*models), Efforts: list(*efforts), Questions: list(*questions)}, Concurrency: *concurrency, Timeout: *timeout, RetryUnknown: *retryUnknown})
		if err := output(summary); err != nil {
			return err
		}
		return runErr
	}
	s, err := storage.Open(*dbPath)
	if err != nil {
		return err
	}
	defer s.Close()
	switch command {
	case "inspect":
		if *hash == "" {
			return errors.New("inspect requires --hash")
		}
		req, err := s.Request(*hash)
		if err != nil {
			return err
		}
		attempts, err := s.Attempts(*hash, "")
		if err != nil {
			return err
		}
		return output(map[string]any{"request": req, "attempts": attempts})
	case "reconcile":
		n, err := bench.Reconcile(ctx, s, client)
		if outErr := output(map[string]any{"reconciled": n}); outErr != nil {
			return outErr
		}
		return err
	case "report":
		var snap dataset.Snapshot
		var sel storage.Selection
		if *runID != "" {
			var saved storage.Run
			if *runID == "latest" {
				saved, err = s.LatestRun()
			} else {
				saved, err = s.Run(*runID)
			}
			if err != nil {
				return err
			}
			if saved.Protocol != bench.Protocol() {
				return errors.New("saved run uses another protocol; reproduce it with the matching code version")
			}
			snap, sel, *runID, *base = saved.Snapshot, saved.Selection, saved.ID, saved.BaseURL
		} else {
			snap, err = dataset.Load(*config)
			if err != nil {
				return err
			}
		}
		doc, err := report.Build(s, snap, *runID, *base, sel)
		if err != nil {
			return err
		}
		if err := report.Write(*dir, doc); err != nil {
			return err
		}
		return output(map[string]any{"report": *dir + "/index.html", "scoring_revision": doc.ScoringRevision, "rows": len(doc.Rows), "outcomes": len(doc.Outcomes)})
	}
	return nil
}
