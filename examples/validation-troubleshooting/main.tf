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

# ==============================================================================
# Comprehensive Validation Troubleshooting Examples
# ==============================================================================

# Example 1: Basic cluster readiness check
data "oai_cluster_validations" "basic_readiness" {
  cluster_id = var.cluster_id
  
  # Only show blocking failures that prevent installation
  validation_types = ["blocking"]
  status_filter    = ["failure"]
}

# Example 2: All cluster validation status (comprehensive view)
data "oai_cluster_validations" "comprehensive" {
  cluster_id = var.cluster_id
}

# Example 3: Network-specific troubleshooting
data "oai_cluster_validations" "network_cluster" {
  cluster_id = var.cluster_id
  categories = ["network"]
}

data "oai_host_validations" "network_hosts" {
  cluster_id = var.cluster_id
  categories = ["network"]
}

# Example 4: Hardware validation across all hosts
data "oai_host_validations" "hardware_check" {
  cluster_id = var.cluster_id
  categories = ["hardware"]
}

# Example 5: Operator requirements validation
data "oai_host_validations" "operator_requirements" {
  cluster_id = var.cluster_id
  validation_names = [
    "lso-requirements-satisfied",
    "odf-requirements-satisfied",
    "cnv-requirements-satisfied", 
    "lvm-requirements-satisfied"
  ]
}

# Example 6: Specific host deep-dive (when you have a problematic host)
data "oai_host_validations" "specific_host_debug" {
  # Use this when troubleshooting a specific host
  count = var.debug_host_id != "" ? 1 : 0
  
  infra_env_id = var.infra_env_id
  host_id      = var.debug_host_id
}

# Example 7: Platform-specific validations (e.g., vSphere)
data "oai_host_validations" "platform_validations" {
  cluster_id = var.cluster_id
  categories = ["platform"]
}

# Example 8: Storage and disk validations
data "oai_host_validations" "storage_validations" {
  cluster_id = var.cluster_id
  categories = ["storage"]
}

# ==============================================================================
# Validation Analysis and Reporting
# ==============================================================================

locals {
  # Basic readiness assessment
  has_blocking_cluster_failures = length(data.oai_cluster_validations.basic_readiness.validations) > 0
  
  # Comprehensive analysis
  all_cluster_validations = data.oai_cluster_validations.comprehensive.validations
  all_host_validations    = data.oai_host_validations.hardware_check.validations
  
  # Network analysis
  network_cluster_issues = [
    for v in data.oai_cluster_validations.network_cluster.validations :
    v if v.status == "failure"
  ]
  
  network_host_issues = [
    for v in data.oai_host_validations.network_hosts.validations :
    v if v.status == "failure"
  ]
  
  # Hardware analysis
  hardware_issues = [
    for v in data.oai_host_validations.hardware_check.validations :
    v if v.status == "failure"
  ]
  
  # Group hardware issues by host
  hardware_issues_by_host = {
    for v in local.hardware_issues :
    v.host_id => v...
  }
  
  # Operator requirements analysis
  operator_failures = [
    for v in data.oai_host_validations.operator_requirements.validations :
    v if v.status == "failure"
  ]
  
  # Group by operator type
  lso_failures = [
    for v in local.operator_failures :
    v if can(regex("lso", v.validation_id))
  ]
  
  odf_failures = [
    for v in local.operator_failures :
    v if can(regex("odf", v.validation_id))
  ]
  
  cnv_failures = [
    for v in local.operator_failures :
    v if can(regex("cnv", v.validation_id))
  ]
  
  # Platform issues (e.g., vSphere-specific)
  platform_issues = [
    for v in data.oai_host_validations.platform_validations.validations :
    v if v.status == "failure"
  ]
  
  # Storage issues
  storage_issues = [
    for v in data.oai_host_validations.storage_validations.validations :
    v if v.status == "failure"
  ]
  
  # Overall assessment
  cluster_validation_summary = {
    total_validations = length(local.all_cluster_validations)
    success_count     = length([for v in local.all_cluster_validations : v if v.status == "success"])
    failure_count     = length([for v in local.all_cluster_validations : v if v.status == "failure"])
    pending_count     = length([for v in local.all_cluster_validations : v if v.status == "pending"])
    disabled_count    = length([for v in local.all_cluster_validations : v if v.status == "disabled"])
  }
  
  # Host-level summary
  host_validation_summary = {
    total_validations = length(local.all_host_validations)
    success_count     = length([for v in local.all_host_validations : v if v.status == "success"])
    failure_count     = length([for v in local.all_host_validations : v if v.status == "failure"])
    pending_count     = length([for v in local.all_host_validations : v if v.status == "pending"])
    
    # Unique hosts with validations
    unique_hosts = distinct([for v in local.all_host_validations : v.host_id])
    hosts_with_failures = distinct([for v in local.all_host_validations : v.host_id if v.status == "failure"])
  }
}

# ==============================================================================
# Detailed Troubleshooting Outputs
# ==============================================================================

output "validation_health_check" {
  description = "High-level validation health assessment"
  value = {
    cluster_ready = !local.has_blocking_cluster_failures
    
    cluster_summary = local.cluster_validation_summary
    host_summary    = local.host_validation_summary
    
    critical_issues = local.has_blocking_cluster_failures ? 
      "❌ Found blocking cluster validation failures" : 
      "✅ No blocking cluster validation failures"
    
    recommendations = local.has_blocking_cluster_failures ? [
      "Review cluster validation failures in detailed outputs below",
      "Fix cluster-level configuration issues", 
      "Ensure network VIPs and CIDRs are properly configured",
      "Verify operator requirements are met"
    ] : [
      "Cluster validations look good",
      "Check host validations for any remaining issues",
      "Proceed with installation if all hosts are ready"
    ]
  }
}

output "network_troubleshooting" {
  description = "Network validation analysis"
  value = {
    cluster_network_issues = [
      for issue in local.network_cluster_issues : {
        validation  = issue.validation_id
        status     = issue.status
        message    = issue.message
        category   = issue.category
        type       = issue.validation_type
      }
    ]
    
    host_network_issues = [
      for issue in local.network_host_issues : {
        host_id    = issue.host_id
        validation = issue.validation_id
        status     = issue.status
        message    = issue.message
        category   = issue.category
        type       = issue.validation_type
      }
    ]
    
    network_ready = length(local.network_cluster_issues) == 0 && length(local.network_host_issues) == 0
    
    troubleshooting_tips = {
      api_vips = "Ensure API VIPs are properly configured and reachable"
      ingress_vips = "Ensure Ingress VIPs are properly configured and not conflicting"
      machine_cidr = "Verify machine CIDR includes all host IP addresses"
      dns_resolution = "Check that cluster API/apps domains resolve correctly from hosts"
      network_connectivity = "Verify hosts can reach each other and have default routes"
    }
  }
}

output "hardware_troubleshooting" {
  description = "Hardware validation analysis by host"
  value = {
    hosts_with_hardware_issues = keys(local.hardware_issues_by_host)
    
    hardware_issues_detail = {
      for host_id, issues in local.hardware_issues_by_host : host_id => [
        for issue in issues : {
          validation = issue.validation_id
          message    = issue.message
          category   = issue.category
          type       = issue.validation_type
        }
      ]
    }
    
    common_hardware_fixes = {
      cpu_cores = "Ensure hosts meet minimum CPU requirements (4 cores for masters, 2+ for workers)"
      memory = "Ensure hosts meet minimum memory requirements (16GB for masters, 8GB+ for workers)"
      disks = "Ensure hosts have sufficient disk space (100GB+ for system disk)"
      inventory = "Check that host inventory was collected successfully during discovery"
    }
  }
}

output "operator_requirements_analysis" {
  description = "Operator-specific validation analysis"
  value = {
    lso_status = {
      ready   = length(local.lso_failures) == 0
      issues  = local.lso_failures
      requirements = "Local Storage Operator requires available block devices on worker nodes"
    }
    
    odf_status = {
      ready   = length(local.odf_failures) == 0  
      issues  = local.odf_failures
      requirements = "OpenShift Data Foundation requires sufficient CPU, memory, and storage devices"
    }
    
    cnv_status = {
      ready   = length(local.cnv_failures) == 0
      issues  = local.cnv_failures
      requirements = "Container Native Virtualization requires CPU virtualization support"
    }
    
    overall_operator_ready = (
      length(local.lso_failures) == 0 &&
      length(local.odf_failures) == 0 &&
      length(local.cnv_failures) == 0
    )
  }
}

output "platform_specific_analysis" {
  description = "Platform-specific validation issues"
  value = {
    platform_issues = [
      for issue in local.platform_issues : {
        host_id    = issue.host_id
        validation = issue.validation_id
        message    = issue.message
        category   = issue.category
      }
    ]
    
    platform_ready = length(local.platform_issues) == 0
    
    # Common platform-specific troubleshooting
    vsphere_tips = {
      disk_uuid = "Ensure disk.EnableUUID=true in VM configuration"
      hardware_version = "Use supported VMware hardware version"
      cpu_features = "Enable CPU virtualization features in VM settings"
    }
    
    baremetal_tips = {
      bmc_access = "Ensure BMC/IPMI is accessible and configured"
      boot_order = "Verify PXE/UEFI boot order is correct"
      hardware_raid = "Configure RAID controllers as needed"
    }
  }
}

output "storage_analysis" {
  description = "Storage and disk validation analysis"
  value = {
    storage_issues = [
      for issue in local.storage_issues : {
        host_id    = issue.host_id
        validation = issue.validation_id
        message    = issue.message
        category   = issue.category
      }
    ]
    
    storage_ready = length(local.storage_issues) == 0
    
    storage_troubleshooting = {
      disk_speed = "Ensure installation disk meets minimum I/O performance requirements"
      disk_size = "Verify installation disk has sufficient capacity (100GB+ recommended)"
      disk_formatting = "Check that disks are not already formatted or in use"
      multipath = "Configure multipath if using shared storage"
    }
  }
}

# Specific host debugging output (only when debugging a specific host)
output "specific_host_debug" {
  description = "Detailed validation information for specific host"
  value = var.debug_host_id != "" ? {
    host_id = var.debug_host_id
    
    validations = [
      for v in data.oai_host_validations.specific_host_debug[0].validations : {
        validation = v.validation_id
        status     = v.status
        message    = v.message
        category   = v.category
        type       = v.validation_type
        group      = v.validation_group
      }
    ]
    
    failure_summary = [
      for v in data.oai_host_validations.specific_host_debug[0].validations :
      "${v.category}/${v.validation_id}: ${v.message}" 
      if v.status == "failure"
    ]
  } : null
}

# ==============================================================================
# Variables
# ==============================================================================

variable "cluster_id" {
  description = "The ID of the cluster to analyze validations for"
  type        = string
}

variable "infra_env_id" {
  description = "Infrastructure environment ID (required for specific host debugging)"
  type        = string
  default     = ""
}

variable "debug_host_id" {
  description = "Specific host ID to debug (optional)"
  type        = string
  default     = ""
}