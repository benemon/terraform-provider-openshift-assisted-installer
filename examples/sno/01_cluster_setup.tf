# ==============================================================================
# Module 1: Cluster Setup
# Creates the cluster definition and infrastructure environment
# Run this first to prepare for host discovery
# ==============================================================================

terraform {
  required_providers {
    openshift_assisted_installer = {
      source  = "benemon/openshift-assisted-installer"
      version = "~> 0.1"
    }
  }
}

provider "openshift_assisted_installer" {
  # Uses OFFLINE_TOKEN environment variable
}

# Variables
variable "pull_secret" {
  description = "Red Hat pull secret"
  type        = string
  sensitive   = true
}

variable "ssh_public_key" {
  description = "SSH public key for cluster access"
  type        = string
}

# Get latest OpenShift version
data "openshift_assisted_installer_versions" "latest" {
  only_latest = true
}

locals {
  cluster_name = "sno-cluster"
  base_domain  = "example.com"
  
  openshift_version = [for v in data.openshift_assisted_installer_versions.latest.versions : v.version 
    if v.support_level == "production"][0]
}

# Create cluster definition
resource "openshift_assisted_installer_cluster" "sno" {
  name              = local.cluster_name
  base_dns_domain   = local.base_domain
  openshift_version = local.openshift_version
  
  # Single node configuration
  control_plane_count  = 1
  schedulable_masters  = true
  
  # Network configuration
  cluster_network_cidr = "10.128.0.0/14"
  service_network_cidr = "172.30.0.0/16"
  
  pull_secret    = var.pull_secret
  ssh_public_key = var.ssh_public_key
  
  # Note: No trigger_installation field anymore
  # Installation is handled by separate resource
}

# Create infrastructure environment for host discovery
resource "openshift_assisted_installer_infra_env" "sno" {
  name              = "${local.cluster_name}-infra"
  cluster_id        = openshift_assisted_installer_cluster.sno.id
  cpu_architecture  = "x86_64"
  pull_secret       = var.pull_secret
  ssh_authorized_key = var.ssh_public_key
  image_type        = "minimal-iso"
}

# Outputs for the next module
output "cluster_id" {
  value = openshift_assisted_installer_cluster.sno.id
}

output "download_url" {
  value = openshift_assisted_installer_infra_env.sno.download_url
}

output "next_steps" {
  value = <<-EOT
    Cluster setup complete! Now:
    1. Download ISO: ${openshift_assisted_installer_infra_env.sno.download_url}
    2. Boot your machine from this ISO
    3. Wait for host to be discovered
    4. Run the installation module (02_cluster_installation.tf)
  EOT
}