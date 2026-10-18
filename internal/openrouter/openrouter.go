// Package openrouter implements the small subset of the API used by the benchmark.
package openrouter

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"konkour-bench/internal/dataset"
)

const Prompt = "گزینهٔ درست را انتخاب کن. فقط پاسخ را مطابق قالب JSON برگردان."

var Tiers = []string{"minimal", "medium", "max"}

type Model struct {
	ID           string `json:"id"`
	Architecture struct {
		Input []string `json:"input_modalities"`
	} `json:"architecture"`
	Parameters []string `json:"supported_parameters"`
	Reasoning  *struct {
		Efforts json.RawMessage `json:"supported_efforts"`
	} `json:"reasoning"`
}

func (m Model) Images() bool { return slices.Contains(m.Architecture.Input, "image") }
func (m Model) Candidates(tier string) ([]string, string) {
	bands := map[string][]string{"minimal": {"minimal", "low"}, "medium": {"medium"}, "max": {"max", "xhigh", "high"}}
	band, ok := bands[tier]
	if !ok {
		return nil, "unknown effort tier"
	}
	if !slices.Contains(m.Parameters, "structured_outputs") {
		return nil, "model does not advertise structured outputs"
	}
	if m.Reasoning == nil || len(m.Reasoning.Efforts) == 0 {
		return nil, "model does not advertise selectable reasoning efforts"
	}
	if string(m.Reasoning.Efforts) == "null" {
		return band, "catalog accepts all gateway efforts"
	}
	var supported []string
	if err := json.Unmarshal(m.Reasoning.Efforts, &supported); err != nil {
		return nil, "invalid reasoning capability metadata"
	}
	var out []string
	for _, effort := range band {
		if slices.Contains(supported, effort) {
			out = append(out, effort)
		}
	}
	if len(out) == 0 {
		return nil, "no supported effort within requested band"
	}
	return out, "catalog supported_efforts"
}

type Catalog struct {
	Models    map[string]Model           `json:"models"`
	Raw       json.RawMessage            `json:"raw"`
	Endpoints map[string]json.RawMessage `json:"endpoints"`
	FetchedAt time.Time                  `json:"fetched_at"`
}

type Client struct {
	BaseURL, Key string
	HTTP         *http.Client
}

func New(baseURL, key string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), Key: key, HTTP: &http.Client{Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment, DialContext: (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout: 30 * time.Second, IdleConnTimeout: 90 * time.Second,
	}, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}}
}

type HTTPResult struct {
	Status     int             `json:"status"`
	Body       json.RawMessage `json:"body"`
	RetryAfter time.Duration   `json:"-"`
	Error      string          `json:"error,omitempty"`
}

func (c *Client) Do(ctx context.Context, method, path string, body []byte) HTTPResult {
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return HTTPResult{Error: err.Error()}
	}
	if c.Key != "" {
		req.Header.Set("Authorization", "Bearer "+c.Key)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Title", "Konkour Bench")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return HTTPResult{Error: c.redact(err.Error())}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	r := HTTPResult{Status: resp.StatusCode, Body: []byte(c.redact(string(b)))}
	if err != nil {
		r.Error = c.redact(err.Error())
	}
	if seconds, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && seconds > 0 {
		r.RetryAfter = time.Duration(seconds) * time.Second
	} else if until, err := http.ParseTime(resp.Header.Get("Retry-After")); err == nil {
		r.RetryAfter = time.Until(until)
	}
	return r
}

func (c *Client) redact(s string) string {
	if c.Key != "" {
		return strings.ReplaceAll(s, c.Key, "[REDACTED]")
	}
	return s
}

func (c *Client) Discover(ctx context.Context, models []string) (Catalog, error) {
	cat := Catalog{Models: map[string]Model{}, Endpoints: map[string]json.RawMessage{}, FetchedAt: time.Now().UTC()}
	r := c.Do(ctx, http.MethodGet, "/models", nil)
	if r.Error != "" || r.Status != 200 {
		return cat, fmt.Errorf("model discovery: HTTP %d %s %s", r.Status, r.Error, string(r.Body))
	}
	var envelope struct {
		Data []Model `json:"data"`
	}
	if err := json.Unmarshal(r.Body, &envelope); err != nil {
		return cat, fmt.Errorf("model discovery: %w", err)
	}
	cat.Raw = r.Body
	for _, m := range envelope.Data {
		cat.Models[m.ID] = m
	}
	for _, id := range models {
		base := strings.TrimSuffix(id, ":exacto")
		if _, ok := cat.Models[base]; !ok {
			continue
		}
		r := c.Do(ctx, http.MethodGet, "/models/"+base+"/endpoints", nil)
		if r.Error != "" || r.Status != 200 {
			return cat, fmt.Errorf("endpoint discovery for %s: HTTP %d %s %s", base, r.Status, r.Error, string(r.Body))
		}
		if !json.Valid(r.Body) {
			return cat, fmt.Errorf("endpoint discovery for %s returned invalid JSON", base)
		}
		cat.Endpoints[base] = r.Body
	}
	return cat, nil
}

type Content struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	Image *struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
}
type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}
type Request struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	Reasoning struct {
		Enabled bool   `json:"enabled"`
		Effort  string `json:"effort"`
	} `json:"reasoning"`
	Provider struct {
		Require bool `json:"require_parameters"`
	} `json:"provider"`
	Format any  `json:"response_format"`
	Stream bool `json:"stream"`
}

func Build(q dataset.Question, model, effort string) ([]byte, error) {
	r := Request{Model: strings.TrimSuffix(model, ":exacto") + ":exacto"}
	r.Reasoning.Enabled, r.Reasoning.Effort, r.Provider.Require = true, effort, true
	r.Format = map[string]any{"type": "json_schema", "json_schema": map[string]any{"name": "answer", "strict": true, "schema": map[string]any{
		"type": "object", "properties": map[string]any{"answer": map[string]any{"type": "integer", "enum": []int{1, 2, 3, 4}}}, "required": []string{"answer"}, "additionalProperties": false,
	}}}
	var text strings.Builder
	text.WriteString(q.Text)
	for _, o := range q.Options {
		fmt.Fprintf(&text, "\n%d. %s", o.Number, o.Text)
	}
	parts := []Content{{Type: "text", Text: text.String()}}
	addImage := func(label string, a *dataset.Asset) {
		if a == nil {
			return
		}
		parts = append(parts, Content{Type: "text", Text: label}, Content{Type: "image_url", Image: &struct {
			URL string `json:"url"`
		}{"data:" + a.MIME + ";base64," + base64.StdEncoding.EncodeToString(a.Bytes)}})
	}
	addImage("تصویر سؤال", q.Asset)
	for _, o := range q.Options {
		addImage(fmt.Sprintf("تصویر گزینهٔ %d", o.Number), o.Asset)
	}
	r.Messages = []Message{{Role: "system", Content: Prompt}, {Role: "user", Content: parts}}
	return json.Marshal(r)
}

type Usage struct {
	Prompt            *int64   `json:"prompt_tokens"`
	Completion        *int64   `json:"completion_tokens"`
	Total             *int64   `json:"total_tokens"`
	Cost              *float64 `json:"cost"`
	CompletionDetails struct {
		Reasoning *int64 `json:"reasoning_tokens"`
	} `json:"completion_tokens_details"`
	PromptDetails struct {
		Cached *int64 `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
}

type Generation struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Usage    Usage  `json:"usage"`
	Choices  []struct {
		Finish  string `json:"finish_reason"`
		Message struct {
			Content json.RawMessage `json:"content"`
			Refusal json.RawMessage `json:"refusal"`
		} `json:"message"`
	} `json:"choices"`
	Error json.RawMessage `json:"error"`
}

func Parse(body []byte) (Generation, int, string, error) {
	var g Generation
	if err := json.Unmarshal(body, &g); err != nil {
		return g, 0, "", fmt.Errorf("invalid API response: %w", err)
	}
	if len(g.Error) > 0 && string(g.Error) != "null" {
		return g, 0, "", fmt.Errorf("API error: %s", g.Error)
	}
	if len(g.Choices) != 1 {
		return g, 0, "", fmt.Errorf("API response must contain one completion choice")
	}
	c := g.Choices[0]
	if c.Finish != "stop" {
		return g, 0, "non-stop finish reason: " + c.Finish, nil
	}
	if len(c.Message.Refusal) > 0 && string(c.Message.Refusal) != "null" && string(c.Message.Refusal) != `""` {
		return g, 0, "model refused", nil
	}
	var content string
	if err := json.Unmarshal(c.Message.Content, &content); err != nil {
		return g, 0, "answer content is not a string", nil
	}
	// Token parsing also rejects duplicate fields, extra fields, and trailing JSON.
	d := json.NewDecoder(strings.NewReader(content))
	tok, err := d.Token()
	if err != nil || tok != json.Delim('{') {
		return g, 0, "answer is not a JSON object", nil
	}
	if !d.More() {
		return g, 0, "answer field missing", nil
	}
	key, err := d.Token()
	if err != nil || key != "answer" {
		return g, 0, "unexpected answer field", nil
	}
	var answer int
	if err := d.Decode(&answer); err != nil || answer < 1 || answer > 4 {
		return g, 0, "answer must be an integer from 1 to 4", nil
	}
	if d.More() {
		return g, 0, "additional or duplicate answer field", nil
	}
	if tok, err = d.Token(); err != nil || tok != json.Delim('}') {
		return g, 0, "invalid answer object", nil
	}
	if _, err = d.Token(); err != io.EOF {
		return g, 0, "trailing answer content", nil
	}
	return g, answer, "", nil
}

func EffortRejected(r HTTPResult) bool {
	if r.Status != 400 && r.Status != 422 {
		return false
	}
	s := strings.ToLower(string(r.Body))
	return strings.Contains(s, "effort") && (strings.Contains(s, "unsupported") || strings.Contains(s, "not support") || strings.Contains(s, "invalid") || strings.Contains(s, "allowed") || strings.Contains(s, "must be") || strings.Contains(s, "one of"))
}

func (c *Client) Generation(ctx context.Context, id string) HTTPResult {
	return c.Do(ctx, http.MethodGet, "/generation?id="+url.QueryEscape(id), nil)
}
