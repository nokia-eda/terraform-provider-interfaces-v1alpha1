terraform {
  required_providers {
    interfaces = {
      source = "github.com/nokia-eda/interfaces-v1alpha1"
    }
  }
}

variable "base_url" {
  type        = string
  description = "The base URL of the EDA cluster that handles REST API calls"
  nullable    = false
}

variable "username" {
  type        = string
  description = "The EDA login username"
  nullable    = false
  default     = "admin"
}

variable "password" {
  type        = string
  description = "The EDA login password"
  nullable    = false
  default     = "admin"
}

variable "client_secret" {
  type        = string
  description = "The EDA client secret"
}

# Provider configuration
provider "interfaces" {
  base_url          = var.base_url      # EDA_BASE_URL
  eda_username      = var.username      # EDA_USERNAME
  eda_password      = var.password      # EDA_PASSWORD
  eda_client_secret = var.client_secret # EDA_CLIENT_SECRET
  tls_skip_verify   = true              # TLS_SKIP_VERIFY
}
