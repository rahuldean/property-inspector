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

variable "project_number" {
  description = "GCP project number (find on GCP Console home page) -- used to construct the Cloud Run service agent email for VPC egress IAM"
  type        = string
}
