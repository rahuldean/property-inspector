package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/labmox/ai-property-inspector-service/inspector"
)

func main() {
	analyzeCmd := flag.NewFlagSet("analyze", flag.ExitOnError)
	analyzeImage := analyzeCmd.String("image", "", "path to room image")
	analyzeRoom := analyzeCmd.String("room", "Unknown Room", "room name")
	analyzeUnit := analyzeCmd.String("unit", "", "floor/unit identifier")

	compareCmd := flag.NewFlagSet("compare", flag.ExitOnError)
	compareBefore := compareCmd.String("before", "", "path to before image")
	compareAfter := compareCmd.String("after", "", "path to after image")
	compareRoom := compareCmd.String("room", "Unknown Room", "room name")
	compareUnit := compareCmd.String("unit", "", "floor/unit identifier")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: inspect <analyze|compare> [flags]\n")
		os.Exit(1)
	}

	baseURL := os.Getenv("LITELLM_URL")
	if baseURL == "" {
		baseURL = "http://localhost:4000"
	}

	client := inspector.NewClient(
		inspector.WithBaseURL(baseURL),
		inspector.WithModel(envOr("LITELLM_MODEL", "inspector")),
		inspector.WithAPIKey(os.Getenv("LITELLM_API_KEY")),
	)

	ctx := context.Background()

	switch os.Args[1] {
	case "analyze":
		analyzeCmd.Parse(os.Args[2:])
		if *analyzeImage == "" {
			fmt.Fprintf(os.Stderr, "error: --image is required\n")
			os.Exit(1)
		}
		meta := inspector.RoomMeta{RoomName: *analyzeRoom, FloorUnit: *analyzeUnit, InspectedAt: time.Now()}
		result, err := client.AnalyzeRoom(ctx, *analyzeImage, meta)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		printJSON(result)

	case "compare":
		compareCmd.Parse(os.Args[2:])
		if *compareBefore == "" || *compareAfter == "" {
			fmt.Fprintf(os.Stderr, "error: --before and --after are required\n")
			os.Exit(1)
		}
		meta := inspector.RoomMeta{RoomName: *compareRoom, FloorUnit: *compareUnit, InspectedAt: time.Now()}
		result, err := client.CompareInspections(ctx, *compareBefore, *compareAfter, meta)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		printJSON(result)

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\nusage: inspect <analyze|compare> [flags]\n", os.Args[1])
		os.Exit(1)
	}
}

func printJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
