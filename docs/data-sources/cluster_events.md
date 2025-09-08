---
page_title: "Data Source: openshift_assisted_installer_cluster_events"
subcategory: "Cluster Management"
---

# openshift_assisted_installer_cluster_events Data Source

Retrieves cluster and host events from the Assisted Service API for monitoring and troubleshooting.

## Example Usage

### Get All Cluster Events

```hcl
data "openshift_assisted_installer_cluster_events" "all" {
  cluster_id = openshift_assisted_installer_cluster.example.id
}

output "recent_events" {
  value = [
    for event in slice(data.openshift_assisted_installer_cluster_events.all.events, 0, 10) :
    "${event.event_time}: ${event.severity} - ${event.message}"
  ]
}
```

### Filter Events by Severity

```hcl
data "openshift_assisted_installer_cluster_events" "errors" {
  cluster_id = openshift_assisted_installer_cluster.example.id
  severities = ["error", "critical"]
}

output "error_events" {
  value = [
    for event in data.openshift_assisted_installer_cluster_events.errors.events :
    "${event.name}: ${event.message}"
  ]
}
```

### Get Host-Specific Events

```hcl
data "openshift_assisted_installer_cluster_events" "host_events" {
  cluster_id = openshift_assisted_installer_cluster.example.id
  host_id    = openshift_assisted_installer_host.master1.id
}

output "host_event_summary" {
  value = {
    total = length(data.openshift_assisted_installer_cluster_events.host_events.events)
    errors = length([
      for e in data.openshift_assisted_installer_cluster_events.host_events.events :
      e if e.severity == "error"
    ])
  }
}
```

### Filter by Event Category

```hcl
data "openshift_assisted_installer_cluster_events" "user_events" {
  cluster_id = openshift_assisted_installer_cluster.example.id
  categories = ["user"]
  limit      = 50
}
```

## Argument Reference

* `cluster_id` - (Optional) The cluster ID to retrieve events for.
* `host_id` - (Optional) Filter events for a specific host.
* `infra_env_id` - (Optional) Filter events for a specific infrastructure environment.
* `severities` - (Optional) List of severities to filter by. Valid values: `info`, `warning`, `error`, `critical`.
* `categories` - (Optional) List of categories to filter by. Valid values: `user`, `system`.
* `message` - (Optional) Filter events containing this message text.
* `cluster_level` - (Optional) Whether to retrieve cluster-level events only.
* `limit` - (Optional) Maximum number of events to retrieve (default: 100).
* `offset` - (Optional) Number of events to skip for pagination.
* `order` - (Optional) Sort order. Valid values: `ascending`, `descending` (default).

## Attribute Reference

* `id` - The data source ID.
* `events` - List of events with the following attributes:
  * `name` - Event name/type.
  * `cluster_id` - Associated cluster ID.
  * `host_id` - Associated host ID (if applicable).
  * `infra_env_id` - Associated infrastructure environment ID.
  * `severity` - Event severity level.
  * `category` - Event category.
  * `message` - Event message.
  * `event_time` - When the event occurred.
  * `request_id` - Associated API request ID.