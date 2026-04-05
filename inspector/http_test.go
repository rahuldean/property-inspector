package inspector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncodeImageToDataURI(t *testing.T) {
	// Create a tiny test file for each supported type.
	tests := []struct {
		filename string
		wantMIME string
	}{
		{"test.jpg", "data:image/jpeg;base64,"},
		{"test.png", "data:image/png;base64,"},
		{"test.webp", "data:image/webp;base64,"},
	}

	dir := t.TempDir()
	for _, tt := range tests {
		path := filepath.Join(dir, tt.filename)
		os.WriteFile(path, []byte("fake image data"), 0644)

		uri, err := encodeImageToDataURI(path)
		if err != nil {
			t.Fatalf("encodeImageToDataURI(%s): %v", tt.filename, err)
		}
		if !strings.HasPrefix(uri, tt.wantMIME) {
			t.Errorf("expected prefix %s, got %s", tt.wantMIME, uri[:40])
		}
	}
}

func TestEncodeImageToDataURIMissingFile(t *testing.T) {
	_, err := encodeImageToDataURI("/does/not/exist.jpg")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestSendChatSuccess(t *testing.T) {
	expected := `{"issues":[],"summary":"all good","overall_condition":"excellent"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request looks right.
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected content-type: %s", r.Header.Get("Content-Type"))
		}

		var req chatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "test-model" {
			t.Errorf("expected model test-model, got %s", req.Model)
		}

		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: expected}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithModel("test-model"))
	result, err := c.sendChat(context.Background(), "system prompt", []contentPart{
		{Type: "text", Text: "hello"},
	})
	if err != nil {
		t.Fatalf("sendChat: %v", err)
	}
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestSendChatAuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer my-key" {
			t.Errorf("expected Authorization 'Bearer my-key', got %q", auth)
		}
		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "ok"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithAPIKey("my-key"))
	_, err := c.sendChat(context.Background(), "test", []contentPart{{Type: "text", Text: "hi"}})
	if err != nil {
		t.Fatalf("sendChat: %v", err)
	}
}

func TestSendChat4xxError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": {"message": "bad request"}}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithMaxRetries(0))
	_, err := c.sendChat(context.Background(), "test", []contentPart{{Type: "text", Text: "hi"}})
	if err == nil {
		t.Fatal("expected error on 400")
	}
	litellmErr, ok := err.(*LiteLLMError)
	if !ok {
		t.Fatalf("expected *LiteLLMError, got %T", err)
	}
	if litellmErr.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", litellmErr.StatusCode)
	}
}

func TestSendChatRetriesOn500(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(500)
			w.Write([]byte(`server error`))
			return
		}
		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "recovered"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithMaxRetries(3))
	result, err := c.sendChat(context.Background(), "test", []contentPart{{Type: "text", Text: "hi"}})
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if result != "recovered" {
		t.Errorf("expected 'recovered', got %q", result)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestSendChatContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	c := NewClient(WithBaseURL(srv.URL), WithMaxRetries(3))
	_, err := c.sendChat(ctx, "test", []contentPart{{Type: "text", Text: "hi"}})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
