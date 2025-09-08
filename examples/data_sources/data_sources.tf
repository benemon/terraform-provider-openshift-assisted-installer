terraform {
  required_providers {
    openshift_assisted_installer = {
      source  = "benemon/openshift-assisted-installer"
      version = "~> 0.1"
    }
  }
}

# Configure the OpenShift Assisted Installer Provider
# Uses OFFLINE_TOKEN environment variable for authentication
provider "openshift_assisted_installer" {
  # offline_token = var.offline_token  # Optional: Use variable instead of env var
  # endpoint = "https://api.openshift.com/api/assisted-install"  # Optional: Override default
  # timeout = "300s"  # Optional: Increase timeout for slow connections
}

# ==============================================================================
# DATA SOURCE 1: OpenShift Versions
# Fetches all available OpenShift versions with their metadata
# ==============================================================================

# Get all OpenShift versions
data "openshift_assisted_installer_openshift_versions" "all_versions" {
  # No filters - get everything
}

# Get only the latest versions
data "openshift_assisted_installer_openshift_versions" "latest_only" {
  only_latest = true
}

# Filter for specific version family
data "openshift_assisted_installer_openshift_versions" "v4_14" {
  version = "4.14"
}

# Filter for 4.15 versions
data "openshift_assisted_installer_openshift_versions" "v4_15" {
  version     = "4.15"
  only_latest = true
}

# ==============================================================================
# DATA SOURCE 2: Supported Operators
# Lists all operators that can be installed during cluster creation
# ==============================================================================

data "openshift_assisted_installer_supported_operators" "all_operators" {
  # No configuration needed - fetches all supported operators
}

# ==============================================================================
# DATA SOURCE 3: Operator Bundles
# Lists operator bundles (collections of related operators)
# 
# Schema matches Swagger specification exactly:
# - id: Bundle identifier (string)
# - title: Bundle title (string) - matches Swagger "title" field
# - operators: List of operator names ([]string) - simple strings, not objects
# ==============================================================================

data "openshift_assisted_installer_operator_bundles" "all_bundles" {
  # No configuration needed - fetches all available bundles
  # 
  # Output format (matches Swagger API exactly):
  # bundles = [
  #   {
  #     id = "virtualization"
  #     title = "Virtualization"
  #     operators = ["cnv", "nmstate", "metallb", ...]
  #   }
  # ]
}

# ==============================================================================
# DATA SOURCE 4: Support Levels
# Shows feature support levels for specific OpenShift versions
# ==============================================================================

# Get support levels for the latest 4.14 version on x86_64 baremetal
data "openshift_assisted_installer_support_levels" "v4_14_x86_baremetal" {
  openshift_version = "4.14.0"
  cpu_architecture  = "x86_64"
  platform_type     = "baremetal"
}

# Get support levels for ARM64 architecture
data "openshift_assisted_installer_support_levels" "v4_14_arm64" {
  openshift_version = "4.14.0"
  cpu_architecture  = "arm64"
  platform_type     = "baremetal"
}

# Get support levels for vSphere platform
data "openshift_assisted_installer_support_levels" "v4_14_vsphere" {
  openshift_version = "4.14.0"
  cpu_architecture  = "x86_64"
  platform_type     = "vsphere"
}

# Get support levels for latest 4.15 on Nutanix
data "openshift_assisted_installer_support_levels" "v4_15_nutanix" {
  openshift_version = "4.15.0"
  cpu_architecture  = "x86_64"
  platform_type     = "nutanix"
}

# ==============================================================================
# OUTPUTS: Comprehensive validation of all data sources
# ==============================================================================

# -----------------------------------------------------------------------------
# OpenShift Versions Outputs
# -----------------------------------------------------------------------------

output "all_versions_count" {
  description = "Total number of OpenShift versions available"
  value       = length(data.openshift_assisted_installer_openshift_versions.all_versions.versions)
}

output "all_versions_sample" {
  description = "First 5 available OpenShift versions"
  value       = slice(data.openshift_assisted_installer_openshift_versions.all_versions.versions, 0, min(5, length(data.openshift_assisted_installer_openshift_versions.all_versions.versions)))
}

output "latest_versions" {
  description = "Latest OpenShift versions only"
  value       = data.openshift_assisted_installer_openshift_versions.latest_only.versions
}

output "v4_14_versions" {
  description = "All 4.14.x versions available"
  value = [for v in data.openshift_assisted_installer_openshift_versions.v4_14.versions : {
    version      = v.version
    display_name = v.display_name
    default      = v.default
    support      = v.support_level
  }]
}

output "v4_15_latest" {
  description = "Latest 4.15.x version"
  value       = data.openshift_assisted_installer_openshift_versions.v4_15.versions
}

# Extract CPU architectures from versions
output "supported_architectures_in_4_14" {
  description = "CPU architectures supported in 4.14 versions"
  value = distinct(flatten([
    for v in data.openshift_assisted_installer_openshift_versions.v4_14.versions : v.cpu_architectures
  ]))
}

# -----------------------------------------------------------------------------
# Supported Operators Outputs
# -----------------------------------------------------------------------------

output "operators_count" {
  description = "Total number of supported operators"
  value       = length(data.openshift_assisted_installer_supported_operators.all_operators.operators)
}

output "operator_names" {
  description = "List of all supported operator names"
  value       = data.openshift_assisted_installer_supported_operators.all_operators.operators
}

output "has_odf_operator" {
  description = "Check if OpenShift Data Foundation operator is available"
  value       = contains(data.openshift_assisted_installer_supported_operators.all_operators.operators, "odf")
}

output "has_cnv_operator" {
  description = "Check if OpenShift Virtualization operator is available"
  value       = contains(data.openshift_assisted_installer_supported_operators.all_operators.operators, "cnv")
}

output "has_mce_operator" {
  description = "Check if Multicluster Engine operator is available"
  value       = contains(data.openshift_assisted_installer_supported_operators.all_operators.operators, "mce")
}

# -----------------------------------------------------------------------------
# Operator Bundles Outputs
# -----------------------------------------------------------------------------

output "bundles_count" {
  description = "Total number of operator bundles available"
  value       = length(data.openshift_assisted_installer_operator_bundles.all_bundles.bundles)
}

output "bundle_list" {
  description = "List of available operator bundles with details"
  value = [for b in data.openshift_assisted_installer_operator_bundles.all_bundles.bundles : {
    id        = b.id
    title     = b.title
    operators = length(b.operators)
  }]
}

output "virtualization_bundle" {
  description = "Details of the virtualization bundle if available"
  value = [for b in data.openshift_assisted_installer_operator_bundles.all_bundles.bundles : b
    if b.id == "virtualization"
  ]
}

output "ai_bundle" {
  description = "Details of the OpenShift AI bundle if available"
  value = [for b in data.openshift_assisted_installer_operator_bundles.all_bundles.bundles : b
    if contains(["openshift-ai", "openshift-ai-nvidia"], b.id)
  ]
}

# -----------------------------------------------------------------------------
# Support Levels Outputs
# -----------------------------------------------------------------------------

output "v4_14_x86_features" {
  description = "Feature support levels for 4.14 on x86_64 baremetal"
  value       = data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.features
}

output "v4_14_x86_architectures" {
  description = "Architecture support levels for 4.14"
  value       = data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.architectures
}

output "v4_14_arm64_features" {
  description = "Feature support levels for 4.14 on ARM64"
  value       = data.openshift_assisted_installer_support_levels.v4_14_arm64.features
}

output "v4_14_vsphere_features" {
  description = "Feature support levels for 4.14 on vSphere"
  value       = data.openshift_assisted_installer_support_levels.v4_14_vsphere.features
}

output "v4_15_nutanix_features" {
  description = "Feature support levels for 4.15 on Nutanix"
  value       = data.openshift_assisted_installer_support_levels.v4_15_nutanix.features
}

# Compare SNO support across platforms
output "sno_support_comparison" {
  description = "Single Node OpenShift support across different configurations"
  value = {
    "4.14_x86_baremetal" = try(data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.features["SNO"], "unknown")
    "4.14_arm64"         = try(data.openshift_assisted_installer_support_levels.v4_14_arm64.features["SNO"], "unknown")
    "4.14_vsphere"       = try(data.openshift_assisted_installer_support_levels.v4_14_vsphere.features["SNO"], "unknown")
    "4.15_nutanix"       = try(data.openshift_assisted_installer_support_levels.v4_15_nutanix.features["SNO"], "unknown")
  }
}

# Compare architecture support
output "architecture_support_4_14" {
  description = "Architecture support levels for OpenShift 4.14"
  value = {
    x86_64  = try(data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.architectures["x86_64"], "unknown")
    arm64   = try(data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.architectures["arm64"], "unknown")
    ppc64le = try(data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.architectures["ppc64le"], "unknown")
    s390x   = try(data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.architectures["s390x"], "unknown")
  }
}

# -----------------------------------------------------------------------------
# Cross-Data Source Analysis
# -----------------------------------------------------------------------------

output "summary" {
  description = "Summary of all data sources"
  value = {
    total_versions        = length(data.openshift_assisted_installer_openshift_versions.all_versions.versions)
    latest_versions_count = length(data.openshift_assisted_installer_openshift_versions.latest_only.versions)
    v4_14_versions_count  = length(data.openshift_assisted_installer_openshift_versions.v4_14.versions)
    v4_15_versions_count  = length(data.openshift_assisted_installer_openshift_versions.v4_15.versions)
    supported_operators   = length(data.openshift_assisted_installer_supported_operators.all_operators.operators)
    operator_bundles      = length(data.openshift_assisted_installer_operator_bundles.all_bundles.bundles)
    v4_14_x86_features    = length(data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.features)
    v4_14_architectures   = length(data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.architectures)
  }
}

# -----------------------------------------------------------------------------
# Validation Checks
# -----------------------------------------------------------------------------

output "data_validation" {
  description = "Validation that all data sources returned data"
  value = {
    versions_loaded        = length(data.openshift_assisted_installer_openshift_versions.all_versions.versions) > 0
    latest_versions_loaded = length(data.openshift_assisted_installer_openshift_versions.latest_only.versions) > 0
    operators_loaded       = length(data.openshift_assisted_installer_supported_operators.all_operators.operators) > 0
    bundles_loaded         = length(data.openshift_assisted_installer_operator_bundles.all_bundles.bundles) > 0
    support_levels_loaded  = length(data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.features) > 0
    architectures_loaded   = length(data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.architectures) > 0
  }
}

# -----------------------------------------------------------------------------
# Useful Derived Information
# -----------------------------------------------------------------------------

# Find the default OpenShift version
output "default_openshift_version" {
  description = "The default OpenShift version for new clusters"
  value = [for v in data.openshift_assisted_installer_openshift_versions.all_versions.versions : v.version
    if try(v.default, false) == true
  ]
}

# List versions that support multi-architecture
output "multi_arch_versions" {
  description = "Versions that support multiple CPU architectures"
  value = [for v in data.openshift_assisted_installer_openshift_versions.all_versions.versions : {
    version       = v.version
    architectures = v.cpu_architectures
    } if length(v.cpu_architectures) > 1
  ]
}

# Find production-ready features for 4.14 on baremetal
output "production_features_4_14" {
  description = "Features with 'supported' status in 4.14 on x86_64 baremetal"
  value = {
    for feature, level in data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.features :
    feature => level if level == "supported"
  }
}

# Find tech-preview features
output "tech_preview_features_4_14" {
  description = "Features in tech-preview for 4.14 on x86_64 baremetal"
  value = {
    for feature, level in data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.features :
    feature => level if level == "tech-preview"
  }
}

# -----------------------------------------------------------------------------
# Example Usage Information
# -----------------------------------------------------------------------------

output "usage_example" {
  description = "Example of how to use this data for cluster configuration"
  value = {
    recommended_version = try(
      [for v in data.openshift_assisted_installer_openshift_versions.latest_only.versions : v.version
        if v.support_level == "supported"
      ][0],
      "No supported version found"
    )
    recommended_operators = slice(
      data.openshift_assisted_installer_supported_operators.all_operators.operators,
      0,
      min(5, length(data.openshift_assisted_installer_supported_operators.all_operators.operators))
    )
    available_bundles = [for b in data.openshift_assisted_installer_operator_bundles.all_bundles.bundles : b.id]
    supported_platforms = keys({
      for feature, level in data.openshift_assisted_installer_support_levels.v4_14_x86_baremetal.features :
      feature => level if startswith(feature, "PLATFORM_")
    })
  }
}