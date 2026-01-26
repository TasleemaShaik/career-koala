output "gke_cluster_name" {
  value = google_container_cluster.primary.name
}

output "gke_location" {
  value = google_container_cluster.primary.location
}

output "gke_endpoint" {
  value = google_container_cluster.primary.endpoint
}

output "gke_ca_certificate" {
  value     = google_container_cluster.primary.master_auth[0].cluster_ca_certificate
  sensitive = true
}

output "cloudsql_instance_name" {
  value = try(google_sql_database_instance.postgres[0].name, "")
}

output "cloudsql_private_ip" {
  value = try(google_sql_database_instance.postgres[0].private_ip_address, "")
}

output "cloudsql_connection_name" {
  value = try(google_sql_database_instance.postgres[0].connection_name, "")
}

output "cloudsql_database" {
  value = try(google_sql_database.app[0].name, "")
}

output "database_url" {
  value = var.enable_cloudsql ? format(
    "postgres://%s:%s@%s:5432/%s?sslmode=disable",
    var.cloudsql_user,
    local.cloudsql_effective_password,
    try(google_sql_database_instance.postgres[0].private_ip_address, ""),
    var.cloudsql_db_name
  ) : ""
  sensitive = true
}

output "cloudsql_password" {
  value     = local.cloudsql_effective_password
  sensitive = true
}
