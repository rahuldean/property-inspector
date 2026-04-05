# Property Inspector Service
[![property-inspector-service](https://github.com/rahuldean/property-inspector/actions/workflows/ci.yml/badge.svg)](https://github.com/rahuldean/property-inspector/actions/workflows/ci.yml)

Analyzes property inspection photos using vision models via a LiteLLM proxy. Upload a room photo and get back a structured list of issues. Upload before/after photos to see what got fixed and what didn't.

## How it works

The server accepts image uploads, base64-encodes them, and sends them to LiteLLM using the OpenAI chat completions format. LiteLLM routes to Claude (or whatever model you configure). The model returns structured JSON describing issues found in the room.

## Setup

You need Docker and an Anthropic (for now) API key.

```bash
cp .env.example .env
# Put your actual key in .env
```

Start everything:

```bash
docker compose up --build
```

This brings up:
- **LiteLLM proxy** on port 4000 (routes to Claude)
- **Go API server** on port 8080

## API

### `POST /analyze`

Upload a single room photo. Returns a list of issues found.

```bash
curl -X POST http://localhost:8080/analyze \
  -F "image=@photo.jpg" \
  -F "room_name=Kitchen" \
  -F "floor_unit=Unit 4B"
```

Response:

```json
{
  "room_meta": {
    "room_name": "Kitchen",
    "floor_unit": "Unit 4B",
    "inspected_at": "2025-12-15T10:30:00Z"
  },
  "issues": [
    {
      "category": "Wall Damage",
      "severity": "moderate",
      "description": "Scuff marks and small dent near the doorframe",
      "location": "east wall by entrance",
      "confidence": 0.85
    }
  ],
  "summary": "Kitchen is in fair condition with minor wall damage near the entrance.",
  "overall_condition": "fair",
  "generated_at": "2025-12-15T10:47:00Z"
}
```

### `POST /compare`

Upload before and after photos of the same room. Returns what changed.

```bash
curl -X POST http://localhost:8080/compare \
  -F "before=@move_in.jpg" \
  -F "after=@move_out.jpg" \
  -F "room_name=Living Room" \
  -F "floor_unit=Unit 4B"
```

Response includes `resolved_issues`, `new_issues`, and `unchanged_issues` arrays plus a summary.

### `GET /health`

Returns `{"status": "ok"}`.

## CLI

There's also a CLI if you want to skip the HTTP server and talk to LiteLLM directly:

```bash
go run ./cmd/inspect analyze --image photo.jpg --room "Kitchen" --unit "2A"
go run ./cmd/inspect compare --before move_in.jpg --after move_out.jpg --room "Kitchen"
```

Set `LITELLM_URL` if the proxy isn't on localhost:4000.

## Switching models

Edit `litellm/config.yaml` to point at a different model. For example, to use GPT-4o:

```yaml
model_list:
  - model_name: inspector
    litellm_params:
      model: openai/gpt-4o
      api_key: os.environ/OPENAI_API_KEY
```

Then set `OPENAI_API_KEY` in your `.env` and restart.

## Using as a Go library

```go
client := inspector.NewClient(
    inspector.WithBaseURL("http://localhost:4000"),
    inspector.WithModel("inspector"),
)

result, err := client.AnalyzeRoom(ctx, "photo.jpg", inspector.RoomMeta{
    RoomName:  "Kitchen",
    FloorUnit: "Unit 4B",
})
```
