# OpenShift Assisted Installer Provider Documentation

## Overview

The OpenShift Assisted Installer provider enables Infrastructure as Code management of OpenShift clusters using the Red Hat OpenShift Assisted Service API. This provider supports the complete cluster lifecycle from definition through installation to post-installation access.

## Provider Configuration

### Basic Configuration

```hcl
terraform {
  required_providers {
    oai = {
      source  = "benemon/openshift-assisted-installer"
      version = "~> 0.1"
    }
  }
}

provider "oai" {
  offline_token = var.offline_token
  endpoint      = "https://api.openshift.com/api/assisted-install"
  timeout       = "60s"
}
```

### Configuration Arguments

| Argument       | Type   | Required | Description |
|----------------|--------|----------|-------------|
| `offline_token` | string | Yes      | Red Hat offline token for API authentication. Can also be provided via `OFFLINE_TOKEN` environment variable. |
| `endpoint`     | string | No       | OpenShift Assisted Service API endpoint. Defaults to `https://api.openshift.com/api/assisted-install`. |
| `timeout`      | string | No       | Default timeout for API requests in Go duration format (e.g., "30s", "5m"). Defaults to "30s". |

### Authentication

The provider uses Red Hat's offline token authentication system:

1. **Get your offline token**: Visit [Red Hat API Tokens](https://console.redhat.com/openshift/token/show)
2. **Configure the provider**: Set the token in the provider configuration or environment variable
3. **Token refresh**: The provider automatically exchanges offline tokens for access tokens and handles refresh

```bash
# Set via environment variable
export OFFLINE_TOKEN="your-offline-token-here"

# Or configure directly in Terraform
provider "oai" {
  offline_token = "your-offline-token-here"
}
```

## Resources

### `oai_cluster`

Manages OpenShift cluster definitions and configurations.

```hcl
resource "oai_cluster" "example" {
  name                 = "my-cluster"
  base_dns_domain      = "example.com"
  openshift_version    = "4.16.0"
  cpu_architecture     = "x86_64"
  control_plane_count  = 3
  
  # Network configuration
  api_vips = [{ ip = "192.168.1.100" }]
  ingress_vips = [{ ip = "192.168.1.101" }]
  
  # Required secrets
  pull_secret    = var.pull_secret
  ssh_public_key = var.ssh_public_key
}
```

**Key Arguments:**
- `name` (Required) - Cluster name
- `openshift_version` (Required) - OpenShift version to install
- `cpu_architecture` (Required) - Target architecture (x86_64, arm64, ppc64le, s390x)
- `pull_secret` (Required) - Red Hat pull secret in JSON format
- `control_plane_count` (Optional) - Number of control plane nodes (1 for SNO, 3+ for HA)
- `schedulable_masters` (Optional) - Allow workloads on control plane nodes

### `oai_cluster_installation`

Triggers and monitors cluster installation progress.

```hcl
resource "oai_cluster_installation" "example" {
  cluster_id          = oai_cluster.example.id
  wait_for_hosts      = true
  expected_host_count = 3
  
  timeouts {
    create = "120m"
  }
}
```

**Key Arguments:**
- `cluster_id` (Required) - ID of the cluster to install
- `wait_for_hosts` (Optional) - Wait for hosts before starting installation
- `expected_host_count` (Optional) - Number of hosts to wait for

### `oai_infra_env`

Manages infrastructure environments for host discovery.

```hcl
resource "oai_infra_env" "example" {
  name              = "my-infra-env"
  cluster_id        = oai_cluster.example.id
  cpu_architecture  = "x86_64"
  pull_secret       = var.pull_secret
  ssh_authorized_key = var.ssh_public_key
  image_type        = "full-iso"
}
```

**Key Arguments:**
- `name` (Required) - Infrastructure environment name
- `cpu_architecture` (Required) - Target architecture
- `pull_secret` (Required) - Red Hat pull secret
- `cluster_id` (Optional) - Associate with a specific cluster
- `image_type` (Optional) - ISO type: "full-iso" or "minimal-iso"

### `oai_host`

Manages individual cluster hosts and their configuration.

```hcl
resource "oai_host" "worker" {
  infra_env_id        = oai_infra_env.example.id
  host_id             = "discovered-host-id"
  host_name           = "worker-1"
  host_role           = "worker"
  installation_disk_id = "/dev/sda"
}
```

### `oai_manifest`

Applies custom configuration manifests to clusters.

```hcl
resource "oai_manifest" "example" {
  cluster_id = oai_cluster.example.id
  folder     = "manifests"
  file_name  = "custom-config.yaml"
  
  content = templatefile("${path.module}/manifests/config.yaml", {
    cluster_name = oai_cluster.example.name
  })
}
```

## Data Sources

### Cluster Information

#### `oai_openshift_versions`

Retrieves available OpenShift versions.

```hcl
data "oai_openshift_versions" "latest" {
  only_latest = true
}

# Use in cluster configuration
resource "oai_cluster" "example" {
  openshift_version = data.oai_openshift_versions.latest.versions[0].version
}
```

#### `oai_supported_operators`

Lists available OLM operators.

```hcl
data "oai_supported_operators" "all" {}

# Filter operators
locals {
  storage_operators = [
    for op in data.oai_supported_operators.all.operators :
    op if contains(["lso", "odf", "ceph"], op)
  ]
}
```

#### `oai_operator_bundles`

Retrieves operator bundle information.

```hcl
data "oai_operator_bundles" "virtualization" {}
```

#### `oai_support_levels`

Checks feature support levels by platform and architecture.

```hcl
data "oai_support_levels" "features" {
  openshift_version = "4.16.0"
  cpu_architecture  = "x86_64"
  platform_type     = "baremetal"
}
```

### Post-Installation Access

#### `oai_cluster_credentials`

Retrieves cluster admin credentials after installation.

```hcl
data "oai_cluster_credentials" "admin" {
  cluster_id = oai_cluster.example.id
  depends_on = [oai_cluster_installation.example]
}

output "cluster_access" {
  value = {
    username    = data.oai_cluster_credentials.admin.username
    password    = data.oai_cluster_credentials.admin.password
    console_url = data.oai_cluster_credentials.admin.console_url
  }
  sensitive = true
}
```

#### `oai_cluster_events`

Retrieves installation and cluster events for monitoring.

```hcl
data "oai_cluster_events" "installation" {
  cluster_id = oai_cluster.example.id
  severities = ["error", "critical"]
  limit      = 100
  order      = "desc"
}

# Monitor installation progress
output "recent_errors" {
  value = [
    for event in data.oai_cluster_events.installation.events :
    "${event.event_time}: ${event.message}"
    if contains(["error", "critical"], event.severity)
  ]
}
```

#### `oai_cluster_logs`

Downloads cluster logs for troubleshooting.

```hcl
data "oai_cluster_logs" "installation" {
  cluster_id = oai_cluster.example.id
  logs_type  = "controller"
}

# Save logs locally
resource "local_file" "logs" {
  content  = data.oai_cluster_logs.installation.content
  filename = "${path.module}/installation.log"
}
```

#### `oai_cluster_files`

Downloads cluster configuration files.

```hcl
# Download kubeconfig
data "oai_cluster_files" "kubeconfig" {
  cluster_id = oai_cluster.example.id
  file_name  = "kubeconfig"
}

# Download install configuration
data "oai_cluster_files" "install_config" {
  cluster_id = oai_cluster.example.id
  file_name  = "install-config.yaml"
}

# Save kubeconfig locally
resource "local_file" "kubeconfig" {
  content  = data.oai_cluster_files.kubeconfig.content
  filename = "${path.module}/kubeconfig"
}
```

**Available Files:**
- `kubeconfig` - Kubernetes configuration for cluster access
- `kubeconfig-noingress` - Kubeconfig without ingress routing
- `kubeadmin-password` - Admin password file
- `install-config.yaml` - Installation configuration
- `bootstrap.ign` - Bootstrap ignition configuration
- `master.ign` - Master node ignition configuration
- `worker.ign` - Worker node ignition configuration
- `metadata.json` - Cluster metadata
- `manifests` - Applied manifests

### Validation Data Sources

#### `oai_cluster_validations`

Checks cluster-level validation status for pre-installation readiness.

```hcl
data "oai_cluster_validations" "ready" {
  cluster_id        = oai_cluster.example.id
  validation_type   = "blocking"
  status            = "failure"
}

# Ensure cluster is ready before installation
locals {
  cluster_ready = length(data.oai_cluster_validations.ready.validations) == 0
}
```

#### `oai_host_validations`

Validates host readiness for cluster installation.

```hcl
data "oai_host_validations" "hosts" {
  cluster_id = oai_cluster.example.id
  categories = ["hardware", "network"]
}

# Check if all hosts pass validation
locals {
  hosts_ready = alltrue([
    for host in data.oai_host_validations.hosts.validations :
    alltrue([for v in host.validations : v.status == "success"])
  ])
}
```

### Resource Information Data Sources

#### `oai_cluster`

Retrieves information about an existing cluster.

```hcl
data "oai_cluster" "existing" {
  cluster_id = var.cluster_id
}

output "cluster_status" {
  value = data.oai_cluster.existing.status
}
```

#### `oai_infra_env`

Gets infrastructure environment details including discovery ISO URL.

```hcl
data "oai_infra_env" "discovery" {
  infra_env_id = oai_infra_env.example.id
}

output "iso_download" {
  value = data.oai_infra_env.discovery.download_url
}
```

#### `oai_host`

Retrieves discovered host information.

```hcl
data "oai_host" "master1" {
  infra_env_id = oai_infra_env.example.id
  host_id      = var.master1_host_id
}

output "host_inventory" {
  value = data.oai_host.master1.inventory
}
```

#### `oai_manifest`

Downloads manifest content from a cluster.

```hcl
data "oai_manifest" "custom" {
  cluster_id = oai_cluster.example.id
  file_name  = "99-custom-config.yaml"
  folder     = "manifests"
}

output "manifest_content" {
  value = data.oai_manifest.custom.content
}
```

## Workflow Patterns

### Complete Cluster Lifecycle

```hcl
# 1. Define cluster
resource "oai_cluster" "production" {
  name                = "prod-cluster"
  openshift_version   = "4.16.0"
  cpu_architecture    = "x86_64" 
  control_plane_count = 3
  
  pull_secret    = var.pull_secret
  ssh_public_key = var.ssh_public_key
  
  # Production network config
  api_vips = [{ ip = "10.0.1.100" }]
  ingress_vips = [{ ip = "10.0.1.101" }]
}

# 2. Create discovery environment
resource "oai_infra_env" "production" {
  name         = "${oai_cluster.production.name}-infra"
  cluster_id   = oai_cluster.production.id
  pull_secret  = var.pull_secret
  image_type   = "full-iso"
}

# 3. Apply custom manifests
resource "oai_manifest" "monitoring" {
  cluster_id = oai_cluster.production.id
  file_name  = "monitoring-config.yaml"
  folder     = "manifests"
  
  content = templatefile("${path.module}/monitoring.yaml", {
    retention_days = 30
  })
}

# 4. Trigger installation
resource "oai_cluster_installation" "production" {
  cluster_id          = oai_cluster.production.id
  wait_for_hosts      = true
  expected_host_count = 6  # 3 masters + 3 workers
  
  timeouts {
    create = "180m"
  }
}

# 5. Access cluster post-installation
data "oai_cluster_credentials" "admin" {
  cluster_id = oai_cluster.production.id
  depends_on = [oai_cluster_installation.production]
}

data "oai_cluster_files" "kubeconfig" {
  cluster_id = oai_cluster.production.id
  file_name  = "kubeconfig"
  depends_on = [oai_cluster_installation.production]
}

# 6. Monitor and troubleshoot
data "oai_cluster_events" "health" {
  cluster_id = oai_cluster.production.id
  severities = ["warning", "error", "critical"]
}
```

### Error Handling and Monitoring

```hcl
# Monitor installation health
locals {
  installation_errors = [
    for event in data.oai_cluster_events.health.events :
    event if contains(["error", "critical"], event.severity)
  ]
  
  has_critical_errors = length(local.installation_errors) > 0
}

# Conditional outputs based on health
output "installation_status" {
  value = local.has_critical_errors ? "FAILED" : "SUCCESS"
}

output "troubleshooting_guide" {
  value = local.has_critical_errors ? {
    errors = local.installation_errors
    logs_command = "terraform output -raw cluster_logs > debug.log"
    support_info = "Check recent errors and review installation logs"
  } : null
}
```

## Best Practices

### Security
- Store pull secrets and offline tokens securely (use Terraform Cloud variables or external secret management)
- Mark sensitive outputs appropriately
- Use dedicated service accounts with minimal required permissions

### Resource Management
- Use `depends_on` for proper resource ordering
- Configure appropriate timeouts for long-running operations
- Plan for cluster lifecycle management (updates, scaling, deletion)

### Monitoring and Troubleshooting
- Always include post-installation monitoring with event data sources
- Download and store kubeconfig files for operational access
- Implement health checks and alerting based on installation events

### Development Workflow
- Use separate environments for development, staging, and production
- Version control your Terraform configurations
- Test configurations with smaller clusters before production deployment

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   - Verify offline token is valid and not expired
   - Check token has necessary permissions

2. **Host Discovery Issues**
   - Verify network connectivity from hosts to API endpoint
   - Check discovery ISO was generated correctly
   - Ensure hosts meet minimum requirements

3. **Installation Timeouts**
   - Review cluster events for specific error messages
   - Check network connectivity and DNS resolution
   - Verify hardware meets OpenShift requirements

4. **Post-Installation Access**
   - Ensure installation completed successfully before accessing credentials
   - Check kubeconfig file permissions and content
   - Verify network connectivity to cluster API

### Debug Commands

```bash
# View detailed installation events
terraform output -json installation_events

# Download cluster logs
terraform output -raw cluster_logs > installation.log

# Export kubeconfig
export KUBECONFIG=$(terraform output -raw kubeconfig_path)
oc whoami
```