terraform {
  required_version = ">= 1.5"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

locals {
  registry_host = "${var.region}-docker.pkg.dev"
  registry_path = "${var.region}-docker.pkg.dev/${var.project_id}/property-inspector"
}

# --------------------------------------------------------------------------
# Artifact Registry
# --------------------------------------------------------------------------

resource "google_artifact_registry_repository" "property_inspector" {
  repository_id = "property-inspector"
  format        = "DOCKER"
  location      = var.region
  description   = "Docker images for the property inspector service"
}

# --------------------------------------------------------------------------
# Secret Manager
# --------------------------------------------------------------------------

resource "google_secret_manager_secret" "litellm_master_key" {
  secret_id = "LITELLM_MASTER_KEY"
  replication {
    auto {}
  }
}

# Virtual key created via the LiteLLM admin UI and stored here for the API to use
resource "google_secret_manager_secret" "litellm_virtual_key" {
  secret_id = "LITELLM_VIRTUAL_KEY"
  replication {
    auto {}
  }
}

# Secret for the full DATABASE_URL -- built from the generated password below
resource "google_secret_manager_secret" "litellm_db_url" {
  secret_id = "LITELLM_DATABASE_URL"
  replication {
    auto {}
  }
}

# --------------------------------------------------------------------------
# Cloud SQL (Postgres) for LiteLLM
# --------------------------------------------------------------------------

resource "random_password" "litellm_db" {
  length  = 32
  special = false
}

resource "google_sql_database_instance" "litellm" {
  name             = "litellm-postgres"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier = "db-f1-micro"

    backup_configuration {
      enabled = true
    }
  }

  deletion_protection = false
}

resource "google_sql_database" "litellm" {
  name     = "litellm"
  instance = google_sql_database_instance.litellm.name
}

resource "google_sql_user" "litellm" {
  name     = "litellm"
  instance = google_sql_database_instance.litellm.name
  password = random_password.litellm_db.result
}

resource "google_secret_manager_secret_version" "litellm_db_url" {
  secret      = google_secret_manager_secret.litellm_db_url.id
  secret_data = "postgresql://litellm:${random_password.litellm_db.result}@localhost/litellm?host=/cloudsql/${google_sql_database_instance.litellm.connection_name}"
}

# --------------------------------------------------------------------------
# Service account for Cloud Run workloads
# --------------------------------------------------------------------------

resource "google_service_account" "cloud_run" {
  account_id   = "property-inspector-run"
  display_name = "Property Inspector Cloud Run SA"
}

resource "google_secret_manager_secret_iam_member" "run_litellm_key" {
  secret_id = google_secret_manager_secret.litellm_master_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run.email}"
}

resource "google_secret_manager_secret_iam_member" "run_litellm_virtual_key" {
  secret_id = google_secret_manager_secret.litellm_virtual_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run.email}"
}

resource "google_secret_manager_secret_iam_member" "run_db_url" {
  secret_id = google_secret_manager_secret.litellm_db_url.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Token used to gate access to the Next.js inspector app (URL token auth)
resource "google_secret_manager_secret" "inspector_app_token" {
  secret_id = "INSPECTOR_APP_TOKEN"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_iam_member" "run_app_token" {
  secret_id = google_secret_manager_secret.inspector_app_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Direct VPC Egress requires compute.networkUser on both the workload SA and the Cloud Run service agent
resource "google_project_iam_member" "run_network_user" {
  project = var.project_id
  role    = "roles/compute.networkUser"
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
}

resource "google_project_iam_member" "cloudrun_agent_network_user" {
  project = var.project_id
  role    = "roles/compute.networkUser"
  member  = "serviceAccount:service-${var.project_number}@serverless-robot-prod.iam.gserviceaccount.com"
}

# Allow the Cloud Run SA to invoke other Cloud Run services (e.g. inspector-app calling the Go API)
resource "google_project_iam_member" "run_invoker" {
  project = var.project_id
  role    = "roles/run.invoker"
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
}

resource "google_project_iam_member" "run_cloudsql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
}

# --------------------------------------------------------------------------
# BigQuery for inspection logging
# --------------------------------------------------------------------------

resource "google_bigquery_dataset" "property_inspector" {
  dataset_id  = "property_inspector"
  location    = "us-east1"
  description = "Inspection results logged by the property inspector API"
}

resource "google_bigquery_table" "inspections" {
  dataset_id          = google_bigquery_dataset.property_inspector.dataset_id
  table_id            = "inspections"
  deletion_protection = false

  schema = jsonencode([
    { name = "id",                 type = "STRING",    mode = "REQUIRED" },
    { name = "room_name",          type = "STRING",    mode = "NULLABLE" },
    { name = "floor_unit",         type = "STRING",    mode = "NULLABLE" },
    { name = "endpoint",           type = "STRING",    mode = "REQUIRED" },
    { name = "model_used",         type = "STRING",    mode = "NULLABLE" },
    { name = "overall_condition",  type = "STRING",    mode = "NULLABLE" },
    { name = "before_issue_count", type = "INTEGER",   mode = "NULLABLE" },
    { name = "after_issue_count",  type = "INTEGER",   mode = "NULLABLE" },
    { name = "response_time_ms",   type = "INTEGER",   mode = "REQUIRED" },
    { name = "error",              type = "BOOLEAN",   mode = "REQUIRED" },
    { name = "inspected_at",       type = "TIMESTAMP", mode = "REQUIRED" },
  ])
}

# Write rows to this dataset only, not all BQ tables in the project
resource "google_bigquery_dataset_iam_member" "run_bq_writer" {
  dataset_id = google_bigquery_dataset.property_inspector.dataset_id
  role       = "roles/bigquery.dataEditor"
  member     = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Required to execute insert jobs
resource "google_project_iam_member" "run_bq_job_user" {
  project = var.project_id
  role    = "roles/bigquery.jobUser"
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
}

# --------------------------------------------------------------------------
# Workload Identity Federation for GitHub Actions
# --------------------------------------------------------------------------

resource "google_iam_workload_identity_pool" "github" {
  workload_identity_pool_id = "github-actions"
  display_name              = "GitHub Actions"
}

resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github-provider"
  display_name                       = "GitHub OIDC provider"

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }

  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.actor"      = "assertion.actor"
    "attribute.repository" = "assertion.repository"
  }

  attribute_condition = "assertion.repository == 'rahuldean/property-inspector'"
}

# Service account used by GitHub Actions to push images and deploy
resource "google_service_account" "github_actions" {
  account_id   = "property-inspector-ci"
  display_name = "Property Inspector GitHub Actions SA"
}

resource "google_service_account_iam_member" "github_wif_binding" {
  service_account_id = google_service_account.github_actions.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/rahuldean/property-inspector"
}

resource "google_project_iam_member" "github_artifact_writer" {
  project = var.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${google_service_account.github_actions.email}"
}

resource "google_project_iam_member" "github_run_developer" {
  project = var.project_id
  role    = "roles/run.developer"
  member  = "serviceAccount:${google_service_account.github_actions.email}"
}

# GitHub Actions needs to be able to act as the Cloud Run SA when deploying
resource "google_service_account_iam_member" "github_impersonate_run_sa" {
  service_account_id = google_service_account.cloud_run.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${google_service_account.github_actions.email}"
}


