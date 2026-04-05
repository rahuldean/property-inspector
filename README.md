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
terraform apply
```

Note: Use a `terraform.tfvars` file to set `project_id` and `region`.

After apply, copy the output values into your GitHub repository secrets:

* `GCP_PROJECT_ID` - your project ID
* `GCP_REGION` - the region you deployed to
* `GCP_SERVICE_ACCOUNT` - value of `github_actions_service_account`
* `GCP_WORKLOAD_IDENTITY_PROVIDER` - value of `workload_identity_provider`

Seed the master key in Secret Manager:

```bash
echo -n "sk-litellm-master-..." | gcloud secrets versions add LITELLM_MASTER_KEY --data-file=-
```

### Deploying

- **API**: triggers automatically on every push to `main`
- **LiteLLM proxy**: triggers on push to `main` when `litellm/**` changes, or manually via the `deploy-litellm` workflow in GitHub Actions

Deploy LiteLLM first on initial setup.

### Post-deploy: add models and create a virtual key

Access the LiteLLM admin UI via the Cloud Run proxy:

```bash
gcloud run services proxy litellm-proxy --region=REGION --project=PROJECT_ID
```

Open `http://localhost:8080/ui`, log in with your master key, add your model API keys and configure models, then create a virtual key for the API service. Store it:

```bash
echo -n "sk-virtual-..." | gcloud secrets versions add LITELLM_VIRTUAL_KEY --data-file=-
```

Then trigger the `deploy-api` workflow to pick up the new secret.

## API

Both services require Cloud Run authentication. Callers must include an identity token:

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

## Switching models

Models and API keys are managed through the LiteLLM admin UI (changes persist in Cloud SQL). Set `LITELLM_MODEL` on the API Cloud Run service to switch which model alias it uses.

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

## FAQ
1. [LiteLLM Bug]: There is an existing [issue](https://github.com/BerriAI/litellm/issues/23741#issuecomment-4122638733) that throws `vector_store_ids: Extra inputs are not permitted`. Follow the issue for the proposed work around.