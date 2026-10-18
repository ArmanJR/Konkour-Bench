// Package report scores stored generations against a versioned answer-key snapshot.
package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"konkour-bench/internal/bench"
	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/storage"
)

type Outcome struct {
	storage.Job
	Dataset  string           `json:"dataset"`
	Category string           `json:"category"`
	Image    bool             `json:"has_image"`
	Expected int              `json:"expected"`
	Answer   int              `json:"answer"`
	Correct  bool             `json:"correct"`
	Invalid  string           `json:"invalid,omitempty"`
	Provider string           `json:"provider,omitempty"`
	Usage    openrouter.Usage `json:"usage"`
}
type Row struct {
	Model             string   `json:"model"`
	Tier              string   `json:"tier"`
	Cohort            string   `json:"cohort"`
	Dataset           string   `json:"dataset"`
	Category          string   `json:"category"`
	Status            string   `json:"status"`
	Total             int      `json:"total"`
	Complete          int      `json:"complete"`
	Correct           int      `json:"correct"`
	Incorrect         int      `json:"incorrect"`
	Unsupported       int      `json:"unsupported"`
	Skipped           int      `json:"skipped"`
	Failed            int      `json:"failed"`
	Accuracy          *float64 `json:"accuracy"`
	Cost              *float64 `json:"cost_usd"`
	CostPerQuestion   *float64 `json:"cost_per_question_usd"`
	Tokens            *int64   `json:"total_tokens"`
	TokensPerQuestion *float64 `json:"tokens_per_question"`
	Prompt            *int64   `json:"prompt_tokens"`
	Completion        *int64   `json:"completion_tokens"`
	Reasoning         *int64   `json:"reasoning_tokens"`
	CachedInput       *int64   `json:"cached_input_tokens"`
	SentEfforts       string   `json:"sent_efforts"`
	CostFrontier      bool     `json:"cost_frontier"`
	TokenFrontier     bool     `json:"token_frontier"`
	Pareto            bool     `json:"pareto_3d"`
}
type Document struct {
	Generated       time.Time         `json:"generated"`
	ScoringRevision string            `json:"scoring_revision"`
	Protocol        string            `json:"protocol"`
	RunID           string            `json:"run_id,omitempty"`
	Snapshot        dataset.Snapshot  `json:"-"`
	Rows            []Row             `json:"rows"`
	Outcomes        []Outcome         `json:"outcomes"`
	Datasets        []string          `json:"datasets"`
	Categories      []string          `json:"categories"`
	Models          []string          `json:"models"`
	Attempts        []storage.Attempt `json:"-"`
	Spending        []RunSpending     `json:"run_spending"`
}

type RunSpending struct {
	RunID          string  `json:"run_id"`
	Attempts       int     `json:"attempts"`
	KnownCost      float64 `json:"known_cost_usd"`
	KnownRetryCost float64 `json:"known_retry_cost_usd"`
	UnknownCost    int     `json:"attempts_with_unknown_cost"`
}

func Build(s *storage.Store, snap dataset.Snapshot, runID, baseURL string, sel storage.Selection) (Document, error) {
	d := Document{Generated: time.Now().UTC(), ScoringRevision: dataset.Hash(snap), Protocol: bench.Protocol(), RunID: runID, Snapshot: snap}
	if len(sel.Models) == 0 {
		sel.Models = snap.Config.Models
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
			for _, tier := range sel.Efforts {
				o := Outcome{Job: storage.Job{Question: q.ID(), Model: model, Tier: tier, Status: "not_run"}, Dataset: q.Dataset, Category: q.Category, Image: q.HasImage(), Expected: q.Correct}
				for _, j := range lookup[model+"\x00"+tier+"\x00"+q.ID()] {
					if j.Hash != "" {
						body, err := openrouter.Build(q, model, j.Effort)
						if err != nil {
							return d, err
						}
						if bench.RequestHash(baseURL, body) != j.Hash {
							continue
						}
					}
					if j.Status == "skipped" && !q.HasImage() {
						continue
					}
					o.Job = j
					if j.Status == "complete" {
						req, err := s.Request(j.Hash)
						if err != nil {
							return d, err
						}
						o.Answer = req.Result.Answer
						o.Correct = o.Answer == q.Correct
						o.Invalid = req.Result.Invalid
						o.Provider = req.Result.Generation.Provider
						o.Usage = req.Result.Generation.Usage
					}
					break
				}
				d.Outcomes = append(d.Outcomes, o)
			}
		}
	}
	sort.Strings(d.Datasets)
	sort.Strings(d.Categories)
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
		var g openrouter.Generation
		if json.Unmarshal([]byte(a.Response), &g) != nil || g.Usage.Cost == nil {
			entry.UnknownCost++
			continue
		}
		entry.KnownCost += *g.Usage.Cost
		req, err := s.Request(a.Hash)
		if err != nil {
			return d, err
		}
		if a.ID != req.AttemptID {
			entry.KnownRetryCost += *g.Usage.Cost
		}
	}
	for _, entry := range ledger {
		d.Spending = append(d.Spending, *entry)
	}
	sort.Slice(d.Spending, func(i, j int) bool { return d.Spending[i].RunID < d.Spending[j].RunID })
	return d, nil
}

func Aggregate(group []Outcome) Row {
	r := Row{Model: group[0].Model, Tier: group[0].Tier, Total: len(group), Status: "complete"}
	var cost float64
	costKnown := true
	var tokens, prompt, completion, reasoning, cached int64
	tokensKnown, promptKnown, completionKnown, reasoningKnown, cachedKnown := true, true, true, true, true
	var efforts []string
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
	}
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

func Write(dir string, d Document) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "results.json"), b, 0644); err != nil {
		return err
	}
	if err := writeCSV(filepath.Join(dir, "leaderboard.csv"), d.Rows); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "report.js"), []byte(script), 0644); err != nil {
		return err
	}
	for _, metric := range []string{"cost", "tokens", "accuracy"} {
		if err := os.WriteFile(filepath.Join(dir, metric+".svg"), []byte(staticSVG(d.Rows, metric)), 0644); err != nil {
			return err
		}
	}
	// encoding/json escapes HTML-significant characters before embedding the data.
	view := struct {
		Document
		Data   template.JS
		Script template.JS
	}{d, template.JS(b), template.JS(script)}
	t, err := template.New("report").Parse(page)
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, "index.html"))
	if err != nil {
		return err
	}
	if err := t.Execute(f, view); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func writeCSV(path string, rows []Row) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err := w.Write([]string{"model", "tier", "sent_efforts", "cohort", "dataset", "category", "status", "questions", "completed", "correct", "incorrect", "accuracy", "cost_usd", "total_tokens", "reasoning_tokens", "cost_frontier", "token_frontier", "pareto_3d"}); err != nil {
		return err
	}
	number := func(p *float64) string {
		if p == nil {
			return ""
		}
		return strconv.FormatFloat(*p, 'g', -1, 64)
	}
	integer := func(p *int64) string {
		if p == nil {
			return ""
		}
		return strconv.FormatInt(*p, 10)
	}
	for _, r := range rows {
		record := []string{r.Model, r.Tier, r.SentEfforts, r.Cohort, r.Dataset, r.Category, r.Status, strconv.Itoa(r.Total), strconv.Itoa(r.Complete), strconv.Itoa(r.Correct), strconv.Itoa(r.Incorrect), number(r.Accuracy), number(r.Cost), integer(r.Tokens), integer(r.Reasoning), strconv.FormatBool(r.CostFrontier), strconv.FormatBool(r.TokenFrontier), strconv.FormatBool(r.Pareto)}
		// Prevent spreadsheet formula execution in user-controlled labels.
		for i := 0; i < 7; i++ {
			if len(record[i]) > 0 && strings.ContainsAny(record[i][:1], "=+-@\t\r") {
				record[i] = "'" + record[i]
			}
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func staticSVG(rows []Row, metric string) string {
	var b strings.Builder
	var points []Row
	for _, r := range rows {
		if r.Cohort == "text-only" && r.Dataset == "" && r.Category == "" && r.Accuracy != nil {
			points = append(points, r)
		}
	}
	height := 600
	if metric == "accuracy" {
		height = max(160, len(points)*28+80)
	}
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="1000" height="%d" viewBox="0 0 1000 %d"><rect width="1000" height="%d" fill="white"/><g font-family="sans-serif" fill="#17212b">`, height, height, height)
	if metric == "accuracy" {
		b.WriteString(`<text x="40" y="30" font-size="20">Text-only accuracy</text>`)
		for i, r := range points {
			y := 55 + i*28
			fmt.Fprintf(&b, `<text x="480" y="%d" text-anchor="end" font-size="13">%s / %s</text><rect x="490" y="%d" width="%.2f" height="19" fill="#246a9b"/><text x="920" y="%d" font-size="13">%.2f%%</text>`, y+14, template.HTMLEscapeString(r.Model), r.Tier, y, *r.Accuracy*410, y+14, *r.Accuracy*100)
		}
		if len(points) == 0 {
			b.WriteString(`<text x="40" y="80">No complete results</text>`)
		}
		b.WriteString(`</g></svg>`)
		return b.String()
	}
	value := func(r Row) *float64 {
		if metric == "cost" {
			return r.CostPerQuestion
		}
		return r.TokensPerQuestion
	}
	frontier := func(r Row) bool {
		if metric == "cost" {
			return r.CostFrontier
		}
		return r.TokenFrontier
	}
	maxX := 0.0
	for _, r := range points {
		if v := value(r); v != nil && *v > maxX {
			maxX = *v
		}
	}
	if maxX == 0 {
		maxX = 1
	}
	label := "Native tokens per question"
	if metric == "cost" {
		label = "USD per question"
	}
	fmt.Fprintf(&b, `<text x="40" y="30" font-size="20">Text-only: accuracy versus %s</text><text x="450" y="590" font-size="14">%s</text><text x="600" y="30" fill="#ad4b00">Orange: Pareto frontier</text>`, metric, label)
	for i := 0; i <= 4; i++ {
		fraction := float64(i) / 4
		y := 540 - fraction*470
		x := 70 + fraction*850
		fmt.Fprintf(&b, `<path d="M70 %.2fH930" stroke="#ddd"/><text x="60" y="%.2f" text-anchor="end" font-size="12">%.0f%%</text><text x="%.2f" y="565" text-anchor="middle" font-size="12">%.5g</text>`, y, y+4, fraction*100, x, fraction*maxX)
	}
	sort.SliceStable(points, func(i, j int) bool {
		a, c := value(points[i]), value(points[j])
		return a != nil && (c == nil || *a < *c)
	})
	var line []string
	for _, r := range points {
		if v := value(r); v != nil && frontier(r) {
			line = append(line, fmt.Sprintf("%.2f,%.2f", 70+*v/maxX*850, 540-*r.Accuracy*470))
		}
	}
	if len(line) > 1 {
		fmt.Fprintf(&b, `<polyline points="%s" fill="none" stroke="#ad4b00" stroke-dasharray="4 3"/>`, strings.Join(line, " "))
	}
	for _, r := range points {
		v := value(r)
		if v == nil {
			continue
		}
		x := 70 + *v/maxX*850
		y := 540 - *r.Accuracy*470
		color := "#246a9b"
		if frontier(r) {
			color = "#ad4b00"
		}
		fmt.Fprintf(&b, `<circle cx="%.2f" cy="%.2f" r="6" fill="%s"><title>%s / %s: %.2f%%, %g</title></circle>`, x, y, color, template.HTMLEscapeString(r.Model), r.Tier, *r.Accuracy*100, *v)
	}
	b.WriteString(`</g></svg>`)
	return b.String()
}
