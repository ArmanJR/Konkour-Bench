package dataset

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T, edit func(map[string]any, map[string]any)) string {
	t.Helper()
	dir := t.TempDir()
	q := map[string]any{"questions": []any{map[string]any{"number": 2, "category": "ریاضی", "question": `مقدار $x^2$؟`, "options": []any{
		map[string]any{"number": 1, "text": "یک"}, map[string]any{"number": 2, "text": "two"}, map[string]any{"number": 3, "text": "ثلاثة"}, map[string]any{"number": 4, "text": nil, "image": "option.png"},
	}}}}
	a := map[string]any{"answers": []any{map[string]any{"number": 2, "correct_option": 4}, map[string]any{"number": 99, "correct_option": 1}}}
	if edit != nil {
		edit(q, a)
	}
	var imageBytes bytes.Buffer
	if err := png.Encode(&imageBytes, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "option.png"), imageBytes.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]any{"questions.json": q, "answers.json": a, "config.json": Config{Datasets: []Source{{ID: "exam", Questions: "questions.json", Answers: "answers.json"}}, Models: []string{"test/model"}}} {
		b, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, "config.json")
}

func TestLoadPreservesContentAndAllowsExtraKeys(t *testing.T) {
	s, err := Load(fixture(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	q := s.Questions[0]
	if q.ID() != "exam:2" || q.Correct != 4 || q.Text != `مقدار $x^2$؟` || q.Options[2].Text != "ثلاثة" || !q.HasImage() || q.Options[3].Asset.MIME != "image/png" {
		t.Fatalf("unexpected question: %+v", q)
	}
	if len(s.Warnings) != 1 || !strings.Contains(s.Warnings[0], "1 extra") {
		t.Fatalf("warnings: %v", s.Warnings)
	}
}
func TestValidationFailures(t *testing.T) {
	cases := []struct {
		name, want string
		edit       func(map[string]any, map[string]any)
	}{
		{"missing key", "missing answer key", func(_ map[string]any, a map[string]any) { a["answers"] = []any{} }},
		{"bad key", "invalid or duplicate answer", func(_ map[string]any, a map[string]any) {
			a["answers"] = []any{map[string]any{"number": 2, "correct_option": 5}}
		}},
		{"duplicate question", "duplicate question", func(q map[string]any, _ map[string]any) {
			qs := q["questions"].([]any)
			q["questions"] = append(qs, qs[0])
		}},
		{"three options", "four options", func(q map[string]any, _ map[string]any) {
			v := q["questions"].([]any)[0].(map[string]any)
			v["options"] = v["options"].([]any)[:3]
		}},
		{"duplicate option", "duplicate option", func(q map[string]any, _ map[string]any) {
			v := q["questions"].([]any)[0].(map[string]any)
			v["options"].([]any)[1].(map[string]any)["number"] = 1
		}},
		{"empty option", "has no content", func(q map[string]any, _ map[string]any) {
			v := q["questions"].([]any)[0].(map[string]any)
			v["options"].([]any)[0].(map[string]any)["text"] = " "
		}},
		{"missing image", "no-such-file", func(q map[string]any, _ map[string]any) {
			q["questions"].([]any)[0].(map[string]any)["image"] = "no-such-file.png"
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(fixture(t, tc.edit))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %q", err, tc.want)
			}
		})
	}
}
func TestCorruptImage(t *testing.T) {
	path := fixture(t, nil)
	b := append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 40)...)
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "option.png"), b, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("corrupt image accepted")
	}
}
func TestSuppliedSample(t *testing.T) {
	s, err := Load("../../benchmark.sample.json")
	if err != nil {
		t.Fatal(err)
	}
	images := 0
	for _, q := range s.Questions {
		if q.HasImage() {
			images++
		}
	}
	if len(s.Questions) != 15 || images != 2 || len(s.Config.Models) != 15 {
		t.Fatalf("unexpected sample: %d questions, %d images", len(s.Questions), images)
	}
}

func TestHumanities1405Booklet(t *testing.T) {
	s, err := Load("../../benchmark.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Questions) != 80 || len(s.Config.Models) != 13 {
		t.Fatalf("unexpected booklet/model count: %d/%d", len(s.Questions), len(s.Config.Models))
	}
	wantImages := map[int]bool{1: true, 9: true, 14: true, 39: true, 52: true, 64: true}
	categories, sections := map[string]int{}, map[string][]int{}
	assets := 0
	for _, q := range s.Questions {
		if q.Dataset != "humanities-1405-booklet-1" || q.Number < 1 || q.Number > 80 || q.HasImage() != wantImages[q.Number] {
			t.Fatalf("unexpected question: %s, image=%v", q.ID(), q.HasImage())
		}
		categories[q.Category]++
		if q.Section != "" {
			sections[q.Section] = append(sections[q.Section], q.Number)
		}
		if q.Asset != nil {
			assets++
		}
		for _, option := range q.Options {
			if option.Asset != nil {
				assets++
			}
		}
		if q.Number == 15 && (q.Options[1].Text != `$\frac{3}{8}$` || q.Correct != 1) {
			t.Fatal("question 15 must preserve the source PDF's option 2 and supplied answer key")
		}
	}
	for category, want := range map[string]int{"ریاضی": 20, "زبان و ادبیات فارسی": 30, "علوم اجتماعی": 15, "روان‌شناسی": 15} {
		if categories[category] != want {
			t.Errorf("category %s: got %d, want %d", category, categories[category], want)
		}
	}
	// Literature follows the booklet's section order; question 45 combines several sections.
	for section, want := range map[string][2]int{"تاریخ ادبیات و سبک\u200cشناسی": {21, 28}, "آرایه\u200cهای ادبی": {29, 36}, "عروض و قافیه": {37, 44}, "ترکیبی": {45, 45}, "قرابت معنایی": {46, 50}} {
		got := sections[section]
		if len(got) != want[1]-want[0]+1 || got[0] != want[0] || got[len(got)-1] != want[1] {
			t.Errorf("section %s: got questions %v, want %d-%d", section, got, want[0], want[1])
		}
	}
	if len(sections) != 5 {
		t.Errorf("unexpected sections: %v", sections)
	}
	if assets != 12 || len(s.Warnings) != 1 || !strings.Contains(s.Warnings[0], "120 extra") {
		t.Fatalf("assets=%d warnings=%v", assets, s.Warnings)
	}
}
