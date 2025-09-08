# ==============================================================================
# Module 2: Cluster Installation
# Triggers and monitors the cluster installation
# Run this AFTER hosts have been discovered from the ISO
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

# Input from the setup module
variable "cluster_id" {
  description = "Cluster ID from the setup module"
  type        = string
}

# Trigger cluster installation
resource "openshift_assisted_installer_cluster_installation" "sno" {
  cluster_id = var.cluster_id
  
  # For SNO, we expect 1 host
  wait_for_hosts      = true
  expected_host_count = 1
  
  timeouts {
    create = "90m"  # Allow up to 90 minutes for installation
  }
}

# Outputs
output "installation_status" {
  value = {
    status              = openshift_assisted_installer_cluster_installation.sno.status
    status_info        = openshift_assisted_installer_cluster_installation.sno.status_info
    started_at         = openshift_assisted_installer_cluster_installation.sno.install_started_at
    completed_at       = openshift_assisted_installer_cluster_installation.sno.install_completed_at
  }
}

# Get cluster credentials after installation completes
data "openshift_assisted_installer_cluster_credentials" "admin" {
  cluster_id = var.cluster_id
  depends_on = [openshift_assisted_installer_cluster_installation.sno]
}

# Download kubeconfig for local access
data "openshift_assisted_installer_cluster_files" "kubeconfig" {
  cluster_id = var.cluster_id  
  file_name  = "kubeconfig"
  depends_on = [openshift_assisted_installer_cluster_installation.sno]
}

# Monitor installation events
data "openshift_assisted_installer_cluster_events" "installation_progress" {
  cluster_id = var.cluster_id
  severities = ["info", "warning", "error", "critical"]
  limit      = 100
  order      = "desc"
}

# Save kubeconfig to local file
resource "local_file" "kubeconfig" {
  content  = data.openshift_assisted_installer_cluster_files.kubeconfig.content
  filename = "${path.module}/kubeconfig-sno"
  
  depends_on = [openshift_assisted_installer_cluster_installation.sno]
}

output "cluster_access" {
  description = "Cluster access information and credentials"
  value = {
    # URLs
    console_url = data.openshift_assisted_installer_cluster_credentials.admin.console_url
    api_url     = "https://api.sno-cluster.example.com:6443"
    
    # Credentials (sensitive)
    username = data.openshift_assisted_installer_cluster_credentials.admin.username
    password = data.openshift_assisted_installer_cluster_credentials.admin.password
    
    # Local files
    kubeconfig_path = local_file.kubeconfig.filename
    
    # Usage instructions
    message = "Cluster installation complete! Use 'export KUBECONFIG=${local_file.kubeconfig.filename}' then 'oc whoami' to verify access."
  }
  sensitive = true
}

output "installation_summary" {
  description = "Installation progress and health summary"
  value = {
    status              = openshift_assisted_installer_cluster_installation.sno.status
    status_info         = openshift_assisted_installer_cluster_installation.sno.status_info  
    install_started_at  = openshift_assisted_installer_cluster_installation.sno.install_started_at
    install_completed_at = openshift_assisted_installer_cluster_installation.sno.install_completed_at
    
    # Event summary
    total_events = length(data.openshift_assisted_installer_cluster_events.installation_progress.events)
    error_events = length([for event in data.openshift_assisted_installer_cluster_events.installation_progress.events : event if event.severity == "error"])
    warning_events = length([for event in data.openshift_assisted_installer_cluster_events.installation_progress.events : event if event.severity == "warning"])
  }
}

# ==============================================================================
# Usage Instructions
# ==============================================================================

# Step 1: First run the cluster setup
# $ terraform apply -target=module.cluster_setup

# Step 2: Boot your machine from the ISO
# Download the ISO URL provided in the setup output and boot your machine

# Step 3: Wait for host discovery
# Check the OpenShift console or API to ensure host is discovered

# Step 4: Run this installation module
# $ terraform apply -target=module.cluster_installation -var="cluster_id=<cluster-id-from-setup>"

# The installation will:
# - Wait for the required number of hosts
# - Trigger the OpenShift installation
# - Monitor progress until completion
# - Provide access information when done