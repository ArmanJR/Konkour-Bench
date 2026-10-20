package claude

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"konkour-bench/internal/openrouter"
)

type cliUsage struct {
	Input   *int64 `json:"input_tokens"`
	Output  *int64 `json:"output_tokens"`
	Created *int64 `json:"cache_creation_input_tokens"`
	Cached  *int64 `json:"cache_read_input_tokens"`
	Details struct {
		Thinking *int64 `json:"thinking_tokens"`
	} `json:"output_tokens_details"`
}
type resultEvent struct {
	Subtype    string                     `json:"subtype"`
	IsError    bool                       `json:"is_error"`
	Result     string                     `json:"result"`
	Structured json.RawMessage            `json:"structured_output"`
	Cost       *float64                   `json:"total_cost_usd"`
	Usage      cliUsage                   `json:"usage"`
	Models     map[string]json.RawMessage `json:"modelUsage"`
	Errors     []string                   `json:"errors"`
	Stop       string                     `json:"stop_reason"`
	Terminal   string                     `json:"terminal_reason"`
}
type streamState struct {
	model, version, session, requestID, actualModel string
	initialized, sawModel                           bool
	result                                          *resultEvent
}

var datedModel = regexp.MustCompile(`^\d{8}$`)

func sameModel(actual, expected string) bool {
	return actual == expected || (strings.HasPrefix(actual, expected+"-") && datedModel.MatchString(strings.TrimPrefix(actual, expected+"-")))
}

func (s *streamState) accept(raw []byte) error {
	var e struct {
		Type    string            `json:"type"`
		Subtype string            `json:"subtype"`
		Session string            `json:"session_id"`
		Request string            `json:"request_id"`
		Model   string            `json:"model"`
		Version string            `json:"claude_code_version"`
		Tools   []string          `json:"tools"`
		MCP     []json.RawMessage `json:"mcp_servers"`
		Message struct {
			Model string `json:"model"`
		} `json:"message"`
		Event json.RawMessage `json:"event"`
		Rate  struct {
			Status  string `json:"status"`
			Overage bool   `json:"isUsingOverage"`
		} `json:"rate_limit_info"`
	}
	if err := json.Unmarshal(raw, &e); err != nil {
		return fmt.Errorf("invalid Claude stream JSON: %w", err)
	}
	if e.Type == "" {
		return fmt.Errorf("Claude stream event has no type")
	}
	if e.Session != "" {
		if s.session != "" && e.Session != s.session {
			return fmt.Errorf("Claude session changed within one question")
		}
		s.session = e.Session
	}
	if e.Request != "" {
		s.requestID = e.Request
	}
	if e.Type == "system" && e.Subtype == "init" {
		if s.initialized {
			return fmt.Errorf("duplicate Claude initialization")
		}
		s.initialized = true
		if e.Version != s.version {
			return fmt.Errorf("Claude CLI version changed after preflight; restart to refresh the cache namespace")
		}
		if !sameModel(e.Model, s.model) {
			return fmt.Errorf("Claude selected model %q instead of %q", e.Model, s.model)
		}
		for _, tool := range e.Tools {
			if tool != "StructuredOutput" {
				return fmt.Errorf("unexpected Claude tool enabled: %s", tool)
			}
		}
		if len(e.MCP) > 0 {
			return fmt.Errorf("unexpected Claude MCP servers enabled")
		}
	}
	var stream struct {
		Message struct {
			Model string `json:"model"`
		} `json:"message"`
	}
	if e.Type == "stream_event" {
		if err := json.Unmarshal(e.Event, &stream); err != nil {
			return fmt.Errorf("invalid Claude partial event: %w", err)
		}
	}
	for _, model := range []string{e.Message.Model, stream.Message.Model} {
		if model == "" || model == "<synthetic>" {
			continue
		}
		if !sameModel(model, s.model) {
			return fmt.Errorf("Claude generated with %q instead of %q; result excluded", model, s.model)
		}
		s.sawModel = true
		s.actualModel = model
	}
	if e.Type == "rate_limit_event" {
		if e.Rate.Overage {
			return fmt.Errorf("Claude reported paid extra usage; disable extra usage in your Claude account before continuing")
		}
		if e.Rate.Status == "rejected" {
			return fmt.Errorf("Claude subscription rate limit reached; wait for the quota reset before resuming")
		}
	}
	if e.Type == "result" {
		if s.result != nil {
			return fmt.Errorf("multiple Claude results for one question")
		}
		var r resultEvent
		if err := json.Unmarshal(raw, &r); err != nil {
			return fmt.Errorf("invalid Claude result: %w", err)
		}
		s.result = &r
		for model := range r.Models {
			if !sameModel(model, s.model) {
				return fmt.Errorf("Claude used unexpected model %q; result excluded", model)
			}
			s.sawModel = true
			s.actualModel = model
		}
	}
	return nil
}

func (s *streamState) normalize() openrouter.HTTPResult {
	r := s.result
	out := openrouter.HTTPResult{GenerationID: s.session, RequestID: s.requestID}
	if r.IsError || r.Subtype != "success" {
		out.Error = fmt.Sprintf("Claude CLI %s: %s %s", r.Subtype, r.Result, strings.Join(r.Errors, "; "))
		out.Fatal = true
		// Only an explicit effort validation failure permits a same-band retry.
		raw, _ := json.Marshal(map[string]string{"error": out.Error})
		rejected := openrouter.HTTPResult{Status: 400, Body: raw}
		if !s.sawModel && openrouter.EffortRejected(rejected) {
			return rejected
		}
		return out
	}
	if !s.initialized || !s.sawModel || s.session == "" {
		out.Error = "Claude result lacks initialization, session, or actual model evidence"
		out.Fatal = true
		return out
	}
	u := openrouter.Usage{EstimatedCost: r.Cost, Completion: r.Usage.Output}
	if r.Usage.Input != nil && r.Usage.Created != nil && r.Usage.Cached != nil {
		n := *r.Usage.Input + *r.Usage.Created + *r.Usage.Cached
		u.Prompt = &n
		if u.Completion != nil {
			total := n + *u.Completion
			u.Total = &total
		}
	}
	u.PromptDetails.Cached = r.Usage.Cached
	u.CompletionDetails.Reasoning = r.Usage.Details.Thinking
	finish := "stop"
	if r.Stop != "end_turn" && r.Stop != "tool_use" && r.Stop != "stop_sequence" {
		finish = r.Stop
	}
	if len(r.Structured) == 0 || string(r.Structured) == "null" {
		finish = "missing_structured_output"
	}
	// Only structured_output is scored. A plausible JSON string in result is
	// never repaired or accepted as a substitute for the schema contract.
	b, _ := json.Marshal(map[string]any{"id": s.session, "provider": "Claude CLI (subscription)", "model": s.actualModel, "usage": u, "choices": []any{map[string]any{"finish_reason": finish, "message": map[string]any{"content": string(r.Structured)}}}})
	out.Status, out.Body = 200, b
	return out
}
