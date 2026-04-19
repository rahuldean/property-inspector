# Property Inspector Service
[![property-inspector-service](https://github.com/rahuldean/property-inspector/actions/workflows/ci.yml/badge.svg)](https://github.com/rahuldean/property-inspector/actions/workflows/ci.yml)

Analyzes property inspection photos using vision models via a LiteLLM proxy. Upload a room photo and get back a structured list of issues, or upload before/after photos to see what changed between inspections.

## Quick start (local)

```bash
cp .env.example .env   # set LITELLM_MASTER_KEY and your model API key
docker compose up --build
```

This starts a local LiteLLM proxy on port 4000 and the API server on port 6000.

## LiteLLM proxy

The API routes all LLM calls through a LiteLLM instance. For local dev, `docker compose` spins one up automatically. For production, you can run LiteLLM anywhere -- the API just needs `LITELLM_URL` pointed at it.

If your LiteLLM instance is behind a Cloudflare Access policy, set the service token credentials so the API can reach it:

```bash
LITELLM_URL=https://your-litellm-host
LITELLM_API_KEY=sk-virtual-...       # virtual key from LiteLLM admin UI
CF_ACCESS_CLIENT_ID=...              # CF-Access-Client-Id service token
CF_ACCESS_CLIENT_SECRET=...          # CF-Access-Client-Secret service token
```

If your LiteLLM instance is not behind Cloudflare Access, only `LITELLM_URL` and `LITELLM_API_KEY` are needed. `CF_ACCESS_CLIENT_ID` and `CF_ACCESS_CLIENT_SECRET` are ignored when empty.

Models and API keys are managed through the LiteLLM admin UI. Create a virtual key there and use it as `LITELLM_API_KEY`. Set `LITELLM_MODEL` to change which model alias the API uses (default: `inspector`).

## GCP deployment

### First-time setup

```bash
cd terraform
terraform init
terraform apply
```

`terraform.tfvars` only needs two variables:

```hcl
# terraform/terraform.tfvars  (gitignored)
project_id = "your-project-id"
region     = "us-east1"
```

Enable required APIs before running `terraform apply`:

```bash
gcloud services enable bigquery.googleapis.com --project=YOUR_PROJECT_ID
```

After apply, copy the output values into your GitHub repository secrets:

* `GCP_PROJECT_ID` - your project ID
* `GCP_REGION` - the region you deployed to
* `GCP_SERVICE_ACCOUNT` - value of `github_actions_service_account`
* `GCP_WORKLOAD_IDENTITY_PROVIDER` - value of `workload_identity_provider`

Populate the secrets in GCP Secret Manager:

```bash
# Virtual key from LiteLLM admin UI
echo -n "sk-virtual-..." | gcloud secrets versions add LITELLM_API_KEY --data-file=- --project=YOUR_PROJECT_ID

# Cloudflare Access service token (if your LiteLLM is behind CF Access)
echo -n "your-client-id" | gcloud secrets versions add CF_ACCESS_CLIENT_ID --data-file=- --project=YOUR_PROJECT_ID
echo -n "your-client-secret" | gcloud secrets versions add CF_ACCESS_CLIENT_SECRET --data-file=- --project=YOUR_PROJECT_ID
```

### Deploying

- **API**: triggers automatically on every push to `main`

Set `LITELLM_URL` in `.github/workflows/deploy-api.yml` to your LiteLLM instance URL before deploying.

## BigQuery logging

Every `/analyze` and `/compare` call is logged asynchronously to `property_inspector.inspections` in BigQuery. The write happens in a background goroutine after the response is sent -- it never blocks the API. If `GOOGLE_CLOUD_PROJECT` is unset (local dev), logging is skipped silently.

Schema: `id`, `room_name`, `floor_unit`, `endpoint`, `model_used`, `overall_condition`, `before_issue_count`, `after_issue_count`, `response_time_ms`, `error`, `inspected_at`. For `/compare`, both `before_issue_count` and `after_issue_count` are populated so change can be computed at query time (`after - before`).

## API

`property-inspector-api` requires Cloud Run authentication. Callers must include an identity token:

```bash
TOKEN=$(gcloud auth print-identity-token)
```

### `POST /analyze`

```bash
curl -X POST https://YOUR_API_URL/analyze \
  -H "Authorization: Bearer $TOKEN" \
  -F "image=@photo.jpg" \
  -F "room_name=Kitchen" \
  -F "floor_unit=Unit 4B"
```

Response: `RoomAnalysis` JSON with `issues[]`, `summary`, `overall_condition`, and `room_meta`.

### `POST /compare`

```bash
curl -X POST https://YOUR_API_URL/compare \
  -H "Authorization: Bearer $TOKEN" \
  -F "before=@move_in.jpg" \
  -F "after=@move_out.jpg" \
  -F "room_name=Living Room" \
  -F "floor_unit=Unit 4B"
```

Response: `ComparisonReport` JSON with `resolved_issues[]`, `new_issues[]`, `unchanged_issues[]`, and a `summary`.

### `GET /health`

Returns `{"status": "ok"}`.

## Using as a Go library

```go
import "github.com/rahuldean/property-inspector/inspector"

client := inspector.NewClient(
    inspector.WithBaseURL("https://your-litellm-host"),
    inspector.WithModel("inspector"),
    inspector.WithAPIKey("sk-virtual-..."),
    // Only needed if LiteLLM is behind Cloudflare Access:
    inspector.WithCFAccessClientID("..."),
    inspector.WithCFAccessClientSecret("..."),
)

result, err := client.AnalyzeRoom(ctx, "photo.jpg", inspector.RoomMeta{
    RoomName:  "Kitchen",
    FloorUnit: "Unit 4B",
})
```
