terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = "duo-streak-widget"
  region = "europe-west1"
}

resource "google_service_account" "function_sa" {
  account_id = "duostreak-badge-sa"
  display_name = "DuoSteak Badge Function Runtime"
}

resource "google_project_iam_member" "firestore_access" {
  project = "571260445375"
  role = "roles/datastore.user"
  member = "serviceAccount:${google_service_account.function_sa.email}"
}

data "archive_file" "function_source" {
  type = "zip"
  source_dir = "../src/cloud-function/" 
  output_path = "/tmp/function-source.zip"
}

# 5. CREATE A STORAGE BUCKET FOR THE ZIPPED CODE
# Terraform must upload the code zip to a bucket before creating the function.
resource "google_storage_bucket" "source_bucket" {
  name          = "duo-streak-widget-gcf-source" 
  location      = "EUROPE-WEST1"
  uniform_bucket_level_access = true
}

# 6. UPLOAD THE ZIPPED CODE TO THE BUCKET
resource "google_storage_bucket_object" "source_object" {
  name   = "source.zip" # The name of the file in the bucket
  bucket = google_storage_bucket.source_bucket.name
  source = data.archive_file.function_source.output_path # The zipped file from step 4
}

# 7. DEFINE THE CLOUD FUNCTION ITSELF
# This ties everything together
resource "google_cloudfunctions2_function" "badge_function" {
  name     = "generateBadge"
  location = "europe-west1"
  description = "Serves a dynamic Duolingo streak badge (Go + Firestore)"

  build_config {
    runtime     = "go121"         # Your language
    entry_point = "GenerateBadge" # Your Go function name
    source {
      storage_source {
        bucket = google_storage_bucket.source_bucket.name
        object = google_storage_bucket_object.source_object.name
      }
    }
  }

  service_config {
    max_instance_count  = 1
    available_memory    = "256Mi"
    timeout_seconds     = 30
    # Tell the function to run as the identity from Step 2
    service_account_email = google_service_account.function_sa.email
  }

  # Make sure this function doesn't get created until its permissions are ready
  depends_on = [
    google_project_iam_member.firestore_access
  ]
}

# 8. MAKE THE FUNCTION PUBLICLY VIEWABLE
# This is the final, critical piece: allow public access.
resource "google_cloudfunctions2_function_iam_member" "public_invoker" {
  project      = google_cloudfunctions2_function.badge_function.project
  location     = google_cloudfunctions2_function.badge_function.location
  cloud_function = google_cloudfunctions2_function.badge_function.name

  role   = "roles/cloudfunctions.invoker"
  member = "allUsers"
}

