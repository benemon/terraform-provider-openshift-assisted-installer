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

# Example variable for a cluster that has already been installed
variable "cluster_id" {
  description = "The ID of an installed cluster"
  type        = string
}

# ==============================================================================
# Post-Installation Data Sources Examples
# ==============================================================================

# Get cluster admin credentials after installation
data "openshift_assisted_installer_cluster_credentials" "admin" {
  cluster_id = var.cluster_id
}

# Get cluster events for monitoring and troubleshooting
data "openshift_assisted_installer_cluster_events" "cluster_events" {
  cluster_id = var.cluster_id
  severities = ["error", "critical"]  # Filter for important events
  limit      = 100
  order      = "desc"                 # Most recent first
}

# Get installation events for all clusters
data "openshift_assisted_installer_cluster_events" "all_events" {
  # No cluster_id filter - gets events from all clusters
  severities     = ["info", "warning", "error"]
  categories     = ["user"]
  cluster_level  = true
  limit          = 50
}

# Download cluster logs for troubleshooting
data "openshift_assisted_installer_cluster_logs" "cluster_logs" {
  cluster_id = var.cluster_id
  logs_type  = "controller"  # Optional: specify log type
}

# Download specific cluster files
data "openshift_assisted_installer_cluster_files" "kubeconfig" {
  cluster_id = var.cluster_id
  file_name  = "kubeconfig"
}

data "openshift_assisted_installer_cluster_files" "install_config" {
  cluster_id = var.cluster_id
  file_name  = "install-config.yaml"
}

data "openshift_assisted_installer_cluster_files" "manifests" {
  cluster_id = var.cluster_id
  file_name  = "manifests"
}

# ==============================================================================
# Outputs
# ==============================================================================

output "cluster_access" {
  description = "Cluster access information"
  value = {
    username    = data.openshift_assisted_installer_cluster_credentials.admin.username
    password    = data.openshift_assisted_installer_cluster_credentials.admin.password  # Sensitive
    console_url = data.openshift_assisted_installer_cluster_credentials.admin.console_url
  }
  sensitive = true
}

output "recent_errors" {
  description = "Recent error events in the cluster"
  value = [for event in data.openshift_assisted_installer_cluster_events.cluster_events.events : {
    time     = event.event_time
    severity = event.severity
    message  = event.message
  } if contains(["error", "critical"], event.severity)]
}

output "installation_summary" {
  description = "Summary of installation progress"
  value = {
    total_events = length(data.openshift_assisted_installer_cluster_events.cluster_events.events)
    error_count  = length([for event in data.openshift_assisted_installer_cluster_events.cluster_events.events : event if event.severity == "error"])
    warning_count = length([for event in data.openshift_assisted_installer_cluster_events.cluster_events.events : event if event.severity == "warning"])
  }
}

# Save kubeconfig to local file (optional)
resource "local_file" "kubeconfig" {
  content  = data.openshift_assisted_installer_cluster_files.kubeconfig.content
  filename = "${path.module}/kubeconfig-${var.cluster_id}"
}

# ==============================================================================
# Usage Examples
# ==============================================================================

# Example 1: Check cluster health after installation
# terraform apply -var="cluster_id=your-cluster-uuid"

# Example 2: Monitor installation progress
# while terraform apply -var="cluster_id=your-cluster-uuid"; do
#   sleep 30
# done

# Example 3: Export cluster credentials
# export KUBECONFIG=$(terraform output -raw kubeconfig_path)
# oc whoami

# Example 4: Troubleshoot installation issues
# Review the recent_errors output to identify problems
# Check the installation_summary for overall health

# ==============================================================================
# Prerequisites
# ==============================================================================
# 1. Set OFFLINE_TOKEN environment variable with your Red Hat token
# 2. Ensure the cluster specified by cluster_id exists and is installed
# 3. Have appropriate permissions to access cluster credentials and logs