package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

func Write(dir string, d Document) error {
	if d.AnalysisVersion == "" {
		d.AnalysisVersion = "3"
	}
	if d.Views == nil {
		d.Views = Analyze(d)
	}
	// Render before replacing any artifact so template/data errors preserve the
	// previous report. HTML embeds compact JSON and all runtime assets.
	compact, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("encode report: %w", err)
	}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	title := "Benchmark results"
	if len(d.Datasets) > 0 {
		title = strings.ReplaceAll(strings.ReplaceAll(strings.Join(d.Datasets, " + "), "-booklet-", " · booklet "), "-", " ")
		runes := []rune(title)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			title = string(runes)
		}
	}
	// encoding/json escapes HTML-significant characters before embedding the data.
	logos, err := logoRuntime()
	if err != nil {
		return err
	}
	runtime := logos + "\n" + chartScript + "\n" + script
	fontCSS := chartFontCSS()
	view := struct {
		Document
		Title      string
		Data       template.JS
		Script     template.JS
		MathScript template.JS
		Style      template.CSS
		MathStyle  template.CSS
		FontStyle  template.CSS
	}{d, title, template.JS(compact), template.JS(runtime), template.JS(mathScript), template.CSS(stylesheet), template.CSS(mathStylesheet), template.CSS(fontCSS)}
	t, err := template.New("report").Parse(page)
	if err != nil {
		return err
	}
	var html bytes.Buffer
	if err := t.Execute(&html, view); err != nil {
		return fmt.Errorf("render report: %w", err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	files := map[string][]byte{"results.json": b, "index.html": html.Bytes(), "report.js": []byte(runtime), "report.css": []byte(fontCSS + "\n" + stylesheet)}
	for _, metric := range []string{"cost", "tokens", "accuracy"} {
		files[metric+".svg"] = []byte(staticSVG(d.Rows, metric))
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), body, 0644); err != nil {
			return fmt.Errorf("write report artifact %s: %w", name, err)
		}
	}
	if err := writeCSV(filepath.Join(dir, "leaderboard.csv"), d.Rows); err != nil {
		return err
	}
	if err := writeAnalysisCSVs(dir, d); err != nil {
		return err
	}
	slog.Info("report generated", "directory", dir, "outcomes", len(d.Outcomes), "analysis_scopes", len(d.Views))
	return nil
}

func writeCSV(path string, rows []Row) error {
	header := []string{"model", "tier", "sent_efforts", "cohort", "dataset", "category", "status", "questions", "completed", "correct", "incorrect", "accuracy", "cost_usd", "total_tokens", "reasoning_tokens", "cost_frontier", "token_frontier", "pareto_3d", "accuracy_ci_low", "accuracy_ci_high", "completion_bound_low", "completion_bound_high", "invalid"}
	for _, metric := range []string{"cost_usd", "tokens", "latency_seconds"} {
		header = append(header, metric+"_sample_n", metric+"_median", metric+"_p90")
	}
	records := make([][]string, 0, len(rows))
	for _, r := range rows {
		record := []string{r.Model, r.Tier, r.SentEfforts, r.Cohort, r.Dataset, r.Category, r.Status, strconv.Itoa(r.Total), strconv.Itoa(r.Complete), strconv.Itoa(r.Correct), strconv.Itoa(r.Incorrect), csvNumber(r.Accuracy), csvNumber(r.Cost), csvInteger(r.Tokens), csvInteger(r.Reasoning), strconv.FormatBool(r.CostFrontier), strconv.FormatBool(r.TokenFrontier), strconv.FormatBool(r.Pareto)}
		// Prevent spreadsheet formula execution in user-controlled labels.
		for i := 0; i < 7; i++ {
			record[i] = csvText(record[i])
		}
		record = append(record, intervalCSV(r.Interval)...)
		record = append(record, intervalCSV(r.Bounds)...)
		record = append(record, strconv.Itoa(r.Invalid))
		for _, d := range []Distribution{r.CostDistribution, r.TokenDistribution, r.Latency} {
			record = append(record, strconv.Itoa(d.N), csvNumber(d.Median), csvNumber(d.P90))
		}
		records = append(records, record)
	}
	return writeRecords(path, header, records)
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
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="1000" height="%d" viewBox="0 0 1000 %d"><style>%s</style><rect width="1000" height="%d" fill="white"/><g font-family="Inter, sans-serif" letter-spacing="0.15" fill="#17212b">`, height, height, template.HTMLEscapeString(chartFontCSS()), height)
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
