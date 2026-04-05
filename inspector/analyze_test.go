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

func TestAnalyzeRoomSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		modelResp := `{
			"issues": [
				{
					"category": "Wall Damage",
					"severity": "minor",
					"description": "Small scuff on north wall",
					"location": "north wall",
					"confidence": 0.8
				}
			],
			"summary": "Room has minor wall damage.",
			"overall_condition": "good"
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

	// Create a fake image file.
	imgPath := filepath.Join(t.TempDir(), "room.jpg")
	os.WriteFile(imgPath, []byte("fake jpg"), 0644)

	c := NewClient(WithBaseURL(srv.URL))
	meta := RoomMeta{RoomName: "Kitchen", FloorUnit: "Unit 2A", InspectedAt: time.Now()}

	result, err := c.AnalyzeRoom(context.Background(), imgPath, meta)
	if err != nil {
		t.Fatalf("AnalyzeRoom: %v", err)
	}

	if result.OverallCondition != "good" {
		t.Errorf("expected condition 'good', got %q", result.OverallCondition)
	}
	if len(result.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(result.Issues))
	}
	if result.Issues[0].Category != "Wall Damage" {
		t.Errorf("expected category 'Wall Damage', got %q", result.Issues[0].Category)
	}
	if result.RoomMeta.RoomName != "Kitchen" {
		t.Errorf("expected room name 'Kitchen', got %q", result.RoomMeta.RoomName)
	}
}

func TestAnalyzeRoomBadImage(t *testing.T) {
	c := NewClient()
	_, err := c.AnalyzeRoom(context.Background(), "/nonexistent/image.jpg", RoomMeta{})
	if err == nil {
		t.Error("expected error for missing image")
	}
}

func TestAnalyzeRoomBadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "this is not json"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	imgPath := filepath.Join(t.TempDir(), "room.jpg")
	os.WriteFile(imgPath, []byte("fake"), 0644)

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.AnalyzeRoom(context.Background(), imgPath, RoomMeta{})
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}
