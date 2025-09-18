package langgraph

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
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
	Role   string         `json:"role"`
	Prompt string         `json:"prompt"`
	Params map[string]any `json:"params,omitempty"`
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

func NewClient() *Client {
	return nil
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
	return nil, nil
}

func (c *Client) AskQuestions(req QuestionsRequest) (*QuestionsResponse, error) {
	// Not supported with direct OpenAI integration; kept for backward compatibility
	return nil, errors.New("AskQuestions is not supported with OpenAI client")
}

// HealthCheck calls GET /v1/models to verify API reachability.
func (c *Client) HealthCheck() error {
	url := c.baseURL + "/v1/models"
	attempts := c.maxRetries + 1
	var lastErr error
	for i := range attempts {
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

func (c *Client) doJSON(method, url string, in any, out any) error {
	return nil
}

func transient(err error) bool {
	var ne net.Error
	if errors.As(err, &ne) {
		return ne.Timeout() || ne.Temporary()
	}
	// treat DNS errors and connection refused as transient
	var opErr *net.OpError

	return errors.As(err, &opErr)
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
