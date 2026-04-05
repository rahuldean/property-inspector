package inspector

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// AnalyzeRoom takes an image path and room metadata, sends it to the
// vision model via LiteLLM, and returns a structured analysis.
func (c *Client) AnalyzeRoom(ctx context.Context, imagePath string, meta RoomMeta) (*RoomAnalysis, error) {
	dataURI, err := encodeImageToDataURI(imagePath)
	if err != nil {
		return nil, err
	}

	parts := []contentPart{
		{Type: "image_url", ImageURL: &imageURL{URL: dataURI}},
		{Type: "text", Text: fmt.Sprintf("Analyze this image of %s in %s.", meta.RoomName, meta.FloorUnit)},
	}

	raw, err := c.sendChat(ctx, analyzeSystemPrompt, parts)
	if err != nil {
		return nil, err
	}

	var result struct {
		Issues           []Issue `json:"issues"`
		Summary          string  `json:"summary"`
		OverallCondition string  `json:"overall_condition"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parsing model response: %w\nraw response: %s", err, raw)
	}

	return &RoomAnalysis{
		RoomMeta:         meta,
		Issues:           result.Issues,
		Summary:          result.Summary,
		OverallCondition: result.OverallCondition,
		GeneratedAt:      time.Now(),
	}, nil
}
