package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/rahuldean/property-inspector/inspector"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	litellmURL := os.Getenv("LITELLM_URL")
	if litellmURL == "" {
		litellmURL = "http://localhost:4000"
	}

	model := os.Getenv("LITELLM_MODEL")
	if model == "" {
		model = "inspector"
	}

	maxUploadMB := 32
	if v := os.Getenv("MAX_UPLOAD_MB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxUploadMB = n
		}
	}

	client := inspector.NewClient(
		inspector.WithBaseURL(litellmURL),
		inspector.WithModel(model),
		inspector.WithAPIKey(os.Getenv("LITELLM_API_KEY")),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /analyze", handleAnalyze(client, int64(maxUploadMB)))
	mux.HandleFunc("POST /compare", handleCompare(client, int64(maxUploadMB)))

	log.Printf("starting server on :%s (litellm: %s, model: %s)", port, litellmURL, model)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleAnalyze(client *inspector.Client, maxUploadMB int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(maxUploadMB << 20); err != nil {
			httpError(w, "failed to parse form", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil {
			httpError(w, "missing 'image' field", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Write to a temp file so the inspector library can read it.
		tmpFile, err := os.CreateTemp("", "inspect-*-"+header.Filename)
		if err != nil {
			httpError(w, "failed to create temp file", http.StatusInternalServerError)
			return
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		if _, err := tmpFile.ReadFrom(file); err != nil {
			httpError(w, "failed to save upload", http.StatusInternalServerError)
			return
		}
		tmpFile.Close()

		meta := inspector.RoomMeta{
			RoomName:    r.FormValue("room_name"),
			FloorUnit:   r.FormValue("floor_unit"),
			InspectedAt: time.Now(),
		}
		if meta.RoomName == "" {
			meta.RoomName = "Unknown Room"
		}

		result, err := client.AnalyzeRoom(r.Context(), tmpFile.Name(), meta)
		if err != nil {
			log.Printf("analyze error: %v", err)
			httpError(w, fmt.Sprintf("analysis failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func handleCompare(client *inspector.Client, maxUploadMB int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(maxUploadMB * 2 << 20); err != nil {
			httpError(w, "failed to parse form", http.StatusBadRequest)
			return
		}

		beforeFile, beforeHeader, err := r.FormFile("before")
		if err != nil {
			httpError(w, "missing 'before' field", http.StatusBadRequest)
			return
		}
		defer beforeFile.Close()

		afterFile, afterHeader, err := r.FormFile("after")
		if err != nil {
			httpError(w, "missing 'after' field", http.StatusBadRequest)
			return
		}
		defer afterFile.Close()

		// Save both to temp files
		beforeTmp, err := saveTempFile(beforeFile, beforeHeader.Filename)
		if err != nil {
			httpError(w, "failed to save before image", http.StatusInternalServerError)
			return
		}
		defer os.Remove(beforeTmp)

		afterTmp, err := saveTempFile(afterFile, afterHeader.Filename)
		if err != nil {
			httpError(w, "failed to save after image", http.StatusInternalServerError)
			return
		}
		defer os.Remove(afterTmp)

		meta := inspector.RoomMeta{
			RoomName:    r.FormValue("room_name"),
			FloorUnit:   r.FormValue("floor_unit"),
			InspectedAt: time.Now(),
		}
		if meta.RoomName == "" {
			meta.RoomName = "Unknown Room"
		}

		result, err := client.CompareInspections(r.Context(), beforeTmp, afterTmp, meta)
		if err != nil {
			log.Printf("compare error: %v", err)
			httpError(w, fmt.Sprintf("comparison failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func saveTempFile(src io.Reader, name string) (string, error) {
	tmp, err := os.CreateTemp("", "inspect-*-"+name)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(tmp, src); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	tmp.Close()
	return tmp.Name(), nil
}

func httpError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
