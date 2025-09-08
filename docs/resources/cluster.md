---
page_title: "Resource: oai_cluster"
subcategory: "Cluster Management"
---

# oai_cluster Resource

Manages an OpenShift cluster using the Assisted Service API. This resource handles the complete cluster lifecycle from initial definition through to installation completion.

## Example Usage

### Basic Cluster

```hcl
resource "oai_cluster" "example" {
  name                 = "production-cluster"
  openshift_version    = "4.16.0"
  pull_secret         = var.pull_secret
  cpu_architecture    = "x86_64"
  control_plane_count = 3
}

# Trigger installation separately
resource "oai_cluster_installation" "example" {
  cluster_id          = oai_cluster.example.id
  wait_for_hosts      = true
  expected_host_count = 3
  
  timeouts {
    create = "120m"
  }
}
```

### Advanced Configuration

```hcl
resource "oai_cluster" "advanced" {
  name                     = "advanced-cluster"
  openshift_version        = "4.16.0"
  pull_secret             = var.pull_secret
  cpu_architecture        = "x86_64"
  control_plane_count     = 3
  
  # Networking Configuration
  base_dns_domain         = "example.com"
  cluster_network_cidr    = "10.128.0.0/14"
  service_network_cidr    = "172.30.0.0/16"
  api_vips               = ["192.168.1.100"]
  ingress_vips           = ["192.168.1.101"]
  user_managed_networking = false
  
  # Proxy Configuration
  http_proxy  = "http://proxy.example.com:8080"
  https_proxy = "http://proxy.example.com:8080"
  no_proxy    = "localhost,127.0.0.1,.example.com"
  
  # Additional Configuration
  ssh_public_key         = var.ssh_public_key
  additional_ntp_source  = "pool.ntp.org"
  hyperthreading        = "all"
  
  # Timeouts
  timeouts {
    create = "120m"
    update = "60m"
  }
}
```

## Argument Reference

### Required Arguments

- `name` (String) - Name of the cluster. Must be unique within your organisation.
- `openshift_version` (String) - OpenShift version to install. Use data source `oai_openshift_versions` to discover available versions.
- `pull_secret` (String, Sensitive) - Red Hat pull secret in JSON format. Obtain from console.redhat.com.
- `cpu_architecture` (String) - Target CPU architecture. Valid values: `x86_64`, `arm64`, `ppc64le`, `s390x`, `multi`.

### Optional Arguments

#### Cluster Configuration

- `control_plane_count` (Number) - Number of control plane nodes. Valid values: 1 (single node), 3, 4, or 5. Default: 3.
- `base_dns_domain` (String) - Base DNS domain for the cluster. Must be a valid DNS domain name.
- `ssh_public_key` (String) - SSH public key for accessing cluster nodes.

#### Networking Configuration

- `cluster_network_cidr` (String) - CIDR range for pod network. Default: `10.128.0.0/14`.
- `cluster_network_host_prefix` (Number) - Host subnet prefix length for pod network.
- `service_network_cidr` (String) - CIDR range for service network. Default: `172.30.0.0/16`.
- `api_vips` (List of String) - Virtual IP addresses for API servers. Required for multi-node clusters with static networking.
- `ingress_vips` (List of String) - Virtual IP addresses for ingress routers.
- `vip_dhcp_allocation` (Boolean) - Whether to allocate VIPs via DHCP. Default: false.
- `user_managed_networking` (Boolean) - Whether networking is user-managed. Default: false.
- `network_type` (String) - Network plugin type. Valid values depend on OpenShift version.

#### Proxy Configuration

- `http_proxy` (String) - HTTP proxy URL for cluster nodes.
- `https_proxy` (String) - HTTPS proxy URL for cluster nodes.
- `no_proxy` (String) - Comma-separated list of hosts to bypass proxy.

#### Additional Configuration

- `additional_ntp_source` (String) - Additional NTP server for time synchronisation.
- `hyperthreading` (String) - Hyperthreading configuration. Valid values: `all`, `masters`, `workers`, `none`.
- `high_availability_mode` (String) - **Deprecated**: Use `control_plane_count` instead.

#### Timeouts

- `timeouts.create` (String) - Timeout for cluster creation and installation. Default: `90m`.
- `timeouts.update` (String) - Timeout for cluster updates. Default: `60m`.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

### Computed Attributes

- `id` (String) - Unique identifier of the cluster.
- `status` (String) - Current cluster status. Possible values:
  - `insufficient` - Cluster definition created but not ready for installation
  - `ready` - All prerequisites met, ready for installation
  - `installing` - Installation in progress
  - `installed` - Installation completed successfully
  - `error` - Installation failed
- `status_info` (String) - Additional information about the current status.
- `install_completed` (Boolean) - Whether installation has completed successfully.
- `kind` (String) - Resource type identifier.
- `href` (String) - API href for the cluster resource.

## Import

Clusters can be imported using their ID:

```shell
terraform import oai_cluster.example 550e8400-e29b-41d4-a716-446655440000
```

## State Management

### Installation States

The cluster progresses through several states during installation:

1. **insufficient** - Initial state after creation, waiting for hosts to be discovered
2. **ready** - All hosts discovered and validated, ready for installation
3. **installing** - Installation triggered and in progress
4. **installed** - Installation completed successfully

### Error Handling

If installation fails, the cluster will transition to an `error` state. The `status_info` attribute provides details about the failure. You can inspect the cluster state and either:

- Address the underlying issue and retry
- Destroy and recreate the cluster
- Import the cluster into a separate configuration for troubleshooting

### Updates

Most cluster configuration can be updated after creation, but before installation begins. Once installation has started, only limited fields can be modified. Configuration changes that require replacement will be clearly indicated by Terraform's plan output.

## Examples

See the [examples directory](https://github.com/benemon/terraform-provider-openshift-assisted-installer/tree/main/examples) for complete configuration examples including:

- Single node clusters
- Three node compact clusters  
- Standard multi-node clusters with workers
- Clusters with custom networking and proxy configuration
