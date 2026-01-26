provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

locals {
  network_name    = "career-koala-vpc"
  subnetwork_name = "career-koala-subnet"
  subnet_cidr     = "10.10.0.0/16"
}

resource "random_password" "cloudsql" {
  count   = var.enable_cloudsql && var.cloudsql_password == "" ? 1 : 0
  length  = 20
  special = true
}

locals {
  cloudsql_effective_password = var.enable_cloudsql ? (var.cloudsql_password != "" ? var.cloudsql_password : random_password.cloudsql[0].result) : ""
}

resource "google_compute_network" "primary" {
  name                    = local.network_name
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "primary" {
  name                     = local.subnetwork_name
  region                   = var.region
  network                  = google_compute_network.primary.id
  ip_cidr_range            = local.subnet_cidr
  private_ip_google_access = true
}

resource "google_project_service" "container" {
  count              = var.enable_services ? 1 : 0
  service            = "container.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "sqladmin" {
  count              = var.enable_services ? 1 : 0
  service            = "sqladmin.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "servicenetworking" {
  count              = var.enable_services ? 1 : 0
  service            = "servicenetworking.googleapis.com"
  disable_on_destroy = false
}

resource "google_container_cluster" "primary" {
  name     = var.gke_name
  location = var.region

  network    = google_compute_network.primary.id
  subnetwork = google_compute_subnetwork.primary.id

  remove_default_node_pool = true
  initial_node_count       = 1

  ip_allocation_policy {}

  deletion_protection = var.gke_deletion_protection

  node_locations = length(var.node_locations) > 0 ? var.node_locations : null

  depends_on = [google_project_service.container]
}

resource "google_container_node_pool" "primary" {
  name       = "primary"
  location   = google_container_cluster.primary.location
  cluster    = google_container_cluster.primary.name
  node_count = var.node_count

  node_config {
    machine_type = var.node_machine_type
    disk_size_gb = var.node_disk_size_gb
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]
  }
}

resource "google_compute_global_address" "private_service_range" {
  count         = var.enable_cloudsql ? 1 : 0
  name          = "${var.gke_name}-psa"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.primary.id

  depends_on = [google_project_service.servicenetworking]
}

resource "google_service_networking_connection" "private_vpc_connection" {
  count                   = var.enable_cloudsql ? 1 : 0
  network                 = google_compute_network.primary.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_service_range[0].name]

  depends_on = [google_project_service.servicenetworking]
}

resource "google_sql_database_instance" "postgres" {
  count            = var.enable_cloudsql ? 1 : 0
  name             = var.cloudsql_instance_name
  region           = var.region
  database_version = "POSTGRES_16"

  settings {
    tier            = var.cloudsql_tier
    disk_size       = var.cloudsql_disk_size_gb
    availability_type = var.cloudsql_availability_type

    ip_configuration {
      ipv4_enabled    = false
      private_network = google_compute_network.primary.id
    }

    backup_configuration {
      enabled = true
    }
  }

  deletion_protection = var.cloudsql_deletion_protection

  depends_on = [
    google_service_networking_connection.private_vpc_connection,
    google_project_service.sqladmin,
  ]
}

resource "google_sql_database" "app" {
  count    = var.enable_cloudsql ? 1 : 0
  name     = var.cloudsql_db_name
  instance = google_sql_database_instance.postgres[0].name
}

resource "google_sql_user" "app" {
  count    = var.enable_cloudsql ? 1 : 0
  name     = var.cloudsql_user
  instance = google_sql_database_instance.postgres[0].name
  password = local.cloudsql_effective_password
}
