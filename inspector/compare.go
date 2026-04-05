package inspector

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CompareInspections takes before/after images of the same room and returns
// a report of what changed between inspections.
func (c *Client) CompareInspections(ctx context.Context, beforePath, afterPath string, meta RoomMeta) (*ComparisonReport, error) {
	beforeURI, err := encodeImageToDataURI(beforePath)
	if err != nil {
		return nil, fmt.Errorf("encoding before image: %w", err)
	}

	afterURI, err := encodeImageToDataURI(afterPath)
	if err != nil {
		return nil, fmt.Errorf("encoding after image: %w", err)
	}

	parts := []contentPart{
		{Type: "text", Text: "Image 1 (BEFORE):"},
		{Type: "image_url", ImageURL: &imageURL{URL: beforeURI}},
		{Type: "text", Text: "Image 2 (AFTER):"},
		{Type: "image_url", ImageURL: &imageURL{URL: afterURI}},
		{Type: "text", Text: fmt.Sprintf("Compare these two inspection photos of %s in %s.", meta.RoomName, meta.FloorUnit)},
	}

	raw, err := c.sendChat(ctx, compareSystemPrompt, parts)
	if err != nil {
		return nil, err
	}

	var result struct {
		BeforeAnalysis  RoomAnalysis `json:"before_analysis"`
		AfterAnalysis   RoomAnalysis `json:"after_analysis"`
		ResolvedIssues  []Issue      `json:"resolved_issues"`
		NewIssues       []Issue      `json:"new_issues"`
		UnchangedIssues []Issue      `json:"unchanged_issues"`
		Summary         string       `json:"summary"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parsing model response: %w\nraw response: %s", err, raw)
	}

	return &ComparisonReport{
		RoomMeta:        meta,
		BeforeAnalysis:  result.BeforeAnalysis,
		AfterAnalysis:   result.AfterAnalysis,
		ResolvedIssues:  result.ResolvedIssues,
		NewIssues:       result.NewIssues,
		UnchangedIssues: result.UnchangedIssues,
		Summary:         result.Summary,
		GeneratedAt:     time.Now(),
	}, nil
}
