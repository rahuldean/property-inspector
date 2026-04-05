output "cloud_run_url" {
  description = "Public URL of the Go API Cloud Run service"
  value       = google_cloud_run_v2_service.api.uri
}

output "litellm_url" {
  description = "Public URL of the LiteLLM proxy Cloud Run service"
  value       = google_cloud_run_v2_service.litellm.uri
}

output "artifact_registry_url" {
  description = "Artifact Registry repository URL"
  value       = "${local.registry_host}/${var.project_id}/property-inspector"
}

output "github_actions_service_account" {
  description = "Service account email to set as GCP_SERVICE_ACCOUNT in GitHub secrets"
  value       = google_service_account.github_actions.email
}

output "workload_identity_provider" {
  description = "Workload Identity Provider resource name to set as GCP_WORKLOAD_IDENTITY_PROVIDER in GitHub secrets"
  value       = google_iam_workload_identity_pool_provider.github.name
}
