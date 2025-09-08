---
page_title: "Data Source: openshift_assisted_installer_infra_env"
subcategory: "Infrastructure Environment"
---

# openshift_assisted_installer_infra_env Data Source

Retrieves information about an existing infrastructure environment from the Assisted Service API.

## Example Usage

### Read Existing Infrastructure Environment

```hcl
data "openshift_assisted_installer_infra_env" "existing" {
  infra_env_id = "650e8400-e29b-41d4-a716-446655440000"
}

output "iso_download_url" {
  value = data.openshift_assisted_installer_infra_env.existing.download_url
}
```

### Get Infrastructure Environment for Cluster

```hcl
data "openshift_assisted_installer_infra_env" "cluster_env" {
  infra_env_id = openshift_assisted_installer_infra_env.example.id
}

output "discovered_hosts" {
  value = data.openshift_assisted_installer_infra_env.cluster_env.host_count
}
```

## Argument Reference

* `infra_env_id` - (Required) The ID of the infrastructure environment to retrieve.

## Attribute Reference

* `id` - The infrastructure environment ID.
* `name` - The infrastructure environment name.
* `cluster_id` - Associated cluster ID (if bound to a cluster).
* `status` - Current status.
* `cpu_architecture` - CPU architecture for the discovery ISO.
* `openshift_version` - OpenShift version.
* `image_type` - ISO image type (full-iso or minimal-iso).
* `download_url` - URL to download the discovery ISO.
* `expires_at` - ISO expiration timestamp.
* `created_at` - Creation timestamp.
* `updated_at` - Last update timestamp.
* `host_count` - Number of discovered hosts.
* `ssh_authorized_key` - SSH public key for host access.
* `proxy` - Proxy configuration.
* `additional_ntp_sources` - Additional NTP sources.
* `static_network_config` - Static network configuration.
* `kernel_arguments` - Kernel arguments for the discovery ISO.