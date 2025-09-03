---
page_title: "Data Source: oai_operator_bundles"
subcategory: "Operator Information"
---

# oai_operator_bundles Data Source

Retrieves information about available operator bundles from the OpenShift Assisted Service. Operator bundles are curated collections of operators that work together to provide specific functionality.

## Example Usage

### List All Available Bundles

```hcl
data "oai_operator_bundles" "all" {}

output "available_bundles" {
  value = data.oai_operator_bundles.all.bundles
}
```

### Get Specific Bundle Information

```hcl
data "oai_operator_bundles" "virtualization" {
  bundle_id = "virtualization"
}

output "virtualization_bundle" {
  value = data.oai_operator_bundles.virtualization.bundles[0]
}
```

## Argument Reference

### Optional Arguments

- `bundle_id` (String) - ID of a specific bundle to retrieve. If not specified, all available bundles are returned.

## Attribute Reference

The following attributes are exported:

- `bundles` (List of Object) - List of available operator bundles. Each bundle object contains:
  - `id` (String) - Unique identifier for the bundle
  - `title` (String) - Human-readable title of the bundle
  - `operators` (List of String) - List of operator names included in this bundle

## Available Bundles

The supported bundles typically include:

### virtualization
- **Title**: OpenShift Virtualization
- **Purpose**: Virtual machine management and containerised workloads
- **Operators**: Includes KubeVirt and related operators for VM lifecycle management

### openshift-ai
- **Title**: Red Hat OpenShift AI
- **Purpose**: Machine learning and artificial intelligence workloads
- **Operators**: Includes operators for data science workflows, model serving, and ML pipelines

## Practical Examples

### Bundle-Based Cluster Configuration

```hcl
data "oai_operator_bundles" "available" {}

variable "enable_virtualization" {
  description = "Enable virtualization bundle"
  type        = bool
  default     = false
}

variable "enable_ai" {
  description = "Enable AI/ML bundle"
  type        = bool
  default     = false
}

locals {
  # Find requested bundles
  virtualization_bundle = [
    for bundle in data.oai_operator_bundles.available.bundles :
    bundle if bundle.id == "virtualization"
  ]
  
  ai_bundle = [
    for bundle in data.oai_operator_bundles.available.bundles :
    bundle if bundle.id == "openshift-ai"
  ]
  
  # Collect operators from enabled bundles
  bundle_operators = concat(
    var.enable_virtualization && length(local.virtualization_bundle) > 0 ? local.virtualization_bundle[0].operators : [],
    var.enable_ai && length(local.ai_bundle) > 0 ? local.ai_bundle[0].operators : []
  )
}

resource "oai_cluster" "with_bundles" {
  name              = "cluster-with-bundles"
  openshift_version = "4.14"
  # ... other configuration
  
  # Include all operators from selected bundles
  olm_operators = [
    for op in local.bundle_operators : {
      name = op
    }
  ]
}

output "selected_operators" {
  description = "Operators included from bundles"
  value       = local.bundle_operators
}
```

### Bundle Compatibility Validation

```hcl
data "oai_operator_bundles" "bundles" {}
data "oai_supported_operators" "supported" {
  openshift_version = var.openshift_version
  cpu_architecture  = var.cpu_architecture
}

locals {
  # Check which bundle operators are supported
  bundle_compatibility = {
    for bundle in data.oai_operator_bundles.bundles.bundles :
    bundle.id => {
      title             = bundle.title
      total_operators   = length(bundle.operators)
      supported_operators = [
        for op in bundle.operators :
        op if contains(data.oai_supported_operators.supported.operators, op)
      ]
    }
  }
  
  # Calculate compatibility percentage
  bundle_compatibility_pct = {
    for bundle_id, info in local.bundle_compatibility :
    bundle_id => {
      title              = info.title
      compatibility_pct  = (length(info.supported_operators) / info.total_operators) * 100
      supported         = info.supported_operators
      unsupported       = setsubtract(
        toset([for bundle in data.oai_operator_bundles.bundles.bundles : bundle.operators if bundle.id == bundle_id][0]),
        toset(info.supported_operators)
      )
    }
  }
}

output "bundle_compatibility" {
  description = "Compatibility analysis for each bundle"
  value       = local.bundle_compatibility_pct
}
```

### Environment-Specific Bundle Selection

```hcl
data "oai_operator_bundles" "available" {}

# Different bundle configurations for different environments
locals {
  environment_bundles = {
    development = ["virtualization"]              # VM testing environment
    staging     = ["virtualization"]              # Pre-production testing
    production  = []                              # Conservative production setup
    ml_platform = ["openshift-ai"]               # Dedicated ML platform
    hybrid      = ["virtualization", "openshift-ai"] # Full-featured environment
  }
  
  selected_bundles = local.environment_bundles[var.environment]
  
  # Get operators for selected bundles
  selected_operators = flatten([
    for bundle_id in local.selected_bundles : [
      for bundle in data.oai_operator_bundles.available.bundles :
      bundle.operators if bundle.id == bundle_id
    ]
  ])
}

variable "environment" {
  description = "Deployment environment"
  type        = string
  validation {
    condition = contains([
      "development", "staging", "production", "ml_platform", "hybrid"
    ], var.environment)
    error_message = "Environment must be one of: development, staging, production, ml_platform, hybrid."
  }
}

resource "oai_cluster" "environment_specific" {
  name = "${var.environment}-cluster"
  # ... other configuration
  
  olm_operators = [
    for op in local.selected_operators : {
      name = op
    }
  ]
}
```

### Bundle Documentation and Planning

```hcl
data "oai_operator_bundles" "documentation" {}

locals {
  # Generate comprehensive bundle documentation
  bundle_details = {
    for bundle in data.oai_operator_bundles.documentation.bundles :
    bundle.id => {
      title       = bundle.title
      description = "Bundle containing ${length(bundle.operators)} operators"
      operators   = bundle.operators
      use_cases = bundle.id == "virtualization" ? [
        "Virtual machine workloads",
        "Legacy application containerisation",
        "Development environment isolation"
      ] : bundle.id == "openshift-ai" ? [
        "Machine learning model training",
        "Data science workflows",
        "AI application deployment"
      ] : [
        "General purpose bundle"
      ]
    }
  }
}

# Generate bundle selection guide
output "bundle_selection_guide" {
  description = "Guide for selecting appropriate bundles"
  value = {
    for bundle_id, details in local.bundle_details :
    bundle_id => {
      title          = details.title
      operator_count = length(details.operators)
      use_cases      = details.use_cases
      operators      = details.operators
    }
  }
}
```

### Bundle Cost and Resource Planning

```hcl
data "oai_operator_bundles" "planning" {}

locals {
  # Estimate resource requirements for bundles
  bundle_resource_estimates = {
    virtualization = {
      min_worker_nodes = 3
      min_cpu_per_node = 8
      min_memory_gb    = 32
      additional_storage = "Required for VM images"
    }
    "openshift-ai" = {
      min_worker_nodes = 2  
      min_cpu_per_node = 16
      min_memory_gb    = 64
      additional_storage = "Required for model storage and datasets"
    }
  }
  
  # Calculate total requirements for selected bundles
  selected_bundle_ids = ["virtualization"] # Example selection
  
  total_requirements = {
    min_worker_nodes = max([
      for bundle_id in local.selected_bundle_ids :
      local.bundle_resource_estimates[bundle_id].min_worker_nodes
      if contains(keys(local.bundle_resource_estimates), bundle_id)
    ]...)
    min_cpu_per_node = max([
      for bundle_id in local.selected_bundle_ids :
      local.bundle_resource_estimates[bundle_id].min_cpu_per_node
      if contains(keys(local.bundle_resource_estimates), bundle_id)
    ]...)
    min_memory_gb = max([
      for bundle_id in local.selected_bundle_ids :
      local.bundle_resource_estimates[bundle_id].min_memory_gb
      if contains(keys(local.bundle_resource_estimates), bundle_id)
    ]...)
  }
}

output "resource_requirements" {
  description = "Estimated resource requirements for selected bundles"
  value       = local.total_requirements
}
```

## Integration Patterns

### Bundle + Individual Operator Selection

```hcl
data "oai_operator_bundles" "bundles" {}
data "oai_supported_operators" "individual" {}

locals {
  # Start with bundle operators
  bundle_operators = flatten([
    for bundle in data.oai_operator_bundles.bundles.bundles :
    bundle.operators if bundle.id == "virtualization"
  ])
  
  # Add individual operators not covered by bundles
  additional_operators = [
    "metallb-operator",     # Load balancing
    "local-storage-operator" # Local storage
  ]
  
  # Combine and deduplicate
  all_operators = distinct(concat(local.bundle_operators, local.additional_operators))
  
  # Validate all operators are supported
  validated_operators = [
    for op in local.all_operators :
    op if contains(data.oai_supported_operators.individual.operators, op)
  ]
}

resource "oai_cluster" "combined" {
  # ... configuration
  olm_operators = [
    for op in local.validated_operators : {
      name = op
    }
  ]
}
```

This data source enables sophisticated operator bundle management, allowing you to deploy curated collections of operators that are designed to work together effectively.