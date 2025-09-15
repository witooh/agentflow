package langgraph

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

type dummyNetErr struct{}

func (dummyNetErr) Error() string   { return "dummy" }
func (dummyNetErr) Timeout() bool   { return true }
func (dummyNetErr) Temporary() bool { return true }

func TestTrimTrailingSlash(t *testing.T) {
	cases := map[string]string{
		"http://localhost:8123/": "http://localhost:8123",
		"http://localhost:8123":  "http://localhost:8123",
		"":                       "",
		"/api/v1///":             "/api/v1",
	}
	for in, want := range cases {
		if got := trimTrailingSlash(in); got != want {
			t.Fatalf("trimTrailingSlash(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTransientNetError(t *testing.T) {
	if !transient(dummyNetErr{}) {
		t.Fatalf("expected transient to treat net.Error with Timeout()=true as transient")
	}
}

func TestClientRunRetryAndAuth(t *testing.T) {
	var runCalls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			w.WriteHeader(404)
			return
		}
		c := atomic.AddInt32(&runCalls, 1)
		// Require auth header and non-empty body on all attempts
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			w.WriteHeader(401)
			w.Write([]byte("missing auth"))
			return
		}
		if r.Body == nil {
			w.WriteHeader(400)
			return
		}
		// We don't need full schema validation; just ensure JSON is present
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(400)
			return
		}
		if _, ok := body["model"]; !ok {
			w.WriteHeader(400)
			return
		}
		// First two attempts fail with 500 to trigger retry
		if c < 3 {
			w.WriteHeader(500)
			w.Write([]byte("temporary"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"id": "chatcmpl-123",
			"choices": []any{
				map[string]any{
					"message": map[string]any{
						"role":    "assistant",
						"content": "ok",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	os.Setenv("OPENAI_API_KEY", "test-token")
	defer os.Unsetenv("OPENAI_API_KEY")

	c := NewClient().WithRetries(3, time.Nanosecond).WithTimeout(2 * time.Second)

	out, err := c.RunAgent(RunRequest{Role: "dev", Prompt: "hello", Params: map[string]any{"model": "gpt-4o-mini"}})
	if err != nil {
		t.Fatalf("RunAgent error: %v", err)
	}
	if out.RunID != "chatcmpl-123" || out.Content != "ok" {
		t.Fatalf("unexpected run response: %+v", out)
	}
	if got := atomic.LoadInt32(&runCalls); got != 3 {
		t.Fatalf("expected 3 calls to /v1/chat/completions, got %d", got)
	}
}

func TestHealthCheckRetriesThenOK(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			w.WriteHeader(404)
			return
		}
		c := atomic.AddInt32(&calls, 1)
		if c < 3 {
			w.WriteHeader(500)
			w.Write([]byte("bad"))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := NewClient().WithRetries(3, time.Nanosecond)
	if err := c.HealthCheck(); err != nil {
		t.Fatalf("HealthCheck error: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 healthz calls, got %d", got)
	}
}
