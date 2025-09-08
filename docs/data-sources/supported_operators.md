---
page_title: "Data Source: oai_supported_operators"
subcategory: "General Information"
---

# oai_supported_operators Data Source

Retrieves the list of OLM operators supported by the OpenShift Assisted Service. Use this data source to discover available operators for installation during cluster deployment.

## Example Usage

### List All Supported Operators

```hcl
data "oai_supported_operators" "all" {}

output "available_operators" {
  value = data.oai_supported_operators.all.operators
}
```

### Filter by OpenShift Version

```hcl
data "oai_supported_operators" "for_version" {
  openshift_version = "4.16"
}

output "operators_4_14" {
  value = data.oai_supported_operators.for_version.operators
}
```

### Filter by Architecture

```hcl
data "oai_supported_operators" "arm64" {
  cpu_architecture = "arm64"
}

output "arm64_operators" {
  value = data.oai_supported_operators.arm64.operators
}
```

### Comprehensive Filtering

```hcl
data "oai_supported_operators" "specific" {
  openshift_version = "4.16"
  cpu_architecture  = "x86_64"
  platform_type     = "baremetal"
}

output "filtered_operators" {
  value = data.oai_supported_operators.specific.operators
}
```

## Argument Reference

### Optional Arguments

- `openshift_version` (String) - Filter operators by OpenShift version compatibility.
- `cpu_architecture` (String) - Filter operators by CPU architecture. Valid values: `x86_64`, `arm64`, `ppc64le`, `s390x`, `multi`.
- `platform_type` (String) - Filter operators by platform type. Valid values: `baremetal`, `nutanix`, `vsphere`, `none`, `external`.

## Attribute Reference

The following attributes are exported:

- `operators` (List of String) - List of supported operator names.

## Available Operators

The supported operators typically include (availability may vary by version and platform):

### Core Infrastructure Operators
- `odf-operator` - OpenShift Data Foundation for persistent storage
- `local-storage-operator` - Local storage management
- `kubernetes-nmstate-operator` - Network configuration management
- `metallb-operator` - Load balancer for bare metal environments

### Monitoring and Observability
- `cluster-logging` - Centralized logging with ElasticSearch
- `elasticsearch-operator` - ElasticSearch cluster management
- `jaeger-product` - Distributed tracing
- `kiali-ossm` - Service mesh observability
- `servicemeshoperator` - Red Hat OpenShift Service Mesh

### Development and CI/CD
- `openshift-gitops-operator` - GitOps workflow management
- `openshift-pipelines-operator-rh` - Tekton-based CI/CD pipelines
- `web-terminal` - Browser-based terminal access

### Security and Compliance
- `rhacs-operator` - Red Hat Advanced Cluster Security
- `compliance-operator` - Security compliance scanning
- `file-integrity-operator` - File integrity monitoring

### Virtualisation and Containers
- `kubevirt-hyperconverged` - Virtual machine management
- `container-security-operator` - Container security scanning

### Networking
- `cluster-network-addons-operator` - Additional networking features
- `sriov-network-operator` - SR-IOV networking support

## Practical Examples

### Select Operators for Cluster Deployment

```hcl
data "oai_supported_operators" "available" {
  openshift_version = "4.16"
  cpu_architecture  = "x86_64"
}

locals {
  # Define required operators for the cluster
  required_operators = [
    "odf-operator",
    "local-storage-operator", 
    "openshift-gitops-operator"
  ]
  
  # Validate all required operators are available
  available_operators = toset(data.oai_supported_operators.available.operators)
  missing_operators = [
    for op in local.required_operators :
    op if !contains(local.available_operators, op)
  ]
}

# Only proceed if all operators are available
resource "oai_cluster" "with_operators" {
  count = length(local.missing_operators) == 0 ? 1 : 0
  
  name              = "cluster-with-operators"
  openshift_version = "4.16"
  # ... other configuration
  
  olm_operators = [
    for op in local.required_operators : {
      name = op
    }
  ]
}

# Output any missing operators for troubleshooting
output "missing_operators" {
  value = local.missing_operators
}
```

### Environment-Specific Operator Selection

```hcl
data "oai_supported_operators" "production" {
  openshift_version = "4.16"
  platform_type     = "baremetal"
}

data "oai_supported_operators" "development" {
  openshift_version = "4.16"
  platform_type     = "baremetal"
}

locals {
  # Production operators - conservative selection
  prod_operators = [
    "odf-operator",
    "local-storage-operator",
    "cluster-logging"
  ]
  
  # Development operators - include additional tools
  dev_operators = concat(local.prod_operators, [
    "openshift-pipelines-operator-rh",
    "web-terminal",
    "kubevirt-hyperconverged"
  ])
}

resource "oai_cluster" "production" {
  # ... basic configuration
  olm_operators = [
    for op in local.prod_operators : {
      name = op
    }
  ]
}

resource "oai_cluster" "development" {
  # ... basic configuration  
  olm_operators = [
    for op in local.dev_operators : {
      name = op
    } if contains(data.oai_supported_operators.development.operators, op)
  ]
}
```

### Architecture-Specific Deployments

```hcl
data "oai_supported_operators" "x86_64" {
  cpu_architecture = "x86_64"
}

data "oai_supported_operators" "arm64" {
  cpu_architecture = "arm64"
}

locals {
  # Operators available on all architectures
  universal_operators = setintersection(
    toset(data.oai_supported_operators.x86_64.operators),
    toset(data.oai_supported_operators.arm64.operators)
  )
}

resource "oai_cluster" "multi_arch" {
  cpu_architecture = "multi"
  # ... other configuration
  
  # Only use operators available on all architectures
  olm_operators = [
    for op in local.universal_operators : {
      name = op
    }
  ]
}
```

### Operator Validation and Documentation

```hcl
data "oai_supported_operators" "current" {
  openshift_version = var.openshift_version
}

# Generate operator documentation
locals {
  operator_categories = {
    storage = [
      "odf-operator",
      "local-storage-operator"
    ]
    monitoring = [
      "cluster-logging",
      "elasticsearch-operator"
    ]
    security = [
      "rhacs-operator",
      "compliance-operator"
    ]
    networking = [
      "metallb-operator",
      "sriov-network-operator"
    ]
  }
  
  available_by_category = {
    for category, ops in local.operator_categories :
    category => [
      for op in ops :
      op if contains(data.oai_supported_operators.current.operators, op)
    ]
  }
}

output "operators_by_category" {
  description = "Available operators organised by category"
  value       = local.available_by_category
}
```

## Integration with Cluster Resources

### Dynamic Operator Selection

```hcl
variable "enable_storage" {
  description = "Enable storage operators"
  type        = bool
  default     = true
}

variable "enable_monitoring" {
  description = "Enable monitoring stack"
  type        = bool
  default     = false
}

data "oai_supported_operators" "available" {
  openshift_version = var.openshift_version
}

locals {
  conditional_operators = concat(
    var.enable_storage ? ["odf-operator", "local-storage-operator"] : [],
    var.enable_monitoring ? ["cluster-logging", "elasticsearch-operator"] : []
  )
  
  # Only include operators that are actually supported
  validated_operators = [
    for op in local.conditional_operators :
    op if contains(data.oai_supported_operators.available.operators, op)
  ]
}

resource "oai_cluster" "conditional" {
  # ... basic configuration
  olm_operators = [
    for op in local.validated_operators : {
      name = op
    }
  ]
}
```

This data source ensures that only supported operators are included in cluster configurations, preventing deployment failures due to operator incompatibilities.