package report

import (
	"encoding/base64"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
	"time"

	"konkour-bench/internal/backbone"
	"konkour-bench/internal/dataset"
	"konkour-bench/internal/storage"
)

// Intervals describe question-sampling uncertainty, not repeated model runs.
type Interval struct {
	Low  float64 `json:"low"`
	High float64 `json:"high"`
}

type Distribution struct {
	N      int      `json:"n"`
	Median *float64 `json:"median"`
	P90    *float64 `json:"p90"`
	Max    *float64 `json:"max"`
}

func Wilson(correct, total int) *Interval {
	if total <= 0 {
		return nil
	}
	n, p, z := float64(total), float64(correct)/float64(total), 1.959963984540054
	denom := 1 + z*z/n
	center := (p + z*z/(2*n)) / denom
	half := z * math.Sqrt(p*(1-p)/n+z*z/(4*n*n)) / denom
	return &Interval{Low: math.Max(0, center-half), High: math.Min(1, center+half)}
}

// Describe uses linear interpolation at (n-1)*p (R type 7).
// Empty samples stay unknown; explicit zero measurements are retained.
func Describe(values []float64) Distribution {
	d := Distribution{N: len(values)}
	if len(values) == 0 {
		return d
	}
	v := slices.Clone(values)
	sort.Float64s(v)
	quantile := func(p float64) *float64 {
		i := p * float64(len(v)-1)
		lo, hi := int(math.Floor(i)), int(math.Ceil(i))
		x := v[lo] + (v[hi]-v[lo])*(i-float64(lo))
		return &x
	}
	d.Median, d.P90, d.Max = quantile(.5), quantile(.9), quantile(1)
	return d
}

type QuestionOption struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
	Image  string `json:"image,omitempty"`
}
type QuestionContent struct {
	ID       string           `json:"id"`
	Dataset  string           `json:"dataset"`
	Category string           `json:"category"`
	Section  string           `json:"section,omitempty"`
	Number   int              `json:"number"`
	Text     string           `json:"text"`
	Image    string           `json:"image,omitempty"`
	Options  []QuestionOption `json:"options"`
}
type SourceRun struct {
	ID             string    `json:"id"`
	Created        time.Time `json:"created"`
	Backend        string    `json:"backend"`
	CLIVersion     string    `json:"cli_version,omitempty"`
	TimeoutSeconds float64   `json:"timeout_seconds"`
	BudgetUSD      float64   `json:"budget_usd"`
	SelectedJobs   int       `json:"selected_jobs"`
	OriginAnswers  int       `json:"origin_answers"`
}

func enrich(d *Document, s *storage.Store) error {
	d.AnalysisVersion = "3"
	d.DatasetMetadata = d.Snapshot.Metadata
	d.Warnings = d.Snapshot.Warnings
	d.QuestionBank = []QuestionContent{}
	selected := map[string]bool{}
	for _, o := range d.Outcomes {
		selected[o.Question] = true
	}
	asset := func(a *dataset.Asset) string {
		if a == nil {
			return ""
		}
		return "data:" + a.MIME + ";base64," + base64.StdEncoding.EncodeToString(a.Bytes)
	}
	for _, q := range d.Snapshot.Questions {
		if !selected[q.ID()] {
			continue
		}
		content := QuestionContent{ID: q.ID(), Dataset: q.Dataset, Category: q.Category, Section: q.Section, Number: q.Number, Text: q.Text, Image: asset(q.Asset), Options: []QuestionOption{}}
		for _, o := range q.Options {
			content.Options = append(content.Options, QuestionOption{Number: o.Number, Text: o.Text, Image: asset(o.Asset)})
		}
		d.QuestionBank = append(d.QuestionBank, content)
	}
	timings, err := s.AttemptTimings()
	if err != nil {
		return fmt.Errorf("load report attempt timing: %w", err)
	}
	sources := map[string]*SourceRun{}
	load := func(id string) (*SourceRun, error) {
		if v, ok := sources[id]; ok {
			return v, nil
		}
		r, err := s.Run(id)
		if err != nil {
			return nil, fmt.Errorf("load report source run %s: %w", id, err)
		}
		v := &SourceRun{ID: id, Created: r.Created, Backend: backbone.Name(r.BaseURL), TimeoutSeconds: r.Timeout.Seconds(), BudgetUSD: r.BudgetUSD}
		endpoint := backbone.EndpointFor(r.BaseURL, "anthropic/claude-opus-5.5")
		if backbone.IsClaude(endpoint) {
			v.CLIVersion = strings.TrimPrefix(endpoint, "claude-cli://v1/")
		}
		sources[id] = v
		return v, nil
	}
	for i := range d.Outcomes {
		o := &d.Outcomes[i]
		if o.RunID != "" {
			source, err := load(o.RunID)
			if err != nil {
				return err
			}
			source.SelectedJobs++
		}
		if o.Status != "complete" {
			continue
		}
		if a, ok := timings[o.AttemptID]; ok {
			o.OriginRun = a.RunID
			source, err := load(a.RunID)
			if err != nil {
				return err
			}
			source.OriginAnswers++
			start, e1 := time.Parse(time.RFC3339Nano, a.Started)
			end, e2 := time.Parse(time.RFC3339Nano, a.Finished)
			if e1 == nil && e2 == nil && !end.Before(start) {
				seconds := end.Sub(start).Seconds()
				o.Duration = &seconds
			}
		}
	}
	d.Sources = []SourceRun{}
	for _, source := range sources {
		d.Sources = append(d.Sources, *source)
	}
	sort.Slice(d.Sources, func(i, j int) bool { return d.Sources[i].Created.Before(d.Sources[j].Created) })
	return nil
}

type Scope struct {
	Cohort   string `json:"cohort"`
	Dataset  string `json:"dataset"`
	Category string `json:"category"`
}
type Coverage struct {
	Questions      int            `json:"questions"`
	Images         int            `json:"images"`
	Models         int            `json:"models"`
	Configurations int            `json:"configurations"`
	Ranked         int            `json:"ranked"`
	Slots          int            `json:"slots"`
	Complete       int            `json:"complete"`
	Invalid        int            `json:"invalid"`
	Statuses       map[string]int `json:"statuses"`
	Reasons        []Count        `json:"reasons"`
	TokenAnomalies int            `json:"token_anomalies"`
}
type Count struct {
	Label string `json:"label"`
	N     int    `json:"n"`
}
type Insight struct {
	Label  string `json:"label"`
	Value  string `json:"value"`
	Detail string `json:"detail"`
}
type EffortComparison struct {
	Model       string   `json:"model"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	Total       int      `json:"total"`
	Paired      int      `json:"paired"`
	Gained      int      `json:"gained"`
	Lost        int      `json:"lost"`
	BothCorrect int      `json:"both_correct"`
	BothWrong   int      `json:"both_wrong"`
	Delta       *float64 `json:"delta"`
	CostRatio   *float64 `json:"cost_ratio"`
	TokenRatio  *float64 `json:"token_ratio"`
}
type QuestionAnalysis struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Section  string `json:"section,omitempty"`
	Image    bool   `json:"has_image"`
	Expected int    `json:"expected"`
	Panel    int    `json:"panel"`
	Correct  int    `json:"correct"`
	Answers  [5]int `json:"answers"` // Index 0 is invalid / no valid answer.
	Review   bool   `json:"review"`
}
type AnalysisView struct {
	Scope
	Coverage   Coverage             `json:"coverage"`
	Insights   []Insight            `json:"insights"`
	Efforts    []EffortComparison   `json:"efforts"`
	Questions  []QuestionAnalysis   `json:"questions"`
	Panel      []PanelConfig        `json:"panel"`
	Ranks      []RankInterval       `json:"ranks"`
	Resamples  int                  `json:"rank_resamples"`
	Separation []Separation         `json:"separation"`
	Length     LengthAnalysis       `json:"length"`
	Estimates  []CompletionEstimate `json:"estimates"`
	Providers  []ProviderEffect     `json:"providers"`
}

func inScope(o Outcome, scope Scope) bool {
	return (scope.Dataset == "" || o.Dataset == scope.Dataset) && (scope.Category == "" || o.Category == scope.Category) &&
		(scope.Cohort == "full" || scope.Cohort == "image-only" && o.Image || scope.Cohort == "text-only" && !o.Image)
}
func configKey(model, tier string) string { return model + "\x00" + tier }

// Analyze uses a fixed panel of complete configurations for item difficulty.
// This avoids changing the model mix between easy and hard questions.
func Analyze(d Document) []AnalysisView {
	groups := map[Scope][]Row{}
	for _, r := range d.Rows {
		scope := Scope{r.Cohort, r.Dataset, r.Category}
		groups[scope] = append(groups[scope], r)
	}
	scopes := make([]Scope, 0, len(groups))
	for scope := range groups {
		scopes = append(scopes, scope)
	}
	sort.Slice(scopes, func(i, j int) bool {
		a, b := scopes[i], scopes[j]
		return strings.Join([]string{a.Cohort, a.Dataset, a.Category}, "\x00") < strings.Join([]string{b.Cohort, b.Dataset, b.Category}, "\x00")
	})
	views := make([]AnalysisView, 0, len(scopes))
	for _, scope := range scopes {
		rows := groups[scope]
		v := AnalysisView{Scope: scope, Coverage: Coverage{Configurations: len(rows), Statuses: map[string]int{}, Reasons: []Count{}}, Insights: []Insight{}, Efforts: []EffortComparison{}, Questions: []QuestionAnalysis{}}
		panel := map[string]bool{}
		models := map[string]bool{}
		var ranked []Row
		for _, r := range rows {
			models[r.Model] = true
			if r.Accuracy != nil {
				panel[configKey(r.Model, r.Tier)] = true
				ranked = append(ranked, r)
			}
		}
		v.Coverage.Models = len(models)
		v.Coverage.Ranked = len(ranked)
		questions := map[string]*QuestionAnalysis{}
		outcomes := map[string]map[string]Outcome{}
		reasons := map[string]int{}
		for _, o := range d.Outcomes {
			if !inScope(o, scope) {
				continue
			}
			v.Coverage.Slots++
			v.Coverage.Statuses[o.Status]++
			if o.Status == "complete" {
				v.Coverage.Complete++
			}
			if o.Invalid != "" {
				v.Coverage.Invalid++
				reasons["Invalid: "+o.Invalid]++
			}
			if o.Status != "complete" {
				label := o.Status
				if o.Note != "" {
					label += " · " + o.Note
				}
				reasons[label]++
			}
			if o.Status == "complete" && o.Usage.Completion != nil && o.Usage.CompletionDetails.Reasoning != nil && *o.Usage.CompletionDetails.Reasoning > *o.Usage.Completion {
				v.Coverage.TokenAnomalies++
			}
			key := configKey(o.Model, o.Tier)
			if outcomes[key] == nil {
				outcomes[key] = map[string]Outcome{}
			}
			outcomes[key][o.Question] = o
			q := questions[o.Question]
			if q == nil {
				q = &QuestionAnalysis{ID: o.Question, Category: o.Category, Section: o.Section, Image: o.Image, Expected: o.Expected}
				questions[o.Question] = q
			}
			if panel[key] {
				q.Panel++
				if o.Correct {
					q.Correct++
				}
				answer := o.Answer
				if o.Invalid != "" || answer < 1 || answer > 4 {
					answer = 0
				}
				q.Answers[answer]++
			}
		}
		for _, q := range questions {
			v.Coverage.Questions++
			if q.Image {
				v.Coverage.Images++
			}
			for answer, n := range q.Answers {
				if answer > 0 && answer != q.Expected && q.Panel >= 3 && n*2 > q.Panel {
					q.Review = true
				}
			}
			v.Questions = append(v.Questions, *q)
		}
		sort.Slice(v.Questions, func(i, j int) bool {
			a, b := v.Questions[i], v.Questions[j]
			if a.Correct != b.Correct {
				return a.Correct < b.Correct
			}
			return a.ID < b.ID
		})
		for label, n := range reasons {
			v.Coverage.Reasons = append(v.Coverage.Reasons, Count{label, n})
		}
		sort.Slice(v.Coverage.Reasons, func(i, j int) bool {
			a, b := v.Coverage.Reasons[i], v.Coverage.Reasons[j]
			if a.N != b.N {
				return a.N > b.N
			}
			return a.Label < b.Label
		})
		modelIDs := make([]string, 0, len(models))
		for model := range models {
			modelIDs = append(modelIDs, model)
		}
		sort.Strings(modelIDs)
		for _, model := range modelIDs {
			for _, pair := range [][2]string{{"minimal", "medium"}, {"medium", "max"}, {"minimal", "max"}} {
				a, b := outcomes[configKey(model, pair[0])], outcomes[configKey(model, pair[1])]
				if len(a) == 0 || len(b) == 0 {
					continue
				}
				e := EffortComparison{Model: model, From: pair[0], To: pair[1], Total: len(questions)}
				var ac, bc, at, bt float64
				costsKnown, tokensKnown := true, true
				// Stable question order also keeps floating-point sums reproducible.
				for _, q := range v.Questions {
					oa := a[q.ID]
					ob, ok := b[q.ID]
					if !ok || oa.Status != "complete" || ob.Status != "complete" {
						continue
					}
					e.Paired++
					switch {
					case oa.Correct && ob.Correct:
						e.BothCorrect++
					case !oa.Correct && ob.Correct:
						e.Gained++
					case oa.Correct && !ob.Correct:
						e.Lost++
					default:
						e.BothWrong++
					}
					if oa.Usage.Cost == nil || ob.Usage.Cost == nil {
						costsKnown = false
					} else {
						ac += *oa.Usage.Cost
						bc += *ob.Usage.Cost
					}
					if oa.Usage.Total == nil || ob.Usage.Total == nil {
						tokensKnown = false
					} else {
						at += float64(*oa.Usage.Total)
						bt += float64(*ob.Usage.Total)
					}
				}
				if e.Paired > 0 {
					delta := float64(e.Gained-e.Lost) / float64(e.Paired)
					e.Delta = &delta
				}
				if costsKnown && ac > 0 {
					ratio := bc / ac
					e.CostRatio = &ratio
				}
				if tokensKnown && at > 0 {
					ratio := bt / at
					e.TokenRatio = &ratio
				}
				v.Efforts = append(v.Efforts, e)
			}
		}
		p := newPanel(ranked, outcomes, v.Questions)
		v.Panel = p.configs
		if v.Panel == nil {
			v.Panel = []PanelConfig{}
		}
		v.Ranks = p.ranks()
		if len(v.Ranks) > 0 {
			v.Resamples = bootstrapResamples
		}
		v.Separation = p.separation()
		v.Length = p.length()
		v.Estimates = p.estimates(rows, outcomes)
		v.Providers = p.providers(outcomes)
		v.Insights = insights(v, ranked)
		v.Insights = append(v.Insights, diagnosticInsights(v, p, outcomes)...)
		views = append(views, v)
	}
	return views
}

func insights(v AnalysisView, rows []Row) []Insight {
	result := []Insight{}
	if len(rows) > 0 {
		sort.Slice(rows, func(i, j int) bool {
			if *rows[i].Accuracy != *rows[j].Accuracy {
				return *rows[i].Accuracy > *rows[j].Accuracy
			}
			return configKey(rows[i].Model, rows[i].Tier) < configKey(rows[j].Model, rows[j].Tier)
		})
		top := rows[0]
		labels := []string{}
		for _, r := range rows {
			if *r.Accuracy == *top.Accuracy {
				labels = append(labels, r.Model+" / "+r.Tier)
			}
		}
		result = append(result, Insight{Label: "Highest observed accuracy", Value: fmt.Sprintf("%.2f%%", *top.Accuracy*100), Detail: fmt.Sprintf("%d/%d correct · %s. One question changes this score by %.2f percentage points.", top.Correct, top.Total, strings.Join(labels, "; "), 100/float64(top.Total))})
	}
	result = append(result, Insight{Label: "Comparable configurations", Value: fmt.Sprintf("%d / %d", v.Coverage.Ranked, v.Coverage.Configurations), Detail: fmt.Sprintf("%d of %d answer slots completed. Only configurations that completed every selected question are ranked; invalid completions count as wrong.", v.Coverage.Complete, v.Coverage.Slots)})
	paired, noGain := 0, 0
	for _, e := range v.Efforts {
		if e.From == "medium" && e.To == "max" && e.Paired == e.Total && e.Total > 0 {
			paired++
			if e.Gained <= e.Lost {
				noGain++
			}
		}
	}
	if paired > 0 {
		result = append(result, Insight{Label: "No gain at maximum effort", Value: fmt.Sprintf("%d / %d", noGain, paired), Detail: "Models with complete medium/max pairs that show no net accuracy gain at max. Gains and regressions below compare the same questions; this is one retained generation per setting."})
	}
	review := 0
	for _, q := range v.Questions {
		if q.Review {
			review++
		}
	}
	if v.Coverage.Ranked >= 3 {
		result = append(result, Insight{Label: "Questions worth reviewing", Value: fmt.Sprint(review), Detail: "A strict majority of the complete configuration panel chose the same non-key answer. Check the source, transcription, and key; correlated model agreement does not establish a key error."})
	}
	return result
}
