terraform {
  backend "gcs" {
    bucket = "duo-streak-widget-tf-state"
    prefix= "terraform/state"
  }

  required_providers {
    google = {
      source = "hashicorp/google"
      version = ">= 4.51.0"
    }
    cloudflare = {
      source = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }
}

provider "google" {
  project = "duo-streak-widget"
  region = "us-central1"
  billing_project = "duo-streak-widget"
  user_project_override = true
}

provider "cloudflare" {
   api_token = var.cloudflare_api_token
}
