package openrouter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStreamAssemblyAndAccounting(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("X-Generation-Id", "gen-stream")
		w.Header().Set("X-Request-Id", "req-stream")
		fmt.Fprint(w, ": OPENROUTER PROCESSING\r\n\r\nevent: message\ndata: {\"id\":\"gen-stream\",\ndata: \"provider\":\"test\",\"choices\":[{\"index\":0,\"delta\":{\"reasoning\":\"تفکر\"}}]}\n\n")
		fmt.Fprint(w, "data: {\"id\":\"gen-stream\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"{\\\"answer\\\":\"}}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"3}\"},\"finish_reason\":\"stop\"}]}\n\n")
		// OpenRouter repeats the terminal choice in its accounting chunk.
		fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":90,\"total_tokens\":100,\"cost\":0.1}}\n\ndata: [DONE]\n\n")
	}))
	defer srv.Close()
	var events []Progress
	res := New(srv.URL, "").Complete(context.Background(), []byte(`{}`), func(p Progress) error { events = append(events, p); return nil })
	g, answer, invalid, err := Parse(res.Body)
	if res.Error != "" || err != nil || invalid != "" || answer != 3 || *g.Usage.Total != 100 || *g.Usage.Cost != 0.1 {
		t.Fatalf("result %+v parse %d %s %v", res, answer, invalid, err)
	}
	if len(events) != 6 || events[0].GenerationID != "gen-stream" || events[0].RequestID != "req-stream" || events[0].Raw != "" || !strings.Contains(events[1].Raw, "تفکر") {
		t.Fatalf("events %+v", events)
	}
}

func TestStreamInterruptionsNeverScore(t *testing.T) {
	cases := map[string]string{
		"truncated":           `data: {"id":"gen-test","choices":[{"index":0,"delta":{"content":"{\"answer\":1}"}}]}` + "\n\n",
		"finish without done": `data: {"id":"gen-test","choices":[{"index":0,"delta":{"content":"{\"answer\":1}"},"finish_reason":"stop"}]}` + "\n\n",
		"done without finish": "data: [DONE]\n\n",
		"invalid json":        "data: {broken}\n\n",
		"midstream error":     `data: {"id":"gen-test","error":{"message":"disconnected"},"choices":[]}` + "\n\n",
		"changed id":          `data: {"id":"gen-other","choices":[]}` + "\n\n",
		"multiple choices":    `data: {"choices":[{"index":0},{"index":1}]}` + "\n\n",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("X-Generation-Id", "gen-test")
				fmt.Fprint(w, body)
			}))
			defer srv.Close()
			res := New(srv.URL, "").Complete(context.Background(), []byte(`{}`), nil)
			if res.Error == "" || res.GenerationID != "gen-test" {
				t.Fatalf("unsafe stream result: %+v", res)
			}
		})
	}
}

func TestStreamDeadlinePersistsHeaderIDBeforeBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("X-Generation-Id", "gen-hanging")
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var first Progress
	res := New(srv.URL, "").Complete(ctx, []byte(`{}`), func(p Progress) error { first = p; cancel(); return nil })
	if first.GenerationID != "gen-hanging" || res.Error == "" || res.GenerationID != "gen-hanging" {
		t.Fatalf("lost interrupted ID: %+v %+v", first, res)
	}
}

func TestNonStreamFallbackAndPersistenceFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Generation-Id", "gen-fallback")
			fmt.Fprint(w, `{"id":"gen-fallback","choices":[{"finish_reason":"stop","message":{"content":"{\"answer\":2}"}}]}`)
		}))
		res := New(srv.URL, "").Complete(context.Background(), []byte(`{}`), func(Progress) error {
			if fail {
				return fmt.Errorf("disk full")
			}
			return nil
		})
		srv.Close()
		if fail {
			if !strings.Contains(res.Error, "disk full") {
				t.Fatal(res)
			}
		} else {
			_, answer, _, err := Parse(res.Body)
			if err != nil || answer != 2 {
				t.Fatalf("%+v %v", res, err)
			}
		}
	}
}

func TestStreamRedactionAndLargeReasoningFrame(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: {\"id\":\"gen-test\",\"choices\":[{\"delta\":{\"reasoning\":\"%ssecret-key\"}}]}\n\n", strings.Repeat("x", 70000))
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"secret-key\"},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n")
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := New(srv.URL, "secret-key").Complete(ctx, []byte(`{}`), func(p Progress) error {
		if strings.Contains(p.Raw, "secret-key") {
			t.Error("secret in audit")
		}
		return nil
	})
	if res.Error != "" || strings.Contains(string(res.Body), "secret-key") {
		t.Fatal("large event failed or secret leaked", res.Error)
	}
}
