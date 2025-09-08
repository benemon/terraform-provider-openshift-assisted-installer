---
page_title: "Data Source: openshift_assisted_installer_host_validations"
subcategory: "Host Management"
---

# openshift_assisted_installer_host_validations Data Source

Retrieves host validation results from the Assisted Service API. This data source supports two modes: cluster-wide host validation checking or single-host validation checking.

## Example Usage

### Get All Host Validations for a Cluster

```hcl
data "openshift_assisted_installer_host_validations" "cluster_hosts" {
  cluster_id = openshift_assisted_installer_cluster.example.id
}

output "host_readiness" {
  value = {
    for v in data.openshift_assisted_installer_host_validations.cluster_hosts.validations :
    v.host_id => {
      status   = v.status
      failures = length([for val in v.validations : val if val.status == "failure"])
    }
  }
}
```

### Check Single Host Validations

```hcl
data "openshift_assisted_installer_host_validations" "single_host" {
  infra_env_id = openshift_assisted_installer_infra_env.example.id
  host_id      = openshift_assisted_installer_host.worker.id
}

output "host_validation_details" {
  value = [
    for v in data.openshift_assisted_installer_host_validations.single_host.validations[0].validations :
    "${v.id}: ${v.status} - ${v.message}"
  ]
}
```

### Filter Hardware Validations

```hcl
data "openshift_assisted_installer_host_validations" "hardware" {
  cluster_id = openshift_assisted_installer_cluster.example.id
  categories = ["hardware"]
}

# Check if all hosts meet hardware requirements
locals {
  hardware_ready = alltrue([
    for host in data.openshift_assisted_installer_host_validations.hardware.validations :
    alltrue([
      for v in host.validations :
      v.status == "success"
    ])
  ])
}
```

### Filter Blocking Host Validations

```hcl
data "openshift_assisted_installer_host_validations" "blocking" {
  cluster_id      = openshift_assisted_installer_cluster.example.id
  validation_type = "blocking"
  status          = "failure"
}

output "blocking_issues" {
  value = flatten([
    for host in data.openshift_assisted_installer_host_validations.blocking.validations : [
      for v in host.validations :
      "${host.host_name}: ${v.id} - ${v.message}"
    ]
  ])
}
```

## Argument Reference

* `cluster_id` - (Optional) The ID of the cluster to retrieve host validations for. Use for cluster-wide checks.
* `infra_env_id` - (Optional) The infrastructure environment ID. Required when checking a single host.
* `host_id` - (Optional) The specific host ID to check. Requires `infra_env_id`.
* `validation_type` - (Optional) Filter by validation type. Valid values: `blocking`, `non-blocking`.
* `status` - (Optional) Filter by validation status. Valid values: `success`, `failure`, `pending`.
* `validation_names` - (Optional) List of specific validation IDs to check.
* `categories` - (Optional) List of validation categories. Valid values: `hardware`, `network`, `operators`, `platform`, `storage`.

**Note:** Either `cluster_id` OR both `infra_env_id` and `host_id` must be specified.

## Attribute Reference

* `id` - The data source ID.
* `validations` - List of host validation results:
  * `host_id` - The host identifier.
  * `host_name` - The hostname.
  * `role` - The host role (`master`, `worker`, `auto-assign`).
  * `status` - Overall host status.
  * `validations` - List of individual validations:
    * `id` - The validation identifier.
    * `status` - The validation status.
    * `message` - Human-readable validation message.
    * `category` - The validation category.

## Host Validation Categories

* **hardware** - CPU, memory, disk requirements
* **network** - Network connectivity and configuration
* **operators** - Operator-specific host requirements
* **platform** - Platform compatibility checks
* **storage** - Storage configuration validations