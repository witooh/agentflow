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
	"strings"
	"time"
	"strconv"
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
    // Default to OpenAI public API; allow override via OPENAI_BASE_URL for tests/dev.
    base := strings.TrimSpace(os.Getenv("OPENAI_BASE_URL"))
    if base == "" {
        base = "https://api.openai.com"
    }
    c := &Client{
        baseURL:    trimTrailingSlash(base),
        hc:         &http.Client{Timeout: 60 * time.Second},
        apiKey:     os.Getenv("OPENAI_API_KEY"),
        maxRetries: 3,
        baseDelay:  500 * time.Millisecond,
        timeout:    60 * time.Second,
    }
    // Optional environment overrides
    if v := strings.TrimSpace(os.Getenv("AGENTFLOW_HTTP_TIMEOUT")); v != "" {
        if d, err := time.ParseDuration(v); err == nil {
            c.WithTimeout(d)
        } else if n, err := strconv.Atoi(v); err == nil {
            c.WithTimeout(time.Duration(n) * time.Second)
        }
    } else if v := strings.TrimSpace(os.Getenv("OPENAI_TIMEOUT")); v != "" { // alias
        if d, err := time.ParseDuration(v); err == nil {
            c.WithTimeout(d)
        } else if n, err := strconv.Atoi(v); err == nil {
            c.WithTimeout(time.Duration(n) * time.Second)
        }
    }
    if v := strings.TrimSpace(os.Getenv("AGENTFLOW_MAX_RETRIES")); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n >= 0 {
            c.maxRetries = n
        }
    }
    if v := strings.TrimSpace(os.Getenv("AGENTFLOW_RETRY_BASE_DELAY_MS")); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n >= 0 {
            c.baseDelay = time.Duration(n) * time.Millisecond
        }
    }
    return c
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
	// Translate RunRequest into OpenAI Chat Completions request
	type chatMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type chatReq struct {
		Model       string        `json:"model"`
		Messages    []chatMessage `json:"messages"`
		Temperature float64       `json:"temperature,omitempty"`
		MaxTokens   int           `json:"max_tokens,omitempty"`
	}
	type chatResp struct {
		ID      string `json:"id"`
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	model := ""
	if req.Params != nil {
		if v, ok := req.Params["model"].(string); ok {
			model = v
		}
	}
	if model == "" {
		model = "gpt-5"
	}
	temperature := 0.0
	if req.Params != nil {
		if v, ok := req.Params["temperature"].(float64); ok {
			temperature = v
		} else if v, ok := req.Params["temperature"].(float32); ok {
			temperature = float64(v)
		} else if v, ok := req.Params["temperature"].(int); ok {
			temperature = float64(v)
		}
	}
	// maxTokens := 0
	// if req.Params != nil {
	// 	if v, ok := req.Params["max_tokens"].(int); ok {
	// 		maxTokens = v
	// 	} else if v, ok := req.Params["max_tokens"].(float64); ok {
	// 		maxTokens = int(v)
	// 	}
	// }

	in := chatReq{
		Model:       model,
		Messages:    []chatMessage{{Role: "user", Content: req.Prompt}},
		Temperature: temperature,
		// MaxTokens:   maxTokens,
	}

	endpoint := c.baseURL + "/v1/chat/completions"
	var out chatResp
	if err := c.doJSON(http.MethodPost, endpoint, in, &out); err != nil {
		return nil, err
	}
	content := ""
	if len(out.Choices) > 0 {
		content = out.Choices[0].Message.Content
	}
	return &RunResponse{RunID: out.ID, Content: content}, nil
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
	var payload []byte
	var err error
	if in != nil {
		payload, err = json.Marshal(in)
		if err != nil {
			return err
		}
	}
	attempts := c.maxRetries + 1
	for i := range attempts {
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
				err = fmt.Errorf("openai error: %s: %s", resp.Status, string(b))
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
