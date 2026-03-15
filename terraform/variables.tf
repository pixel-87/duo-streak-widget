variable "billing_account_id" {
  type        = string
  description = "The GCP Billing Account ID"
}

variable "cloudflare_api_token" {
  type        = string
  sensitive   = true
}

variable "cloudflare_zone_id" {
  type        = string
}
