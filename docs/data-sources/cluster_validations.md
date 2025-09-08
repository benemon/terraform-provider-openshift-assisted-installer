---
page_title: "Data Source: oai_cluster_validations"
subcategory: "Cluster Management"
---

# oai_cluster_validations Data Source

Retrieves cluster validation results from the Assisted Service API. Use this data source to check pre-installation validation status and troubleshoot cluster readiness issues.

## Example Usage

### Get All Cluster Validations

```hcl
data "oai_cluster_validations" "cluster_checks" {
  cluster_id = oai_cluster.example.id
}

output "validation_summary" {
  value = {
    total    = length(data.oai_cluster_validations.cluster_checks.validations)
    failures = length([for v in data.oai_cluster_validations.cluster_checks.validations : v if v.status == "failure"])
  }
}
```

### Filter Blocking Validations Only

```hcl
data "oai_cluster_validations" "blocking" {
  cluster_id        = oai_cluster.example.id
  validation_type   = "blocking"
}

# Check if cluster is ready to install
locals {
  ready_to_install = length([
    for v in data.oai_cluster_validations.blocking.validations :
    v if v.status == "failure"
  ]) == 0
}
```

### Check Specific Validations

```hcl
data "oai_cluster_validations" "network_checks" {
  cluster_id       = oai_cluster.example.id
  validation_names = [
    "api-vips-defined",
    "api-vips-valid",
    "ingress-vips-defined",
    "ingress-vips-valid",
    "no-cidrs-overlapping"
  ]
}
```

### Filter by Category

```hcl
data "oai_cluster_validations" "operators" {
  cluster_id = oai_cluster.example.id
  categories = ["operators"]
}

output "operator_validations" {
  value = [
    for v in data.oai_cluster_validations.operators.validations :
    "${v.id}: ${v.status} - ${v.message}"
  ]
}
```

## Argument Reference

* `cluster_id` - (Required) The ID of the cluster to retrieve validations for.
* `validation_type` - (Optional) Filter by validation type. Valid values: `blocking`, `non-blocking`.
* `status` - (Optional) Filter by validation status. Valid values: `success`, `failure`, `pending`.
* `validation_names` - (Optional) List of specific validation IDs to check.
* `categories` - (Optional) List of validation categories to filter by. Valid values: `network`, `cluster`, `operators`, `hardware`, `platform`, `storage`.

## Attribute Reference

* `id` - The data source ID (same as cluster_id).
* `validations` - List of validation results with the following attributes:
  * `id` - The validation identifier.
  * `status` - The validation status (`success`, `failure`, `pending`).
  * `message` - Human-readable validation message.
  * `category` - The validation category.

## Validation Categories

* **network** - Network configuration validations (VIPs, CIDR ranges, connectivity)
* **cluster** - Cluster-level validations (host count, roles, general requirements)
* **operators** - OLM operator compatibility and requirements
* **hardware** - Hardware requirements and compatibility
* **platform** - Platform-specific validations (vSphere, bare metal, etc.)
* **storage** - Storage configuration and requirements