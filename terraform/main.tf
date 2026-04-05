terraform {
  required_version = ">= 1.5"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
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

# --------------------------------------------------------------------------
# Cloud Run: LiteLLM proxy
# --------------------------------------------------------------------------

resource "google_cloud_run_v2_service" "litellm" {
  name     = "litellm-proxy"
  location = var.region

  template {
    service_account = google_service_account.cloud_run.email

    scaling {
      min_instance_count = 0
      max_instance_count = 5
    }

    containers {
      image = "${local.registry_path}/litellm:${var.image_tag}"
      ports {
        container_port = 4000
      }

      env {
        name = "LITELLM_MASTER_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.litellm_master_key.secret_id
            version = "latest"
          }
        }
      }

    }
  }
}

resource "google_cloud_run_v2_service_iam_member" "litellm_public" {
  location = google_cloud_run_v2_service.litellm.location
  name     = google_cloud_run_v2_service.litellm.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# --------------------------------------------------------------------------
# Cloud Run: Go API server
# --------------------------------------------------------------------------

resource "google_cloud_run_v2_service" "api" {
  name     = "property-inspector-api"
  location = var.region

  template {
    service_account = google_service_account.cloud_run.email

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }

    containers {
      image = "${local.registry_path}/api:${var.image_tag}"
      ports {
        container_port = 8080
      }

      env {
        name  = "PORT"
        value = "8080"
      }

      env {
        name  = "LITELLM_URL"
        value = google_cloud_run_v2_service.litellm.uri
      }

      env {
        name  = "LITELLM_MODEL"
        value = "inspector"
      }

      env {
        name = "LITELLM_API_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.litellm_virtual_key.secret_id
            version = "latest"
          }
        }
      }
    }
  }

  depends_on = [google_cloud_run_v2_service.litellm]
}

resource "google_cloud_run_v2_service_iam_member" "api_public" {
  location = google_cloud_run_v2_service.api.location
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
