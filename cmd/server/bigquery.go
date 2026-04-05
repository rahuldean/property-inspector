package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/bigquery"
)

const (
	bqDataset = "property_inspector"
	bqTable   = "inspections"
)

type inspectionRow struct {
	ID               string    `bigquery:"id"`
	RoomName         string    `bigquery:"room_name"`
	FloorUnit        string    `bigquery:"floor_unit"`
	Endpoint         string    `bigquery:"endpoint"`
	ModelUsed        string    `bigquery:"model_used"`
	OverallCondition string    `bigquery:"overall_condition"`
	BeforeIssueCount bigquery.NullInt64 `bigquery:"before_issue_count"`
	AfterIssueCount  bigquery.NullInt64 `bigquery:"after_issue_count"`
	ResponseTimeMs   int64     `bigquery:"response_time_ms"`
	Error            bool      `bigquery:"error"`
	InspectedAt      time.Time `bigquery:"inspected_at"`
}

type BQLogger struct {
	inserter *bigquery.Inserter
}

func newBQLogger(ctx context.Context, projectID string) (*BQLogger, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	inserter := client.Dataset(bqDataset).Table(bqTable).Inserter()
	return &BQLogger{inserter: inserter}, nil
}

// LogAsync fires a goroutine to insert the row so the caller is never blocked.
// If the insert fails, it logs the error and moves on.
func (l *BQLogger) LogAsync(row inspectionRow) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := l.inserter.Put(ctx, row); err != nil {
			log.Printf("bigquery insert failed: %v", err)
		}
	}()
}

func newUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
