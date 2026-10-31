package report

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func csvText(s string) string {
	trimmed := strings.TrimLeft(s, " \t\r\n")
	if strings.ContainsAny(s, "\t\r\n") || len(trimmed) > 0 && strings.ContainsAny(trimmed[:1], "=+-@") {
		return "'" + s
	}
	return s
}
func csvNumber(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'g', -1, 64)
}
func csvInteger(v *int64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatInt(*v, 10)
}
func intervalCSV(v *Interval) []string {
	if v == nil {
		return []string{"", ""}
	}
	return []string{csvNumber(&v.Low), csvNumber(&v.High)}
}

func writeRecords(path string, header []string, records [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create CSV %s: %w", path, err)
	}
	w := csv.NewWriter(f)
	w.WriteAll(append([][]string{header}, records...))
	writeErr, closeErr := w.Error(), f.Close()
	if writeErr != nil {
		return fmt.Errorf("write CSV %s: %w", path, writeErr)
	}
	return closeErr
}

func csvCounts(counts []Count, denominator int) string {
	parts := make([]string, 0, len(counts))
	for _, c := range counts {
		part := c.Label + " " + strconv.Itoa(c.N)
		if denominator > 0 {
			part += "/" + strconv.Itoa(denominator)
		}
		parts = append(parts, part)
	}
	return csvText(strings.Join(parts, "; "))
}

func writeAnalysisCSVs(dir string, d Document) error {
	var questions, efforts, outcomes, ranks, separation, length, estimates, providers [][]string
	for _, v := range d.Views {
		scope := []string{csvText(v.Cohort), csvText(v.Dataset), csvText(v.Category)}
		row := func(values ...string) []string { return append(append([]string{}, scope...), values...) }
		for _, q := range v.Questions {
			r := row(csvText(q.ID), csvText(q.Category), strconv.FormatBool(q.Image), strconv.Itoa(q.Expected), strconv.Itoa(q.Panel), strconv.Itoa(q.Correct))
			for _, n := range q.Answers {
				r = append(r, strconv.Itoa(n))
			}
			questions = append(questions, append(r, strconv.FormatBool(q.Review), csvText(q.Section)))
		}
		for _, r := range v.Ranks {
			ranks = append(ranks, row(csvText(r.Model), csvText(r.Tier), strconv.Itoa(r.Rank), strconv.Itoa(r.Low), strconv.Itoa(r.Median), strconv.Itoa(r.High), csvNumber(&r.First)))
		}
		for _, s := range v.Separation {
			separation = append(separation, row(s.Level, csvText(s.Category), csvText(s.Section), strconv.Itoa(s.Questions), strconv.Itoa(s.Ceiling), strconv.Itoa(s.Floor), csvNumber(&s.Mean), csvNumber(&s.Min), csvNumber(&s.Max), csvNumber(s.Share)))
		}
		for _, l := range v.Length.Rows {
			length = append(length, row(csvText(l.Model), csvText(l.Tier), strconv.Itoa(l.Correct), strconv.Itoa(l.Wrong), csvNumber(l.CorrectMedian), csvNumber(l.WrongMedian), csvNumber(l.Ratio), strconv.Itoa(l.Longest), strconv.Itoa(l.LongestRight), csvNumber(v.Length.Ratio)))
		}
		for _, e := range v.Estimates {
			low, high := float64(e.Correct)/float64(e.Total), float64(e.Correct+e.Unresolved)/float64(e.Total)
			estimates = append(estimates, row(csvText(e.Model), csvText(e.Tier), strconv.Itoa(e.Total), strconv.Itoa(e.Complete), strconv.Itoa(e.Correct), strconv.Itoa(e.Unresolved), csvNumber(&low), csvNumber(&high), csvNumber(e.MissingDifficulty), csvNumber(e.CompletedDifficulty), csvNumber(e.Expected), csvNumber(e.Spread), csvNumber(e.Estimate), csvNumber(e.CheckPredicted), csvCounts(e.SameModel, e.Unresolved)))
		}
		for _, e := range v.Providers {
			accuracy := float64(e.Correct) / float64(e.Answers)
			providers = append(providers, append(row(csvText(e.Model), csvText(e.Provider), csvCounts(e.Tiers, 0), strconv.Itoa(e.Answers), strconv.Itoa(e.Correct), strconv.Itoa(e.Invalid), csvNumber(&accuracy), csvNumber(e.Expected), csvNumber(e.Excess), strconv.Itoa(e.Others), csvNumber(e.Gap)), intervalCSV(e.GapInterval)...))
		}
		for _, e := range v.Efforts {
			r := append([]string{}, scope...)
			r = append(r, csvText(e.Model), csvText(e.From), csvText(e.To), strconv.Itoa(e.Total), strconv.Itoa(e.Paired), strconv.Itoa(e.Gained), strconv.Itoa(e.Lost), strconv.Itoa(e.BothCorrect), strconv.Itoa(e.BothWrong), csvNumber(e.Delta), csvNumber(e.CostRatio), csvNumber(e.TokenRatio))
			efforts = append(efforts, r)
		}
	}
	for _, o := range d.Outcomes {
		r := []string{o.Question, o.Dataset, o.Category, o.Model, o.Tier, o.Effort, o.Backend, o.Status, o.Invalid, o.Provider, o.RunID, o.OriginRun, o.Hash}
		for i := range r {
			r[i] = csvText(r[i])
		}
		r = append(r, strconv.FormatBool(o.Image), strconv.Itoa(o.Expected), strconv.Itoa(o.Answer), strconv.FormatBool(o.Correct), csvNumber(o.Usage.Cost), csvInteger(o.Usage.Total), csvInteger(o.Usage.Completion), csvInteger(o.Usage.CompletionDetails.Reasoning), csvNumber(o.Duration), csvText(o.Section))
		outcomes = append(outcomes, r)
	}
	for _, export := range []struct {
		name   string
		header []string
		rows   [][]string
	}{
		{"questions.csv", []string{"cohort", "dataset", "category", "question", "question_category", "has_image", "expected", "panel_configurations", "correct", "invalid_answers", "option_1", "option_2", "option_3", "option_4", "review_candidate", "question_section"}, questions},
		{"effort.csv", []string{"cohort", "dataset", "category", "model", "from", "to", "selected_questions", "paired_questions", "gained", "lost", "both_correct", "both_wrong", "accuracy_delta", "cost_ratio", "token_ratio"}, efforts},
		{"outcomes.csv", []string{"question", "dataset", "category", "model", "tier", "sent_effort", "backend", "status", "invalid", "provider", "selected_run", "origin_run", "request_hash", "has_image", "expected", "answer", "correct", "cost_usd", "total_tokens", "completion_tokens", "reasoning_tokens", "duration_seconds", "section"}, outcomes},
		{"ranks.csv", []string{"cohort", "dataset", "category", "model", "tier", "rank", "rank_p05", "rank_median", "rank_p95", "first_share"}, ranks},
		{"separation.csv", []string{"cohort", "dataset", "category", "level", "subject", "section", "questions", "solved_by_all", "solved_by_none", "mean_accuracy", "min_accuracy", "max_accuracy", "variance_share"}, separation},
		{"length.csv", []string{"cohort", "dataset", "category", "model", "tier", "correct_answers", "wrong_answers", "correct_median_output_tokens", "wrong_median_output_tokens", "wrong_to_correct_ratio", "longest_decile_answers", "longest_decile_correct", "scope_within_question_ratio"}, length},
		{"estimates.csv", []string{"cohort", "dataset", "category", "model", "tier", "questions", "completed", "correct", "unresolved", "bound_low", "bound_high", "unresolved_panel_accuracy", "completed_panel_accuracy", "expected_correct", "expected_correct_sd", "estimate", "check_predicted", "same_model_on_unresolved"}, estimates},
		{"providers.csv", []string{"cohort", "dataset", "category", "model", "provider", "tiers", "answers", "correct", "invalid", "accuracy", "expected_accuracy", "excess", "other_answers", "gap", "gap_ci_low", "gap_ci_high"}, providers},
	} {
		if err := writeRecords(filepath.Join(dir, export.name), export.header, export.rows); err != nil {
			return err
		}
	}
	return nil
}
