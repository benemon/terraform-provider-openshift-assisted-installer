terraform {
  required_providers {
    oai = {
      source = "benemon/oai"
    }
  }
}

# Configure the OpenShift Assisted Installer Provider
# Set OFFLINE_TOKEN environment variable for authentication
provider "oai" {
  # endpoint = "https://api.openshift.com/api/assisted-install"  # Optional: Override default
  # timeout = "600s"  # Optional: Increase timeout for installation
}

# ==============================================================================
# Single Node OpenShift (SNO) Cluster
# Minimal configuration for a single node cluster suitable for edge/development
# ==============================================================================

# Get the latest OpenShift version
data "oai_openshift_versions" "latest" {
  only_latest = true
}

locals {
  cluster_name = "sno-cluster"
  base_domain  = "example.com"
  
  # Use the latest supported version
  openshift_version = [for v in data.oai_openshift_versions.latest.versions : v.version 
    if v.support_level == "production"][0]
}

# Create a single node OpenShift cluster
resource "oai_cluster" "sno_cluster" {
  # Basic cluster configuration
  name              = local.cluster_name
  base_dns_domain   = local.base_domain
  openshift_version = local.openshift_version
  
  # SNO-specific configuration
  cpu_architecture     = "x86_64"
  control_plane_count  = 1        # Single node = 1 control plane
  schedulable_masters  = true     # Allow workloads on control plane
  
  # Network configuration
  cluster_network_cidr = "10.128.0.0/14"
  service_network_cidr = "172.30.0.0/16"
  
  # Pull secret (required)
  # Store this in pull-secret.json file or use variable
  pull_secret = file("pull-secret.json")
  
  # SSH access key for troubleshooting
  ssh_public_key = file("~/.ssh/id_rsa.pub")
  
  # Optional: Add minimal operators for functionality
  olm_operators = [
    {
      name = "lso"  # Local Storage Operator for storage
    }
  ]
  
  # Optional: Configure user-managed networking for edge scenarios
  user_managed_networking = false  # Set to true for custom networking
  
  # VIP configuration for single node (optional)
  # api_vips = [
  #   { ip = "192.168.1.100" }
  # ]
  # 
  # ingress_vips = [
  #   { ip = "192.168.1.101" }
  # ]
}

# Create infrastructure environment for host discovery
resource "oai_infra_env" "sno_infra" {
  name              = "${local.cluster_name}-infra"
  cluster_id        = oai_cluster.sno_cluster.id
  cpu_architecture  = "x86_64"
  pull_secret       = file("pull-secret.json")
  
  # SSH key for discovered hosts
  ssh_authorized_key = file("~/.ssh/id_rsa.pub")
  
  # Use minimal ISO for faster downloads
  image_type = "minimal-iso"
}

# ==============================================================================
# Outputs
# ==============================================================================

output "cluster_info" {
  description = "Single node cluster information"
  value = {
    cluster_id    = oai_cluster.sno_cluster.id
    cluster_name  = oai_cluster.sno_cluster.name
    base_domain   = oai_cluster.sno_cluster.base_dns_domain
    version       = oai_cluster.sno_cluster.openshift_version
    architecture  = oai_cluster.sno_cluster.cpu_architecture
    node_count    = oai_cluster.sno_cluster.control_plane_count
  }
}

output "infra_env_info" {
  description = "Infrastructure environment for host discovery"
  value = {
    infra_env_id  = oai_infra_env.sno_infra.id
    download_url  = oai_infra_env.sno_infra.download_url
    image_type    = oai_infra_env.sno_infra.image_type
    expires_at    = oai_infra_env.sno_infra.expires_at
  }
}

output "next_steps" {
  description = "Next steps to complete the installation"
  value = {
    step_1 = "Download the discovery ISO from: ${oai_infra_env.sno_infra.download_url}"
    step_2 = "Boot your physical/virtual machine from the ISO"
    step_3 = "Wait for the host to appear and be validated"
    step_4 = "Installation will begin automatically once host is ready"
    step_5 = "Access cluster at: https://console-openshift-console.apps.${local.cluster_name}.${local.base_domain}"
  }
}

# ==============================================================================
# Usage Notes
# ==============================================================================

# This example creates a Single Node OpenShift cluster with:
# - 1 control plane node that also runs workloads
# - Minimal operator set for basic functionality
# - Standard networking configuration
# - Discovery ISO for automatic host registration
#
# Prerequisites:
# 1. Set OFFLINE_TOKEN environment variable
# 2. Create pull-secret.json with your Red Hat pull secret
# 3. Ensure ~/.ssh/id_rsa.pub exists for SSH access
# 4. Have physical/virtual machine ready to boot from ISO
#
# Hardware Requirements:
# - CPU: 4 vCPUs minimum, 8 vCPUs recommended
# - Memory: 16 GB RAM minimum, 32 GB recommended
# - Storage: 120 GB minimum
# - Network: Internet connectivity for image pulls