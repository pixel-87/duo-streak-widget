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
      resources {
        limits = {
          cpu = "1000m" # 1 vCPU
          memory = "128Mi" # smallest mem
        }
        cpu_idle = true
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

resource "google_storage_bucket" "tf_state" {
   name = "duo-streak-widget-tf-state"
   location = "us-central1"

   # Force uniform access 
   uniform_bucket_level_access = true

   versioning {
     enabled = true
   }

   lifecycle_rule {
     action {
       type = "Delete"
     }
     condition {
       num_newer_versions = 10
     }
   }

   # Prevent tf from destroying state bucket
   lifecycle {
     prevent_destroy = true
   }
}


# Google Cloud Budget Alert
resource "google_billing_budget" "budget" {
  billing_account = var.billing_account_id
  display_name = "Duo Widget $1 Budget"

  budget_filter {
    projects = ["projects/duo-streak-widget"]
  }

  amount {
     specified_amount {
      currency_code = "USD"
      units = "1"
    }
  }

  threshold_rules {
    threshold_percent = 1.0
    spend_basis = "FORECASTED_SPEND"
  }
}

# Cloudflare Rate Limiting
resource "cloudflare_ruleset" "api_limit" {
  zone_id = var.cloudflare_zone_id
  name = "Prevent api abuse"
  kind = "zone"
  phase = "http_ratelimit"

  rules {
    action = "block"
    ratelimit {
      characteristics = ["ip.src", "cf.colo.id"]
      period = 10
      requests_per_period = 4 # 4 badges/min per person
      mitigation_timeout = 10 # 10 sec timeout if exceeding requests
    }
    expression = "(http.request.uri.path contains \"/api/\")"
  }
}
