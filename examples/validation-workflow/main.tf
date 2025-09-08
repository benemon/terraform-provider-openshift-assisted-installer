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
# Validation-Driven Cluster Deployment
# ==============================================================================

# Create cluster definition
resource "openshift_assisted_installer_cluster" "production" {
  name                = "prod-cluster"
  base_dns_domain     = "example.com" 
  openshift_version   = "4.15.20"
  cpu_architecture    = "x86_64"
  control_plane_count = 3
  
  pull_secret    = var.pull_secret
  ssh_public_key = var.ssh_public_key
  
  # Network configuration
  api_vips = [{ ip = "10.0.1.100" }]
  ingress_vips = [{ ip = "10.0.1.101" }]
  
  # Enable operators that require validation
  olm_operators = [
    { name = "lso" },  # Local Storage Operator
    { name = "odf" },  # OpenShift Data Foundation
  ]
}

# Create infra environment
resource "openshift_assisted_installer_infra_env" "production" {
  name              = "${openshift_assisted_installer_cluster.production.name}-infra"
  cluster_id        = openshift_assisted_installer_cluster.production.id
  cpu_architecture  = "x86_64"
  pull_secret       = var.pull_secret
  ssh_authorized_key = var.ssh_public_key
  image_type        = "full-iso"
}

# ==============================================================================
# Validation Checks Using openshift_assisted_installer_cluster_validations and openshift_assisted_installer_host_validations
# ==============================================================================

# Check cluster-level validations for installation readiness
data "openshift_assisted_installer_cluster_validations" "readiness" {
  cluster_id = openshift_assisted_installer_cluster.production.id
  
  # Only check blocking validations that would prevent installation
  validation_types = ["blocking"]
  status_filter    = ["failure", "pending"]
}

# Check all cluster validations for comprehensive status
data "openshift_assisted_installer_cluster_validations" "all_cluster" {
  cluster_id = openshift_assisted_installer_cluster.production.id
}

# Check host-level validations for all hosts in the cluster
data "openshift_assisted_installer_host_validations" "all_hosts" {
  cluster_id = openshift_assisted_installer_cluster.production.id
  
  # Focus on blocking validations that prevent installation
  validation_types = ["blocking"]
  status_filter    = ["failure", "pending"]
}

# Check specific validations for storage operator requirements
data "openshift_assisted_installer_host_validations" "storage_readiness" {
  cluster_id = openshift_assisted_installer_cluster.production.id
  
  # Filter for storage operator requirements only
  validation_names = [
    "lso-requirements-satisfied",
    "odf-requirements-satisfied", 
    "has-min-valid-disks",
    "sufficient-installation-diskspeed"
  ]
}

# Check network-related validations across cluster and hosts
data "openshift_assisted_installer_cluster_validations" "network_cluster" {
  cluster_id = openshift_assisted_installer_cluster.production.id
  categories = ["network"]
}

data "openshift_assisted_installer_host_validations" "network_hosts" {
  cluster_id = openshift_assisted_installer_cluster.production.id
  categories = ["network"]
  status_filter = ["failure", "pending"]
}

# ==============================================================================
# Conditional Logic Based on Validations
# ==============================================================================

locals {
  # Parse validation results - cluster blocking failures
  cluster_blocking_failures = [
    for validation in data.openshift_assisted_installer_cluster_validations.readiness.validations :
    validation if validation.status == "failure"  # Already filtered for blocking + failure/pending
  ]
  
  # Parse validation results - host blocking failures  
  host_blocking_failures = [
    for validation in data.openshift_assisted_installer_host_validations.all_hosts.validations :
    validation if validation.status == "failure"  # Already filtered for blocking + failure/pending
  ]
  
  # Parse validation results - storage operator failures
  storage_validation_failures = [
    for validation in data.openshift_assisted_installer_host_validations.storage_readiness.validations :
    validation if validation.status == "failure"
  ]
  
  # Network validation failures (cluster + hosts combined)
  network_failures = concat(
    [for v in data.openshift_assisted_installer_cluster_validations.network_cluster.validations : v if v.status == "failure"],
    [for v in data.openshift_assisted_installer_host_validations.network_hosts.validations : v if v.status == "failure"]
  )
  
  # Installation readiness flags
  cluster_ready = length(local.cluster_blocking_failures) == 0
  hosts_ready   = length(local.host_blocking_failures) == 0  
  storage_ready = length(local.storage_validation_failures) == 0
  network_ready = length(local.network_failures) == 0
  
  # Overall readiness assessment
  installation_ready = (
    local.cluster_ready && 
    local.hosts_ready && 
    local.storage_ready && 
    local.network_ready
  )
  
  # Validation summary for reporting
  total_blocking_failures = (
    length(local.cluster_blocking_failures) + 
    length(local.host_blocking_failures)
  )
  
  # Host failure breakdown by host ID
  host_failures_by_host = {
    for validation in data.openshift_assisted_installer_host_validations.all_hosts.validations :
    validation.host_id => validation...
    if validation.status == "failure"
  }
}

# ==============================================================================
# Conditional Installation Based on Validations
# ==============================================================================

# Only trigger installation if all validations pass
resource "openshift_assisted_installer_cluster_installation" "production" {
  # Conditional creation based on validation results
  count = local.installation_ready ? 1 : 0
  
  cluster_id          = openshift_assisted_installer_cluster.production.id
  wait_for_hosts      = true
  expected_host_count = 3
  
  timeouts {
    create = "120m"
  }
  
  # Ensure validations are checked first
  depends_on = [
    data.openshift_assisted_installer_cluster_validations.readiness,
    data.openshift_assisted_installer_host_validations.all_hosts,
    data.openshift_assisted_installer_host_validations.storage_readiness
  ]
}

# ==============================================================================
# Validation-Based Outputs and Troubleshooting
# ==============================================================================

output "validation_status" {
  description = "Comprehensive validation status"
  value = {
    cluster_ready     = local.cluster_ready
    hosts_ready       = local.hosts_ready
    storage_ready     = local.storage_ready
    installation_ready = local.installation_ready
    
    # Detailed failure information
    cluster_failures = local.cluster_blocking_failures
    host_failures    = local.host_blocking_failures
    storage_failures = local.storage_validation_failures
  }
}

output "troubleshooting_guide" {
  description = "Validation failure troubleshooting"
  value = !local.installation_ready ? {
    message = "Installation blocked by validation failures"
    summary = "Found ${local.total_blocking_failures} blocking validation failures"
    
    cluster_issues = [
      for failure in local.cluster_blocking_failures :
      "❌ Cluster ${failure.validation_id} (${failure.category}): ${failure.message}"
    ]
    
    host_issues = [
      for failure in local.host_blocking_failures :
      "❌ Host ${failure.host_id} - ${failure.validation_id} (${failure.category}): ${failure.message}"
    ]
    
    storage_issues = [
      for failure in local.storage_validation_failures :
      "❌ Storage on Host ${failure.host_id} - ${failure.validation_id}: ${failure.message}"
    ]
    
    network_issues = [
      for failure in local.network_failures :
      "❌ Network ${can(failure.host_id) ? "Host ${failure.host_id}" : "Cluster"} - ${failure.validation_id}: ${failure.message}"
    ]
    
    # Host-specific breakdown
    hosts_with_issues = keys(local.host_failures_by_host)
    
    next_steps = [
      "1. Review validation failures listed above by category",
      "2. Fix cluster-level issues (network, VIPs, operators)",
      "3. Fix host-level issues (hardware, connectivity, inventory)", 
      "4. Address storage requirements if using LSO/ODF operators",
      "5. Wait for automatic re-validation (hosts check every few minutes)",
      "6. Run 'terraform plan' to verify readiness before 'terraform apply'"
    ]
  } : {
    message = "✅ All critical validations passed - installation can proceed"
    summary = "Cluster and all hosts are ready for installation"
  }
}

# Pre-installation health check
output "pre_installation_health" {
  description = "Pre-installation health summary"
  value = {
    # Overall status
    installation_ready = local.installation_ready
    
    # Validation counts (all validations, not just blocking)
    total_cluster_validations = length(data.openshift_assisted_installer_cluster_validations.all_cluster.validations)
    total_host_validations    = length(data.openshift_assisted_installer_host_validations.all_hosts.validations)
    total_validations = (
      length(data.openshift_assisted_installer_cluster_validations.all_cluster.validations) +
      length(data.openshift_assisted_installer_host_validations.all_hosts.validations)
    )
    
    # Success counts
    passing_cluster_validations = length([
      for v in data.openshift_assisted_installer_cluster_validations.all_cluster.validations : v 
      if v.status == "success"
    ])
    passing_host_validations = length([
      for v in data.openshift_assisted_installer_host_validations.all_hosts.validations : v 
      if v.status == "success"  
    ])
    
    # Failure analysis
    blocking_failures      = local.total_blocking_failures
    storage_failures       = length(local.storage_validation_failures)
    network_failures       = length(local.network_failures)
    
    # Readiness by component
    cluster_component_ready = local.cluster_ready
    hosts_component_ready   = local.hosts_ready
    storage_component_ready = local.storage_ready
    network_component_ready = local.network_ready
    
    # Overall health percentage
    readiness_percentage = local.installation_ready ? 100 : (
      (local.cluster_ready ? 25 : 0) +
      (local.hosts_ready ? 25 : 0) +
      (local.storage_ready ? 25 : 0) +
      (local.network_ready ? 25 : 0)
    )
    
    # Host breakdown
    total_hosts = length(local.host_failures_by_host) > 0 ? 
      length(distinct([for v in data.openshift_assisted_installer_host_validations.all_hosts.validations : v.host_id])) :
      0
    hosts_with_failures = length(keys(local.host_failures_by_host))
  }
}

# ==============================================================================
# Advanced Validation Patterns
# ==============================================================================

# Operator-specific validation checks
locals {
  # Check if ODF requirements are met before enabling
  odf_validated = length([
    for v in data.openshift_assisted_installer_host_validations.storage_readiness.validations :
    v if v.validation_id == "odf-requirements-satisfied" && v.status == "success"
  ]) > 0
  
  # Check if LSO requirements are met
  lso_validated = length([
    for v in data.openshift_assisted_installer_host_validations.storage_readiness.validations :
    v if v.validation_id == "lso-requirements-satisfied" && v.status == "success"
  ]) > 0
}

# Conditional manifest application based on validation
resource "openshift_assisted_installer_manifest" "storage_config" {
  # Only apply storage config if storage validations pass
  count = local.storage_ready ? 1 : 0
  
  cluster_id = openshift_assisted_installer_cluster.production.id
  file_name  = "storage-configuration.yaml"
  folder     = "manifests"
  
  content = templatefile("${path.module}/manifests/storage.yaml", {
    enable_odf = local.odf_validated
    enable_lso = local.lso_validated
  })
}

# ==============================================================================
# Validation Monitoring During Installation
# ==============================================================================

# Monitor validation status changes over time
data "openshift_assisted_installer_cluster_events" "validation_events" {
  cluster_id = openshift_assisted_installer_cluster.production.id
  
  # Filter for validation-related events
  message = "validation"
  limit   = 100
}

output "validation_timeline" {
  description = "Timeline of validation status changes"
  value = [
    for event in data.openshift_assisted_installer_cluster_events.validation_events.events :
    {
      timestamp = event.event_time
      message   = event.message
      severity  = event.severity
    }
    if can(regex("validation|requirement|ready", event.message))
  ]
}