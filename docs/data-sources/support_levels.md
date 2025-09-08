---
page_title: "Data Source: openshift_assisted_installer_support_levels"
subcategory: "General Information"
---

# openshift_assisted_installer_support_levels Data Source

Retrieves feature support level information from the OpenShift Assisted Service API. Use this data source to determine which features are supported for specific OpenShift versions, CPU architectures, and platform types.

## Example Usage

### Get Support Levels for Specific Version

```hcl
data "openshift_assisted_installer_support_levels" "current" {
  openshift_version = "4.16"
}

output "feature_support" {
  value = data.openshift_assisted_installer_support_levels.current.features
}
```

### Filter by Architecture and Platform

```hcl
data "openshift_assisted_installer_support_levels" "baremetal_x86" {
  openshift_version = "4.16"
  cpu_architecture  = "x86_64"
  platform_type     = "baremetal"
}

output "baremetal_features" {
  value = data.openshift_assisted_installer_support_levels.baremetal_x86.features
}
```

### Multi-Architecture Support Analysis

```hcl
data "openshift_assisted_installer_support_levels" "arm64" {
  openshift_version = "4.16"
  cpu_architecture  = "arm64"
}

data "openshift_assisted_installer_support_levels" "x86_64" {
  openshift_version = "4.16"
  cpu_architecture  = "x86_64"
}

locals {
  # Compare feature support between architectures
  arm64_features   = keys(data.openshift_assisted_installer_support_levels.arm64.features)
  x86_64_features  = keys(data.openshift_assisted_installer_support_levels.x86_64.features)
  
  common_features = setintersection(
    toset(local.arm64_features),
    toset(local.x86_64_features)
  )
  
  x86_only_features = setsubtract(
    toset(local.x86_64_features),
    toset(local.arm64_features)
  )
}

output "architecture_comparison" {
  value = {
    common_features   = local.common_features
    x86_only_features = local.x86_only_features
  }
}
```

## Argument Reference

### Required Arguments

- `openshift_version` (String) - OpenShift version to check support levels for.

### Optional Arguments

- `cpu_architecture` (String) - Filter by CPU architecture. Valid values: `x86_64`, `arm64`, `ppc64le`, `s390x`, `multi`.
- `platform_type` (String) - Filter by platform type. Valid values: `baremetal`, `nutanix`, `vsphere`, `none`, `external`.

## Attribute Reference

The following attributes are exported:

- `features` (Map of String) - Map of feature names to their support levels. Support levels include:
  - `supported` - Fully supported feature
  - `dev-preview` - Developer preview, not suitable for production
  - `tech-preview` - Technical preview, limited support
  - `unavailable` - Feature not available for this configuration

## Common Features

The support levels data source typically includes information about:

### Core Platform Features
- `SNO` - Single Node OpenShift support
- `VIP_AUTO_ALLOC` - Automatic VIP allocation
- `DISK_ENCRYPTION` - Full disk encryption
- `PROXY` - HTTP/HTTPS proxy support

### Networking Features
- `USER_MANAGED_NETWORKING` - User-managed networking mode
- `DUAL_STACK` - IPv4/IPv6 dual stack networking
- `PLATFORM_MANAGED_NETWORKING` - Platform-managed networking

### Storage Features
- `ODF` - OpenShift Data Foundation support
- `LOCAL_STORAGE` - Local storage operator support
- `MULTIPATH` - Multipath storage device support

### Advanced Features
- `CUSTOM_MANIFESTS` - Custom manifest support
- `SCHEDULABLE_MASTERS` - Workloads on control plane nodes
- `ARM64_ARCHITECTURE` - ARM64 CPU architecture support

## Practical Examples

### Feature-Based Deployment Decisions

```hcl
data "openshift_assisted_installer_support_levels" "target_config" {
  openshift_version = var.openshift_version
  cpu_architecture  = var.cpu_architecture
  platform_type     = var.platform_type
}

locals {
  # Check if required features are supported
  required_features = ["SNO", "USER_MANAGED_NETWORKING", "CUSTOM_MANIFESTS"]
  
  feature_support_status = {
    for feature in local.required_features :
    feature => lookup(data.openshift_assisted_installer_support_levels.target_config.features, feature, "unavailable")
  }
  
  # Determine if deployment is viable
  unsupported_features = [
    for feature, support in local.feature_support_status :
    feature if support == "unavailable"
  ]
  
  can_deploy = length(local.unsupported_features) == 0
}

# Only create cluster if all required features are available
resource "openshift_assisted_installer_cluster" "conditional" {
  count = local.can_deploy ? 1 : 0
  
  name              = "feature-validated-cluster"
  openshift_version = var.openshift_version
  cpu_architecture  = var.cpu_architecture
  # ... other configuration
}

output "deployment_analysis" {
  value = {
    can_deploy           = local.can_deploy
    feature_status       = local.feature_support_status
    unsupported_features = local.unsupported_features
  }
}
```

### Single Node OpenShift Validation

```hcl
data "openshift_assisted_installer_support_levels" "sno_check" {
  openshift_version = var.openshift_version
  cpu_architecture  = var.cpu_architecture
}

locals {
  sno_support = lookup(
    data.openshift_assisted_installer_support_levels.sno_check.features,
    "SNO",
    "unavailable"
  )
  
  sno_supported = contains(["supported", "dev-preview", "tech-preview"], local.sno_support)
}

resource "openshift_assisted_installer_cluster" "single_node" {
  count = local.sno_supported ? 1 : 0
  
  name                = "single-node-cluster"
  control_plane_count = 1  # SNO configuration
  # ... other configuration
}

output "sno_compatibility" {
  value = {
    support_level = local.sno_support
    supported     = local.sno_supported
    message = local.sno_supported ? 
      "Single Node OpenShift is ${local.sno_support}" :
      "Single Node OpenShift is not available for this configuration"
  }
}
```

### Platform Migration Planning

```hcl
# Check feature support across different platforms
data "openshift_assisted_installer_support_levels" "baremetal" {
  openshift_version = "4.16"
  platform_type     = "baremetal"
}

data "openshift_assisted_installer_support_levels" "vsphere" {
  openshift_version = "4.16" 
  platform_type     = "vsphere"
}

data "openshift_assisted_installer_support_levels" "nutanix" {
  openshift_version = "4.16"
  platform_type     = "nutanix"
}

locals {
  platforms = {
    baremetal = data.openshift_assisted_installer_support_levels.baremetal.features
    vsphere   = data.openshift_assisted_installer_support_levels.vsphere.features
    nutanix   = data.openshift_assisted_installer_support_levels.nutanix.features
  }
  
  # Features we need for our workload
  required_features = [
    "USER_MANAGED_NETWORKING",
    "CUSTOM_MANIFESTS", 
    "ODF",
    "PROXY"
  ]
  
  # Analyse platform compatibility
  platform_compatibility = {
    for platform, features in local.platforms :
    platform => {
      compatible = alltrue([
        for feature in local.required_features :
        contains(["supported", "dev-preview"], lookup(features, feature, "unavailable"))
      ])
      feature_status = {
        for feature in local.required_features :
        feature => lookup(features, feature, "unavailable")
      }
    }
  }
  
  # Find best platform option
  compatible_platforms = [
    for platform, status in local.platform_compatibility :
    platform if status.compatible
  ]
}

output "platform_analysis" {
  value = {
    compatibility       = local.platform_compatibility
    compatible_platforms = local.compatible_platforms
    recommendation = length(local.compatible_platforms) > 0 ?
      "Recommended platforms: ${join(", ", local.compatible_platforms)}" :
      "No platforms support all required features"
  }
}
```

### Version Upgrade Compatibility

```hcl
variable "current_version" {
  description = "Current OpenShift version"
  type        = string
  default     = "4.15"
}

variable "target_version" {
  description = "Target OpenShift version"
  type        = string
  default     = "4.16"
}

data "openshift_assisted_installer_support_levels" "current" {
  openshift_version = var.current_version
  cpu_architecture  = var.cpu_architecture
}

data "openshift_assisted_installer_support_levels" "target" {
  openshift_version = var.target_version
  cpu_architecture  = var.cpu_architecture
}

locals {
  current_features = keys(data.openshift_assisted_installer_support_levels.current.features)
  target_features  = keys(data.openshift_assisted_installer_support_levels.target.features)
  
  # Features that might be lost in upgrade
  deprecated_features = setsubtract(
    toset(local.current_features),
    toset(local.target_features)
  )
  
  # New features available in target version
  new_features = setsubtract(
    toset(local.target_features),
    toset(local.current_features)
  )
  
  # Check if currently used features remain supported
  upgrade_impact = {
    deprecated_features = local.deprecated_features
    new_features       = local.new_features
    upgrade_safe = length(local.deprecated_features) == 0
  }
}

output "upgrade_analysis" {
  value = {
    from_version        = var.current_version
    to_version         = var.target_version
    deprecated_features = local.deprecated_features
    new_features       = local.new_features
    safe_to_upgrade    = local.upgrade_impact.upgrade_safe
    recommendation = local.upgrade_impact.upgrade_safe ?
      "Upgrade appears safe - no features will be lost" :
      "Review deprecated features before upgrading: ${join(", ", local.deprecated_features)}"
  }
}
```

### Production Readiness Assessment

```hcl
data "openshift_assisted_installer_support_levels" "production_check" {
  openshift_version = var.openshift_version
  cpu_architecture  = var.cpu_architecture
  platform_type     = var.platform_type
}

locals {
  # Features critical for production
  production_features = [
    "CUSTOM_MANIFESTS",
    "PROXY", 
    "DISK_ENCRYPTION"
  ]
  
  # Evaluate production readiness
  production_assessment = {
    for feature in local.production_features :
    feature => {
      support_level = lookup(data.openshift_assisted_installer_support_levels.production_check.features, feature, "unavailable")
      production_ready = lookup(data.openshift_assisted_installer_support_levels.production_check.features, feature, "unavailable") == "supported"
    }
  }
  
  production_ready_count = length([
    for feature, status in local.production_assessment :
    feature if status.production_ready
  ])
  
  overall_production_ready = local.production_ready_count == length(local.production_features)
}

output "production_readiness" {
  value = {
    feature_assessment     = local.production_assessment
    production_ready_count = "${local.production_ready_count}/${length(local.production_features)}"
    overall_ready         = local.overall_production_ready
    recommendation = local.overall_production_ready ?
      "Configuration is production-ready" :
      "Some features are not fully supported for production use"
  }
}
```

This data source is essential for making informed decisions about OpenShift deployments, ensuring that required features are available and appropriately supported for your specific use case and environment.