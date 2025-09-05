---
page_title: "Data Source: oai_host"
subcategory: "Host Management"
---

# oai_host Data Source

Retrieves information about a discovered host from the Assisted Service API.

## Example Usage

### Read Host Information

```hcl
data "oai_host" "worker" {
  infra_env_id = oai_infra_env.example.id
  host_id      = "750e8400-e29b-41d4-a716-446655440000"
}

output "host_details" {
  value = {
    hostname = data.oai_host.worker.requested_hostname
    role     = data.oai_host.worker.role
    status   = data.oai_host.worker.status
  }
}
```

### Check Host Hardware Inventory

```hcl
data "oai_host" "master" {
  infra_env_id = var.infra_env_id
  host_id      = var.host_id
}

output "hardware_info" {
  value = {
    cpu_cores = data.oai_host.master.inventory.cpu.count
    memory_gb = data.oai_host.master.inventory.memory.physical_bytes / 1073741824
    disk_count = length(data.oai_host.master.inventory.disks)
  }
}
```

## Argument Reference

* `infra_env_id` - (Required) The infrastructure environment ID containing the host.
* `host_id` - (Required) The ID of the host to retrieve.

## Attribute Reference

* `id` - The host ID.
* `infra_env_id` - The infrastructure environment ID.
* `cluster_id` - Associated cluster ID (if bound to a cluster).
* `status` - Current host status.
* `status_info` - Detailed status information.
* `role` - Host role (master, worker, auto-assign).
* `requested_hostname` - Requested hostname.
* `discovered_hostname` - Discovered hostname.
* `installation_disk_id` - Selected installation disk ID.
* `inventory` - Hardware inventory information:
  * `cpu` - CPU information (cores, architecture, frequency).
  * `memory` - Memory information.
  * `disks` - List of discovered disks.
  * `interfaces` - Network interfaces.
  * `system_vendor` - System vendor information.
* `progress` - Installation progress.
* `validations_info` - Host validation results.
* `created_at` - Discovery timestamp.
* `updated_at` - Last update timestamp.
* `checked_in_at` - Last check-in timestamp.