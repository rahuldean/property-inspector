variable "project_id" {
  description = "GCP project ID"
  type        = string
  default     = "YOUR_GCP_PROJECT_ID"
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "image_tag" {
  description = "Docker image tag to deploy (e.g. git SHA)"
  type        = string
}
