terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = ">= 4.51.0"
    }
  }
}

provider "google" {
  project = "duo-streak-widget"
  region = "us-central1"
}
