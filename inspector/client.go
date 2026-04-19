package inspector

import (
	"net/http"
	"time"
)

type Client struct {
	baseURL              string
	apiKey               string
	model                string
	timeout              time.Duration
	maxRetries           int
	httpClient           *http.Client
	cfAccessClientID     string
	cfAccessClientSecret string
}

type Option func(*Client)

func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:    "http://localhost:4000",
		model:      "inspector",
		timeout:    30 * time.Second,
		maxRetries: 3,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: c.timeout}
	}
	return c
}

func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

func WithAPIKey(key string) Option {
	return func(c *Client) { c.apiKey = key }
}

func WithModel(model string) Option {
	return func(c *Client) { c.model = model }
}

func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

func WithMaxRetries(n int) Option {
	return func(c *Client) { c.maxRetries = n }
}

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

func WithCFAccessClientID(id string) Option {
	return func(c *Client) { c.cfAccessClientID = id }
}

func WithCFAccessClientSecret(secret string) Option {
	return func(c *Client) { c.cfAccessClientSecret = secret }
}
