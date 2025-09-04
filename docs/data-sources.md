# Data Sources Reference

## Cluster Information Data Sources

### `oai_openshift_versions`

Retrieves available OpenShift versions and their metadata.

```hcl
data "oai_openshift_versions" "all" {
  # Optional filters
  only_latest = false  # Set to true to get only latest versions
}

# Access version information
output "available_versions" {
  value = [
    for version in data.oai_openshift_versions.all.versions : {
      version      = version.version
      display_name = version.display_name
      support_level = version.support_level
      cpu_architectures = version.cpu_architectures
    }
  ]
}
```

**Attributes:**
- `versions` - List of available versions with metadata
  - `version` - Version string (e.g., "4.15.20")
  - `display_name` - Human-readable name
  - `support_level` - Support level (production, dev-preview, etc.)
  - `cpu_architectures` - Supported CPU architectures
  - `default` - Whether this is the default version

### `oai_supported_operators`

Lists OLM operators available for cluster installation.

```hcl
data "oai_supported_operators" "all" {}

# Filter operators by name pattern
locals {
  storage_operators = [
    for op in data.oai_supported_operators.all.operators :
    op if can(regex("storage|lso|odf|ceph", op))
  ]
}
```

**Attributes:**
- `operators` - List of operator names available for installation

### `oai_operator_bundles`

Retrieves information about operator bundles.

```hcl
data "oai_operator_bundles" "all" {
  # Optional: specific bundle
  bundle_id = "virtualization-bundle"
}
```

### `oai_support_levels`

Checks feature support levels by platform and architecture.

```hcl
data "oai_support_levels" "baremetal" {
  openshift_version = "4.15.20"
  cpu_architecture  = "x86_64"
  platform_type     = "baremetal"
}

# Check if a feature is supported
locals {
  sno_supported = lookup(data.oai_support_levels.baremetal.features, "SNO", "unavailable") != "unavailable"
}
```

**Arguments:**
- `openshift_version` (Required) - OpenShift version to check
- `cpu_architecture` (Optional) - CPU architecture filter
- `platform_type` (Optional) - Platform type filter

**Attributes:**
- `features` - Map of feature names to support levels

## Post-Installation Data Sources

### `oai_cluster_credentials`

**Purpose:** Retrieve admin credentials after successful cluster installation.

```hcl
data "oai_cluster_credentials" "admin" {
  cluster_id = oai_cluster.example.id
  
  # Ensure installation is complete
  depends_on = [oai_cluster_installation.example]
}

# Use credentials in outputs or other resources
resource "kubernetes_secret" "admin_access" {
  metadata {
    name = "cluster-admin-creds"
  }
  
  data = {
    username = data.oai_cluster_credentials.admin.username
    password = data.oai_cluster_credentials.admin.password
  }
  
  type = "Opaque"
}
```

**Arguments:**
- `cluster_id` (Required) - ID of the installed cluster

**Attributes:**
- `username` - Admin username (typically "kubeadmin")
- `password` - Admin password (sensitive)
- `console_url` - OpenShift web console URL

**Important Notes:**
- Only available after cluster installation completes
- Password is marked as sensitive in Terraform state
- Use `depends_on` to ensure installation finishes first

### `oai_cluster_events`

**Purpose:** Monitor installation progress and troubleshoot issues through cluster events.

```hcl
# Get all recent events
data "oai_cluster_events" "all" {
  cluster_id = oai_cluster.example.id
  limit      = 200
  order      = "desc"  # Most recent first
}

# Filter for errors only
data "oai_cluster_events" "errors" {
  cluster_id = oai_cluster.example.id
  severities = ["error", "critical"]
  limit      = 50
}

# Monitor specific host events
data "oai_cluster_events" "host_events" {
  cluster_id = oai_cluster.example.id
  host_id    = "specific-host-id"
}

# Get events across all clusters (no cluster_id filter)
data "oai_cluster_events" "global" {
  severities    = ["critical"]
  cluster_level = true
}
```

**Arguments:**
- `cluster_id` (Optional) - Filter events by cluster ID
- `host_id` (Optional) - Filter events by host ID  
- `infra_env_id` (Optional) - Filter events by infrastructure environment
- `severities` (Optional) - List of severities: ["info", "warning", "error", "critical"]
- `categories` (Optional) - List of categories: ["user", "metrics"]
- `message` (Optional) - Filter by message pattern
- `order` (Optional) - Order by event_time: "asc" or "desc"
- `limit` (Optional) - Maximum number of events to return
- `offset` (Optional) - Number of events to skip
- `cluster_level` (Optional) - Include cluster-level events

**Attributes:**
- `events` - List of events matching filter criteria
  - `name` - Event name
  - `cluster_id` - Associated cluster ID
  - `host_id` - Associated host ID (if applicable)
  - `infra_env_id` - Associated infrastructure environment ID
  - `severity` - Event severity (info, warning, error, critical)
  - `category` - Event category (user, metrics)
  - `message` - Human-readable event message
  - `event_time` - Timestamp when event occurred
  - `request_id` - Request ID that triggered the event
  - `props` - Additional event properties (JSON string)

**Common Use Cases:**

```hcl
# Monitor installation progress
locals {
  installation_progress = {
    total_events = length(data.oai_cluster_events.all.events)
    errors       = [for e in data.oai_cluster_events.all.events : e if e.severity == "error"]
    warnings     = [for e in data.oai_cluster_events.all.events : e if e.severity == "warning"]
    latest_event = length(data.oai_cluster_events.all.events) > 0 ? data.oai_cluster_events.all.events[0] : null
  }
}

# Troubleshooting output
output "troubleshooting" {
  value = length(local.installation_progress.errors) > 0 ? {
    status = "Issues detected"
    errors = [for e in local.installation_progress.errors : "${e.event_time}: ${e.message}"]
  } : {
    status = "Installation proceeding normally"
    latest = local.installation_progress.latest_event != null ? local.installation_progress.latest_event.message : "No events"
  }
}
```

### `oai_cluster_logs`

**Purpose:** Download cluster logs for detailed troubleshooting and analysis.

```hcl
# Get controller logs
data "oai_cluster_logs" "controller" {
  cluster_id = oai_cluster.example.id
  logs_type  = "controller"
}

# Get logs for specific host
data "oai_cluster_logs" "host" {
  cluster_id = oai_cluster.example.id
  logs_type  = "host"
  host_id    = "specific-host-id"
}

# Save logs to file for analysis
resource "local_file" "installation_logs" {
  content  = data.oai_cluster_logs.controller.content
  filename = "${path.module}/logs/installation-${formatdate("YYYY-MM-DD", timestamp())}.log"
}
```

**Arguments:**
- `cluster_id` (Required) - ID of the cluster to download logs for
- `logs_type` (Optional) - Type of logs to download (controller, host, etc.)
- `host_id` (Optional) - Specific host ID when downloading host logs

**Attributes:**
- `content` - Raw log content as a string

### `oai_cluster_files`

**Purpose:** Download cluster configuration files and certificates.

```hcl
# Download kubeconfig
data "oai_cluster_files" "kubeconfig" {
  cluster_id = oai_cluster.example.id
  file_name  = "kubeconfig"
}

# Download kubeconfig without ingress
data "oai_cluster_files" "kubeconfig_noingress" {
  cluster_id = oai_cluster.example.id
  file_name  = "kubeconfig-noingress"
}

# Download installation configuration
data "oai_cluster_files" "install_config" {
  cluster_id = oai_cluster.example.id
  file_name  = "install-config.yaml"
}

# Download manifests
data "oai_cluster_files" "manifests" {
  cluster_id = oai_cluster.example.id
  file_name  = "manifests"
}

# Download ignition configs
data "oai_cluster_files" "bootstrap_ign" {
  cluster_id = oai_cluster.example.id
  file_name  = "bootstrap.ign"
}

# Save files locally
resource "local_file" "kubeconfig" {
  content         = data.oai_cluster_files.kubeconfig.content
  filename        = "${path.module}/kubeconfig"
  file_permission = "0600"  # Secure permissions
}

resource "local_file" "install_config" {
  content  = data.oai_cluster_files.install_config.content
  filename = "${path.module}/install-config.yaml"
}
```

**Arguments:**
- `cluster_id` (Required) - ID of the cluster to download files from
- `file_name` (Required) - Name of the file to download
- `logs_type` (Optional) - Used when file_name is "logs" to specify log type

**Available Files:**
- `kubeconfig` - Standard kubeconfig for cluster access
- `kubeconfig-noingress` - Kubeconfig without ingress routing
- `kubeadmin-password` - Admin password file
- `install-config.yaml` - Installation configuration used
- `bootstrap.ign` - Bootstrap node ignition configuration
- `master.ign` - Master nodes ignition configuration  
- `worker.ign` - Worker nodes ignition configuration
- `metadata.json` - Cluster metadata and information
- `manifests` - All applied manifests
- `logs` - Combined log files (requires logs_type parameter)

**Attributes:**
- `content` - Raw file content as a string

**Usage Patterns:**

```hcl
# Create complete cluster access package
locals {
  cluster_files = {
    kubeconfig     = data.oai_cluster_files.kubeconfig.content
    install_config = data.oai_cluster_files.install_config.content  
    admin_password = data.oai_cluster_credentials.admin.password
    console_url    = data.oai_cluster_credentials.admin.console_url
  }
}

# Export as archive
data "archive_file" "cluster_access" {
  type        = "zip"
  output_path = "${path.module}/cluster-access-${oai_cluster.example.name}.zip"
  
  source {
    content  = local.cluster_files.kubeconfig
    filename = "kubeconfig"
  }
  
  source {
    content  = local.cluster_files.install_config
    filename = "install-config.yaml"
  }
  
  source {
    content  = "Username: ${data.oai_cluster_credentials.admin.username}\nPassword: ${local.cluster_files.admin_password}\nConsole: ${local.cluster_files.console_url}"
    filename = "access-info.txt"
  }
}
```

## Best Practices

### Timing and Dependencies
- Always use `depends_on` with post-installation data sources to ensure installation completes
- Use event monitoring during installation to track progress
- Download files immediately after installation for backup purposes

### Security
- Mark credential outputs as sensitive
- Store kubeconfig files with secure permissions (0600)
- Use Terraform remote state encryption for sensitive data

### Error Handling
- Monitor events with appropriate severity filters
- Implement conditional logic based on installation status
- Provide clear troubleshooting outputs for operators

### Performance
- Use appropriate limits on event queries to avoid large datasets
- Cache downloaded files locally when possible
- Filter events by time ranges for historical analysis