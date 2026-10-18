package openrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"konkour-bench/internal/dataset"
)

func TestRequestContract(t *testing.T) {
	q := dataset.Question{Text: `سؤال $x^2$`, Number: 79, Category: "SECRET_CATEGORY", Dataset: "SECRET_EXAM", Correct: 4, Asset: &dataset.Asset{MIME: "image/png", Bytes: []byte{1, 2, 3}}, Options: []dataset.Option{
		{Number: 1, Text: "one"}, {Number: 2, Text: "اثنان"}, {Number: 3, Text: "سه"}, {Number: 4, Image: "SECRET_FILENAME", Asset: &dataset.Asset{MIME: "image/png", Bytes: []byte{4, 5, 6}}},
	}}
	b, err := Build(q, "test/model:exacto", "max")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"SECRET", "max_tokens", "max_completion_tokens", "temperature", "top_p", "seed", "correct_option"} {
		if strings.Contains(string(b), forbidden) {
			t.Fatalf("request leaked %s", forbidden)
		}
	}
	var v map[string]any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if v["model"] != "test/model:exacto" {
		t.Fatal(v["model"])
	}
	if !v["provider"].(map[string]any)["require_parameters"].(bool) {
		t.Fatal("provider parameters not required")
	}
	parts := v["messages"].([]any)[1].(map[string]any)["content"].([]any)
	if len(parts) != 5 {
		t.Fatalf("want text + two labeled images, got %d parts", len(parts))
	}
	if !strings.Contains(string(b), "data:image/png;base64,AQID") {
		t.Fatal("missing image bytes")
	}
	q.Correct = 2
	q.Category = "changed"
	q.Dataset = "other"
	b2, _ := Build(q, "test/model", "max")
	if string(b) != string(b2) {
		t.Fatal("internal metadata altered request")
	}
}
func TestStrictAnswers(t *testing.T) {
	cases := []struct {
		content string
		valid   bool
	}{
		{`{"answer":3}`, true}, {` { "answer" : 1 } `, true}, {`{"answer":0}`, false}, {`{"answer":5}`, false},
		{`{"answer":"3"}`, false}, {`{"answer":3.0}`, false}, {`{"answer":3,"explanation":"x"}`, false},
		{`{"answer":3,"answer":2}`, false}, {"```json\n{\"answer\":3}\n```", false}, {`{"answer":3} {}`, false}, {`null`, false}, {`{}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.content, func(t *testing.T) {
			body, _ := json.Marshal(map[string]any{"id": "gen-test", "choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"content": tc.content}}}})
			_, answer, invalid, err := Parse(body)
			if err != nil {
				t.Fatal(err)
			}
			if (answer > 0 && invalid == "") != tc.valid {
				t.Fatalf("answer=%d invalid=%s", answer, invalid)
			}
		})
	}
	for _, raw := range []string{`{"choices":[{"finish_reason":"length","message":{"content":"{\"answer\":1}"}}]}`, `{"choices":[{"finish_reason":"stop","message":{"refusal":"no","content":null}}]}`} {
		_, a, invalid, err := Parse([]byte(raw))
		if err != nil || a != 0 || invalid == "" {
			t.Fatalf("unexpected refusal/truncation: %d %s %v", a, invalid, err)
		}
	}
}
func TestEffortBands(t *testing.T) {
	var m Model
	if err := json.Unmarshal([]byte(`{"supported_parameters":["reasoning","structured_outputs"],"reasoning":{"supported_efforts":["low","high"]}}`), &m); err != nil {
		t.Fatal(err)
	}
	for tier, want := range map[string]string{"minimal": "low", "max": "high", "medium": ""} {
		v, _ := m.Candidates(tier)
		if strings.Join(v, ",") != want {
			t.Fatalf("%s: %v", tier, v)
		}
	}
	m.Reasoning.Efforts = nil
	if v, _ := m.Candidates("max"); len(v) != 0 {
		t.Fatal("nonselectable reasoning accepted")
	}
	m.Reasoning.Efforts = json.RawMessage("null")
	if v, _ := m.Candidates("max"); strings.Join(v, ",") != "max,xhigh,high" {
		t.Fatal(v)
	}
}
func TestEffortErrorsAreSpecific(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   bool
	}{{400, `unsupported reasoning effort max`, true}, {422, `effort must be one of low, high`, true}, {400, `invalid JSON schema`, false}, {500, `unsupported effort`, false}} {
		if got := EffortRejected(HTTPResult{Status: tc.status, Body: []byte(tc.body)}); got != tc.want {
			t.Fatalf("%+v: %v", tc, got)
		}
	}
}
func TestCredentialRedactionAndRedirectProtection(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-secret" {
			t.Error("missing authorization")
		}
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "http://example.invalid", 302)
			return
		}
		fmtBody := `{"error":"test-secret"}`
		_, _ = w.Write([]byte(fmtBody))
	}))
	defer s.Close()
	c := New(s.URL, "test-secret")
	r := c.Do(context.Background(), "GET", "/", nil)
	if strings.Contains(string(r.Body), "test-secret") {
		t.Fatal("secret leaked")
	}
	r = c.Do(context.Background(), "GET", "/redirect", nil)
	if r.Status != 302 {
		t.Fatalf("redirect followed: %+v", r)
	}
}
