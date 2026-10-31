package report

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"
)

type diagnosticConfig struct {
	model, tier string
	correct     []bool
	status      map[int]string // Question index -> non-complete status.
	provider    func(i int) string
	tokens      func(i int) *int64
}

// diagnosticDocument builds one full-exam scope over questions exam:1..n.
func diagnosticDocument(configs []diagnosticConfig, category, section func(i int) string) Document {
	d := Document{}
	for _, c := range configs {
		var group []Outcome
		for i, ok := range c.correct {
			o := outcome(ok)
			o.Model, o.Tier = c.model, c.tier
			o.Question, o.Dataset, o.Expected = fmt.Sprintf("exam:%d", i+1), "exam", 1
			o.Category, o.Section = category(i), section(i)
			o.Answer = 2
			if ok {
				o.Answer = 1
			}
			if status := c.status[i]; status != "" {
				o.Status, o.Correct, o.Answer = status, false, 0
			}
			if c.provider != nil {
				o.Provider = c.provider(i)
			}
			if c.tokens != nil {
				o.Usage.Completion = c.tokens(i)
			}
			group = append(group, o)
			d.Outcomes = append(d.Outcomes, o)
		}
		r := Aggregate(group)
		r.Cohort = "full"
		d.Rows = append(d.Rows, r)
	}
	d.Views = Analyze(d)
	return d
}

func pattern(s string) []bool {
	v := make([]bool, len(s))
	for i, c := range s {
		v[i] = c == '1'
	}
	return v
}

func single(string) func(int) string { return func(int) string { return "subject" } }
func none(int) string                { return "" }

func TestQuestionOrderIsNumericWithinDataset(t *testing.T) {
	ids := []string{"b:1", "a:10", "a:2", "a:1"}
	slices.SortFunc(ids, func(x, y string) int {
		if questionLess(x, y) {
			return -1
		}
		return 1
	})
	if strings.Join(ids, ",") != "a:1,a:2,a:10,b:1" {
		t.Fatal(ids)
	}
}

func TestRankIntervalsArePairedDeterministicAndShareTies(t *testing.T) {
	d := diagnosticDocument([]diagnosticConfig{
		{model: "test/a", tier: "max", correct: pattern("111111111111")},
		{model: "test/b", tier: "max", correct: pattern("101010101010")},
		{model: "test/c", tier: "max", correct: pattern("101010101010")},
	}, single(""), none)
	v := d.Views[0]
	if len(v.Ranks) != 3 {
		t.Fatalf("ranks: %+v", v.Ranks)
	}
	a, b, c := v.Ranks[0], v.Ranks[1], v.Ranks[2]
	if a.Model != "test/a" || a.Rank != 1 || a.Low != 1 || a.High != 1 || a.First != 1 {
		t.Fatalf("dominant configuration must always rank first: %+v", a)
	}
	// Identical answers are paired, so the configurations tie in every resample.
	if b.Rank != 2 || b.Low != c.Low || b.Median != c.Median || b.High != c.High || b.First != c.First || b.Low != 2 {
		t.Fatalf("tied configurations: %+v %+v", b, c)
	}
	if again := Analyze(d)[0].Ranks; !reflect.DeepEqual(again, v.Ranks) {
		t.Fatal("bootstrap is not reproducible")
	}
	found := false
	for _, insight := range v.Insights {
		found = found || insight.Label == "Could rank first" && insight.Value == "1 / 3"
	}
	if !found {
		t.Fatalf("leader insight: %+v", v.Insights)
	}
}

func TestSeparationSharesPartitionVariance(t *testing.T) {
	category := func(i int) string {
		if i < 2 {
			return "math"
		}
		return "lit"
	}
	section := func(i int) string {
		switch {
		case i < 2:
			return ""
		case i < 4:
			return "devices"
		}
		return "meter"
	}
	d := diagnosticDocument([]diagnosticConfig{
		{model: "test/a", tier: "max", correct: pattern("111110")},
		{model: "test/b", tier: "max", correct: pattern("101100")},
		{model: "test/c", tier: "max", correct: pattern("111000")},
		{model: "test/d", tier: "max", correct: pattern("101111")},
	}, category, section)
	v := d.Views[0]
	if len(v.Separation) != 4 {
		t.Fatalf("separation rows: %+v", v.Separation)
	}
	math0, lit, devices, meter := v.Separation[0], v.Separation[1], v.Separation[2], v.Separation[3]
	if math0.Category != "math" || math0.Level != "subject" || lit.Category != "lit" || devices.Section != "devices" || meter.Section != "meter" || devices.Level != "section" {
		t.Fatal("subjects and sections must follow booklet order", v.Separation)
	}
	if math.Abs(*math0.Share+*lit.Share-1) > 1e-12 || math.Abs(*devices.Share+*meter.Share-*lit.Share) > 1e-12 {
		t.Fatalf("shares do not partition variance: %v %v %v %v", *math0.Share, *lit.Share, *devices.Share, *meter.Share)
	}
	// Questions 1 and 3 are answered correctly by every configuration.
	if math0.Ceiling != 1 || math0.Floor != 0 || meter.Floor != 0 || devices.Ceiling != 1 || meter.Ceiling != 0 {
		t.Fatalf("ceiling/floor: %+v %+v %+v", math0, devices, meter)
	}
	if !reflect.DeepEqual(math0.Correct, []int{2, 1, 2, 1}) || !reflect.DeepEqual(v.Panel[0], PanelConfig{"test/a", "max"}) {
		t.Fatalf("scores not aligned with panel: %+v %+v", math0.Correct, v.Panel)
	}
	if math0.Min != .5 || math0.Max != 1 || math.Abs(math0.Mean-.75) > 1e-12 {
		t.Fatalf("spread: %+v", math0)
	}
	equal := diagnosticDocument([]diagnosticConfig{
		{model: "test/a", tier: "max", correct: pattern("10")},
		{model: "test/b", tier: "max", correct: pattern("01")},
	}, single(""), none)
	if s := equal.Views[0].Separation; len(s) != 1 || s[0].Share != nil {
		t.Fatalf("equal totals have no variance to attribute: %+v", s)
	}
}

func TestLengthFixedEffectsRecoverWithinQuestionRatio(t *testing.T) {
	base := []float64{100, 400, 50, 900, 200, 300}
	configs := []diagnosticConfig{
		{model: "test/a", tier: "max", correct: pattern("011111")},
		{model: "test/b", tier: "max", correct: pattern("101110")},
		{model: "test/c", tier: "max", correct: pattern("110101")},
		{model: "test/d", tier: "max", correct: pattern("111100")},
	}
	for c := range configs {
		multiplier, correct := float64(c+1), configs[c].correct
		configs[c].tokens = func(i int) *int64 {
			if c == 3 && i == 0 {
				return nil // Unknown accounting leaves an unbalanced panel.
			}
			v := base[i] * multiplier
			if !correct[i] {
				v *= 3
			}
			n := int64(v)
			return &n
		}
	}
	l := diagnosticDocument(configs, single(""), none).Views[0].Length
	if l.Ratio == nil || math.Abs(*l.Ratio-3) > 1e-9 {
		t.Fatalf("within-question ratio: %v", l.Ratio)
	}
	// Question 4 is correct everywhere, so it has no within-question contrast.
	if l.Answers != 23 || l.Longest != 4 || len(l.Rows) != 4 || l.Contrasts != 5 {
		t.Fatalf("length summary: %+v", l)
	}
	// test/a: correct 400,50,900,200,300; wrong 100*3. Longest is the 900-token correct answer.
	a := l.Rows[0]
	if a.Correct != 5 || a.Wrong != 1 || *a.CorrectMedian != 300 || *a.WrongMedian != 300 || *a.Ratio != 1 || a.Longest != 1 || a.LongestRight != 1 {
		t.Fatalf("per-configuration summary: %+v", a)
	}
	longest := 0
	for _, r := range l.Rows {
		longest += r.LongestRight
	}
	if l.LongestCorrect != longest {
		t.Fatal("pooled longest decile does not match rows")
	}
}

func TestRaschAbilitySolvesExpectedScore(t *testing.T) {
	difficulty := []float64{-2, -1, 0, 1, 2}
	questions := []int{0, 1, 2, 3, 4}
	theta := raschAbility(difficulty, questions, 3)
	sum := 0.0
	for _, b := range difficulty {
		sum += logistic(theta - b)
	}
	if math.Abs(sum-3) > 1e-9 {
		t.Fatalf("ability %v gives %v expected correct", theta, sum)
	}
	if raschAbility(difficulty, questions, 5) < 9.9 || raschAbility(difficulty, questions, 0) > -9.9 {
		t.Fatal("perfect and zero scores must stay bounded")
	}
}

func TestCompletionEstimatesUseDifficultyAndOnlyUnresolvedGaps(t *testing.T) {
	panel := []diagnosticConfig{
		{model: "test/a", tier: "medium", correct: pattern("11111100")},
		{model: "test/b", tier: "max", correct: pattern("11111110")},
		{model: "test/c", tier: "max", correct: pattern("11111000")},
		{model: "test/d", tier: "max", correct: pattern("11110101")},
	}
	incomplete := diagnosticConfig{model: "test/a", tier: "max", correct: pattern("11101100"), status: map[int]string{6: "infra_error", 7: "infra_error"}}
	skipped := diagnosticConfig{model: "test/e", tier: "max", correct: pattern("11111100"), status: map[int]string{7: "skipped"}}
	v := diagnosticDocument(append(panel, incomplete, skipped), single(""), none).Views[0]
	if len(v.Estimates) != 1 {
		t.Fatalf("only unresolved gaps are estimated: %+v", v.Estimates)
	}
	e := v.Estimates[0]
	if e.Model != "test/a" || e.Tier != "max" || e.Correct != 5 || e.Complete != 6 || e.Unresolved != 2 || e.Total != 8 {
		t.Fatalf("estimate counts: %+v", e)
	}
	// Panel accuracy is 1 on questions 1-4, 3/4 on 5-6, and 1/4 on the unresolved 7-8.
	if math.Abs(*e.MissingDifficulty-.25) > 1e-12 || math.Abs(*e.CompletedDifficulty-5.5/6) > 1e-12 {
		t.Fatalf("difficulty: %v %v", *e.MissingDifficulty, *e.CompletedDifficulty)
	}
	if *e.Estimate < 5.0/8 || *e.Estimate > 7.0/8 || math.Abs(*e.Estimate-*e.Expected/8) > 1e-12 || *e.Spread <= 0 {
		t.Fatalf("estimate outside bounds: %+v", e)
	}
	if !reflect.DeepEqual(e.SameModel, []Count{{"medium", 0}}) {
		t.Fatalf("same-model evidence: %+v", e.SameModel)
	}
	if e.CheckPredicted == nil || *e.CheckPredicted <= 0 || *e.CheckPredicted >= 1 {
		t.Fatalf("held-out check: %v", e.CheckPredicted)
	}
	found := false
	for _, insight := range v.Insights {
		// Two unresolved slots on questions with 25% panel accuracy; 75% across all questions.
		found = found || insight.Label == "Difficulty of unresolved answers" && insight.Value == "25% vs 75%" && strings.Contains(insight.Detail, " 2 unresolved")
	}
	if !found {
		t.Fatalf("unresolved insight: %+v", v.Insights)
	}
}

func TestProviderEffectsUseLeaveModelOutDifficulty(t *testing.T) {
	n := 24
	other := func(model string, s string) diagnosticConfig {
		return diagnosticConfig{model: model, tier: "max", correct: pattern(strings.Repeat(s, n/len(s)))}
	}
	served := func(i int) string {
		if i < 12 {
			return "Slow"
		}
		return "Fast"
	}
	// test/a answers questions 1-12 through Slow and misses most of them.
	a := func(tier, s string) diagnosticConfig {
		return diagnosticConfig{model: "test/a", tier: tier, correct: pattern(s), provider: served}
	}
	d := diagnosticDocument([]diagnosticConfig{
		a("max", "000100010001111111111111"),
		a("medium", "001000100011111111111110"),
		other("test/b", "110"),
		other("test/c", "101"),
	}, single(""), none)
	v := d.Views[0]
	if len(v.Providers) != 2 {
		t.Fatalf("single-provider models must be omitted: %+v", v.Providers)
	}
	fast, slow := v.Providers[0], v.Providers[1]
	if slow.Provider != "Slow" || fast.Provider != "Fast" || slow.Answers != 24 || slow.Others != 24 || !reflect.DeepEqual(slow.Tiers, []Count{{"medium", 12}, {"max", 12}}) {
		t.Fatalf("provider groups: %+v %+v", slow, fast)
	}
	// test/b and test/c are both right on 1/3 of questions and split on the rest: 2/3 expected.
	rates := []float64{1, .5, .5}
	expected := 0.0
	for i := range 12 {
		expected += rates[i%3] / 12
	}
	if math.Abs(*slow.Expected-expected) > 1e-12 || math.Abs(*slow.Excess-(float64(slow.Correct)/24-expected)) > 1e-12 {
		t.Fatalf("leave-model-out expectation: %v / %v", *slow.Expected, expected)
	}
	if math.Abs(*slow.Gap+*fast.Gap) > 1e-12 || math.Abs(*slow.Gap-(*slow.Excess-*fast.Excess)) > 1e-12 {
		t.Fatalf("gap: %v %v", *slow.Gap, *fast.Gap)
	}
	// Independent clustered standard error for the difference in mean residuals.
	type answer struct {
		question int
		slow     bool
		residual float64
	}
	var answers []answer
	for _, o := range d.Outcomes {
		if o.Model != "test/a" {
			continue
		}
		var q int
		fmt.Sscanf(o.Question, "exam:%d", &q)
		r := -rates[(q-1)%3]
		if o.Correct {
			r++
		}
		answers = append(answers, answer{q, o.Provider == "Slow", r})
	}
	meanOf := func(slow bool) float64 {
		s, n := 0.0, 0.0
		for _, a := range answers {
			if a.slow == slow {
				s += a.residual
				n++
			}
		}
		return s / n
	}
	ms, mf := meanOf(true), meanOf(false)
	clusters := map[int]float64{}
	for _, a := range answers {
		if a.slow {
			clusters[a.question] += (a.residual - ms) / 24
		} else {
			clusters[a.question] -= (a.residual - mf) / 24
		}
	}
	sum := 0.0
	for _, v := range clusters {
		sum += v * v
	}
	se := math.Sqrt(sum * float64(len(clusters)) / float64(len(clusters)-1))
	if slow.GapInterval == nil || math.Abs(slow.GapInterval.Low-(*slow.Gap-1.959963984540054*se)) > 1e-12 || math.Abs(slow.GapInterval.High-(*slow.Gap+1.959963984540054*se)) > 1e-12 {
		t.Fatalf("clustered interval: %+v, se %v", slow.GapInterval, se)
	}
	found := false
	for _, insight := range v.Insights {
		found = found || insight.Label == "Largest provider gap" && strings.Contains(insight.Detail, "test/a served by Slow")
	}
	if !found {
		t.Fatalf("provider insight: %+v", v.Insights)
	}
	// Small groups get no interval: their own variability is not estimable.
	small := diagnosticDocument([]diagnosticConfig{
		{model: "test/a", tier: "max", correct: pattern("110111"), provider: func(i int) string {
			if i == 0 {
				return ""
			}
			return "Main"
		}},
		{model: "test/b", tier: "max", correct: pattern("110110")},
		{model: "test/c", tier: "max", correct: pattern("101101")},
	}, single(""), none).Views[0].Providers
	if len(small) != 2 || small[0].Provider != "Main" || small[1].Provider != "Unknown" || small[0].Gap == nil || small[0].GapInterval != nil || small[1].GapInterval != nil {
		t.Fatalf("small provider groups: %+v", small)
	}
}
