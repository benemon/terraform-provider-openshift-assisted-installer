---
page_title: "Data Source: oai_cluster"
subcategory: "Cluster Management"
---

# oai_cluster Data Source

Retrieves information about an existing OpenShift cluster from the Assisted Service API.

## Example Usage

### Read Existing Cluster

```hcl
data "oai_cluster" "existing" {
  cluster_id = "550e8400-e29b-41d4-a716-446655440000"
}

output "cluster_info" {
  value = {
    name   = data.oai_cluster.existing.name
    status = data.oai_cluster.existing.status
    api_vip = data.oai_cluster.existing.api_vips[0].ip
  }
}
```

### Check Cluster Installation Status

```hcl
data "oai_cluster" "installation" {
  cluster_id = var.cluster_id
}

locals {
  is_installed = data.oai_cluster.installation.status == "installed"
  is_ready     = data.oai_cluster.installation.status == "ready"
  is_error     = data.oai_cluster.installation.status == "error"
}

output "installation_progress" {
  value = data.oai_cluster.installation.progress
}
```

### Get Cluster Network Configuration

```hcl
data "oai_cluster" "network" {
  cluster_id = var.cluster_id
}

output "network_config" {
  value = {
    cluster_cidr = data.oai_cluster.network.cluster_network_cidr
    service_cidr = data.oai_cluster.network.service_network_cidr
    api_vips     = data.oai_cluster.network.api_vips
    ingress_vips = data.oai_cluster.network.ingress_vips
  }
}
```

## Argument Reference

* `cluster_id` - (Required) The ID of the cluster to retrieve.

## Attribute Reference

* `id` - The cluster ID.
* `name` - The cluster name.
* `openshift_version` - The OpenShift version.
* `base_dns_domain` - The base DNS domain.
* `status` - Current cluster status.
* `status_info` - Detailed status information.
* `progress` - Installation progress information.
* `created_at` - Cluster creation timestamp.
* `updated_at` - Last update timestamp.
* `installed_at` - Installation completion timestamp.
* `cpu_architecture` - CPU architecture (x86_64, arm64, etc.).
* `platform` - Platform configuration.
* `cluster_network_cidr` - Cluster network CIDR.
* `service_network_cidr` - Service network CIDR.
* `api_vips` - List of API VIP configurations.
* `ingress_vips` - List of Ingress VIP configurations.
* `vip_dhcp_allocation` - Whether DHCP is used for VIP allocation.
* `ssh_public_key` - SSH public key for cluster access.
* `user_managed_networking` - Whether networking is user-managed.
* `host_count` - Number of hosts in the cluster.
* `enabled_host_count` - Number of enabled hosts.
* `monitored_operators` - List of monitored operators with their status.
* `image_info` - Discovery image information.
* `validations_info` - Validation results (use `oai_cluster_validations` for detailed filtering).
