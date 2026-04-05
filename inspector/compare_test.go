package inspector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCompareInspectionsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		modelResp := `{
			"before_analysis": {
				"issues": [{"category": "Flooring", "severity": "moderate", "description": "Scratched hardwood", "location": "center", "confidence": 0.9}],
				"summary": "Floor damage noted.",
				"overall_condition": "fair"
			},
			"after_analysis": {
				"issues": [],
				"summary": "Floor has been refinished.",
				"overall_condition": "good"
			},
			"resolved_issues": [{"category": "Flooring", "severity": "moderate", "description": "Scratched hardwood", "location": "center", "confidence": 0.9}],
			"new_issues": [],
			"unchanged_issues": [],
			"summary": "Floor damage was repaired between inspections."
		}`
		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: modelResp}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	dir := t.TempDir()
	beforePath := filepath.Join(dir, "before.jpg")
	afterPath := filepath.Join(dir, "after.jpg")
	os.WriteFile(beforePath, []byte("fake before"), 0644)
	os.WriteFile(afterPath, []byte("fake after"), 0644)

	c := NewClient(WithBaseURL(srv.URL))
	meta := RoomMeta{RoomName: "Living Room", FloorUnit: "Unit 1A", InspectedAt: time.Now()}

	report, err := c.CompareInspections(context.Background(), beforePath, afterPath, meta)
	if err != nil {
		t.Fatalf("CompareInspections: %v", err)
	}

	if len(report.ResolvedIssues) != 1 {
		t.Errorf("expected 1 resolved issue, got %d", len(report.ResolvedIssues))
	}
	if len(report.NewIssues) != 0 {
		t.Errorf("expected 0 new issues, got %d", len(report.NewIssues))
	}
	if report.Summary != "Floor damage was repaired between inspections." {
		t.Errorf("unexpected summary: %s", report.Summary)
	}
	if report.RoomMeta.RoomName != "Living Room" {
		t.Errorf("expected room name 'Living Room', got %q", report.RoomMeta.RoomName)
	}
}

func TestCompareInspectionsMissingBefore(t *testing.T) {
	afterPath := filepath.Join(t.TempDir(), "after.jpg")
	os.WriteFile(afterPath, []byte("fake"), 0644)

	c := NewClient()
	_, err := c.CompareInspections(context.Background(), "/nonexistent.jpg", afterPath, RoomMeta{})
	if err == nil {
		t.Error("expected error for missing before image")
	}
}

func TestCompareInspectionsMissingAfter(t *testing.T) {
	beforePath := filepath.Join(t.TempDir(), "before.jpg")
	os.WriteFile(beforePath, []byte("fake"), 0644)

	c := NewClient()
	_, err := c.CompareInspections(context.Background(), beforePath, "/nonexistent.jpg", RoomMeta{})
	if err == nil {
		t.Error("expected error for missing after image")
	}
}
