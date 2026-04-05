# Property Inspector Service
[![property-inspector-service](https://github.com/rahuldean/property-inspector/actions/workflows/ci.yml/badge.svg)](https://github.com/rahuldean/property-inspector/actions/workflows/ci.yml)

Analyzes property inspection photos using vision models via a LiteLLM proxy. Upload a room photo and get back a structured list of issues, or upload before/after photos to see what changed between inspections.

## Quick start (local)

```bash
cp .env.example .env   # set LITELLM_MASTER_KEY and your model API key
docker compose up --build
```

This starts the LiteLLM proxy on port 4000 and the API server on port 6000.

## GCP deployment

### First-time setup

```bash
cd terraform
terraform init
terraform apply -var="image_tag=latest"
```
Note: You can also pass the above variables using a terraform.tfvars file.

After apply, copy the output values into your GitHub repository secrets:


* `GCP_PROJECT_ID` - your project ID
* `GCP_REGION` - the region you deployed to
* `GCP_SERVICE_ACCOUNT` - value of `github_actions_service_account`
* `GCP_WORKLOAD_IDENTITY_PROVIDER` - value of `workload_identity_provider`

Seed the required secrets in Secret Manager:

```bash
echo -n "sk-litellm-master-..." | gcloud secrets versions add LITELLM_MASTER_KEY --data-file=-
```

After first deploy, open the LiteLLM admin UI, add your model API keys (Anthropic, etc.) there, then create a virtual key for the API service and store it:

```bash
echo -n "sk-virtual-..." | gcloud secrets versions add LITELLM_VIRTUAL_KEY --data-file=-
```

Then redeploy (or update the Cloud Run service) so the API picks up the new secret version.

### Deploying

Push to `main` to trigger the deploy workflow. The workflow builds both images, pushes them to Artifact Registry, and deploys them to Cloud Run.

## API

### `POST /analyze`

```bash
curl -X POST https://YOUR_DEMO_URL/analyze \
  -F "image=@photo.jpg" \
  -F "room_name=Kitchen" \
  -F "floor_unit=Unit 4B"
```

Response: `RoomAnalysis` JSON with `issues[]`, `summary`, `overall_condition`, and `room_meta`.

### `POST /compare`

```bash
curl -X POST https://YOUR_DEMO_URL/compare \
  -F "before=@move_in.jpg" \
  -F "after=@move_out.jpg" \
  -F "room_name=Living Room" \
  -F "floor_unit=Unit 4B"
```

Response: `ComparisonReport` JSON with `resolved_issues[]`, `new_issues[]`, `unchanged_issues[]`, and a `summary`.

### `GET /health`

Returns `{"status": "ok"}`.

## Switching models

Edit `litellm/config.yaml` and add the corresponding API key to Secret Manager. The current config has two routes:

- `inspector` - Claude 3.5 Sonnet (default)
- `inspector-gemini` - Gemini 1.5 Pro

Set the `LITELLM_MODEL` env var on the API Cloud Run service to switch routes.

## Using as a Go library

```go
import "github.com/rahuldean/property-inspector/inspector"

client := inspector.NewClient(
    inspector.WithBaseURL("http://localhost:4000"),
    inspector.WithModel("inspector"),
)

result, err := client.AnalyzeRoom(ctx, "photo.jpg", inspector.RoomMeta{
    RoomName:  "Kitchen",
    FloorUnit: "Unit 4B",
})
```
