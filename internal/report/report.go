// Package report scores stored generations against a versioned answer-key snapshot.
package report

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"konkour-bench/internal/backbone"
	"konkour-bench/internal/bench"
	"konkour-bench/internal/claude"
	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/storage"
)

type Outcome struct {
	storage.Job
	Dataset   string           `json:"dataset"`
	Category  string           `json:"category"`
	Section   string           `json:"section,omitempty"`
	Image     bool             `json:"has_image"`
	Expected  int              `json:"expected"`
	Answer    int              `json:"answer"`
	Correct   bool             `json:"correct"`
	Invalid   string           `json:"invalid,omitempty"`
	Provider  string           `json:"provider,omitempty"`
	Usage     openrouter.Usage `json:"usage"`
	Duration  *float64         `json:"duration_seconds"`
	OriginRun string           `json:"origin_run,omitempty"`
}
type Row struct {
	Model             string       `json:"model"`
	Tier              string       `json:"tier"`
	Cohort            string       `json:"cohort"`
	Dataset           string       `json:"dataset"`
	Category          string       `json:"category"`
	Status            string       `json:"status"`
	Total             int          `json:"total"`
	Complete          int          `json:"complete"`
	Correct           int          `json:"correct"`
	Incorrect         int          `json:"incorrect"`
	Unsupported       int          `json:"unsupported"`
	Skipped           int          `json:"skipped"`
	Failed            int          `json:"failed"`
	Accuracy          *float64     `json:"accuracy"`
	Cost              *float64     `json:"cost_usd"`
	CostPerQuestion   *float64     `json:"cost_per_question_usd"`
	Tokens            *int64       `json:"total_tokens"`
	TokensPerQuestion *float64     `json:"tokens_per_question"`
	Prompt            *int64       `json:"prompt_tokens"`
	Completion        *int64       `json:"completion_tokens"`
	Reasoning         *int64       `json:"reasoning_tokens"`
	CachedInput       *int64       `json:"cached_input_tokens"`
	SentEfforts       string       `json:"sent_efforts"`
	CostFrontier      bool         `json:"cost_frontier"`
	TokenFrontier     bool         `json:"token_frontier"`
	Pareto            bool         `json:"pareto_3d"`
	Interval          *Interval    `json:"accuracy_interval"`
	Bounds            *Interval    `json:"completion_bounds"`
	Invalid           int          `json:"invalid"`
	Latency           Distribution `json:"latency_seconds"`
	TokenDistribution Distribution `json:"token_distribution"`
	CostDistribution  Distribution `json:"cost_distribution"`
}
type Document struct {
	Backend         string                     `json:"backend"`
	CostBasis       string                     `json:"cost_basis"`
	Generated       time.Time                  `json:"generated"`
	ScoringRevision string                     `json:"scoring_revision"`
	Protocol        string                     `json:"protocol"`
	RunID           string                     `json:"run_id,omitempty"`
	Snapshot        dataset.Snapshot           `json:"-"`
	Rows            []Row                      `json:"rows"`
	Outcomes        []Outcome                  `json:"outcomes"`
	Datasets        []string                   `json:"datasets"`
	Categories      []string                   `json:"categories"`
	Models          []string                   `json:"models"`
	Attempts        []storage.Attempt          `json:"-"`
	Spending        []RunSpending              `json:"run_spending"`
	AnalysisVersion string                     `json:"analysis_version"`
	QuestionBank    []QuestionContent          `json:"question_bank"`
	DatasetMetadata map[string]json.RawMessage `json:"dataset_metadata"`
	Warnings        []string                   `json:"warnings"`
	Sources         []SourceRun                `json:"source_runs"`
	Views           []AnalysisView             `json:"analysis"`
}

type RunSpending struct {
	SubscriptionAttempts int     `json:"subscription_attempts,omitempty"`
	RunID                string  `json:"run_id"`
	Attempts             int     `json:"attempts"`
	KnownCost            float64 `json:"known_cost_usd"`
	KnownRetryCost       float64 `json:"known_retry_cost_usd"`
	UnknownCost          int     `json:"attempts_with_unknown_cost"`
}

func Build(s *storage.Store, snap dataset.Snapshot, runID, baseURL string, sel storage.Selection) (Document, error) {
	if baseURL == "claude-cli://latest" || baseURL == "mixed://latest" {
		saved, err := s.LatestRunForProtocol(bench.ProtocolFor(baseURL))
		if err != nil {
			return Document{}, fmt.Errorf("load latest %s run for report: %w", backbone.Name(baseURL), err)
		}
		baseURL = saved.BaseURL
	}
	d := Document{Generated: time.Now().UTC(), ScoringRevision: dataset.Hash(snap), Protocol: bench.ProtocolFor(baseURL), RunID: runID, Snapshot: snap, Backend: "openrouter", CostBasis: "reported API charges"}
	if backbone.IsClaude(baseURL) {
		d.Backend = "claude"
		d.CostBasis = "subscription usage; actual cost unavailable; API cost estimates are not charges and are excluded from cost rankings"
	}
	if backbone.IsMixed(baseURL) {
		d.Backend = "mixed"
		d.CostBasis = "OpenRouter charges for other models; Claude CLI subscription usage has no reported actual cost and is excluded from cost rankings"
	}
	if _, err := s.RecoverStale(); err != nil {
		return d, err
	}
	if len(sel.Models) == 0 {
		sel.Models = snap.Config.Models
		if backbone.IsClaude(baseURL) {
			sel.Models = nil
			for _, id := range snap.Config.Models {
				if claude.Supports(id) {
					sel.Models = append(sel.Models, id)
				}
			}
		}
	}
	if len(sel.Efforts) == 0 {
		sel.Efforts = openrouter.Tiers
	}
	d.Models = sel.Models
	jobs, err := s.Jobs(runID, d.Protocol)
	if err != nil {
		return d, err
	}
	lookup := map[string][]storage.Job{}
	for _, j := range jobs {
		key := j.Model + "\x00" + j.Tier + "\x00" + j.Question
		lookup[key] = append(lookup[key], j)
	}
	for _, q := range snap.Questions {
		if len(sel.Questions) > 0 && !slices.Contains(sel.Questions, q.ID()) {
			continue
		}
		if !slices.Contains(d.Datasets, q.Dataset) {
			d.Datasets = append(d.Datasets, q.Dataset)
		}
		if !slices.Contains(d.Categories, q.Category) {
			d.Categories = append(d.Categories, q.Category)
		}
		for _, model := range sel.Models {
			endpoint := backbone.EndpointFor(baseURL, model)
			for _, tier := range sel.Efforts {
				o := Outcome{Job: storage.Job{Question: q.ID(), Model: model, Tier: tier, Status: "not_run"}, Dataset: q.Dataset, Category: q.Category, Section: q.Section, Image: q.HasImage(), Expected: q.Correct}
				for _, j := range lookup[model+"\x00"+tier+"\x00"+q.ID()] {
					if j.Hash != "" {
						body, err := backbone.Build(baseURL, q, model, j.Effort, false)
						if err != nil {
							return d, err
						}
						if bench.RequestHash(endpoint, body) != j.Hash {
							streamBody, err := backbone.Build(baseURL, q, model, j.Effort, true)
							if err != nil {
								return d, err
							}
							if bench.RequestHash(endpoint, streamBody) != j.Hash {
								continue
							}
						}
					}
					if j.Status == "skipped" && !q.HasImage() {
						continue
					}
					o.Job = j
					if j.Status == "pending" {
						req, err := s.Request(j.Hash)
						if err != nil {
							return d, err
						}
						if req.State == "complete" && j.AttemptID != 0 && j.AttemptID == req.AttemptID {
							o.Status = "complete"
						} else if req.State != "pending" {
							o.Status = "infra_error"
							o.Note = req.Note
						}
					}
					if o.Status == "complete" {
						req, err := s.Request(j.Hash)
						if err != nil {
							return d, err
						}
						o.Answer = req.Result.Answer
						o.Correct = o.Answer == q.Correct
						o.Invalid = req.Result.Invalid
						o.Provider = req.Result.Generation.Provider
						o.Usage = req.Result.Generation.Usage
						o.AttemptID = req.AttemptID
					}
					break
				}
				d.Outcomes = append(d.Outcomes, o)
			}
		}
	}
	sort.Strings(d.Datasets)
	sort.Strings(d.Categories)
	if err := enrich(&d, s); err != nil {
		return d, err
	}
	// Each outcome contributes to eight groups, independent of dataset size.
	type groupKey struct{ Model, Tier, Cohort, Dataset, Category string }
	groups := map[groupKey][]Outcome{}
	for _, o := range d.Outcomes {
		cohort := "text-only"
		if o.Image {
			cohort = "image-only"
		}
		for _, ds := range []string{"", o.Dataset} {
			for _, category := range []string{"", o.Category} {
				for _, c := range []string{"full", cohort} {
					key := groupKey{o.Model, o.Tier, c, ds, category}
					groups[key] = append(groups[key], o)
				}
			}
		}
	}
	for key, group := range groups {
		row := Aggregate(group)
		row.Cohort = key.Cohort
		row.Dataset = key.Dataset
		row.Category = key.Category
		d.Rows = append(d.Rows, row)
	}
	Frontiers(d.Rows)
	sort.SliceStable(d.Rows, func(i, j int) bool {
		a, b := d.Rows[i], d.Rows[j]
		if a.Accuracy != nil && b.Accuracy == nil {
			return true
		}
		if a.Accuracy == nil && b.Accuracy != nil {
			return false
		}
		if a.Accuracy != nil && b.Accuracy != nil && *a.Accuracy != *b.Accuracy {
			return *a.Accuracy > *b.Accuracy
		}
		return strings.Join([]string{a.Cohort, a.Dataset, a.Category, a.Model, a.Tier}, "\x00") < strings.Join([]string{b.Cohort, b.Dataset, b.Category, b.Model, b.Tier}, "\x00")
	})
	d.Attempts, err = s.Attempts("", runID)
	if err != nil {
		return d, err
	}
	// A current-history view has one evaluation backend. Its spending ledger
	// must not make another backend's charges look like subscription usage.
	if runID == "" {
		compatible := map[string]bool{}
		var filtered []storage.Attempt
		for _, a := range d.Attempts {
			match, known := compatible[a.RunID]
			if !known {
				r, err := s.Run(a.RunID)
				if err != nil {
					return d, err
				}
				match = r.Protocol == d.Protocol
				compatible[a.RunID] = match
			}
			if match {
				filtered = append(filtered, a)
			}
		}
		d.Attempts = filtered
	}
	ledger := map[string]*RunSpending{}
	if runID != "" {
		ledger[runID] = &RunSpending{RunID: runID}
	}
	for _, a := range d.Attempts {
		entry := ledger[a.RunID]
		if entry == nil {
			entry = &RunSpending{RunID: a.RunID}
			ledger[a.RunID] = entry
		}
		entry.Attempts++
		if a.Details.Backend == "claude" {
			entry.SubscriptionAttempts++
			if backbone.IsMixed(baseURL) {
				continue
			}
		}
		u := a.Usage()
		if u.Cost == nil {
			entry.UnknownCost++
			continue
		}
		entry.KnownCost += *u.Cost
		req, err := s.Request(a.Hash)
		if err != nil {
			return d, err
		}
		if a.ID != req.AttemptID {
			entry.KnownRetryCost += *u.Cost
		}
	}
	for _, entry := range ledger {
		d.Spending = append(d.Spending, *entry)
	}
	sort.Slice(d.Spending, func(i, j int) bool { return d.Spending[i].RunID < d.Spending[j].RunID })
	d.Views = Analyze(d)
	return d, nil
}

func Aggregate(group []Outcome) Row {
	if len(group) == 0 {
		return Row{Status: "empty"}
	}
	r := Row{Model: group[0].Model, Tier: group[0].Tier, Total: len(group), Status: "complete"}
	var cost float64
	costKnown := true
	var tokens, prompt, completion, reasoning, cached int64
	tokensKnown, promptKnown, completionKnown, reasoningKnown, cachedKnown := true, true, true, true, true
	var efforts []string
	var durations, tokenValues, costValues []float64
	add := func(total *int64, known *bool, v *int64) {
		if v == nil {
			*known = false
		} else {
			*total += *v
		}
	}
	for _, o := range group {
		if o.Effort != "" && !slices.Contains(efforts, o.Effort) {
			efforts = append(efforts, o.Effort)
		}
		switch o.Status {
		case "complete":
			r.Complete++
			if o.Invalid != "" {
				r.Invalid++
			}
			if o.Duration != nil {
				durations = append(durations, *o.Duration)
			}
			if o.Usage.Total != nil {
				tokenValues = append(tokenValues, float64(*o.Usage.Total))
			}
			if o.Usage.Cost != nil {
				costValues = append(costValues, *o.Usage.Cost)
			}
			if o.Correct {
				r.Correct++
			} else {
				r.Incorrect++
			}
		case "unsupported":
			r.Unsupported++
		case "skipped":
			r.Skipped++
		default:
			r.Failed++
		}
		if o.Status != "complete" {
			costKnown = false
			tokensKnown = false
			promptKnown = false
			completionKnown = false
			reasoningKnown = false
			cachedKnown = false
			continue
		}
		if o.Usage.Cost == nil {
			costKnown = false
		} else {
			cost += *o.Usage.Cost
		}
		add(&tokens, &tokensKnown, o.Usage.Total)
		add(&prompt, &promptKnown, o.Usage.Prompt)
		add(&completion, &completionKnown, o.Usage.Completion)
		add(&reasoning, &reasoningKnown, o.Usage.CompletionDetails.Reasoning)
		add(&cached, &cachedKnown, o.Usage.PromptDetails.Cached)
	}
	sort.Strings(efforts)
	r.SentEfforts = strings.Join(efforts, ", ")
	if r.Failed > 0 {
		r.Status = "incomplete"
	} else if r.Skipped > 0 {
		r.Status = "ineligible"
	} else if r.Unsupported > 0 {
		r.Status = "unsupported"
	}
	if r.Status == "complete" {
		a := float64(r.Correct) / float64(r.Total)
		r.Accuracy = &a
		r.Interval = Wilson(r.Correct, r.Total)
	} else if r.Failed > 0 && r.Skipped == 0 && r.Unsupported == 0 {
		r.Bounds = &Interval{Low: float64(r.Correct) / float64(r.Total), High: float64(r.Correct+r.Failed) / float64(r.Total)}
	}
	r.Latency = Describe(durations)
	r.TokenDistribution = Describe(tokenValues)
	r.CostDistribution = Describe(costValues)
	if costKnown {
		r.Cost = &cost
		v := cost / float64(r.Total)
		r.CostPerQuestion = &v
	}
	if tokensKnown {
		r.Tokens = &tokens
		v := float64(tokens) / float64(r.Total)
		r.TokensPerQuestion = &v
	}
	if promptKnown {
		r.Prompt = &prompt
	}
	if completionKnown {
		r.Completion = &completion
	}
	if reasoningKnown {
		r.Reasoning = &reasoning
	}
	if cachedKnown {
		r.CachedInput = &cached
	}
	return r
}

func Frontiers(rows []Row) {
	groups := map[[3]string][]int{}
	for i, r := range rows {
		key := [3]string{r.Cohort, r.Dataset, r.Category}
		groups[key] = append(groups[key], i)
	}
	for i := range rows {
		a := &rows[i]
		a.CostFrontier, a.TokenFrontier, a.Pareto = false, false, false
		if a.Status != "complete" || a.Accuracy == nil {
			continue
		}
		a.CostFrontier = a.Cost != nil
		a.TokenFrontier = a.Tokens != nil
		a.Pareto = a.Cost != nil && a.Tokens != nil
		for _, j := range groups[[3]string{a.Cohort, a.Dataset, a.Category}] {
			b := rows[j]
			if i == j || b.Status != "complete" || b.Accuracy == nil || a.Cohort != b.Cohort || a.Dataset != b.Dataset || a.Category != b.Category || *b.Accuracy < *a.Accuracy {
				continue
			}
			betterAccuracy := *b.Accuracy > *a.Accuracy
			costDominates := a.Cost != nil && b.Cost != nil && *b.Cost <= *a.Cost
			tokenDominates := a.Tokens != nil && b.Tokens != nil && *b.Tokens <= *a.Tokens
			if costDominates && (betterAccuracy || *b.Cost < *a.Cost) {
				a.CostFrontier = false
			}
			if tokenDominates && (betterAccuracy || *b.Tokens < *a.Tokens) {
				a.TokenFrontier = false
			}
			if costDominates && tokenDominates && (betterAccuracy || *b.Cost < *a.Cost || *b.Tokens < *a.Tokens) {
				a.Pareto = false
			}
		}
	}
}
