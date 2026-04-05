package inspector

import "fmt"

// LiteLLMError wraps errors from the LiteLLM proxy so callers
// can inspect the HTTP status and upstream message.
type LiteLLMError struct {
	StatusCode int
	Model      string
	Message    string
}

func (e *LiteLLMError) Error() string {
	return fmt.Sprintf("litellm: status %d from model %q: %s", e.StatusCode, e.Model, e.Message)
}
