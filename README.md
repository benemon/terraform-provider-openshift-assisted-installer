# Terraform Provider for OpenShift Assisted Installer

This Terraform provider enables Infrastructure as Code management of OpenShift clusters using the [OpenShift Assisted Service API](https://api.openshift.com/api/assisted-install).

## Features

- **Complete Cluster Lifecycle**: Create, install, and manage OpenShift clusters end-to-end
- **Infrastructure Environments**: Generate discovery ISOs for host bootstrapping 
- **Host Management**: Discover, configure, and manage cluster hosts
- **Custom Manifests**: Apply custom configuration manifests to clusters
- **Installation Monitoring**: Track installation progress and troubleshoot issues
- **Post-Installation Access**: Retrieve cluster credentials, logs, and configuration files
- **Data Sources**: Query supported OpenShift versions, operators, and cluster information
- **Asynchronous Operations**: Handle long-running installations with configurable timeouts

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23
- OpenShift Assisted Service API token

## Installation

### From Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    oai = {
      source  = "benemon/oai"
      version = "~> 1.0"
    }
  }
}
```

### Local Development

1. Clone this repository
2. Build the provider:
   ```bash
   go install
   ```
3. Create a `.terraformrc` file in your home directory:
   ```hcl
   provider_installation {
     dev_overrides {
       "benemon/oai" = "<GOPATH>/bin"
     }
     direct {}
   }
   ```

## Authentication

The provider supports authentication via Red Hat offline tokens:

1. **Provider configuration**:
   ```hcl
   provider "oai" {
     offline_token = "your-offline-token"
   }
   ```

2. **Environment variable**:
   ```bash
   export OFFLINE_TOKEN="your-offline-token"
   ```

Get your offline token from the [Red Hat Console](https://console.redhat.com/openshift/token/show).

## Usage

### Complete Cluster Workflow

```hcl
# Provider configuration
provider "oai" {
  # Uses OFFLINE_TOKEN environment variable
}

# Get latest OpenShift version
data "oai_openshift_versions" "latest" {
  only_latest = true
}

# Create cluster definition
resource "oai_cluster" "example" {
  name                 = "my-cluster"
  base_dns_domain      = "example.com"
  openshift_version    = data.oai_openshift_versions.latest.versions[0].version
  cpu_architecture     = "x86_64"
  control_plane_count  = 3
  schedulable_masters  = false
  
  # Network configuration
  api_vips = [{ ip = "192.168.1.100" }]
  ingress_vips = [{ ip = "192.168.1.101" }]
  
  # Required secrets
  pull_secret    = var.pull_secret
  ssh_public_key = var.ssh_public_key
}

# Create infrastructure environment for host discovery
resource "oai_infra_env" "example" {
  name              = "${oai_cluster.example.name}-infra"
  cluster_id        = oai_cluster.example.id
  cpu_architecture  = "x86_64"
  pull_secret       = var.pull_secret
  ssh_authorized_key = var.ssh_public_key
  image_type        = "full-iso"
}

# Trigger cluster installation once hosts are ready
resource "oai_cluster_installation" "example" {
  cluster_id          = oai_cluster.example.id
  wait_for_hosts      = true
  expected_host_count = 3
  
  timeouts {
    create = "120m"
  }
}

# Get cluster credentials after installation
data "oai_cluster_credentials" "admin" {
  cluster_id = oai_cluster.example.id
  depends_on = [oai_cluster_installation.example]
}

# Monitor installation progress
data "oai_cluster_events" "progress" {
  cluster_id = oai_cluster.example.id
  severities = ["info", "warning", "error"]
  limit      = 50
}
```

### Post-Installation Access

```hcl
# Access cluster credentials
output "cluster_access" {
  value = {
    username    = data.oai_cluster_credentials.admin.username
    password    = data.oai_cluster_credentials.admin.password
    console_url = data.oai_cluster_credentials.admin.console_url
  }
  sensitive = true
}

# Download kubeconfig file
data "oai_cluster_files" "kubeconfig" {
  cluster_id = oai_cluster.example.id
  file_name  = "kubeconfig"
  depends_on = [oai_cluster_installation.example]
}

# Save kubeconfig locally
resource "local_file" "kubeconfig" {
  content  = data.oai_cluster_files.kubeconfig.content
  filename = "${path.module}/kubeconfig"
}

# Get cluster logs for troubleshooting
data "oai_cluster_logs" "installation" {
  cluster_id = oai_cluster.example.id
  logs_type  = "controller"
}
```

### Custom Manifests

```hcl
resource "oai_manifest" "example" {
  cluster_id = oai_cluster.example.id
  folder     = "manifests"
  file_name  = "custom-config.yaml"
  
  content = templatefile("${path.module}/manifests/custom-config.yaml", {
    cluster_name = oai_cluster.example.name
  })
}
```

## Resources

- **`oai_cluster`** - OpenShift cluster definition and configuration
- **`oai_cluster_installation`** - Trigger and monitor cluster installation
- **`oai_infra_env`** - Infrastructure environment for host discovery
- **`oai_host`** - Individual host configuration and management
- **`oai_manifest`** - Custom cluster manifests and configurations

## Data Sources

### Cluster Information
- **`oai_openshift_versions`** - Available OpenShift versions and release information
- **`oai_supported_operators`** - Supported OLM operators for cluster installation
- **`oai_operator_bundles`** - Available operator bundles and dependencies
- **`oai_support_levels`** - Feature support levels by platform and architecture

### Post-Installation Access
- **`oai_cluster_credentials`** - Cluster admin credentials (username, password, console URL)
- **`oai_cluster_events`** - Installation and cluster events for monitoring and troubleshooting
- **`oai_cluster_logs`** - Cluster installation and runtime logs
- **`oai_cluster_files`** - Cluster configuration files (kubeconfig, manifests, ignition configs)

## Configuration Reference

### Provider Configuration

| Argument       | Description                        | Default                                       |
|----------------|------------------------------------|-----------------------------------------------|
| `endpoint`     | Assisted Service API endpoint      | `https://api.openshift.com/api/assisted-install` |
| `offline_token`| Red Hat offline token for authentication | Required (or via `OFFLINE_TOKEN` env var) |
| `timeout`      | Default request timeout            | `30s`                                         |

## Examples

The `examples/` directory contains complete working examples:

- **`examples/sno/`** - Single Node OpenShift (SNO) cluster with modular approach
- **`examples/3no/`** - Compact 3-node cluster configuration
- **`examples/post-installation/`** - Post-installation data source usage examples

Each example includes:
- Complete Terraform configuration
- Variable definitions and defaults  
- Comprehensive documentation
- Usage instructions and prerequisites

## Development

### Building

```bash
go build -v ./...
```

### Testing

```bash
# Unit tests
go test -v ./...

# Acceptance tests (requires API credentials)
TF_ACC=1 go test -v ./...
```

### Linting

```bash
golangci-lint run
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run tests and linting
6. Submit a pull request

## License

This project is licensed under the Mozilla Public License v2.0 - see the [LICENSE](LICENSE) file for details.

## Support

For issues and questions:
- Create an issue in this repository
- Check the [OpenShift Assisted Service documentation](https://github.com/openshift/assisted-service)
- Review the [Terraform Provider Development documentation](https://developer.hashicorp.com/terraform/plugin)
