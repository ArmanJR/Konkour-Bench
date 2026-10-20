package claude

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"konkour-bench/internal/openrouter"
)

type Client struct{ Binary, Version string }

func (c *Client) Endpoint() string { return "claude-cli://v1/" + c.Version }

// Keep only runtime/authentication essentials. In particular, inherited API
// keys, provider overrides, thinking budgets, and agent customizations cannot
// silently change subscription billing or the benchmark protocol.
func environment() []string {
	allowed := []string{"HOME", "PATH", "USER", "LOGNAME", "TMPDIR", "TMP", "TEMP", "LANG", "LC_ALL", "LC_CTYPE", "TERM", "SYSTEMROOT", "WINDIR", "XDG_CONFIG_HOME", "XDG_CACHE_HOME", "HTTPS_PROXY", "HTTP_PROXY", "NO_PROXY", "https_proxy", "http_proxy", "no_proxy", "SSL_CERT_FILE", "SSL_CERT_DIR", "NODE_EXTRA_CA_CERTS", "CLAUDE_CONFIG_DIR", "CLAUDE_CODE_OAUTH_TOKEN"}
	var env []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if slices.Contains(allowed, key) {
			env = append(env, entry)
		}
	}
	return append(env, "DISABLE_AUTOUPDATER=1", "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1", "CLAUDE_CODE_DISABLE_NONSTREAMING_FALLBACK=1")
}

func command(ctx context.Context, binary, dir string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir, cmd.Env = dir, environment()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = 2 * time.Second
	return cmd
}

func redact(text string) string {
	for _, entry := range os.Environ() {
		key, value, _ := strings.Cut(entry, "=")
		if len(value) >= 8 && (strings.Contains(key, "TOKEN") || strings.Contains(key, "KEY") || strings.Contains(key, "SECRET")) {
			text = strings.ReplaceAll(text, value, "[REDACTED]")
		}
	}
	return text
}

func (c *Client) Discover(ctx context.Context, selected []string) (openrouter.Catalog, error) {
	cat := openrouter.Catalog{Models: map[string]openrouter.Model{}, FetchedAt: time.Now().UTC()}
	for _, id := range selected {
		if !Supports(id) {
			return cat, fmt.Errorf("Claude CLI does not support model %q", id)
		}
	}
	path, err := exec.LookPath(c.Binary)
	if err != nil {
		return cat, fmt.Errorf("Claude CLI executable: %w", err)
	}
	c.Binary, err = filepath.Abs(path)
	if err != nil {
		return cat, err
	}
	dir, err := os.MkdirTemp("", "konkour-claude-preflight-")
	if err != nil {
		return cat, err
	}
	defer os.RemoveAll(dir)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	b, err := command(ctx, c.Binary, dir, "--version").Output()
	if err != nil {
		return cat, fmt.Errorf("Claude CLI version check: %w", err)
	}
	var major, minor, patch int
	if _, err := fmt.Sscanf(string(b), "%d.%d.%d", &major, &minor, &patch); err != nil || major < 2 || (major == 2 && (minor < 1 || (minor == 1 && patch < 296))) {
		return cat, fmt.Errorf("Claude CLI 2.1.296 or newer is required for isolated subscription sessions")
	}
	c.Version = fmt.Sprintf("%d.%d.%d", major, minor, patch)
	b, err = command(ctx, c.Binary, dir, "--safe-mode", "--restricted", "auth", "status", "--json").Output()
	if err != nil {
		return cat, fmt.Errorf("Claude subscription authentication unavailable; run claude auth login: %w", err)
	}
	var auth struct {
		LoggedIn bool   `json:"loggedIn"`
		Method   string `json:"authMethod"`
	}
	if err := json.Unmarshal(b, &auth); err != nil {
		return cat, fmt.Errorf("Claude auth status returned invalid JSON")
	}
	if !auth.LoggedIn || (auth.Method != "claude.ai" && auth.Method != "oauth_token") {
		return cat, fmt.Errorf("Claude subscription login required; run claude auth login (API-key billing is not accepted)")
	}
	// Never persist the auth response: it may contain account identifiers.
	cat.Raw, _ = json.Marshal(map[string]string{"backend": "claude", "version": c.Version, "capability_source": "explicit CLI capabilities; account model access checked at inference"})
	for _, id := range selected {
		b, _ := json.Marshal(map[string]any{"id": id, "architecture": map[string]any{"input_modalities": []string{"text", "image"}}, "supported_parameters": []string{"structured_outputs", "reasoning"}, "reasoning": map[string]any{"supported_efforts": []string{"low", "medium", "max", "xhigh", "high"}}})
		var m openrouter.Model
		_ = json.Unmarshal(b, &m)
		cat.Models[id] = m
	}
	return cat, nil
}

// cappedBuffer drains stderr without allowing an unbounded in-memory log.
type cappedBuffer struct {
	data      bytes.Buffer
	truncated bool
}

func (b *cappedBuffer) Len() int { return b.data.Len() }
func (b *cappedBuffer) String() string {
	if b.truncated {
		return b.data.String() + "\n[stderr truncated at 64 KiB]"
	}
	return b.data.String()
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	left := 64*1024 - b.Len()
	if left > 0 {
		_, _ = b.data.Write(p[:min(left, n)])
	}
	if n > left {
		b.truncated = true
	}
	return n, nil
}

func (c *Client) Complete(ctx context.Context, body []byte, save func(openrouter.Progress) error) (out openrouter.HTTPResult) {
	var req Request
	if err := json.Unmarshal(body, &req); err != nil || !Supports(req.Model) {
		return openrouter.HTTPResult{Error: "invalid Claude CLI request", Fatal: true}
	}
	dir, err := os.MkdirTemp("", "konkour-claude-")
	if err != nil {
		return openrouter.HTTPResult{Error: err.Error(), Fatal: true}
	}
	defer os.RemoveAll(dir)
	callCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	cmd := command(callCtx, c.Binary, dir, req.Args...)
	cmd.Stdin = bytes.NewReader(append(append([]byte(nil), req.Input...), '\n'))
	var stderr cappedBuffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return openrouter.HTTPResult{Error: err.Error(), Fatal: true}
	}
	if err := cmd.Start(); err != nil {
		return openrouter.HTTPResult{Error: err.Error(), Fatal: true}
	}
	state := streamState{model: models[req.Model], version: c.Version}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	var streamErr error
	for scanner.Scan() {
		line := redact(scanner.Text())
		if strings.TrimSpace(line) == "" {
			continue
		}
		parseErr := state.accept([]byte(line))
		p := openrouter.Progress{Raw: line, GenerationID: state.session, RequestID: state.requestID, Provider: "Claude CLI (subscription)"}
		if err := save(p); err != nil {
			streamErr = fmt.Errorf("persist Claude event: %w", err)
			cancel()
			break
		}
		if parseErr != nil {
			streamErr = parseErr
			cancel()
			break
		}
	}
	if err := scanner.Err(); err != nil && streamErr == nil {
		streamErr = fmt.Errorf("read Claude stream: %w", err)
		cancel()
	}
	// Wait only after draining stdout. The context kills the whole process group
	// so timeout/cancellation cannot leave a child generating in the background.
	waitErr := cmd.Wait()
	if stderr.Len() > 0 {
		raw, _ := json.Marshal(map[string]string{"type": "stderr", "text": redact(stderr.String())})
		if err := save(openrouter.Progress{Raw: string(raw), GenerationID: state.session}); err != nil && streamErr == nil {
			streamErr = fmt.Errorf("persist Claude stderr: %w", err)
		}
	}
	out.GenerationID, out.RequestID = state.session, state.requestID
	if ctx.Err() != nil {
		out.Error = ctx.Err().Error()
		return
	}
	if streamErr != nil {
		out.Error = streamErr.Error()
		out.Fatal = true
		return
	}
	if state.result == nil {
		out.Error = fmt.Sprintf("Claude CLI ended without a result (exit: %v): %s", waitErr, redact(stderr.String()))
		out.Fatal = true
		return
	}
	out = state.normalize()
	if waitErr != nil && out.Error == "" && out.Status == 200 {
		out.Error = fmt.Sprintf("Claude CLI exit: %v: %s", waitErr, redact(stderr.String()))
		out.Fatal = true
	}
	return
}
