package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"konkour-bench/internal/bench"
	"konkour-bench/internal/dashboard"
	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/report"
	"konkour-bench/internal/storage"
	"konkour-bench/internal/supervisor"
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
		return errors.New("usage: konkour-bench <validate|models|run|status|supervisor|report|serve|inspect|reconcile> [flags]; use COMMAND --help")
	}
	command := args[0]
	valid := map[string]bool{"validate": true, "models": true, "run": true, "status": true, "supervisor": true, "report": true, "serve": true, "inspect": true, "reconcile": true}
	if command == "--help" || command == "help" {
		fmt.Println("Konkour Bench: validate, models, run, status, supervisor, report, serve, inspect, reconcile. Use COMMAND --help for flags.")
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
	timeout := f.Duration("timeout", time.Hour, "per-request emergency deadline; zero disables it")
	warnAfter := f.Duration("warn-after", 10*time.Minute, "log a slow-request warning; zero disables it")
	budget := f.Float64("budget-usd", 50, "new-attempt spending threshold; stops dispatch, not in-flight billing; zero disables it")
	envFile := f.String("env-file", "", "read a literal OPENROUTER_API_KEY assignment from this file")
	serviceFile := f.String("service-file", "data/supervisor/konkour-bench.plist", "launchd configuration to prepare without starting")
	attemptID := f.Int64("attempt", 0, "local attempt ID for reconciliation or stream inspection")
	generationID := f.String("generation-id", "", "OpenRouter generation ID to reconcile with --attempt")
	retryUnknown := f.Bool("retry-unknown", false, "explicitly allow resending uncertain requests; may incur duplicate charges")
	runID := f.String("run", "", "report a saved run ID or 'latest'; default uses current dataset and compatible history")
	dir := f.String("out", "reports/latest", "report output directory")
	hash := f.String("hash", "", "request hash to inspect")
	jsonLogs := f.Bool("log-json", false, "write structured JSON logs to stderr")
	address := f.String("addr", "127.0.0.1:8080", "dashboard listen address")
	assets := f.String("assets", "dashboard/dist", "built React dashboard directory")
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
	if *timeout < 0 || *warnAfter < 0 || *budget < 0 || math.IsNaN(*budget) || math.IsInf(*budget, 0) {
		return errors.New("budget and durations must be finite and nonnegative")
	}
	if command == "supervisor" {
		if runtime.GOOS != "darwin" {
			return errors.New("supervisor preparation currently supports macOS launchd")
		}
		if *envFile == "" {
			return errors.New("supervisor requires --env-file; secrets are never embedded in the plist")
		}
		abs := func(p string) string {
			v, err := filepath.Abs(p)
			if err != nil {
				return p
			}
			return v
		}
		binary, err := os.Executable()
		if err != nil {
			return err
		}
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		args := []string{"run", "--config", abs(*config), "--db", abs(*dbPath), "--base-url", *base, "--env-file", abs(*envFile), "--timeout", timeout.String(), "--warn-after", warnAfter.String(), "--budget-usd", strconv.FormatFloat(*budget, 'g', -1, 64), "--concurrency", strconv.Itoa(*concurrency), "--out", abs(*dir), "--log-json"}
		for _, pair := range [][2]string{{"--model", *models}, {"--effort", *efforts}, {"--question", *questions}} {
			if pair[1] != "" {
				args = append(args, pair[0], pair[1])
			}
		}
		if *retryUnknown {
			args = append(args, "--retry-unknown")
		}
		if err := supervisor.Prepare(abs(*serviceFile), binary, cwd, args); err != nil {
			return err
		}
		return output(map[string]any{"service_file": abs(*serviceFile), "label": supervisor.Label, "started": false})
	}
	key := os.Getenv("OPENROUTER_API_KEY")
	if *envFile != "" && (command == "run" || command == "reconcile") {
		var err error
		key, err = keyFromFile(*envFile)
		if err != nil {
			return err
		}
	}
	client := openrouter.New(*base, key)
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
		summary, runErr := r.Run(ctx, snap, bench.Options{Selection: storage.Selection{Models: list(*models), Efforts: list(*efforts), Questions: list(*questions)}, Concurrency: *concurrency, Timeout: *timeout, WarnAfter: *warnAfter, BudgetUSD: *budget, RetryUnknown: *retryUnknown})
		if summary.RunID != "" {
			saved, err := s.Run(summary.RunID)
			if err == nil {
				var doc report.Document
				doc, err = report.Build(s, saved.Snapshot, saved.ID, saved.BaseURL, saved.Selection)
				if err == nil {
					err = report.Write(*dir, doc)
				}
			}
			if err != nil {
				runErr = errors.Join(runErr, fmt.Errorf("final report: %w", err))
			} else {
				slog.Info("report saved", "path", filepath.Join(*dir, "index.html"))
			}
		}
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
	case "status":
		id := *runID
		if id == "latest" {
			saved, err := s.LatestRun()
			if err != nil {
				return err
			}
			id = saved.ID
		}
		activity, err := s.Activity(id)
		if err != nil {
			return err
		}
		return output(map[string]any{"run_id": id, "requests": activity})
	case "serve":
		return dashboard.Serve(ctx, dashboard.Source{Store: s, Config: *config, RunID: *runID, BaseURL: *base}, *address, *assets)
	case "inspect":
		if *attemptID > 0 {
			events, err := s.Events(*attemptID)
			if err != nil {
				return err
			}
			return output(map[string]any{"attempt": *attemptID, "stream_events": events})
		}
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
		if *attemptID != 0 || *generationID != "" {
			if err := bench.ReconcileID(ctx, s, client, *attemptID, *generationID); err != nil {
				return err
			}
			return output(map[string]any{"reconciled_attempt": *attemptID, "generation_id": *generationID})
		}
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
