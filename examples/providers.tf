terraform {
  required_providers {
    interfaces-v1alpha1 = {
      source = "nokia-eda/interfaces-v1alpha1"
      # version = "0.1.0" # Uncomment to specify provider version
    }
  }
}

# Provider configuration
provider "interfaces-v1alpha1" {
  base_url        = ""      # Example: https://eda.mydomain.com:9443 Env var: BASE_URL
  username        = "admin" # Env var: USERNAME
  password        = "admin" # Env var: PASSWORD
  tls_skip_verify = true    # Env var: TLS_SKIP_VERIFY

  # Client secret will be fetched automatically from Keycloak if not set
  # using keycloak credentials
  keycloak_admin_username = "admin" # Env var: KEYCLOAK_ADMIN_USERNAME
  keycloak_admin_password = "admin" # Env var: KEYCLOAK_ADMIN_PASSWORD
  client_secret           = ""      # Env var: CLIENT_SECRET
}
