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
  description = "GCP project number -- find it on the GCP Console home page under your project name"
  type        = string
}
