terraform {
  required_providers {
    oai = {
      source  = "benemon/openshift-assisted-installer"
      version = "~> 0.1"
    }
  }
}

# Configure the OpenShift Assisted Installer Provider  
# Set OFFLINE_TOKEN environment variable for authentication
provider "oai" {
  # endpoint = "https://api.openshift.com/api/assisted-install"  # Optional: Override default
  # timeout = "900s"  # Optional: Increase timeout for large installations
}

# ==============================================================================
# Full Production OpenShift Cluster
# 3 dedicated control plane nodes + multiple worker nodes
# Production-ready with comprehensive operator suite and advanced configuration
# ==============================================================================

# Get the latest OpenShift version
data "oai_openshift_versions" "latest" {
  only_latest = true
}

# Get operator bundles for comprehensive setup
data "oai_operator_bundles" "available" {}

locals {
  cluster_name = "production-cluster"
  base_domain  = "example.com"
  
  # Use the latest production version
  openshift_version = [for v in data.oai_openshift_versions.latest.versions : v.version 
    if v.support_level == "production"][0]
    
  # Network configuration
  machine_network = "192.168.1.0/24"
  api_vip         = "192.168.1.100"
  ingress_vip     = "192.168.1.101"
}

# Create a full production OpenShift cluster
resource "oai_cluster" "production_cluster" {
  # Basic cluster configuration
  name              = local.cluster_name
  base_dns_domain   = local.base_domain
  openshift_version = local.openshift_version
  
  # Production cluster topology
  cpu_architecture     = "x86_64"
  control_plane_count  = 3        # 3 dedicated control plane nodes
  schedulable_masters  = false    # Dedicated control plane (no workloads)
  
  # Network configuration
  cluster_network_cidr = "10.128.0.0/14"
  service_network_cidr = "172.30.0.0/16"
  
  # Machine networks for discovered hosts
  machine_networks = [
    {
      cidr = local.machine_network
    }
  ]
  
  # Load balancer configuration - cluster-managed
  load_balancer = {
    type = "cluster-managed"
  }
  
  # API and Ingress VIPs (required for multi-node)
  api_vips = [
    { ip = local.api_vip }
  ]
  
  ingress_vips = [
    { ip = local.ingress_vip }
  ]
  
  # Pull secret (required)
  pull_secret = file("pull-secret.json")
  
  # SSH access key
  ssh_public_key = file("~/.ssh/id_rsa.pub")
  
  # Comprehensive operator suite for production
  olm_operators = [
    # Storage operators
    {
      name = "lso"  # Local Storage Operator
    },
    {
      name = "odf"  # OpenShift Data Foundation
    },
    
    # Virtualization and compute
    {
      name = "cnv"  # OpenShift Virtualization  
    },
    {
      name = "mtv"  # Migration Toolkit for Virtualization
    },
    
    # AI/ML platform
    {
      name = "openshift-ai"  # OpenShift AI
    },
    
    # Service mesh and serverless
    {
      name = "servicemesh"   # Red Hat OpenShift Service Mesh
    },
    {
      name = "serverless"    # OpenShift Serverless
    },
    
    # CI/CD and GitOps
    {
      name = "pipelines"     # OpenShift Pipelines (Tekton)
    },
    
    # Observability and monitoring
    {
      name = "cluster-observability"  # Cluster Observability Operator
    },
    
    # Node management
    {
      name = "node-healthcheck"      # Node Health Check Operator
    },
    {
      name = "node-maintenance"      # Node Maintenance Operator
    }
  ]
  
  # Network management
  user_managed_networking = false  # OpenShift-managed networking
  
  # VIP allocation via DHCP (set false for static IPs)
  vip_dhcp_allocation = false
  
  # NTP configuration for time synchronization
  additional_ntp_source = "pool.ntp.org"
  
  # Optional: Proxy configuration for restricted environments
  # proxy = {
  #   http_proxy  = "http://proxy.example.com:8080"
  #   https_proxy = "http://proxy.example.com:8080"
  #   no_proxy    = "localhost,127.0.0.1,.example.com,.cluster.local,.svc,10.128.0.0/14,172.30.0.0/16"
  # }
}

# Create infrastructure environment for host discovery
resource "oai_infra_env" "production_infra" {
  name              = "${local.cluster_name}-infra"
  cluster_id        = oai_cluster.production_cluster.id
  cpu_architecture  = "x86_64"
  pull_secret       = file("pull-secret.json")
  
  # SSH key for discovered hosts
  ssh_authorized_key = file("~/.ssh/id_rsa.pub")
  
  # Use full ISO for maximum hardware compatibility
  image_type = "full-iso"
  
  # Optional: Kernel arguments for specialized hardware
  # kernel_arguments = [
  #   {
  #     operation = "append"
  #     value     = "intel_iommu=on"
  #   }
  # ]
  
  # Optional: Static network configuration for hosts
  # Uncomment and customize for static IP configuration
  # static_network_config = [
  #   {
  #     network_yaml = file("master1-network.yaml")
  #   },
  #   {
  #     network_yaml = file("master2-network.yaml")
  #   },
  #   {
  #     network_yaml = file("master3-network.yaml")
  #   },
  #   {
  #     network_yaml = file("worker1-network.yaml")
  #   },
  #   {
  #     network_yaml = file("worker2-network.yaml")
  #   }
  # ]
}

# Production-grade manifests
resource "oai_manifest" "cluster_monitoring_config" {
  cluster_id = oai_cluster.production_cluster.id
  file_name  = "cluster-monitoring-config.yaml"
  folder     = "manifests"
  
  content = <<-EOT
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: cluster-monitoring-config
      namespace: openshift-monitoring
    data:
      config.yaml: |
        enableUserWorkload: true
        prometheusK8s:
          retention: 7d
          volumeClaimTemplate:
            spec:
              storageClassName: ocs-storagecluster-ceph-rbd
              resources:
                requests:
                  storage: 100Gi
        alertmanagerMain:
          volumeClaimTemplate:
            spec:
              storageClassName: ocs-storagecluster-ceph-rbd
              resources:
                requests:
                  storage: 20Gi
  EOT
}

resource "oai_manifest" "oauth_config" {
  cluster_id = oai_cluster.production_cluster.id
  file_name  = "oauth-htpasswd.yaml"
  folder     = "manifests"
  
  content = <<-EOT
    apiVersion: config.openshift.io/v1
    kind: OAuth
    metadata:
      name: cluster
    spec:
      identityProviders:
      - name: htpasswd
        mappingMethod: claim
        type: HTPasswd
        htpasswd:
          fileData:
            name: htpass-secret
    ---
    apiVersion: v1
    kind: Secret
    metadata:
      name: htpass-secret
      namespace: openshift-config
    type: Opaque
    data:
      htpasswd: YWRtaW46JGFwcjEkLjF6T3RHVHEkbDBCZ1c5YlNuTWk5Nm1qVVEvUWo5MAo=  # admin:admin
  EOT
}

resource "oai_manifest" "ingress_controller" {
  cluster_id = oai_cluster.production_cluster.id
  file_name  = "ingress-controller.yaml"
  folder     = "manifests"
  
  content = <<-EOT
    apiVersion: operator.openshift.io/v1
    kind: IngressController
    metadata:
      name: default
      namespace: openshift-ingress-operator
    spec:
      replicas: 2
      endpointPublishingStrategy:
        type: LoadBalancerService
        loadBalancer:
          scope: External
      nodePlacement:
        nodeSelector:
          matchLabels:
            node-role.kubernetes.io/worker: ""
        tolerations:
        - key: node-role.kubernetes.io/worker
          operator: Exists
  EOT
}

# ==============================================================================
# Outputs
# ==============================================================================

output "cluster_info" {
  description = "Production cluster information"
  value = {
    cluster_id          = oai_cluster.production_cluster.id
    cluster_name        = oai_cluster.production_cluster.name
    base_domain         = oai_cluster.production_cluster.base_dns_domain
    version             = oai_cluster.production_cluster.openshift_version
    architecture        = oai_cluster.production_cluster.cpu_architecture
    control_nodes       = oai_cluster.production_cluster.control_plane_count
    dedicated_masters   = !oai_cluster.production_cluster.schedulable_masters
    api_vip            = oai_cluster.production_cluster.api_vips[0].ip
    ingress_vip        = oai_cluster.production_cluster.ingress_vips[0].ip
  }
}

output "network_config" {
  description = "Network configuration details"
  value = {
    cluster_cidr       = oai_cluster.production_cluster.cluster_network_cidr
    service_cidr       = oai_cluster.production_cluster.service_network_cidr  
    machine_networks   = oai_cluster.production_cluster.machine_networks
    load_balancer      = oai_cluster.production_cluster.load_balancer
    vip_dhcp          = oai_cluster.production_cluster.vip_dhcp_allocation
    ntp_source        = oai_cluster.production_cluster.additional_ntp_source
  }
}

output "access_info" {
  description = "Cluster access information"
  value = {
    api_endpoint    = "https://api.${local.cluster_name}.${local.base_domain}:6443"
    console_url     = "https://console-openshift-console.apps.${local.cluster_name}.${local.base_domain}"
    oauth_endpoint  = "https://oauth-openshift.apps.${local.cluster_name}.${local.base_domain}"
    default_user    = "admin"
    default_pass    = "admin"  # Change in production!
  }
}

output "infra_env_info" {
  description = "Infrastructure environment for host discovery"
  value = {
    infra_env_id  = oai_infra_env.production_infra.id
    download_url  = oai_infra_env.production_infra.download_url
    image_type    = oai_infra_env.production_infra.image_type
    expires_at    = oai_infra_env.production_infra.expires_at
  }
}

output "operators_installed" {
  description = "Operators that will be installed"
  value = [for op in oai_cluster.production_cluster.olm_operators : op.name]
}

output "installation_guide" {
  description = "Complete installation guide"
  value = {
    step_1  = "Download discovery ISO: ${oai_infra_env.production_infra.download_url}"
    step_2  = "Boot 3+ machines from ISO (3 for masters, 2+ for workers recommended)"
    step_3  = "Wait for hosts to be discovered and validated"
    step_4  = "Assign master role to 3 hosts, worker role to remaining hosts"
    step_5  = "Ensure all hosts pass validation checks"
    step_6  = "Installation begins when cluster requirements are met"
    step_7  = "Monitor installation progress in web console or CLI"
    step_8  = "Access cluster: https://console-openshift-console.apps.${local.cluster_name}.${local.base_domain}"
    step_9  = "Login with admin/admin (change password immediately!)"
    step_10 = "Verify all operators are installed and healthy"
  }
}

# ==============================================================================
# Usage Notes  
# ==============================================================================

# This example creates a Full Production OpenShift cluster with:
# - 3 dedicated control plane nodes (no workloads)
# - Multiple worker nodes (2+ recommended)
# - Comprehensive operator suite including AI/ML, virtualization, service mesh
# - Advanced monitoring, authentication, and ingress configuration
# - Production-grade storage and networking setup
#
# Prerequisites:
# 1. Set OFFLINE_TOKEN environment variable
# 2. Create pull-secret.json with your Red Hat pull secret
# 3. Ensure ~/.ssh/id_rsa.pub exists for SSH access
# 4. Have 5+ machines ready (3 masters + 2+ workers minimum)
# 5. Network infrastructure with available VIPs
# 6. DNS configuration for wildcard apps domain
#
# Hardware Requirements:
# Control Plane Nodes (3x):
# - CPU: 4 vCPUs minimum, 8 vCPUs recommended
# - Memory: 16 GB RAM minimum, 32 GB recommended
# - Storage: 120 GB minimum
#
# Worker Nodes (2+ recommended):
# - CPU: 8 vCPUs minimum, 16 vCPUs recommended  
# - Memory: 32 GB RAM minimum, 64 GB recommended
# - Storage: 200 GB minimum (more for workloads)
#
# Network Requirements:
# - API VIP: 192.168.1.100 (must be available)
# - Ingress VIP: 192.168.1.101 (must be available)
# - Machine network: 192.168.1.0/24
# - DNS: *.apps.production-cluster.example.com → Ingress VIP
# - Internet connectivity for all nodes
# - NTP access for time synchronization
#
# Storage:
# - OpenShift Data Foundation provides shared storage
# - Local Storage Operator for node-local storage
# - Consider external storage solutions for production workloads
#
# Security Considerations:
# - Change default admin password immediately after installation
# - Configure proper RBAC and identity providers
# - Review and harden cluster security policies
# - Enable audit logging and monitoring
# - Consider network policies and pod security standards