terraform {
  required_providers {
    oai = {
      source = "benemon/oai"
    }
  }
}

provider "oai" {
  # Configuration options
  endpoint = "https://api.openshift.com/api/assisted-install"
  token    = var.oai_token  # Set via environment variable or terraform.tfvars
  timeout  = "30s"
}

variable "oai_token" {
  description = "OpenShift Assisted Service API token"
  type        = string
  sensitive   = true
}
