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
# Compact 3-Node OpenShift Cluster
# 3 control plane nodes that also run workloads (no dedicated workers)
# Ideal for resource-constrained environments or smaller deployments
# ==============================================================================

# Get the latest OpenShift version
data "oai_openshift_versions" "latest" {
  only_latest = true
}

# Get available operators for optional installation
data "oai_supported_operators" "available" {}

locals {
  cluster_name = "compact-cluster"
  base_domain  = "example.com"
  
  # Use the latest production version
  openshift_version = [for v in data.oai_openshift_versions.latest.versions : v.version 
    if v.support_level == "production"][0]
}

# Create a compact 3-node OpenShift cluster
resource "oai_cluster" "compact_cluster" {
  # Basic cluster configuration
  name              = local.cluster_name
  base_dns_domain   = local.base_domain
  openshift_version = local.openshift_version
  
  # Compact cluster configuration
  cpu_architecture     = "x86_64"
  control_plane_count  = 3        # 3 control plane nodes
  schedulable_masters  = true     # Allow workloads on control plane nodes
  
  # Network configuration
  cluster_network_cidr = "10.128.0.0/14"
  service_network_cidr = "172.30.0.0/16"
  
  # Machine network for discovered hosts
  machine_networks = [
    {
      cidr = "192.168.1.0/24"
    }
  ]
  
  # API and Ingress VIPs (required for multi-node)
  api_vips = [
    { ip = "192.168.1.100" }
  ]
  
  ingress_vips = [
    { ip = "192.168.1.101" }
  ]
  
  # Pull secret (required)
  pull_secret = file("pull-secret.json")
  
  # SSH access key
  ssh_public_key = file("~/.ssh/id_rsa.pub")
  
  # Common operators for compact clusters
  olm_operators = [
    {
      name = "lso"  # Local Storage Operator
    },
    {
      name = "odf"  # OpenShift Data Foundation
    }
  ]
  
  # Network management
  user_managed_networking = false  # Let OpenShift manage networking
  
  # Optional: NTP configuration
  # additional_ntp_source = "pool.ntp.org"
  
  # Optional: Proxy configuration for restricted networks
  # proxy = {
  #   http_proxy  = "http://proxy.example.com:8080"
  #   https_proxy = "http://proxy.example.com:8080"
  #   no_proxy    = "localhost,127.0.0.1,.example.com"
  # }
}

# Create infrastructure environment for host discovery
resource "oai_infra_env" "compact_infra" {
  name              = "${local.cluster_name}-infra"
  cluster_id        = oai_cluster.compact_cluster.id
  cpu_architecture  = "x86_64"
  pull_secret       = file("pull-secret.json")
  
  # SSH key for discovered hosts
  ssh_authorized_key = file("~/.ssh/id_rsa.pub")
  
  # Use full ISO for better hardware compatibility
  image_type = "full-iso"
  
  # Optional: Static network configuration for hosts
  # static_network_config = [
  #   {
  #     network_yaml = file("host1-network.yaml")
  #   },
  #   {
  #     network_yaml = file("host2-network.yaml")
  #   },
  #   {
  #     network_yaml = file("host3-network.yaml")
  #   }
  # ]
}

# Example custom manifest for additional configuration
resource "oai_manifest" "custom_config" {
  cluster_id = oai_cluster.compact_cluster.id
  file_name  = "custom-storage-class.yaml"
  folder     = "manifests"  # or "openshift" for cluster-level manifests
  
  content = <<-EOT
    apiVersion: storage.k8s.io/v1
    kind: StorageClass
    metadata:
      name: compact-cluster-storage
      annotations:
        storageclass.kubernetes.io/is-default-class: "true"
    provisioner: kubernetes.io/no-provisioner
    volumeBindingMode: WaitForFirstConsumer
    allowVolumeExpansion: true
  EOT
}

# ==============================================================================
# Outputs
# ==============================================================================

output "cluster_info" {
  description = "Compact cluster information"
  value = {
    cluster_id       = oai_cluster.compact_cluster.id
    cluster_name     = oai_cluster.compact_cluster.name
    base_domain      = oai_cluster.compact_cluster.base_dns_domain
    version          = oai_cluster.compact_cluster.openshift_version
    architecture     = oai_cluster.compact_cluster.cpu_architecture
    control_nodes    = oai_cluster.compact_cluster.control_plane_count
    api_vip          = oai_cluster.compact_cluster.api_vips[0].ip
    ingress_vip      = oai_cluster.compact_cluster.ingress_vips[0].ip
    schedulable      = oai_cluster.compact_cluster.schedulable_masters
  }
}

output "network_config" {
  description = "Network configuration details"
  value = {
    cluster_cidr     = oai_cluster.compact_cluster.cluster_network_cidr
    service_cidr     = oai_cluster.compact_cluster.service_network_cidr
    machine_networks = oai_cluster.compact_cluster.machine_networks
    api_endpoint     = "https://api.${local.cluster_name}.${local.base_domain}:6443"
    console_url      = "https://console-openshift-console.apps.${local.cluster_name}.${local.base_domain}"
  }
}

output "infra_env_info" {
  description = "Infrastructure environment for host discovery"
  value = {
    infra_env_id  = oai_infra_env.compact_infra.id
    download_url  = oai_infra_env.compact_infra.download_url
    image_type    = oai_infra_env.compact_infra.image_type
    expires_at    = oai_infra_env.compact_infra.expires_at
  }
}

output "installation_steps" {
  description = "Steps to complete the installation"
  value = {
    step_1 = "Download discovery ISO: ${oai_infra_env.compact_infra.download_url}"
    step_2 = "Boot 3 physical/virtual machines from the ISO"
    step_3 = "Wait for all 3 hosts to be discovered and validated"
    step_4 = "Ensure hosts are assigned master role (automatic with schedulable_masters=true)"
    step_5 = "Installation begins when all hosts are ready"
    step_6 = "Access cluster console at: https://console-openshift-console.apps.${local.cluster_name}.${local.base_domain}"
  }
}

output "operators_installed" {
  description = "Operators that will be installed"
  value = [for op in oai_cluster.compact_cluster.olm_operators : op.name]
}

# ==============================================================================
# Usage Notes
# ==============================================================================

# This example creates a Compact 3-Node OpenShift cluster with:
# - 3 control plane nodes that also run workloads
# - No dedicated worker nodes (schedulable_masters=true)
# - Basic storage and data foundation operators
# - Standard networking with VIPs
# - Custom storage class manifest
#
# Prerequisites:
# 1. Set OFFLINE_TOKEN environment variable
# 2. Create pull-secret.json with your Red Hat pull secret
# 3. Ensure ~/.ssh/id_rsa.pub exists for SSH access
# 4. Have 3 machines ready to boot from ISO
# 5. Ensure network has available IPs for VIPs (192.168.1.100-101)
#
# Hardware Requirements (per node):
# - CPU: 4 vCPUs minimum, 8 vCPUs recommended
# - Memory: 16 GB RAM minimum, 32 GB recommended
# - Storage: 120 GB minimum per node
# - Network: Same subnet for all nodes, internet connectivity
#
# Network Requirements:
# - API VIP: 192.168.1.100 (must be available)
# - Ingress VIP: 192.168.1.101 (must be available)  
# - Machine network: 192.168.1.0/24
# - DNS: Ensure *.apps.compact-cluster.example.com resolves to Ingress VIP