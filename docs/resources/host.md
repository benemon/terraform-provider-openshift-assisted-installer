---
page_title: "Resource: openshift_assisted_installer_host"
subcategory: "Host Management"
---

# openshift_assisted_installer_host Resource

Manages a discovered host within an infrastructure environment. Hosts are automatically discovered when they boot from the infrastructure environment's discovery ISO.

## Example Usage

### Basic Host Management

```hcl
resource "openshift_assisted_installer_host" "control_plane_1" {
  infra_env_id = openshift_assisted_installer_infra_env.example.id
  host_id      = "550e8400-e29b-41d4-a716-446655440001"
  host_name    = "control-plane-1"
  host_role    = "master"
}
```

### Advanced Host Configuration

```hcl
resource "openshift_assisted_installer_host" "worker_1" {
  infra_env_id = openshift_assisted_installer_infra_env.example.id
  host_id      = "550e8400-e29b-41d4-a716-446655440002"
  
  # Host Identity
  host_name = "worker-1.example.com"
  host_role = "worker"
  
  # Disk Configuration
  installation_disk_id = "/dev/sda"
  disks_skip_formatting = [
    "/dev/sdb",  # Preserve data disk
    "/dev/sdc"   # Preserve additional storage
  ]
}
```

## Argument Reference

### Required Arguments

- `infra_env_id` (String) - ID of the infrastructure environment containing this host.
- `host_id` (String) - Unique identifier of the host. This is provided by the discovery process when the host boots from the discovery ISO.

### Optional Arguments

#### Host Configuration

- `host_name` (String) - Hostname to assign to the host. If not specified, a hostname will be automatically generated.
- `host_role` (String) - Role for the host in the cluster. Valid values: `master`, `worker`, `auto-assign`. Default: `auto-assign`.

#### Disk Configuration

- `installation_disk_id` (String) - Disk device path to use for OpenShift installation (e.g., `/dev/sda`). If not specified, the system will automatically select the most suitable disk.
- `disks_skip_formatting` (List of String) - List of disk device paths to preserve during installation. These disks will not be formatted or partitioned.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

### Computed Attributes

- `id` (String) - Unique identifier of the host resource (combination of infra_env_id and host_id).
- `status` (String) - Current host status. Possible values:
  - `discovering` - Host is being discovered and inventoried
  - `known` - Host discovered but not yet validated
  - `insufficient` - Host discovered but missing requirements
  - `pending-for-input` - Host requires additional configuration
  - `known-unbound` - Host discovered but not bound to a cluster
  - `bound` - Host bound to a cluster and being validated
  - `ready` - Host validated and ready for installation
  - `installing` - Installation in progress on host
  - `installing-in-progress` - Installation actively running
  - `installing-pending-user-action` - Installation paused waiting for user input
  - `installed` - Installation completed successfully
  - `error` - Host in error state
  - `resetting` - Host being reset
  - `resetting-pending-user-action` - Reset paused waiting for user input
- `status_info` (String) - Additional information about the current status.
- `progress` (Object) - Installation progress information. Structure:
  - `current_stage` (String) - Current installation stage
  - `progress_info` (String) - Detailed progress information
  - `stage_started_at` (String) - Timestamp when current stage started
  - `stage_updated_at` (String) - Timestamp of last progress update
- `inventory` (Object) - Hardware inventory discovered from the host. Contains detailed information about CPU, memory, disks, and network interfaces.

## Import

Hosts can be imported using the format `infra_env_id/host_id`:

```shell
terraform import openshift_assisted_installer_host.example 550e8400-e29b-41d4-a716-446655440000/550e8400-e29b-41d4-a716-446655440001
```

## Host Discovery Process

### Discovery Workflow

1. **Boot from Discovery ISO** - Host boots from the infrastructure environment's discovery ISO
2. **Hardware Inventory** - Discovery agent inventories CPU, memory, disks, and network interfaces
3. **Host Registration** - Host registers with the Assisted Service API
4. **Validation** - System validates hardware meets OpenShift requirements
5. **Ready State** - Host becomes available for cluster assignment and installation

### Finding Host IDs

Host IDs are generated during the discovery process. To find available hosts:

```bash
# List hosts in an infrastructure environment
terraform show -json | jq -r '.values.root_module.resources[] | select(.type=="openshift_assisted_installer_infra_env") | .values.hosts[].id'

# Or use the Assisted Service API directly
curl -H "Authorization: Bearer $TOKEN" \
  "https://api.openshift.com/api/assisted-install/v2/infra-envs/$INFRA_ENV_ID/hosts"
```

### Automatic vs Manual Host Management

**Automatic Management** (Recommended for most cases):
- Let the cluster resource automatically assign roles based on `control_plane_count`
- First N hosts become masters, remaining hosts become workers
- No explicit host resources needed

**Manual Management** (Use when you need specific control):
- Create explicit `openshift_assisted_installer_host` resources for each discovered host
- Specify exact roles, hostnames, and disk configurations
- Required for complex disk layouts or specific host assignments

## Host Roles

### Master Nodes
- Run the OpenShift control plane (etcd, API server, scheduler)
- Minimum 4 CPU cores and 16 GB RAM required
- Typically 3 masters for high availability

### Worker Nodes  
- Run application workloads
- Minimum 2 CPU cores and 8 GB RAM required
- Can be added or removed from clusters dynamically

### Auto-Assign
- System automatically assigns role based on cluster requirements
- Masters assigned first up to `control_plane_count`
- Remaining hosts become workers

## Disk Management

### Installation Disk Selection

The installation disk hosts the OpenShift operating system and container storage:

```hcl
resource "openshift_assisted_installer_host" "example" {
  installation_disk_id = "/dev/nvme0n1"  # Use fastest available disk
  # System will automatically partition and format this disk
}
```

### Preserving Data Disks

To preserve existing data on specific disks:

```hcl
resource "openshift_assisted_installer_host" "example" {
  disks_skip_formatting = [
    "/dev/sdb",  # Database storage
    "/dev/sdc"   # Application data
  ]
  # These disks will not be touched during installation
}
```

## Troubleshooting

### Host Not Discovered

If a host doesn't appear after booting from the discovery ISO:

1. Verify network connectivity from the discovery environment
2. Check that the host meets minimum hardware requirements
3. Review console logs on the host during boot
4. Ensure the discovery ISO matches the infrastructure environment's CPU architecture

### Host in Error State

If a host transitions to error state:

1. Check the `status_info` attribute for specific error details
2. Verify hardware compatibility with OpenShift requirements
3. Address any network or storage issues identified
4. Consider resetting the host and retrying discovery

### Installation Failures

If installation fails on a specific host:

1. Review installation logs from the host console
2. Check for disk space or hardware compatibility issues
3. Verify network connectivity during installation
4. Consider excluding problematic hosts and adding replacement hosts

## Examples

See the [examples directory](https://github.com/benemon/terraform-provider-openshift-assisted-installer/tree/main/examples) for complete configuration examples including host role assignment and disk management scenarios.