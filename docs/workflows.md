# OpenShift Assisted Installer Workflows

## Overview

This document provides complete workflows for different OpenShift cluster deployment scenarios using the Terraform provider.

## Prerequisites

Before starting any workflow:

1. **Red Hat Account**: Valid Red Hat account with OpenShift entitlements
2. **Offline Token**: Get your offline token from [console.redhat.com](https://console.redhat.com/openshift/token/show)
3. **Pull Secret**: Download pull secret from [console.redhat.com](https://console.redhat.com/openshift/install/pull-secret)
4. **SSH Key**: Generate SSH key pair for cluster access
5. **Network Planning**: Plan IP addresses for API/Ingress VIPs and host networks

## Workflow 1: Single Node OpenShift (SNO)

Perfect for edge computing, development, and resource-constrained environments.

### Step 1: Define Variables

```hcl
# variables.tf
variable "pull_secret" {
  description = "Red Hat pull secret (JSON format)"
  type        = string
  sensitive   = true
}

variable "ssh_public_key" {
  description = "SSH public key for cluster access"
  type        = string
}

variable "cluster_name" {
  description = "Name of the OpenShift cluster"
  type        = string
  default     = "sno-cluster"
}

variable "base_domain" {
  description = "Base DNS domain for the cluster"
  type        = string
  default     = "example.com"
}
```

### Step 2: Create Cluster Definition

```hcl
# main.tf
terraform {
  required_providers {
    oai = {
      source  = "benemon/openshift-assisted-installer"
      version = "~> 0.1"
    }
  }
}

provider "oai" {
  # Uses OFFLINE_TOKEN environment variable
}

# Get latest production OpenShift version
data "oai_openshift_versions" "latest" {
  only_latest = true
}

locals {
  openshift_version = [
    for v in data.oai_openshift_versions.latest.versions : v.version 
    if v.support_level == "production"
  ][0]
}

# Create SNO cluster
resource "oai_cluster" "sno" {
  name                 = var.cluster_name
  base_dns_domain      = var.base_domain
  openshift_version    = local.openshift_version
  cpu_architecture     = "x86_64"
  control_plane_count  = 1
  schedulable_masters  = true
  
  pull_secret    = var.pull_secret
  ssh_public_key = var.ssh_public_key
  
  # SNO doesn't need VIPs
  user_managed_networking = true
}
```

### Step 3: Create Discovery Environment

```hcl
# Create infrastructure environment for host discovery
resource "oai_infra_env" "sno" {
  name              = "${var.cluster_name}-infra"
  cluster_id        = oai_cluster.sno.id
  cpu_architecture  = "x86_64"
  pull_secret       = var.pull_secret
  ssh_authorized_key = var.ssh_public_key
  image_type        = "minimal-iso"
}
```

### Step 4: Install Cluster

```hcl
# Trigger installation after host discovery
resource "oai_cluster_installation" "sno" {
  cluster_id          = oai_cluster.sno.id
  wait_for_hosts      = true
  expected_host_count = 1
  
  timeouts {
    create = "90m"
  }
}
```

### Step 5: Post-Installation Access

```hcl
# Get cluster credentials
data "oai_cluster_credentials" "admin" {
  cluster_id = oai_cluster.sno.id
  depends_on = [oai_cluster_installation.sno]
}

# Download kubeconfig
data "oai_cluster_files" "kubeconfig" {
  cluster_id = oai_cluster.sno.id
  file_name  = "kubeconfig"
  depends_on = [oai_cluster_installation.sno]
}

# Monitor installation
data "oai_cluster_events" "progress" {
  cluster_id = oai_cluster.sno.id
  severities = ["info", "warning", "error", "critical"]
  limit      = 100
}

# Save kubeconfig locally
resource "local_file" "kubeconfig" {
  content         = data.oai_cluster_files.kubeconfig.content
  filename        = "${path.module}/kubeconfig"
  file_permission = "0600"
  
  depends_on = [oai_cluster_installation.sno]
}
```

### Step 6: Outputs

```hcl
# outputs.tf
output "cluster_info" {
  description = "SNO cluster information"
  value = {
    cluster_id   = oai_cluster.sno.id
    cluster_name = oai_cluster.sno.name
    version      = oai_cluster.sno.openshift_version
    status       = oai_cluster_installation.sno.status
  }
}

output "download_iso" {
  description = "Download URL for discovery ISO"
  value       = oai_infra_env.sno.download_url
}

output "cluster_access" {
  description = "Cluster access information"
  value = {
    username       = data.oai_cluster_credentials.admin.username
    password       = data.oai_cluster_credentials.admin.password
    console_url    = data.oai_cluster_credentials.admin.console_url
    kubeconfig     = local_file.kubeconfig.filename
  }
  sensitive = true
}

output "installation_health" {
  description = "Installation health summary"
  value = {
    total_events = length(data.oai_cluster_events.progress.events)
    errors       = length([for e in data.oai_cluster_events.progress.events : e if e.severity == "error"])
    warnings     = length([for e in data.oai_cluster_events.progress.events : e if e.severity == "warning"])
  }
}
```

### Execution Steps

```bash
# Set environment variables
export OFFLINE_TOKEN="your-offline-token"
export TF_VAR_pull_secret='{"auths": {...}}'  # Your pull secret JSON
export TF_VAR_ssh_public_key="$(cat ~/.ssh/id_rsa.pub)"

# Initialize and plan
terraform init
terraform plan

# Apply cluster definition and infra env
terraform apply

# Download and boot from ISO
ISO_URL=$(terraform output -raw download_iso)
echo "Download ISO from: $ISO_URL"
echo "Boot your machine from this ISO and wait for discovery..."

# Wait for host discovery, then complete installation
# Check host discovery status in Red Hat console or via events
terraform apply  # This will trigger installation once host is discovered

# Access your cluster
export KUBECONFIG=$(terraform output -raw cluster_access | jq -r '.kubeconfig')
oc whoami
```

## Workflow 2: Compact 3-Node Cluster

Ideal for production workloads with high availability but resource constraints.

### Cluster Definition

```hcl
# Get supported operators
data "oai_supported_operators" "available" {}

locals {
  # Common operators for compact clusters
  selected_operators = [
    "lso",    # Local Storage Operator
    "odf",    # OpenShift Data Foundation
  ]
  
  # Filter to only include available operators
  cluster_operators = [
    for op in local.selected_operators :
    { name = op }
    if contains(data.oai_supported_operators.available.operators, op)
  ]
}

resource "oai_cluster" "compact" {
  name                = "compact-cluster"
  base_dns_domain     = "production.example.com"
  openshift_version   = local.openshift_version
  cpu_architecture    = "x86_64"
  control_plane_count = 3
  schedulable_masters = true  # Allow workloads on control plane
  
  # Network configuration
  api_vips = [{ ip = "10.0.1.100" }]
  ingress_vips = [{ ip = "10.0.1.101" }]
  
  machine_networks = [
    { cidr = "10.0.1.0/24" }
  ]
  
  pull_secret    = var.pull_secret
  ssh_public_key = var.ssh_public_key
  
  # Install operators during deployment
  olm_operators = local.cluster_operators
}
```

### Infrastructure Environment with Static Networking

```hcl
resource "oai_infra_env" "compact" {
  name              = "${oai_cluster.compact.name}-infra"
  cluster_id        = oai_cluster.compact.id
  cpu_architecture  = "x86_64"
  pull_secret       = var.pull_secret
  ssh_authorized_key = var.ssh_public_key
  image_type        = "full-iso"
  
  # Optional: Static network configuration for each host
  static_network_config = [
    {
      network_yaml = templatefile("${path.module}/network-configs/host1.yaml", {
        hostname = "master-1"
        ip       = "10.0.1.10"
        gateway  = "10.0.1.1"
        dns      = "10.0.1.2"
      })
    },
    {
      network_yaml = templatefile("${path.module}/network-configs/host2.yaml", {
        hostname = "master-2"
        ip       = "10.0.1.11"
        gateway  = "10.0.1.1"
        dns      = "10.0.1.2"
      })
    },
    {
      network_yaml = templatefile("${path.module}/network-configs/host3.yaml", {
        hostname = "master-3"
        ip       = "10.0.1.12"
        gateway  = "10.0.1.1"
        dns      = "10.0.1.2"
      })
    }
  ]
}
```

### Custom Manifests

```hcl
# Apply custom storage class
resource "oai_manifest" "storage_class" {
  cluster_id = oai_cluster.compact.id
  file_name  = "default-storage-class.yaml"
  folder     = "manifests"
  
  content = templatefile("${path.module}/manifests/storage-class.yaml", {
    storage_class_name = "compact-local-storage"
    is_default         = true
  })
}

# Configure monitoring retention
resource "oai_manifest" "monitoring_config" {
  cluster_id = oai_cluster.compact.id
  file_name  = "monitoring-config.yaml"
  folder     = "openshift"
  
  content = templatefile("${path.module}/manifests/monitoring.yaml", {
    retention_days    = 15
    storage_size      = "40Gi"
    storage_class     = "compact-local-storage"
  })
}
```

## Workflow 3: Full Production Cluster

Complete high-availability cluster with dedicated control plane and worker nodes.

### Multi-Node Cluster Configuration

```hcl
resource "oai_cluster" "production" {
  name                = "production-cluster"
  base_dns_domain     = "prod.company.com"
  openshift_version   = local.openshift_version
  cpu_architecture    = "x86_64"
  control_plane_count = 3
  schedulable_masters = false  # Dedicated control plane
  
  # Production network configuration
  api_vips = [{ ip = "10.1.100.10" }]
  ingress_vips = [{ ip = "10.1.100.11" }]
  
  machine_networks = [
    { cidr = "10.1.100.0/24" }
  ]
  
  # Additional network settings
  cluster_network_cidr = "10.128.0.0/14"
  service_network_cidr = "172.30.0.0/16"
  
  # Security configuration
  pull_secret    = var.pull_secret
  ssh_public_key = var.ssh_public_key
  
  # Production operators
  olm_operators = [
    { name = "lso" },
    { name = "odf" },
    { name = "elasticsearch-operator" },
    { name = "cluster-logging" }
  ]
  
  # Optional: Proxy configuration for restricted networks
  http_proxy  = var.http_proxy
  https_proxy = var.https_proxy
  no_proxy    = var.no_proxy
}
```

### Host Management

```hcl
# Infrastructure environment for discovery
resource "oai_infra_env" "production" {
  name              = "${oai_cluster.production.name}-infra"
  cluster_id        = oai_cluster.production.id
  cpu_architecture  = "x86_64"
  pull_secret       = var.pull_secret
  ssh_authorized_key = var.ssh_public_key
  image_type        = "full-iso"
}

# Manage individual hosts (after discovery)
resource "oai_host" "control_plane" {
  count = 3
  
  infra_env_id        = oai_infra_env.production.id
  host_id             = var.control_plane_host_ids[count.index]
  host_name           = "master-${count.index + 1}"
  host_role           = "master"
  installation_disk_id = "/dev/sda"
  
  # Skip formatting additional disks for local storage
  disks_skip_formatting = ["/dev/sdb", "/dev/sdc"]
}

resource "oai_host" "workers" {
  count = 3
  
  infra_env_id        = oai_infra_env.production.id
  host_id             = var.worker_host_ids[count.index]
  host_name           = "worker-${count.index + 1}"
  host_role           = "worker"
  installation_disk_id = "/dev/sda"
  
  disks_skip_formatting = ["/dev/sdb", "/dev/sdc", "/dev/sdd"]
}
```

### Installation and Monitoring

```hcl
# Trigger installation
resource "oai_cluster_installation" "production" {
  cluster_id          = oai_cluster.production.id
  wait_for_hosts      = true
  expected_host_count = 6  # 3 masters + 3 workers
  
  timeouts {
    create = "180m"  # 3 hours for large cluster
  }
  
  depends_on = [
    oai_host.control_plane,
    oai_host.workers
  ]
}

# Comprehensive monitoring
data "oai_cluster_events" "installation" {
  cluster_id = oai_cluster.production.id
  severities = ["info", "warning", "error", "critical"]
  limit      = 500
}

data "oai_cluster_events" "errors_only" {
  cluster_id = oai_cluster.production.id
  severities = ["error", "critical"]
}

# Download all important files
data "oai_cluster_files" "kubeconfig" {
  cluster_id = oai_cluster.production.id
  file_name  = "kubeconfig"
  depends_on = [oai_cluster_installation.production]
}

data "oai_cluster_files" "install_config" {
  cluster_id = oai_cluster.production.id
  file_name  = "install-config.yaml"
  depends_on = [oai_cluster_installation.production]
}

data "oai_cluster_logs" "installation_logs" {
  cluster_id = oai_cluster.production.id
  logs_type  = "controller"
  depends_on = [oai_cluster_installation.production]
}
```

## Workflow 4: Troubleshooting and Recovery

### Monitoring Installation Progress

```hcl
# Real-time installation monitoring
data "oai_cluster_events" "live_progress" {
  cluster_id = oai_cluster.example.id
  severities = ["info", "warning", "error", "critical"]
  limit      = 100
  order      = "desc"
}

locals {
  # Parse installation progress
  installation_status = {
    latest_event = length(data.oai_cluster_events.live_progress.events) > 0 ? 
                  data.oai_cluster_events.live_progress.events[0] : null
    
    error_events = [
      for event in data.oai_cluster_events.live_progress.events :
      event if contains(["error", "critical"], event.severity)
    ]
    
    warning_events = [
      for event in data.oai_cluster_events.live_progress.events :
      event if event.severity == "warning"
    ]
    
    progress_events = [
      for event in data.oai_cluster_events.live_progress.events :
      event if can(regex("progress|started|completed", event.message))
    ]
  }
}

# Troubleshooting outputs
output "installation_status" {
  value = {
    current_phase = local.installation_status.latest_event != null ? 
                   local.installation_status.latest_event.message : "No events"
    
    has_errors = length(local.installation_status.error_events) > 0
    
    health_summary = {
      total_events = length(data.oai_cluster_events.live_progress.events)
      errors       = length(local.installation_status.error_events) 
      warnings     = length(local.installation_status.warning_events)
    }
    
    recent_errors = [
      for event in local.installation_status.error_events :
      "${event.event_time}: ${event.message}"
    ]
  }
}
```

### Recovery and Debugging

```hcl
# Download diagnostic logs
data "oai_cluster_logs" "debug" {
  cluster_id = oai_cluster.example.id
  logs_type  = "controller"
}

# Save logs for analysis
resource "local_file" "debug_logs" {
  content  = data.oai_cluster_logs.debug.content
  filename = "${path.module}/debug/installation-${formatdate("YYYY-MM-DD-hhmm", timestamp())}.log"
}

# Generate troubleshooting report
resource "local_file" "troubleshooting_report" {
  content = templatefile("${path.module}/templates/troubleshooting.md", {
    cluster_name    = oai_cluster.example.name
    cluster_id      = oai_cluster.example.id
    installation_status = local.installation_status
    recent_events   = data.oai_cluster_events.live_progress.events
  })
  
  filename = "${path.module}/troubleshooting-report.md"
}
```

## Best Practices by Workflow

### Development/Testing
- Use SNO for rapid development cycles
- Implement automated testing with `terraform destroy` and recreate
- Use minimal resource requirements for cost efficiency

### Production Staging
- Use compact clusters to simulate production constraints
- Test all manifests and operators before production deployment
- Implement monitoring and alerting early

### Production Deployment
- Use dedicated control plane and worker nodes
- Implement comprehensive backup strategies
- Plan for disaster recovery and cluster updates
- Use Infrastructure as Code for all cluster modifications

### Operations and Maintenance
- Regular monitoring of cluster events
- Automated log collection and analysis  
- Planned maintenance windows for updates
- Documentation of all customizations and procedures

## Integration Patterns

### CI/CD Integration

```hcl
# Integrate with external systems
resource "kubernetes_secret" "ci_cd_access" {
  depends_on = [oai_cluster_installation.production]
  
  metadata {
    name      = "ci-cd-cluster-access"
    namespace = "ci-cd"
  }
  
  data = {
    kubeconfig = base64encode(data.oai_cluster_files.kubeconfig.content)
    username   = data.oai_cluster_credentials.admin.username
    password   = data.oai_cluster_credentials.admin.password
  }
}
```

### Monitoring Integration

```hcl
# Send cluster information to external monitoring
resource "http" "register_cluster" {
  depends_on = [oai_cluster_installation.production]
  
  url    = "https://monitoring.company.com/api/clusters"
  method = "POST"
  
  request_headers = {
    "Content-Type"  = "application/json"
    "Authorization" = "Bearer ${var.monitoring_token}"
  }
  
  request_body = jsonencode({
    cluster_name = oai_cluster.production.name
    cluster_id   = oai_cluster.production.id
    console_url  = data.oai_cluster_credentials.admin.console_url
    version      = oai_cluster.production.openshift_version
  })
}
```