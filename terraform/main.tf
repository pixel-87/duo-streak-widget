terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.51.0"
    }
  }
}

provider "google" {
  project = "duo-streak-widget"
  region  = "us-central1"
}

resource "google_cloud_run_v2_service" "default" {
  name     = "duo-streak-widget"
  location = "us-central1"
  ingress  = "INGRESS_TRAFFIC_ALL"
  deletion_protection = false

  template {
    containers {
      image = "gcr.io/duo-streak-widget/duo-streak-widget:latest"
      ports {
        container_port = 8080
      }
    }
    scaling {
      min_instance_count = 0
      max_instance_count = 2
    }
  }
}

resource "google_cloud_run_v2_service_iam_member" "noauth" {
  location = google_cloud_run_v2_service.default.location
  name     = google_cloud_run_v2_service.default.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
