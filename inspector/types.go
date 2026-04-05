package inspector

import "time"

type RoomMeta struct {
	RoomName    string    `json:"room_name"`
	FloorUnit   string    `json:"floor_unit"`
	InspectedAt time.Time `json:"inspected_at"`
}

type Issue struct {
	Category    string  `json:"category"`    // "Wall Damage", "Flooring", "Appliance", "Fixture"
	Severity    string  `json:"severity"`    // "minor", "moderate", "severe"
	Description string  `json:"description"`
	Location    string  `json:"location"`    // "north wall near window"
	Confidence  float64 `json:"confidence"`  // 0.0 - 1.0
}

type RoomAnalysis struct {
	RoomMeta         RoomMeta `json:"room_meta"`
	Issues           []Issue  `json:"issues"`
	Summary          string   `json:"summary"`
	OverallCondition string   `json:"overall_condition"` // "excellent", "good", "fair", "poor"
	GeneratedAt      time.Time `json:"generated_at"`
}

type ComparisonReport struct {
	RoomMeta        RoomMeta     `json:"room_meta"`
	BeforeAnalysis  RoomAnalysis `json:"before_analysis"`
	AfterAnalysis   RoomAnalysis `json:"after_analysis"`
	ResolvedIssues  []Issue      `json:"resolved_issues"`
	NewIssues       []Issue      `json:"new_issues"`
	UnchangedIssues []Issue      `json:"unchanged_issues"`
	Summary         string       `json:"summary"`
	GeneratedAt     time.Time    `json:"generated_at"`
}
