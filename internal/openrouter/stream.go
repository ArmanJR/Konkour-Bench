package openrouter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Progress is persisted before reading the next event, including response headers.
// Raw contains SSE data (including reasoning), never authorization headers.
type Progress struct {
	GenerationID string `json:"generation_id,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
	Provider     string `json:"provider,omitempty"`
	HTTPStatus   int    `json:"http_status"`
	Raw          string `json:"raw,omitempty"`
}

func (c *Client) Complete(ctx context.Context, body []byte, save func(Progress) error) (result HTTPResult) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		result.Error = err.Error()
		return
	}
	if c.Key != "" {
		req.Header.Set("Authorization", "Bearer "+c.Key)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Title", "Konkour Bench")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		result.Error = c.redact(err.Error())
		return
	}
	defer resp.Body.Close()
	result.Status = resp.StatusCode
	result.GenerationID = c.redact(resp.Header.Get("X-Generation-Id"))
	result.RequestID = c.redact(resp.Header.Get("X-Request-Id"))
	if n, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && n > 0 {
		result.RetryAfter = time.Duration(n) * time.Second
	} else if until, err := http.ParseTime(resp.Header.Get("Retry-After")); err == nil {
		result.RetryAfter = time.Until(until)
	}
	persist := func(provider, raw string) error {
		if save == nil {
			return nil
		}
		return save(Progress{GenerationID: result.GenerationID, RequestID: result.RequestID, Provider: provider, HTTPStatus: result.Status, Raw: c.redact(raw)})
	}
	if err := persist("", ""); err != nil {
		result.Error = "persist response headers: " + err.Error()
		return
	}
	media, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if result.Status != 200 || media != "text/event-stream" {
		b, err := io.ReadAll(resp.Body)
		result.Body = []byte(c.redact(string(b)))
		if err != nil {
			result.Error = c.redact(err.Error())
			return
		}
		var g Generation
		if json.Unmarshal(result.Body, &g) == nil && g.ID != "" {
			if result.GenerationID != "" && result.GenerationID != g.ID {
				result.Error = "response generation ID changed"
				return
			}
			result.GenerationID = g.ID
			if err := persist(g.Provider, ""); err != nil {
				result.Error = err.Error()
			}
		}
		return
	}
	var g Generation
	var content, refusal strings.Builder
	var finish string
	var done bool
	defer func() {
		g.ID = result.GenerationID
		// An interrupted stream can retain partial content for audit, never for scoring.
		b, _ := json.Marshal(map[string]any{"id": g.ID, "provider": g.Provider, "model": g.Model, "usage": g.Usage, "error": g.Error,
			"choices": []any{map[string]any{"finish_reason": finish, "message": map[string]any{"content": content.String(), "refusal": refusal.String()}}}})
		result.Body = b
	}()
	consume := func(raw string) error {
		raw = c.redact(raw)
		if raw == "[DONE]" {
			if err := persist(g.Provider, raw); err != nil {
				return err
			}
			done = true
			if finish == "" {
				return fmt.Errorf("stream ended without a finish reason")
			}
			return nil
		}
		var chunk struct {
			ID       string          `json:"id"`
			Provider string          `json:"provider"`
			Model    string          `json:"model"`
			Usage    *Usage          `json:"usage"`
			Error    json.RawMessage `json:"error"`
			Choices  []struct {
				Index  int    `json:"index"`
				Finish string `json:"finish_reason"`
				Delta  struct {
					Content *string `json:"content"`
					Refusal *string `json:"refusal"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(raw), &chunk); err != nil {
			if saveErr := persist(g.Provider, raw); saveErr != nil {
				return saveErr
			}
			return fmt.Errorf("malformed SSE JSON: %w", err)
		}
		if chunk.ID != "" {
			if result.GenerationID != "" && result.GenerationID != chunk.ID {
				return fmt.Errorf("stream generation ID changed")
			}
			result.GenerationID = chunk.ID
		}
		if chunk.Provider != "" {
			g.Provider = chunk.Provider
		}
		if chunk.Model != "" {
			g.Model = chunk.Model
		}
		if err := persist(g.Provider, raw); err != nil {
			return err
		}
		if chunk.Usage != nil {
			g.Usage = *chunk.Usage
		}
		if len(chunk.Error) > 0 && string(chunk.Error) != "null" {
			g.Error = chunk.Error
			return fmt.Errorf("stream API error: %s", chunk.Error)
		}
		if len(chunk.Choices) > 1 {
			return fmt.Errorf("stream contains multiple choices")
		}
		for _, choice := range chunk.Choices {
			if choice.Index != 0 {
				return fmt.Errorf("stream choice index must be zero")
			}
			if finish != "" && ((choice.Delta.Content != nil && *choice.Delta.Content != "") || (choice.Delta.Refusal != nil && *choice.Delta.Refusal != "")) {
				return fmt.Errorf("stream content after finish")
			}
			if choice.Delta.Content != nil {
				content.WriteString(*choice.Delta.Content)
			}
			if choice.Delta.Refusal != nil {
				refusal.WriteString(*choice.Delta.Refusal)
			}
			if choice.Finish != "" {
				if finish != "" && finish != choice.Finish {
					return fmt.Errorf("stream finish reason changed")
				}
				finish = choice.Finish
				if finish == "error" {
					return fmt.Errorf("stream finished with error")
				}
			}
		}
		return nil
	}
	// Scanner's default 64 KiB limit is too small for reasoning events. This bound
	// limits one wire frame, not the model's token budget or entire response.
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 4096), 16<<20)
	var data []string
	size := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if len(data) == 0 {
				continue
			}
			if err := consume(strings.Join(data, "\n")); err != nil {
				result.Error = c.redact(err.Error())
				return
			}
			data = nil
			size = 0
			if done {
				return
			}
		} else if strings.HasPrefix(line, "data:") {
			value := strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " ")
			size += len(value)
			if size > 16<<20 {
				result.Error = "SSE frame exceeds 16 MiB"
				return
			}
			data = append(data, value)
		}
	}
	if err := scanner.Err(); err != nil {
		result.Error = c.redact(err.Error())
	} else {
		result.Error = "stream disconnected before [DONE]"
	}
	return
}
