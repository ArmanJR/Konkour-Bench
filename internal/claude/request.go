// Package claude runs isolated Claude Code print sessions using subscription login.
package claude

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
)

const Schema = `{"type":"object","properties":{"answer":{"type":"integer","enum":[1,2,3,4]}},"required":["answer"],"additionalProperties":false}`

// Explicit capabilities avoid silently benchmarking a CLI alias or a model
// whose CLI clamps unsupported effort levels. Add models after verifying them.
var models = map[string]string{
	"anthropic/claude-opus-5.5":   "claude-opus-5-5",
	"anthropic/claude-sonnet-5.5": "claude-sonnet-5-5",
}

func Supports(model string) bool { _, ok := models[model]; return ok }

type Request struct {
	Model string          `json:"model"`
	Args  []string        `json:"args"`
	Input json.RawMessage `json:"stdin"`
}

func Build(q dataset.Question, model, effort string) ([]byte, error) {
	cliModel, ok := models[model]
	if !ok {
		return nil, fmt.Errorf("Claude CLI model %q is not supported; use an explicit supported Claude model without :exacto", model)
	}
	if !slices.Contains([]string{"low", "medium", "max", "xhigh", "high"}, effort) {
		return nil, fmt.Errorf("unsupported Claude CLI effort %q", effort)
	}
	// Reuse the exact question text, option labels, and image ordering. Gateway
	// routing, exam metadata, and answer keys are not part of the CLI input.
	b, err := openrouter.Build(q, model, effort)
	if err != nil {
		return nil, err
	}
	var source struct {
		Messages []struct{ Content json.RawMessage }
	}
	if err := json.Unmarshal(b, &source); err != nil {
		return nil, err
	}
	var parts []openrouter.Content
	if err := json.Unmarshal(source.Messages[1].Content, &parts); err != nil {
		return nil, err
	}
	content := []any{}
	for _, p := range parts {
		if p.Type == "text" {
			content = append(content, map[string]any{"type": "text", "text": p.Text})
			continue
		}
		mime, data, ok := strings.Cut(strings.TrimPrefix(p.Image.URL, "data:"), ";base64,")
		if !ok {
			return nil, fmt.Errorf("invalid inline image")
		}
		content = append(content, map[string]any{"type": "image", "source": map[string]string{"type": "base64", "media_type": mime, "data": data}})
	}
	input, err := json.Marshal(map[string]any{"type": "user", "message": map[string]any{"role": "user", "content": content}, "parent_tool_use_id": nil})
	if err != nil {
		return nil, err
	}
	args := []string{"--safe-mode", "--restricted", "-p", "--model", cliModel,
		"--effort", effort, "--tools", "", "--disable-slash-commands",
		"--strict-mcp-config", "--mcp-config", `{"mcpServers":{}}`,
		"--setting-sources", "", "--settings", `{"alwaysThinkingEnabled":true,"disableAllHooks":true,"autoMemoryEnabled":false,"ultracode":false}`,
		"--permission-mode", "dontAsk", "--no-session-persistence", "--no-chrome",
		"--system-prompt", openrouter.Prompt, "--json-schema", Schema,
		"--input-format", "stream-json", "--output-format", "stream-json", "--verbose", "--include-partial-messages"}
	return json.Marshal(Request{Model: model, Args: args, Input: input})
}
