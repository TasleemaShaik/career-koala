variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  description = "GCP region"
  default     = "us-central1"
}

variable "zone" {
  type        = string
  description = "GCP zone"
  default     = "us-central1-a"
}

variable "gke_name" {
  type        = string
  description = "GKE cluster name"
  default     = "career-koala-gke"
}

variable "node_machine_type" {
  type        = string
  description = "GKE node machine type"
  default     = "e2-standard-4"
}

variable "node_count" {
  type        = number
  description = "Number of GKE nodes"
  default     = 2
}

variable "node_disk_size_gb" {
  type        = number
  description = "Node disk size (GB)"
  default     = 50
}

variable "node_locations" {
  type        = list(string)
  description = "Optional node locations for regional cluster"
  default     = []
}

variable "gke_deletion_protection" {
  type        = bool
  description = "Protect GKE cluster from deletion"
  default     = false
}

variable "enable_services" {
  type        = bool
  description = "Enable required Google APIs"
  default     = true
}

variable "enable_cloudsql" {
  type        = bool
  description = "Create Cloud SQL instance"
  default     = true
}

variable "cloudsql_instance_name" {
  type        = string
  description = "Cloud SQL instance name"
  default     = "career-koala-postgres"
}

variable "cloudsql_tier" {
  type        = string
  description = "Cloud SQL tier"
  default     = "db-custom-2-4096"
}

variable "cloudsql_disk_size_gb" {
  type        = number
  description = "Cloud SQL disk size (GB)"
  default     = 20
}

variable "cloudsql_availability_type" {
  type        = string
  description = "Cloud SQL availability type (ZONAL or REGIONAL)"
  default     = "ZONAL"
}

variable "cloudsql_deletion_protection" {
  type        = bool
  description = "Protect Cloud SQL instance from deletion"
  default     = false
}

variable "cloudsql_db_name" {
  type        = string
  description = "Cloud SQL database name"
  default     = "career_koala"
}

variable "cloudsql_user" {
  type        = string
  description = "Cloud SQL user"
  default     = "postgres"
}

variable "cloudsql_password" {
  type        = string
  description = "Cloud SQL user password (leave empty to auto-generate)"
  sensitive   = true
  default     = ""

  validation {
    condition     = var.enable_cloudsql ? (length(var.cloudsql_password) == 0 || length(var.cloudsql_password) >= 8) : true
    error_message = "cloudsql_password must be at least 8 characters when set."
  }
}
