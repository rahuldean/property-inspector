package inspector

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

// chatRequest is the OpenAI-compatible request body that LiteLLM expects.
type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []contentPart
}

type contentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL string `json:"url"`
}

// chatResponse is the subset of the OpenAI response we care about.
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// encodeImageToDataURI reads an image file and returns a base64 data URI.
func encodeImageToDataURI(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading image %s: %w", path, err)
	}

	mime := "image/jpeg"
	if strings.HasSuffix(strings.ToLower(path), ".png") {
		mime = "image/png"
	} else if strings.HasSuffix(strings.ToLower(path), ".webp") {
		mime = "image/webp"
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mime, encoded), nil
}

// sendChat sends a request to the LiteLLM proxy and returns the raw
// text content from the model's response. Retries on 5xx/429 with
// exponential backoff.
// TODO: Support streaming
func (c *Client) sendChat(ctx context.Context, systemPrompt string, parts []contentPart) (string, error) {
	messages := []chatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: parts},
	}

	body := chatRequest{
		Model:          c.model,
		Messages:       messages,
		ResponseFormat: &responseFormat{Type: "json_object"},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * 500 * time.Millisecond
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/chat/completions", bytes.NewReader(payload))
		if err != nil {
			return "", fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		if c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response: %w", err)
			continue
		}

		// Retry on rate limit or server errors
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = &LiteLLMError{
				StatusCode: resp.StatusCode,
				Model:      c.model,
				Message:    string(respBody),
			}
			continue
		}

		if resp.StatusCode != 200 {
			return "", &LiteLLMError{
				StatusCode: resp.StatusCode,
				Model:      c.model,
				Message:    string(respBody),
			}
		}

		var chatResp chatResponse
		if err := json.Unmarshal(respBody, &chatResp); err != nil {
			return "", fmt.Errorf("parsing response JSON: %w", err)
		}

		if chatResp.Error != nil {
			return "", &LiteLLMError{
				StatusCode: resp.StatusCode,
				Model:      c.model,
				Message:    chatResp.Error.Message,
			}
		}

		if len(chatResp.Choices) == 0 {
			return "", fmt.Errorf("empty response from model %s", c.model)
		}

		return stripCodeFences(chatResp.Choices[0].Message.Content), nil
	}

	return "", fmt.Errorf("all %d retries exhausted: %w", c.maxRetries, lastErr)
}

// stripCodeFences removes markdown code fences that models sometimes
// wrap around JSON despite being told not to.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// Remove opening fence (```json or just ```)
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		// Remove closing fence
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	return s
}
