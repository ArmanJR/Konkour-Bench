package report

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"sort"
	"strconv"
	"strings"

	"konkour-bench/internal/openrouter"
)

// A fixed seed keeps bootstrap summaries identical for identical records.
const bootstrapResamples = 2000

// Provider gaps get an interval only with this many answers on each side.
const minProviderAnswers = 10

type PanelConfig struct {
	Model string `json:"model"`
	Tier  string `json:"tier"`
}

// RankInterval summarizes competition ranks over paired question resamples.
// Ties share the better rank, as in the leaderboard.
type RankInterval struct {
	Model  string  `json:"model"`
	Tier   string  `json:"tier"`
	Rank   int     `json:"rank"`
	Low    int     `json:"low"`
	Median int     `json:"median"`
	High   int     `json:"high"`
	First  float64 `json:"first"`
}

// Separation attributes between-configuration variance in total correct
// answers to question groups: cov(group, total) / var(total). Shares of a
// partition sum to one.
type Separation struct {
	Level     string   `json:"level"` // "subject", or "section" within a subject
	Category  string   `json:"category"`
	Section   string   `json:"section,omitempty"` // Empty at section level: untagged questions.
	Questions int      `json:"questions"`
	Ceiling   int      `json:"ceiling"`
	Floor     int      `json:"floor"`
	Mean      float64  `json:"mean"`
	Min       float64  `json:"min"`
	Max       float64  `json:"max"`
	Share     *float64 `json:"variance_share"`
	Correct   []int    `json:"correct"` // Aligned with AnalysisView.Panel.
}

type LengthRow struct {
	Model         string   `json:"model"`
	Tier          string   `json:"tier"`
	Correct       int      `json:"correct"`
	Wrong         int      `json:"wrong"`
	CorrectMedian *float64 `json:"correct_median_tokens"`
	WrongMedian   *float64 `json:"wrong_median_tokens"`
	Ratio         *float64 `json:"ratio"`
	Longest       int      `json:"longest"`
	LongestRight  int      `json:"longest_correct"`
}

// LengthAnalysis compares output tokens of correct and wrong answers. Ratio
// is exp(beta) from log tokens on a wrong-answer indicator with question and
// configuration fixed effects.
type LengthAnalysis struct {
	Rows           []LengthRow `json:"rows"`
	Answers        int         `json:"answers"`
	AnswersCorrect int         `json:"answers_correct"`
	Contrasts      int         `json:"contrast_questions"`
	Ratio          *float64    `json:"ratio"`
	Longest        int         `json:"longest"`
	LongestCorrect int         `json:"longest_correct"`
}

// CompletionEstimate fills unresolved answers with a Rasch model: panel item
// difficulty and an ability fitted to the configuration's completed answers.
// It is descriptive and never enters rankings.
type CompletionEstimate struct {
	Model               string   `json:"model"`
	Tier                string   `json:"tier"`
	Total               int      `json:"total"`
	Complete            int      `json:"complete"`
	Correct             int      `json:"correct"`
	Unresolved          int      `json:"unresolved"`
	MissingDifficulty   *float64 `json:"unresolved_panel_accuracy"`
	CompletedDifficulty *float64 `json:"completed_panel_accuracy"`
	Expected            *float64 `json:"expected_correct"`
	Spread              *float64 `json:"expected_correct_sd"`
	Estimate            *float64 `json:"estimate"`
	SameModel           []Count  `json:"same_model"` // Complete tiers of the same model: correct on the unresolved questions.
	// The model's held-out prediction for complete configurations on the same
	// unresolved questions. Their actual accuracy there is MissingDifficulty.
	CheckPredicted *float64 `json:"check_predicted"`
}

// ProviderEffect compares a model's answers from one serving provider with the
// leave-model-out panel accuracy on the same questions.
type ProviderEffect struct {
	Model       string    `json:"model"`
	Provider    string    `json:"provider"`
	Tiers       []Count   `json:"tiers"`
	Answers     int       `json:"answers"`
	Correct     int       `json:"correct"`
	Invalid     int       `json:"invalid"`
	Expected    *float64  `json:"expected"`
	Excess      *float64  `json:"excess"`
	Others      int       `json:"other_answers"`
	Gap         *float64  `json:"gap"`
	GapInterval *Interval `json:"gap_interval"`
}

type panelMatrix struct {
	configs   []PanelConfig
	questions []QuestionAnalysis // Dataset, then booklet order.
	index     map[string]int
	correct   [][]bool // [configuration][question]
	answers   [][]Outcome
	rate      []float64 // Panel accuracy per question.
}

func tierOrder(tier string) int {
	if i := slices.Index(openrouter.Tiers, tier); i >= 0 {
		return i
	}
	return len(openrouter.Tiers)
}

func configLess(a, b PanelConfig) bool {
	if a.Model != b.Model {
		return a.Model < b.Model
	}
	if tierOrder(a.Tier) != tierOrder(b.Tier) {
		return tierOrder(a.Tier) < tierOrder(b.Tier)
	}
	return a.Tier < b.Tier
}

// questionLess orders IDs of the form dataset:number numerically within a dataset.
func questionLess(a, b string) bool {
	ad, an, _ := strings.Cut(a, ":")
	bd, bn, _ := strings.Cut(b, ":")
	if ad != bd {
		return ad < bd
	}
	ai, aerr := strconv.Atoi(an)
	bi, berr := strconv.Atoi(bn)
	if aerr == nil && berr == nil && ai != bi {
		return ai < bi
	}
	return a < b
}

func newPanel(ranked []Row, outcomes map[string]map[string]Outcome, questions []QuestionAnalysis) panelMatrix {
	p := panelMatrix{questions: slices.Clone(questions), index: map[string]int{}}
	sort.Slice(p.questions, func(i, j int) bool { return questionLess(p.questions[i].ID, p.questions[j].ID) })
	for i, q := range p.questions {
		p.index[q.ID] = i
	}
	for _, r := range ranked {
		p.configs = append(p.configs, PanelConfig{r.Model, r.Tier})
	}
	sort.Slice(p.configs, func(i, j int) bool { return configLess(p.configs[i], p.configs[j]) })
	p.rate = make([]float64, len(p.questions))
	for _, c := range p.configs {
		correct, answers := make([]bool, len(p.questions)), make([]Outcome, len(p.questions))
		for i, q := range p.questions {
			o := outcomes[configKey(c.Model, c.Tier)][q.ID]
			answers[i] = o
			correct[i] = o.Status == "complete" && o.Correct
			if correct[i] {
				p.rate[i]++
			}
		}
		p.correct = append(p.correct, correct)
		p.answers = append(p.answers, answers)
	}
	if len(p.configs) > 0 {
		for i := range p.rate {
			p.rate[i] /= float64(len(p.configs))
		}
	}
	return p
}

// ordered lists in-scope outcomes by configuration and booklet order, so that
// floating-point sums are reproducible.
func (p panelMatrix) ordered(outcomes map[string]map[string]Outcome) []Outcome {
	var list []Outcome
	for _, byQuestion := range outcomes {
		for _, o := range byQuestion {
			if _, ok := p.index[o.Question]; ok {
				list = append(list, o)
			}
		}
	}
	sort.Slice(list, func(i, j int) bool {
		a, b := list[i], list[j]
		if a.Model != b.Model || a.Tier != b.Tier {
			return configLess(PanelConfig{a.Model, a.Tier}, PanelConfig{b.Model, b.Tier})
		}
		return p.index[a.Question] < p.index[b.Question]
	})
	return list
}

func (p panelMatrix) score(c int, questions []int) int {
	n := 0
	for _, q := range questions {
		if p.correct[c][q] {
			n++
		}
	}
	return n
}

// nearestRank returns the p-quantile of sorted integers without interpolation.
func nearestRank(sorted []int, p float64) int {
	i := max(0, int(math.Ceil(p*float64(len(sorted))))-1)
	return sorted[min(i, len(sorted)-1)]
}

func (p panelMatrix) ranks() []RankInterval {
	k, n := len(p.configs), len(p.questions)
	result := []RankInterval{}
	if k < 2 || n == 0 {
		return result
	}
	competition := func(scores []int, c int) int {
		rank := 1
		for _, s := range scores {
			if s > scores[c] {
				rank++
			}
		}
		return rank
	}
	all := make([]int, n)
	for i := range all {
		all[i] = i
	}
	observed := make([]int, k)
	for c := range k {
		observed[c] = p.score(c, all)
	}
	rng := rand.New(rand.NewPCG(1405, uint64(n)))
	samples := make([][]int, k)
	for c := range samples {
		samples[c] = make([]int, bootstrapResamples)
	}
	scores, draw := make([]int, k), make([]int, n)
	for b := range bootstrapResamples {
		for i := range draw {
			draw[i] = rng.IntN(n)
		}
		for c := range k {
			scores[c] = p.score(c, draw)
		}
		for c := range k {
			samples[c][b] = competition(scores, c)
		}
	}
	for c, cfg := range p.configs {
		sorted := slices.Clone(samples[c])
		slices.Sort(sorted)
		first := 0
		for _, r := range sorted {
			if r == 1 {
				first++
			}
		}
		result = append(result, RankInterval{Model: cfg.Model, Tier: cfg.Tier, Rank: competition(observed, c), Low: nearestRank(sorted, .05), Median: nearestRank(sorted, .5), High: nearestRank(sorted, .95), First: float64(first) / bootstrapResamples})
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Rank < result[j].Rank })
	return result
}

func (p panelMatrix) separation() []Separation {
	k := len(p.configs)
	result := []Separation{}
	if k < 2 {
		return result
	}
	all := make([]int, len(p.questions))
	for i := range all {
		all[i] = i
	}
	totals := make([]float64, k)
	for c := range k {
		totals[c] = float64(p.score(c, all))
	}
	mean := func(v []float64) float64 {
		s := 0.0
		for _, x := range v {
			s += x
		}
		return s / float64(len(v))
	}
	totalMean := mean(totals)
	variance := 0.0
	for _, t := range totals {
		variance += (t - totalMean) * (t - totalMean)
	}
	group := func(level, category, section string, questions []int) Separation {
		g := Separation{Level: level, Category: category, Section: section, Questions: len(questions), Correct: make([]int, k), Min: 1}
		scores := make([]float64, k)
		for c := range k {
			g.Correct[c] = p.score(c, questions)
			scores[c] = float64(g.Correct[c])
			accuracy := scores[c] / float64(len(questions))
			g.Mean += accuracy / float64(k)
			g.Min, g.Max = math.Min(g.Min, accuracy), math.Max(g.Max, accuracy)
		}
		for _, q := range questions {
			switch p.rate[q] {
			case 1:
				g.Ceiling++
			case 0:
				g.Floor++
			}
		}
		if variance > 0 {
			scoreMean, covariance := mean(scores), 0.0
			for c := range k {
				covariance += (scores[c] - scoreMean) * (totals[c] - totalMean)
			}
			share := covariance / variance
			g.Share = &share
		}
		return g
	}
	// Booklet order keeps subjects and their sections in exam order.
	var categories []string
	byCategory := map[string][]int{}
	for i, q := range p.questions {
		if _, ok := byCategory[q.Category]; !ok {
			categories = append(categories, q.Category)
		}
		byCategory[q.Category] = append(byCategory[q.Category], i)
	}
	for _, category := range categories {
		questions := byCategory[category]
		result = append(result, group("subject", category, "", questions))
		var sections []string
		bySection := map[string][]int{}
		tagged := false
		for _, q := range questions {
			section := p.questions[q].Section
			tagged = tagged || section != ""
			if _, ok := bySection[section]; !ok {
				sections = append(sections, section)
			}
			bySection[section] = append(bySection[section], q)
		}
		if !tagged {
			continue
		}
		for _, section := range sections {
			result = append(result, group("section", category, section, bySection[section]))
		}
	}
	return result
}

func (p panelMatrix) length() LengthAnalysis {
	result := LengthAnalysis{Rows: []LengthRow{}}
	type observation struct {
		config, question int
		tokens           float64
		wrong            bool
	}
	var observations []observation
	for c, cfg := range p.configs {
		row := LengthRow{Model: cfg.Model, Tier: cfg.Tier}
		var right, wrong []float64
		var known []observation
		for q, o := range p.answers[c] {
			if o.Status != "complete" || o.Usage.Completion == nil || *o.Usage.Completion <= 0 {
				continue
			}
			v := observation{c, q, float64(*o.Usage.Completion), !p.correct[c][q]}
			known = append(known, v)
			if v.wrong {
				wrong = append(wrong, v.tokens)
			} else {
				right = append(right, v.tokens)
			}
		}
		row.Correct, row.Wrong = len(right), len(wrong)
		row.CorrectMedian, row.WrongMedian = Describe(right).Median, Describe(wrong).Median
		if row.CorrectMedian != nil && row.WrongMedian != nil {
			ratio := *row.WrongMedian / *row.CorrectMedian
			row.Ratio = &ratio
		}
		// Ties at the cutoff resolve by booklet order so the decile is reproducible.
		sort.SliceStable(known, func(i, j int) bool { return known[i].tokens > known[j].tokens })
		row.Longest = int(math.Ceil(float64(len(known)) / 10))
		for _, v := range known[:row.Longest] {
			if !v.wrong {
				row.LongestRight++
			}
		}
		result.Rows = append(result.Rows, row)
		result.Longest += row.Longest
		result.LongestCorrect += row.LongestRight
		result.AnswersCorrect += row.Correct
		observations = append(observations, known...)
	}
	result.Answers = len(observations)
	outcomesByQuestion := map[int][2]bool{}
	for _, v := range observations {
		seen := outcomesByQuestion[v.question]
		seen[boolIndex(v.wrong)] = true
		outcomesByQuestion[v.question] = seen
	}
	for _, seen := range outcomesByQuestion {
		if seen[0] && seen[1] {
			result.Contrasts++
		}
	}
	if result.Contrasts == 0 {
		return result
	}
	// Alternating projections remove both fixed effects, including in unbalanced panels.
	y, x := make([]float64, len(observations)), make([]float64, len(observations))
	for i, v := range observations {
		y[i] = math.Log(v.tokens)
		if v.wrong {
			x[i] = 1
		}
	}
	demean := func(values []float64, key func(observation) int, groups int) float64 {
		sums, counts := make([]float64, groups), make([]float64, groups)
		for i, v := range observations {
			sums[key(v)] += values[i]
			counts[key(v)]++
		}
		change := 0.0
		for i, v := range observations {
			shift := sums[key(v)] / counts[key(v)]
			values[i] -= shift
			change = math.Max(change, math.Abs(shift))
		}
		return change
	}
	byQuestion := func(v observation) int { return v.question }
	byConfig := func(v observation) int { return v.config }
	for range 10000 {
		change := 0.0
		for _, values := range [][]float64{y, x} {
			change = math.Max(change, demean(values, byQuestion, len(p.questions)))
			change = math.Max(change, demean(values, byConfig, len(p.configs)))
		}
		if change < 1e-12 {
			break
		}
	}
	xy, xx := 0.0, 0.0
	for i := range x {
		xy += x[i] * y[i]
		xx += x[i] * x[i]
	}
	if xx > 1e-12 {
		ratio := math.Exp(xy / xx)
		result.Ratio = &ratio
	}
	return result
}

func boolIndex(v bool) int {
	if v {
		return 1
	}
	return 0
}

func logistic(v float64) float64 { return 1 / (1 + math.Exp(-v)) }

// raschDifficulty converts smoothed panel accuracy into item difficulty,
// optionally without one configuration so that it cannot inform its own check.
func (p panelMatrix) raschDifficulty(exclude int) []float64 {
	difficulty := make([]float64, len(p.questions))
	for q := range p.questions {
		correct, total := 0.0, 0.0
		for c := range p.configs {
			if c == exclude {
				continue
			}
			total++
			if p.correct[c][q] {
				correct++
			}
		}
		// Smoothing keeps difficulty finite at 0/k and k/k.
		smoothed := (correct + .5) / (total + 1)
		difficulty[q] = math.Log((1 - smoothed) / smoothed)
	}
	return difficulty
}

// raschAbility solves sum(P(correct)) = observed correct, which is monotone in ability.
func raschAbility(difficulty []float64, questions []int, correct int) float64 {
	low, high := -10.0, 10.0
	for range 100 {
		theta, expected := (low+high)/2, 0.0
		for _, q := range questions {
			expected += logistic(theta - difficulty[q])
		}
		if expected < float64(correct) {
			low = theta
		} else {
			high = theta
		}
	}
	return (low + high) / 2
}

// estimates covers configurations whose only missing answers are unresolved
// requests. Capability exclusions are not imputable answers. The held-out
// check refits the same model for each complete configuration on the same
// unresolved question set, where the true answers are known.
func (p panelMatrix) estimates(rows []Row, outcomes map[string]map[string]Outcome) []CompletionEstimate {
	result := []CompletionEstimate{}
	k := len(p.configs)
	if k < 3 {
		return result
	}
	difficulty := p.raschDifficulty(-1)
	var incomplete []Row
	for _, r := range rows {
		if r.Bounds != nil {
			incomplete = append(incomplete, r)
		}
	}
	sort.Slice(incomplete, func(i, j int) bool {
		return configLess(PanelConfig{incomplete[i].Model, incomplete[i].Tier}, PanelConfig{incomplete[j].Model, incomplete[j].Tier})
	})
	meanRate := func(questions []int) *float64 {
		if len(questions) == 0 {
			return nil
		}
		s := 0.0
		for _, q := range questions {
			s += p.rate[q]
		}
		v := s / float64(len(questions))
		return &v
	}
	for _, r := range incomplete {
		e := CompletionEstimate{Model: r.Model, Tier: r.Tier, Total: r.Total, Complete: r.Complete, Correct: r.Correct, Unresolved: r.Failed, SameModel: []Count{}}
		var done, missing []int
		for q, question := range p.questions {
			o, ok := outcomes[configKey(r.Model, r.Tier)][question.ID]
			if !ok {
				continue
			}
			if o.Status == "complete" {
				done = append(done, q)
			} else {
				missing = append(missing, q)
			}
		}
		e.MissingDifficulty, e.CompletedDifficulty = meanRate(missing), meanRate(done)
		theta, expected, variance := raschAbility(difficulty, done, e.Correct), float64(e.Correct), 0.0
		for _, q := range missing {
			v := logistic(theta - difficulty[q])
			expected += v
			variance += v * (1 - v)
		}
		spread, estimate := math.Sqrt(variance), expected/float64(e.Total)
		e.Expected, e.Spread, e.Estimate = &expected, &spread, &estimate
		for c, cfg := range p.configs {
			if cfg.Model == r.Model {
				e.SameModel = append(e.SameModel, Count{cfg.Tier, p.score(c, missing)})
			}
		}
		if k >= 4 && len(missing) > 0 {
			unresolved := make([]bool, len(p.questions))
			for _, q := range missing {
				unresolved[q] = true
			}
			var known []int
			for q := range p.questions {
				if !unresolved[q] {
					known = append(known, q)
				}
			}
			predicted := 0.0
			for c := range p.configs {
				heldOut := p.raschDifficulty(c)
				ability := raschAbility(heldOut, known, p.score(c, known))
				for _, q := range missing {
					predicted += logistic(ability - heldOut[q])
				}
			}
			predicted /= float64(k * len(missing))
			e.CheckPredicted = &predicted
		}
		result = append(result, e)
	}
	return result
}

// providers adjusts for question difficulty with the panel excluding every
// configuration of the same model, so a model cannot explain its own results.
func (p panelMatrix) providers(outcomes map[string]map[string]Outcome) []ProviderEffect {
	result := []ProviderEffect{}
	panelCorrect, panelConfigs := make([]float64, len(p.questions)), map[string]int{}
	modelCorrect := map[string][]float64{}
	for c, cfg := range p.configs {
		panelConfigs[cfg.Model]++
		if modelCorrect[cfg.Model] == nil {
			modelCorrect[cfg.Model] = make([]float64, len(p.questions))
		}
		for q, ok := range p.correct[c] {
			if ok {
				panelCorrect[q]++
				modelCorrect[cfg.Model][q]++
			}
		}
	}
	type answer struct {
		provider, tier   string
		question         int
		correct, invalid bool
		residual         float64
	}
	byModel, comparable := map[string][]answer{}, map[string]bool{}
	for _, o := range p.ordered(outcomes) {
		if o.Status != "complete" {
			continue
		}
		q := p.index[o.Question]
		a := answer{provider: o.Provider, tier: o.Tier, question: q, correct: o.Correct, invalid: o.Invalid != ""}
		if a.provider == "" {
			a.provider = "Unknown"
		}
		others := len(p.configs) - panelConfigs[o.Model]
		comparable[o.Model] = others > 0
		if others > 0 {
			own := 0.0
			if modelCorrect[o.Model] != nil {
				own = modelCorrect[o.Model][q]
			}
			a.residual = -(panelCorrect[q] - own) / float64(others)
			if a.correct {
				a.residual++
			}
		}
		byModel[o.Model] = append(byModel[o.Model], a)
	}
	models := make([]string, 0, len(byModel))
	for model := range byModel {
		models = append(models, model)
	}
	sort.Strings(models)
	for _, model := range models {
		answers := byModel[model]
		groups := map[string]*ProviderEffect{}
		tiers := map[string]map[string]int{}
		for _, a := range answers {
			e := groups[a.provider]
			if e == nil {
				e = &ProviderEffect{Model: model, Provider: a.provider, Tiers: []Count{}}
				groups[a.provider] = e
				tiers[a.provider] = map[string]int{}
			}
			e.Answers++
			tiers[a.provider][a.tier]++
			if a.correct {
				e.Correct++
			}
			if a.invalid {
				e.Invalid++
			}
		}
		if len(groups) < 2 {
			continue
		}
		var rows []ProviderEffect
		for provider, e := range groups {
			for tier, n := range tiers[provider] {
				e.Tiers = append(e.Tiers, Count{tier, n})
			}
			sort.Slice(e.Tiers, func(i, j int) bool {
				a, b := e.Tiers[i].Label, e.Tiers[j].Label
				return tierOrder(a) < tierOrder(b) || tierOrder(a) == tierOrder(b) && a < b
			})
			e.Others = len(answers) - e.Answers
			if comparable[model] {
				ownSum, restSum := 0.0, 0.0
				for _, a := range answers {
					if a.provider == provider {
						ownSum += a.residual
					} else {
						restSum += a.residual
					}
				}
				ownMean, restMean := ownSum/float64(e.Answers), restSum/float64(e.Others)
				expected := float64(e.Correct)/float64(e.Answers) - ownMean
				gap := ownMean - restMean
				e.Excess, e.Expected, e.Gap = &ownMean, &expected, &gap
				// One model's tiers share question-specific difficulty, so the
				// difference in means uses standard errors clustered by question.
				clusters, used := make([]float64, len(p.questions)), make([]bool, len(p.questions))
				for _, a := range answers {
					used[a.question] = true
					if a.provider == provider {
						clusters[a.question] += (a.residual - ownMean) / float64(e.Answers)
					} else {
						clusters[a.question] -= (a.residual - restMean) / float64(e.Others)
					}
				}
				g, sum := 0, 0.0
				for q, v := range clusters {
					if used[q] {
						g++
						sum += v * v
					}
				}
				// A group's own variability cannot be estimated from a handful of answers.
				if g >= 2 && e.Answers >= minProviderAnswers && e.Others >= minProviderAnswers {
					se := math.Sqrt(sum * float64(g) / float64(g-1))
					e.GapInterval = &Interval{Low: gap - 1.959963984540054*se, High: gap + 1.959963984540054*se}
				}
			}
			rows = append(rows, *e)
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].Answers != rows[j].Answers {
				return rows[i].Answers > rows[j].Answers
			}
			return rows[i].Provider < rows[j].Provider
		})
		result = append(result, rows...)
	}
	return result
}

// unresolvedDifficulty compares panel accuracy behind unresolved slots with
// panel accuracy across all questions in scope.
func (p panelMatrix) unresolvedDifficulty(outcomes map[string]map[string]Outcome) (slots int, unresolved, overall float64) {
	if len(p.configs) < 3 || len(p.questions) == 0 {
		return 0, 0, 0
	}
	for _, rate := range p.rate {
		overall += rate / float64(len(p.rate))
	}
	for _, o := range p.ordered(outcomes) {
		if o.Status == "complete" || o.Status == "unsupported" || o.Status == "skipped" {
			continue
		}
		slots++
		unresolved += p.rate[p.index[o.Question]]
	}
	if slots > 0 {
		unresolved /= float64(slots)
	}
	return slots, unresolved, overall
}

func sectionName(s Separation) string {
	if s.Level == "subject" {
		return s.Category
	}
	if s.Section == "" {
		return s.Category + " · untagged questions"
	}
	return s.Category + " · " + s.Section
}

func diagnosticInsights(v AnalysisView, p panelMatrix, outcomes map[string]map[string]Outcome) []Insight {
	result := []Insight{}
	percent := func(v float64) string { return fmt.Sprintf("%.1f%%", v*100) }
	if len(v.Ranks) >= 2 {
		leaders := 0
		for _, r := range v.Ranks {
			if r.Low == 1 {
				leaders++
			}
		}
		result = append(result, Insight{Label: "Could rank first", Value: fmt.Sprintf("%d / %d", leaders, len(v.Ranks)), Detail: fmt.Sprintf("Complete configurations whose 90%% rank interval includes first place across %d paired question resamples. Resampling questions does not measure run-to-run variation.", bootstrapResamples)})
	}
	// Report the finest partition of the scope that has more than one group.
	level, subjects := "section", 0
	for _, s := range v.Separation {
		if s.Level == "subject" {
			subjects++
		}
	}
	if subjects > 1 {
		level = "subject"
	}
	var top *Separation
	for i, s := range v.Separation {
		if s.Level == level && s.Share != nil && (top == nil || *s.Share > *top.Share) {
			top = &v.Separation[i]
		}
	}
	if top != nil {
		ceiling := 0
		for _, rate := range p.rate {
			if rate == 1 {
				ceiling++
			}
		}
		result = append(result, Insight{Label: "Where scores separate", Value: percent(*top.Share), Detail: fmt.Sprintf("of the variance in total correct across %d complete configurations comes from %s (%d of %d questions). %d questions were answered correctly by every complete configuration.", len(p.configs), sectionName(*top), top.Questions, len(p.questions), ceiling)})
	}
	if slots, unresolved, overall := p.unresolvedDifficulty(outcomes); slots > 0 {
		result = append(result, Insight{Label: "Difficulty of unresolved answers", Value: fmt.Sprintf("%.0f%% vs %.0f%%", unresolved*100, overall*100), Detail: fmt.Sprintf("Mean panel accuracy is %s on the questions behind %d unresolved answer slots, versus %s across all questions in scope. Lower means harder; when unresolved questions are harder, accuracy on completed answers overstates an unranked configuration.", percent(unresolved), slots, percent(overall))})
	}
	if l := v.Length; l.Longest > 0 {
		detail := fmt.Sprintf("Each complete configuration's longest 10%% of answers by output tokens (%d answers), versus %s across all %d.", l.Longest, percent(float64(l.AnswersCorrect)/float64(l.Answers)), l.Answers)
		if l.Ratio != nil {
			detail += fmt.Sprintf(" Holding question and configuration fixed, wrong answers used %.2f× the output tokens of correct answers.", *l.Ratio)
		}
		result = append(result, Insight{Label: "Accuracy of the longest answers", Value: percent(float64(l.LongestCorrect) / float64(l.Longest)), Detail: detail})
	}
	var worst *ProviderEffect
	for i, e := range v.Providers {
		if e.GapInterval != nil && e.GapInterval.High < 0 && (worst == nil || *e.Gap < *worst.Gap) {
			worst = &v.Providers[i]
		}
	}
	if worst != nil {
		result = append(result, Insight{Label: "Largest provider gap", Value: fmt.Sprintf("%+.1f pp", *worst.Gap*100), Detail: fmt.Sprintf("%s served by %s versus its other providers (%d vs %d answers), adjusted for question difficulty; approximate 95%% interval %+.1f to %+.1f pp. %d of these answers were invalid.", worst.Model, worst.Provider, worst.Answers, worst.Others, worst.GapInterval.Low*100, worst.GapInterval.High*100, worst.Invalid)})
	}
	return result
}
