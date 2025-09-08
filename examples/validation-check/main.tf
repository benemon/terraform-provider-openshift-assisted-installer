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

# ==============================================================================
# Validation Checking Examples
# ==============================================================================

# Example: Check cluster-level validations for readiness
data "openshift_assisted_installer_cluster_validations" "readiness_check" {
  cluster_id = var.cluster_id
  
  # Only check blocking validations that would prevent installation
  validation_types = ["blocking"]
  
  # Only show failures and pending validations
  status_filter = ["failure", "pending"]
}

# Example: Check host-level validations for all hosts in cluster
data "openshift_assisted_installer_host_validations" "cluster_hosts" {
  cluster_id = var.cluster_id
  
  # Focus on blocking validations
  validation_types = ["blocking"]
  
  # Show failures and pending validations
  status_filter = ["failure", "pending"]
}

# Example: Check specific operator requirements
data "openshift_assisted_installer_host_validations" "storage_requirements" {
  cluster_id = var.cluster_id
  
  # Filter for storage operator requirements only
  validation_names = [
    "lso-requirements-satisfied",
    "odf-requirements-satisfied", 
    "has-min-valid-disks"
  ]
}

# Example: Check network-related validations
data "openshift_assisted_installer_cluster_validations" "network_readiness" {
  cluster_id = var.cluster_id
  
  # Filter by network category
  categories = ["network"]
}

# Example: Check hardware readiness for specific host
data "openshift_assisted_installer_host_validations" "specific_host" {
  infra_env_id = var.infra_env_id
  host_id      = var.host_id
  
  # Focus on hardware-related validations
  categories = ["hardware"]
}

# ==============================================================================
# Local computations for validation analysis
# ==============================================================================

locals {
  # Count blocking cluster failures
  cluster_blocking_failures = [
    for validation in data.openshift_assisted_installer_cluster_validations.readiness_check.validations :
    validation if validation.status == "failure" && validation.validation_type == "blocking"
  ]
  
  # Count blocking host failures
  host_blocking_failures = [
    for validation in data.openshift_assisted_installer_host_validations.cluster_hosts.validations :
    validation if validation.status == "failure" && validation.validation_type == "blocking"
  ]
  
  # Check storage validation status
  storage_validation_failures = [
    for validation in data.openshift_assisted_installer_host_validations.storage_requirements.validations :
    validation if validation.status == "failure"
  ]
  
  # Overall readiness assessment
  cluster_ready = length(local.cluster_blocking_failures) == 0
  hosts_ready   = length(local.host_blocking_failures) == 0
  storage_ready = length(local.storage_validation_failures) == 0
  
  # Installation readiness flag
  installation_ready = local.cluster_ready && local.hosts_ready && local.storage_ready
}

# ==============================================================================
# Outputs for validation status reporting
# ==============================================================================

output "validation_summary" {
  description = "Comprehensive validation status summary"
  value = {
    cluster_ready = local.cluster_ready
    hosts_ready   = local.hosts_ready
    storage_ready = local.storage_ready
    
    installation_ready = local.installation_ready
    
    # Failure details
    cluster_failures = [
      for failure in local.cluster_blocking_failures :
      {
        validation_id = failure.validation_id
        message       = failure.message
        category      = failure.category
      }
    ]
    
    host_failures = [
      for failure in local.host_blocking_failures :
      {
        host_id       = failure.host_id
        validation_id = failure.validation_id
        message       = failure.message
        category      = failure.category
      }
    ]
    
    storage_failures = [
      for failure in local.storage_validation_failures :
      {
        host_id       = failure.host_id
        validation_id = failure.validation_id
        message       = failure.message
      }
    ]
  }
}

output "installation_readiness" {
  description = "Simple installation readiness indicator"
  value = {
    ready = local.installation_ready
    
    message = local.installation_ready ? 
      "✅ Cluster is ready for installation" : 
      "❌ Cluster has validation failures - installation blocked"
    
    blocking_issues_count = (
      length(local.cluster_blocking_failures) + 
      length(local.host_blocking_failures) + 
      length(local.storage_validation_failures)
    )
  }
}

output "troubleshooting_guide" {
  description = "Validation failure troubleshooting information"
  value = !local.installation_ready ? {
    message = "Installation blocked by validation failures"
    
    cluster_issues = [
      for failure in local.cluster_blocking_failures :
      "❌ Cluster: ${failure.validation_id} - ${failure.message}"
    ]
    
    host_issues = [
      for failure in local.host_blocking_failures :
      "❌ Host ${failure.host_id}: ${failure.validation_id} - ${failure.message}"
    ]
    
    storage_issues = [
      for failure in local.storage_validation_failures :
      "❌ Storage on Host ${failure.host_id}: ${failure.validation_id} - ${failure.message}"
    ]
    
    next_steps = [
      "1. Review validation failures listed above",
      "2. Fix identified issues (hardware, network, configuration)",
      "3. Wait for validation status to update (hosts re-validate automatically)",
      "4. Re-run 'terraform plan/apply' when ready"
    ]
  } : {
    message = "✅ All critical validations passed - installation can proceed"
    next_steps = [
      "1. Proceed with cluster installation",
      "2. Monitor installation progress through cluster events"
    ]
  }
}

# ==============================================================================
# Network validation breakdown
# ==============================================================================

output "network_validation_details" {
  description = "Detailed network validation status"
  value = {
    validations = [
      for validation in data.openshift_assisted_installer_cluster_validations.network_readiness.validations :
      {
        id       = validation.validation_id
        status   = validation.status
        message  = validation.message
        type     = validation.validation_type
        blocking = validation.validation_type == "blocking"
      }
    ]
    
    # Network readiness summary
    total_network_validations = length(data.openshift_assisted_installer_cluster_validations.network_readiness.validations)
    
    passing_network_validations = length([
      for v in data.openshift_assisted_installer_cluster_validations.network_readiness.validations :
      v if v.status == "success"
    ])
    
    failing_network_validations = length([
      for v in data.openshift_assisted_installer_cluster_validations.network_readiness.validations :
      v if v.status == "failure"
    ])
    
    network_ready = length([
      for v in data.openshift_assisted_installer_cluster_validations.network_readiness.validations :
      v if v.status == "failure" && v.validation_type == "blocking"
    ]) == 0
  }
}

# ==============================================================================
# Variables
# ==============================================================================

variable "cluster_id" {
  description = "The ID of the cluster to check validations for"
  type        = string
}

variable "infra_env_id" {
  description = "Infrastructure environment ID (optional, for specific host checks)"
  type        = string
  default     = ""
}

variable "host_id" {
  description = "Specific host ID to check (optional, requires infra_env_id)"  
  type        = string
  default     = ""
}