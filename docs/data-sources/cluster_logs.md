---
page_title: "Data Source: openshift_assisted_installer_cluster_logs"
subcategory: "Cluster Management"
---

# openshift_assisted_installer_cluster_logs Data Source

Retrieves installation logs from the Assisted Service API for debugging and monitoring.

## Example Usage

### Get Cluster Installation Logs

```hcl
data "openshift_assisted_installer_cluster_logs" "installation" {
  cluster_id = openshift_assisted_installer_cluster.example.id
  log_type   = "controller"
}

output "controller_logs_url" {
  value = data.openshift_assisted_installer_cluster_logs.installation.download_url
}
```

### Get Host-Specific Logs

```hcl
data "openshift_assisted_installer_cluster_logs" "host_logs" {
  cluster_id = openshift_assisted_installer_cluster.example.id
  host_id    = openshift_assisted_installer_host.master1.id
  log_type   = "host"
}

resource "local_file" "host_logs" {
  content  = data.openshift_assisted_installer_cluster_logs.host_logs.content
  filename = "host-${openshift_assisted_installer_host.master1.id}.log"
}
```

### Retrieve All Available Logs

```hcl
data "openshift_assisted_installer_cluster_logs" "all" {
  cluster_id = openshift_assisted_installer_cluster.example.id
  log_type   = "all"
}

output "available_logs" {
  value = data.openshift_assisted_installer_cluster_logs.all.logs_info
}
```

## Argument Reference

* `cluster_id` - (Required) The cluster ID to retrieve logs for.
* `host_id` - (Optional) Specific host ID for host logs.
* `log_type` - (Optional) Type of logs to retrieve. Valid values:
  * `controller` - Assisted service controller logs
  * `host` - Host discovery and installation logs
  * `all` - All available logs
  * `bootstrap` - Bootstrap node logs
* `pull_secret` - (Optional) Pull secret for accessing logs (if required).

## Attribute Reference

* `id` - The data source ID.
* `cluster_id` - The cluster ID.
* `log_type` - The type of logs retrieved.
* `download_url` - URL to download the logs.
* `content` - Log content (if available and not too large).
* `logs_info` - Information about available logs:
  * `name` - Log name.
  * `type` - Log type.
  * `size` - Log file size.
  * `collected_at` - When the logs were collected.
* `expires_at` - When the log download URL expires.