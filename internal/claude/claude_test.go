package claude

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
)

const testModel = "anthropic/claude-sonnet-5.5"
const initEvent = `{"type":"system","subtype":"init","session_id":"session-test","model":"claude-sonnet-5-5","claude_code_version":"2.1.296","tools":["StructuredOutput"],"mcp_servers":[]}`
const resultLine = `{"type":"result","subtype":"success","session_id":"session-test","is_error":false,"stop_reason":"tool_use","terminal_reason":"completed","structured_output":{"answer":2},"total_cost_usd":0.125,"usage":{"input_tokens":2,"output_tokens":7,"cache_creation_input_tokens":10,"cache_read_input_tokens":20,"output_tokens_details":{"thinking_tokens":3}},"modelUsage":{"claude-sonnet-5-5":{}}}`

func question() dataset.Question {
	return dataset.Question{Text: `متن $x^2$`, Dataset: "SECRET-EXAM", Category: "SECRET-CATEGORY", Correct: 3,
		Options: []dataset.Option{{Number: 1, Text: "one"}, {Number: 2, Text: "two"}, {Number: 3, Text: "three"}, {Number: 4, Text: "four"}}}
}

func TestRequestIsolationAndImages(t *testing.T) {
	q := question()
	q.Asset = &dataset.Asset{MIME: "image/png", Bytes: []byte{1, 2, 3}}
	q.Options[2].Asset = &dataset.Asset{MIME: "image/jpeg", Bytes: []byte{4, 5, 6}}
	b, err := Build(q, testModel, "max")
	if err != nil {
		t.Fatal(err)
	}
	var req Request
	if err := json.Unmarshal(b, &req); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"SECRET-EXAM", "SECRET-CATEGORY", "correct_option", "max_tokens", "max-tokens", ":exacto", "--resume", "--continue", "--bare"} {
		if strings.Contains(string(b), forbidden) {
			t.Fatalf("request contains %q", forbidden)
		}
	}
	for _, flag := range []string{"--safe-mode", "--restricted", "--no-session-persistence", "--disable-slash-commands", "--include-partial-messages"} {
		if !slices.Contains(req.Args, flag) {
			t.Fatalf("missing %s", flag)
		}
	}
	for flag, want := range map[string]string{"--model": "claude-sonnet-5-5", "--effort": "max", "--tools": "", "--system-prompt": openrouter.Prompt, "--json-schema": Schema, "--input-format": "stream-json"} {
		i := slices.Index(req.Args, flag)
		if i < 0 || req.Args[i+1] != want {
			t.Fatalf("%s: %v", flag, req.Args)
		}
	}
	var input struct {
		Type    string
		Message struct {
			Role    string
			Content []struct {
				Type, Text string
				Source     struct {
					Type string
					MIME string `json:"media_type"`
					Data string
				}
			}
		}
	}
	if err := json.Unmarshal(req.Input, &input); err != nil {
		t.Fatal(err)
	}
	p := input.Message.Content
	if input.Type != "user" || input.Message.Role != "user" || len(p) != 5 || !strings.Contains(p[0].Text, `متن $x^2$`) || p[2].Source.Data != "AQID" || p[2].Source.MIME != "image/png" || p[4].Source.Data != "BAUG" || !strings.Contains(p[3].Text, "3") {
		t.Fatalf("wrong input: %s", req.Input)
	}
	q.Correct = 1
	q.Category = "changed"
	q.Dataset = "changed"
	again, _ := Build(q, testModel, "max")
	if string(again) != string(b) {
		t.Fatal("internal metadata changes request")
	}
	q.Asset.Bytes = []byte{9}
	changed, _ := Build(q, testModel, "max")
	if string(changed) == string(b) {
		t.Fatal("image bytes missing from cache identity")
	}
	for _, model := range []string{"openai/gpt-6.1-sol", "anthropic/claude-sonnet-5.5:exacto", "sonnet"} {
		if _, err := Build(q, model, "medium"); err == nil {
			t.Fatalf("accepted unsupported model %s", model)
		}
	}
	if _, err := Build(q, testModel, "minimal"); err == nil {
		t.Fatal("CLI minimal must map to low through capabilities")
	}
}

func TestStreamNormalizationAndStrictAnswers(t *testing.T) {
	for _, tc := range []struct {
		name, output string
		answer       int
		invalid      bool
	}{
		{"valid", `{"answer":2}`, 2, false},
		{"out of range", `{"answer":5}`, 0, true},
		{"extra field", `{"answer":2,"reason":"yes"}`, 0, true},
		{"duplicate", `{"answer":2,"answer":1}`, 0, true},
		{"missing structured", `null`, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := streamState{model: "claude-sonnet-5-5", version: "2.1.296"}
			// Real CLI lifecycle events use a string event, whereas partial messages
			// use an object. Regresses the first live image-run parser failure.
			for _, line := range []string{`{"type":"system","subtype":"ui_invalidate","event":"ui.render","session_id":"session-test"}`, initEvent, strings.Replace(resultLine, `{"answer":2}`, tc.output, 1)} {
				if err := s.accept([]byte(line)); err != nil {
					t.Fatal(err)
				}
			}
			r := s.normalize()
			if r.Error != "" || r.Status != 200 {
				t.Fatalf("%+v", r)
			}
			g, answer, invalid, err := openrouter.Parse(r.Body)
			if err != nil || answer != tc.answer || (invalid != "") != tc.invalid {
				t.Fatalf("answer=%d invalid=%s err=%v", answer, invalid, err)
			}
			u := g.Usage
			if u.Cost != nil || u.EstimatedCost == nil || *u.EstimatedCost != 0.125 || *u.Prompt != 32 || *u.Completion != 7 || *u.Total != 39 || *u.PromptDetails.Cached != 20 || *u.CompletionDetails.Reasoning != 3 {
				t.Fatalf("wrong usage: %+v", u)
			}
		})
	}
}

func TestStreamRejectsContaminationAndQuota(t *testing.T) {
	for _, tc := range []struct{ name, line, contains string }{
		{"model switch", `{"type":"assistant","message":{"model":"claude-opus-5-5"}}`, "instead"},
		{"partial model switch", `{"type":"stream_event","event":{"message":{"model":"claude-opus-5-5"}}}`, "instead"},
		{"session switch", `{"type":"system","session_id":"another"}`, "session changed"},
		{"bad JSON", `{`, "JSON"},
		{"quota", `{"type":"rate_limit_event","rate_limit_info":{"status":"rejected"}}`, "rate limit"},
		{"overage", `{"type":"rate_limit_event","rate_limit_info":{"status":"allowed","isUsingOverage":true}}`, "extra usage"},
		{"hidden fallback", strings.Replace(resultLine, "claude-sonnet-5-5", "claude-opus-5-5", 1), "unexpected model"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := streamState{model: "claude-sonnet-5-5", version: "2.1.296"}
			if err := s.accept([]byte(initEvent)); err != nil {
				t.Fatal(err)
			}
			if err := s.accept([]byte(tc.line)); err == nil || !strings.Contains(err.Error(), tc.contains) {
				t.Fatalf("got %v", err)
			}
		})
	}
	for _, line := range []string{strings.Replace(initEvent, "StructuredOutput", "Read", 1), strings.Replace(initEvent, "2.1.296", "2.1.297", 1), strings.Replace(initEvent, `"mcp_servers":[]`, `"mcp_servers":[{}]`, 1)} {
		s := streamState{model: "claude-sonnet-5-5", version: "2.1.296"}
		if s.accept([]byte(line)) == nil {
			t.Fatalf("unsafe init accepted: %s", line)
		}
	}
	if sameModel("claude-sonnet-5-5-other", "claude-sonnet-5-5") || !sameModel("claude-sonnet-5-5-20260928", "claude-sonnet-5-5") {
		t.Fatal("model revision matching")
	}
}

// The executable fixture is this Go test binary behind a tiny shell launcher.
// Real subprocess pipes, exit codes, cancellation and environment are tested
// without a network or Claude installation.
func executable(t *testing.T, scenario string) string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "claude")
	script := "#!/bin/sh\nexec '" + strings.ReplaceAll(exe, "'", "'\\''") + "' -test.run=^TestCLIHelperProcess$ -- '" + scenario + "' \"$@\"\n"
	if err := os.WriteFile(path, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCLIHelperProcess(t *testing.T) {
	i := slices.Index(os.Args, "--")
	if i < 0 {
		return
	}
	scenario, args := os.Args[i+1], os.Args[i+2:]
	if scenario == "child" {
		time.Sleep(10 * time.Second)
		os.Exit(0)
	}
	if slices.Contains(args, "--version") {
		if scenario == "old" {
			fmt.Println("2.1.1 (Claude Code)")
		} else {
			fmt.Println("2.1.296 (Claude Code)")
		}
		os.Exit(0)
	}
	if slices.Contains(args, "auth") {
		if scenario == "api" {
			fmt.Println(`{"loggedIn":true,"authMethod":"api_key"}`)
		} else {
			fmt.Println(`{"loggedIn":true,"authMethod":"claude.ai","email":"private@example.test"}`)
		}
		os.Exit(0)
	}
	if os.Getenv("ANTHROPIC_API_KEY") != "" || os.Getenv("OPENROUTER_API_KEY") != "" || os.Getenv("MAX_THINKING_TOKENS") != "" || os.Getenv("CLAUDE_CODE_EFFORT_LEVEL") != "" {
		fmt.Fprintln(os.Stderr, "inherited inference override")
		os.Exit(2)
	}
	input, err := io.ReadAll(os.Stdin)
	if err != nil || !json.Valid(input) {
		fmt.Fprintln(os.Stderr, "invalid stdin")
		os.Exit(2)
	}
	cwd, _ := os.Getwd()
	files, _ := os.ReadDir(cwd)
	if len(files) != 0 || !strings.Contains(filepath.Base(cwd), "konkour-claude-") {
		fmt.Fprintln(os.Stderr, "not isolated")
		os.Exit(2)
	}
	fmt.Println(initEvent)
	switch scenario {
	case "timeout", "deadline":
		time.Sleep(10 * time.Second)
	case "descendant":
		child := exec.Command(os.Args[0], "-test.run=^TestCLIHelperProcess$", "--", "child")
		child.Stdout, child.Stderr = os.Stdout, os.Stderr
		if err := child.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Println(`{"type":"system","subtype":"descendant_started"}`)
		_ = child.Wait()
	case "no-result":
		os.Exit(0)
	case "nonzero":
		fmt.Fprintln(os.Stderr, "fixture credential-should-be-hidden")
		os.Exit(7)
	case "malformed":
		fmt.Println("not-json")
		os.Exit(0)
	case "large":
		fmt.Printf("{\"type\":\"system\",\"message_text\":%q}\n", strings.Repeat("a", 100000))
	case "large-stderr":
		fmt.Fprint(os.Stderr, strings.Repeat("e", 100000))
	case "auth-error":
		fmt.Println(`{"type":"result","subtype":"success","is_error":true,"result":"OAuth access token has expired"}`)
		os.Exit(1)
	case "effort-error":
		fmt.Println(`{"type":"result","subtype":"success","is_error":true,"result":"unsupported reasoning effort max"}`)
		os.Exit(1)
	}
	fmt.Println(resultLine)
	if scenario == "success-nonzero" {
		os.Exit(7)
	}
	os.Exit(0)
}

func TestDiscovery(t *testing.T) {
	for _, scenario := range []string{"success", "old", "api"} {
		t.Run(scenario, func(t *testing.T) {
			c := Client{Binary: executable(t, scenario)}
			cat, err := c.Discover(context.Background(), []string{testModel})
			if scenario != "success" {
				if err == nil {
					t.Fatal("invalid preflight accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if c.Endpoint() != "claude-cli://v1/2.1.296" || strings.Contains(string(cat.Raw), "private@example.test") {
				t.Fatalf("metadata %s", cat.Raw)
			}
			m := cat.Models[testModel]
			if !m.Images() {
				t.Fatal("vision not advertised")
			}
			for tier, want := range map[string]string{"minimal": "low", "medium": "medium", "max": "max"} {
				got, _ := m.Candidates(tier)
				if len(got) == 0 || got[0] != want {
					t.Fatalf("%s=%v", tier, got)
				}
			}
		})
	}
	c := Client{Binary: "nonexistent"}
	if _, err := c.Discover(context.Background(), []string{"openai/gpt-6.1-sol"}); err == nil {
		t.Fatal("non-Claude accepted")
	}
}

func TestProcessExecution(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "credential-should-be-hidden")
	t.Setenv("OPENROUTER_API_KEY", "unused-key")
	t.Setenv("MAX_THINKING_TOKENS", "1")
	t.Setenv("CLAUDE_CODE_EFFORT_LEVEL", "low")
	body, _ := Build(question(), testModel, "medium")
	for _, scenario := range []string{"success", "large", "large-stderr", "timeout", "deadline", "descendant", "no-result", "nonzero", "success-nonzero", "malformed", "auth-error", "persist-failure", "effort-error"} {
		t.Run(scenario, func(t *testing.T) {
			c := Client{Binary: executable(t, scenario), Version: "2.1.296"}
			limit := 3 * time.Second
			if scenario == "deadline" {
				limit = 100 * time.Millisecond
			}
			ctx, cancel := context.WithTimeout(context.Background(), limit)
			defer cancel()
			var events []openrouter.Progress
			started := time.Now()
			out := c.Complete(ctx, body, func(p openrouter.Progress) error {
				events = append(events, p)
				if scenario == "timeout" {
					cancel()
				}
				if scenario == "descendant" && strings.Contains(p.Raw, "descendant_started") {
					cancel()
				}
				if scenario == "persist-failure" {
					return errors.New("disk full")
				}
				return nil
			})
			if scenario == "success" || scenario == "large" || scenario == "large-stderr" {
				if out.Error != "" || out.Status != 200 || len(events) < 2 {
					t.Fatalf("%+v events=%d", out, len(events))
				}
				if scenario == "large-stderr" {
					raw := events[len(events)-1].Raw
					if len(raw) > 66000 || !strings.Contains(raw, "stderr truncated") {
						t.Fatalf("stderr was not bounded: %d bytes", len(raw))
					}
				}
				return
			}
			if scenario == "effort-error" {
				if out.Error != "" || !openrouter.EffortRejected(out) {
					t.Fatalf("explicit effort rejection lost: %+v", out)
				}
				return
			}
			if out.Error == "" || strings.Contains(out.Error, "credential-should-be-hidden") {
				t.Fatalf("unsafe/missing error: %+v", out)
			}
			if (scenario == "timeout" || scenario == "descendant") && out.Error != context.Canceled.Error() {
				t.Fatalf("cancel: %+v", out)
			}
			if scenario == "deadline" && out.Error != context.DeadlineExceeded.Error() {
				t.Fatalf("deadline: %+v", out)
			}
			if scenario == "descendant" && time.Since(started) > 2*time.Second {
				t.Fatal("child survived cancellation and kept stdout open")
			}
			if scenario != "timeout" && scenario != "deadline" && scenario != "descendant" && !out.Fatal {
				t.Fatalf("not fatal: %+v", out)
			}
			if scenario == "nonzero" {
				last := events[len(events)-1].Raw
				if !strings.Contains(last, "[REDACTED]") || strings.Contains(last, "credential-should-be-hidden") {
					t.Fatalf("stderr leaked: %s", last)
				}
			}
		})
	}
}
