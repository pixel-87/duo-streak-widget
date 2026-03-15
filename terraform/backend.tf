 terraform {
  backend "gcs" {
    bucket = "duo-streak-widget-tf-state"
    prefix = "terraform/state"
  }
}
