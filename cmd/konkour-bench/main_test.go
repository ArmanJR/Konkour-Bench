package main

import (
	"context"
	"encoding/xml"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestEnvFileIsLiteralAndNeverExecuted(t *testing.T) {
	for _, value := range []string{"test-key", "'test-key'", "\"test-key\""} {
		path := filepath.Join(t.TempDir(), ".env")
		os.WriteFile(path, []byte("# comment\nexport OPENROUTER_API_KEY="+value+"\n"), 0600)
		key, err := keyFromFile(path)
		if err != nil || key != "test-key" {
			t.Fatal(key, err)
		}
	}
	for _, text := range []string{"OPENROUTER_API_KEY=$(touch nope)", "OPENROUTER_API_KEY=one\nOPENROUTER_API_KEY=two", "OPENROUTER_API_KEY=", "OTHER=value"} {
		path := filepath.Join(t.TempDir(), ".env")
		os.WriteFile(path, []byte(text), 0600)
		if _, err := keyFromFile(path); err == nil {
			t.Fatal("accepted invalid env assignment")
		}
	}
}

func TestInvalidSafetyFlagsFailBeforeNetwork(t *testing.T) {
	for _, args := range [][]string{{"run", "--budget-usd", "NaN"}, {"run", "--budget-usd", "-1"}, {"run", "--timeout", "-1s"}, {"run", "--warn-after", "-1s"}} {
		err := run(context.Background(), args)
		if err == nil || !strings.Contains(err.Error(), "nonnegative") {
			t.Fatalf("%v: %v", args, err)
		}
	}
}

func TestBackendFlagsAndClaudeSelection(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"run", "--backend", "invalid"}, "unknown backend"},
		{[]string{"run", "--backend", "claude", "--budget-usd", "1"}, "subscription"},
		{[]string{"reconcile", "--backend", "claude"}, "no OpenRouter"},
		{[]string{"run", "--backend", "claude", "--config", "does-not-exist.json"}, "does-not-exist.json"},
	} {
		if err := run(context.Background(), tc.args); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%v: %v", tc.args, err)
		}
	}
	got := claudeModels([]string{"openai/gpt-6.1-sol", "anthropic/claude-opus-5.5", "anthropic/claude-sonnet-5.5", "anthropic/claude-sonnet-5.5:exacto"})
	if !slices.Equal(got, []string{"anthropic/claude-opus-5.5", "anthropic/claude-sonnet-5.5"}) {
		t.Fatalf("selection: %v", got)
	}
}

func TestClaudeSupervisorNeedsNoAPIKey(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("launchd requires macOS")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "worker.plist")
	if err := run(context.Background(), []string{"supervisor", "--backend", "claude", "--claude-bin", exe, "--service-file", path}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var job struct {
		Args []string `xml:"dict>array>string"`
	}
	if err := xml.Unmarshal(b, &job); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(job.Args, "--env-file") {
		t.Fatal("Claude worker requires API credentials")
	}
	for flag, want := range map[string]string{"--backend": "claude", "--claude-bin": exe, "--concurrency": "1", "--budget-usd": "0"} {
		i := slices.Index(job.Args, flag)
		if i < 0 || job.Args[i+1] != want {
			t.Fatalf("%s: %v", flag, job.Args)
		}
	}
}

func TestSupervisorRequestTimeout(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("supervisor preparation requires macOS")
	}
	for _, tc := range []struct {
		name, override, want string
	}{
		{"default", "", "20m0s"},
		{"override", "5m", "5m0s"},
		{"disabled", "0", "0s"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "job.plist")
			args := []string{"supervisor", "--env-file", filepath.Join(dir, ".env"), "--service-file", path}
			if tc.override != "" {
				args = append(args, "--timeout", tc.override)
			}
			if err := run(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var job struct {
				Arguments []string `xml:"dict>array>string"`
			}
			if err := xml.Unmarshal(raw, &job); err != nil {
				t.Fatal(err)
			}
			i := slices.Index(job.Arguments, "--timeout")
			if i < 0 || i+1 >= len(job.Arguments) || job.Arguments[i+1] != tc.want {
				t.Fatalf("expected request timeout %s in worker arguments, got %v", tc.want, job.Arguments)
			}
		})
	}
}
