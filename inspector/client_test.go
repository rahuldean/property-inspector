package inspector

import (
	"net/http"
	"testing"
	"time"
)

func TestNewClientDefaults(t *testing.T) {
	c := NewClient()

	if c.baseURL != "http://localhost:4000" {
		t.Errorf("expected default baseURL http://localhost:4000, got %s", c.baseURL)
	}
	if c.model != "inspector" {
		t.Errorf("expected default model inspector, got %s", c.model)
	}
	if c.maxRetries != 3 {
		t.Errorf("expected default maxRetries 3, got %d", c.maxRetries)
	}
	if c.timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %s", c.timeout)
	}
	if c.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	c := NewClient(
		WithBaseURL("http://proxy:9000"),
		WithAPIKey("test-key"),
		WithModel("gpt-4o"),
		WithTimeout(10*time.Second),
		WithMaxRetries(5),
		WithHTTPClient(custom),
	)

	if c.baseURL != "http://proxy:9000" {
		t.Errorf("expected baseURL http://proxy:9000, got %s", c.baseURL)
	}
	if c.apiKey != "test-key" {
		t.Errorf("expected apiKey test-key, got %s", c.apiKey)
	}
	if c.model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %s", c.model)
	}
	if c.maxRetries != 5 {
		t.Errorf("expected maxRetries 5, got %d", c.maxRetries)
	}
	if c.httpClient != custom {
		t.Error("expected custom httpClient to be used")
	}
}
