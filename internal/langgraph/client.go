package langgraph

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
)

type Client struct {
	baseURL    string
	hc         *http.Client
	apiKey     string
	maxRetries int
	baseDelay  time.Duration
	timeout    time.Duration
}

type RunRequest struct {
	Role   string                 `json:"role"`
	Prompt string                 `json:"prompt"`
	Params map[string]interface{} `json:"params,omitempty"`
}

type RunResponse struct {
	RunID   string `json:"runId"`
	Content string `json:"content"`
}

type QuestionsRequest struct {
	RunID     string   `json:"runId"`
	Questions []string `json:"questions"`
}

type QuestionsResponse struct {
	Status string `json:"status"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    trimTrailingSlash(baseURL),
		hc:         &http.Client{Timeout: 30 * time.Second},
		apiKey:     os.Getenv("LANGGRAPH_API_KEY"),
		maxRetries: 3,
		baseDelay:  300 * time.Millisecond,
		timeout:    30 * time.Second,
	}
}

// WithRetries allows overriding retry behavior.
func (c *Client) WithRetries(maxRetries int, baseDelay time.Duration) *Client {
	c.maxRetries = maxRetries
	c.baseDelay = baseDelay
	return c
}

// WithTimeout sets per-request timeout on the underlying HTTP client.
func (c *Client) WithTimeout(d time.Duration) *Client {
	c.timeout = d
	c.hc.Timeout = d
	return c
}

func (c *Client) RunAgent(req RunRequest) (*RunResponse, error) {
	endpoint := c.baseURL + "/agents/run"
	var out RunResponse
	if err := c.doJSON(http.MethodPost, endpoint, req, &out); err != nil {
		return nil, err
	}
	if out.RunID == "" && out.Content == "" {
		return nil, errors.New("invalid response from LangGraph")
	}
	return &out, nil
}

func (c *Client) AskQuestions(req QuestionsRequest) (*QuestionsResponse, error) {
	endpoint := c.baseURL + "/agents/questions"
	var out QuestionsResponse
	if err := c.doJSON(http.MethodPost, endpoint, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// HealthCheck calls GET /healthz and returns nil if status OK (2xx) and optionally parses a {status} body.
func (c *Client) HealthCheck() error {
	url := c.baseURL + "/healthz"
	attempts := c.maxRetries + 1
	var lastErr error
	for i := 0; i < attempts; i++ {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		if c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
		}
		resp, err := c.hc.Do(req)
		if err != nil {
			if transient(err) && i < attempts-1 {
				backoffSleep(c.baseDelay, i)
				continue
			}
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		b, _ := io.ReadAll(resp.Body)
		lastErr = fmt.Errorf("healthcheck error: %s: %s", resp.Status, string(b))
		if i < attempts-1 {
			backoffSleep(c.baseDelay, i)
			continue
		}
	}
	if lastErr == nil {
		lastErr = errors.New("healthcheck failed")
	}
	return lastErr
}

func (c *Client) doJSON(method, url string, in interface{}, out interface{}) error {
	var payload []byte
	var err error
	if in != nil {
		payload, err = json.Marshal(in)
		if err != nil {
			return err
		}
	}
	attempts := c.maxRetries + 1
	for i := 0; i < attempts; i++ {
		var body io.Reader
		if payload != nil {
			body = bytes.NewReader(payload)
		}
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		if c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
		}
		resp, err := c.hc.Do(req)
		if err != nil {
			if transient(err) && i < attempts-1 {
				backoffSleep(c.baseDelay, i)
				continue
			}
			return err
		}
		func() {
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				b, _ := io.ReadAll(resp.Body)
				err = fmt.Errorf("langgraph error: %s: %s", resp.Status, string(b))
				return
			}
			if out != nil {
				err = json.NewDecoder(resp.Body).Decode(out)
			}
		}()
		if err == nil {
			return nil
		}
		if i < attempts-1 {
			backoffSleep(c.baseDelay, i)
			continue
		}
		return err
	}
	return errors.New("unreachable")
}

func transient(err error) bool {
	var ne net.Error
	if errors.As(err, &ne) {
		return ne.Timeout() || ne.Temporary()
	}
	// treat DNS errors and connection refused as transient
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	return false
}

func backoffSleep(base time.Duration, attempt int) {
	// exponential backoff with jitter
	d := time.Duration(1<<attempt) * base
	jitter := time.Duration(rand.Int63n(int64(base)))
	time.Sleep(d + jitter)
}

func trimTrailingSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
