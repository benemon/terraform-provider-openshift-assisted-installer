---
page_title: "Resource: oai_infra_env"
subcategory: "Core Resources"
---

# oai_infra_env Resource

Manages an infrastructure environment for the OpenShift Assisted Installer. Infrastructure environments generate discovery ISOs that hosts boot from to register themselves with the cluster.

## Example Usage

### Basic Infrastructure Environment

```hcl
resource "oai_infra_env" "example" {
  name             = "cluster-infra"
  pull_secret     = var.pull_secret
  cpu_architecture = "x86_64"
  cluster_id      = oai_cluster.example.id
  ssh_authorized_key = var.ssh_public_key
}
```

### Advanced Configuration

```hcl
resource "oai_infra_env" "advanced" {
  name             = "production-infra"
  pull_secret     = var.pull_secret
  cpu_architecture = "x86_64"
  cluster_id      = oai_cluster.example.id
  
  # SSH Configuration
  ssh_authorized_key = var.ssh_public_key
  
  # Image Configuration
  image_type = "full-iso"
  
  # Network Configuration
  proxy {
    http_proxy  = "http://proxy.example.com:8080"
    https_proxy = "http://proxy.example.com:8080"
    no_proxy    = "localhost,127.0.0.1,.example.com"
  }
  
  # Static Network Configuration
  static_network_config {
    network_yaml = yamlencode({
      dns-resolver = {
        config = {
          server = ["8.8.8.8", "8.8.4.4"]
        }
      }
      routes = {
        config = [{
          destination      = "0.0.0.0/0"
          next_hop_address = "192.168.1.1"
          next_hop_interface = "ens3"
        }]
      }
      interfaces = [{
        name = "ens3"
        type = "ethernet"
        state = "up"
        ipv4 = {
          enabled = true
          address = [{
            ip           = "192.168.1.100"
            prefix_length = 24
          }]
        }
      }]
    })
    
    mac_interface_map {
      mac_address      = "52:54:00:12:34:56"
      logical_nic_name = "ens3"
    }
  }
  
  # Additional Configuration
  additional_ntp_sources = "pool.ntp.org,time.google.com"
  additional_trust_bundle = file("${path.module}/ca-bundle.crt")
  
  # Kernel Arguments
  kernel_arguments {
    operation = "append"
    value    = "console=ttyS0,115200n8"
  }
  
  # Ignition Override
  ignition_config_override = file("${path.module}/ignition-override.json")
}
```

## Argument Reference

### Required Arguments

- `name` (String) - Name of the infrastructure environment.
- `pull_secret` (String, Sensitive) - Red Hat pull secret in JSON format.
- `cpu_architecture` (String) - Target CPU architecture. Valid values: `x86_64`, `arm64`, `ppc64le`, `s390x`, `multi`.

### Optional Arguments

#### Cluster Association

- `cluster_id` (String) - ID of the cluster to associate with this infrastructure environment. If specified, discovered hosts will be automatically associated with the cluster.
- `openshift_version` (String) - OpenShift version override. If not specified, uses the cluster's version.

#### SSH Configuration

- `ssh_authorized_key` (String) - SSH public key to inject into discovered hosts for debugging access.

#### Image Configuration

- `image_type` (String) - Type of discovery image to generate. Valid values: `full-iso` (includes all dependencies), `minimal-iso` (requires network access). Default: `minimal-iso`.

#### Network Configuration

- `proxy` (Block) - Proxy configuration for discovered hosts. Structure:
  - `http_proxy` (String) - HTTP proxy URL
  - `https_proxy` (String) - HTTPS proxy URL  
  - `no_proxy` (String) - Comma-separated list of hosts to bypass proxy

- `static_network_config` (Block Set) - Static network configuration for hosts. Multiple blocks can be specified for different hosts. Structure:
  - `network_yaml` (String) - Network configuration in YAML format using NetworkManager syntax
  - `mac_interface_map` (Block Set) - Mapping of MAC addresses to logical interface names. Structure:
    - `mac_address` (String) - MAC address of the network interface
    - `logical_nic_name` (String) - Logical name to assign to the interface

#### Additional Configuration

- `additional_ntp_sources` (String) - Comma-separated list of additional NTP servers.
- `additional_trust_bundle` (String) - Additional certificate bundle to trust during discovery and installation.

#### Kernel Configuration

- `kernel_arguments` (Block Set) - Kernel argument modifications. Structure:
  - `operation` (String) - Operation type. Valid values: `append`, `replace`, `delete`
  - `value` (String) - Kernel argument value

#### Advanced Configuration

- `ignition_config_override` (String) - Custom Ignition configuration to merge with the generated configuration.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

### Computed Attributes

- `id` (String) - Unique identifier of the infrastructure environment.
- `download_url` (String) - URL to download the generated discovery ISO.
- `expires_at` (String) - Expiration timestamp for the discovery ISO.
- `size_bytes` (Number) - Size of the discovery ISO in bytes.

## Import

Infrastructure environments can be imported using their ID:

```shell
terraform import oai_infra_env.example 550e8400-e29b-41d4-a716-446655440000
```

## Discovery ISO Usage

### Downloading the ISO

Once the infrastructure environment is created, download the discovery ISO:

```bash
curl -L -o discovery.iso "$(terraform output -raw infra_env_download_url)"
```

### Creating Bootable Media

Create bootable USB drives for each target host:

```bash
# Linux
sudo dd if=discovery.iso of=/dev/sdX bs=4M status=progress

# macOS  
sudo dd if=discovery.iso of=/dev/diskX bs=4m

# Windows
# Use tools like Rufus or balenaEtcher
```

### Host Boot Process

1. Insert the bootable USB drive into the target host
2. Boot the host from the USB drive
3. The discovery agent automatically starts and inventories hardware
4. Host registers with the infrastructure environment
5. Remove USB drive once host appears in the cluster (typically 3-5 minutes)

### Sequential Host Discovery

You can use a single USB drive to boot multiple hosts sequentially:

1. Boot first host and wait for discovery completion
2. Remove USB drive and insert into second host
3. Repeat for all target hosts
4. All hosts will remain registered with the cluster

## Static Network Configuration

For environments without DHCP, static network configuration can be provided using NetworkManager YAML syntax:

```hcl
static_network_config {
  network_yaml = yamlencode({
    interfaces = [{
      name  = "ens3"
      type  = "ethernet" 
      state = "up"
      ipv4 = {
        enabled = true
        address = [{
          ip            = "192.168.1.100"
          prefix_length = 24
        }]
      }
    }]
    routes = {
      config = [{
        destination         = "0.0.0.0/0"
        next_hop_address   = "192.168.1.1"
        next_hop_interface = "ens3"
      }]
    }
    dns-resolver = {
      config = {
        server = ["8.8.8.8", "8.8.4.4"]
      }
    }
  })
  
  mac_interface_map {
    mac_address      = "52:54:00:12:34:56"
    logical_nic_name = "ens3"
  }
}
```

## Examples

See the [examples directory](https://github.com/benemon/terraform-provider-openshift-assisted-installer/tree/main/examples) for complete configuration examples including static networking and proxy configurations.