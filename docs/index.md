---
page_title: "Provider: OpenShift Assisted Installer"
---

# OpenShift Assisted Installer Provider

The OpenShift Assisted Installer provider enables Infrastructure as Code management of OpenShift cluster deployments using the Red Hat OpenShift Assisted Service API.

## Features

- Complete cluster lifecycle management from definition to installation
- Infrastructure environment configuration for host discovery
- Host management and role assignment
- Custom manifest deployment
- Comprehensive data sources for version and operator information

## Authentication

The provider supports authentication via offline tokens obtained from the Red Hat Hybrid Cloud Console.

### Obtaining an Offline Token

1. Navigate to [console.redhat.com](https://console.redhat.com)
2. Select "API Tokens" from the user menu
3. Generate an offline token for the Assisted Service API
4. Configure the provider with this token

## Provider Configuration

```hcl
terraform {
  required_providers {
    oai = {
      source = "benemon/openshift-assisted-installer"
    }
  }
}

provider "oai" {
  endpoint      = "https://api.openshift.com/api/assisted-install"
  offline_token = var.offline_token
  timeout       = "30s"
}
```

### Configuration Reference

- `endpoint` (Optional) - The API endpoint URL. Defaults to the Red Hat production endpoint.
- `offline_token` (Optional) - Offline token for authentication. Can also be provided via the `OFFLINE_TOKEN` environment variable.
- `timeout` (Optional) - HTTP request timeout duration. Defaults to 30 seconds.

## Environment Variables

- `OFFLINE_TOKEN` - Alternative method for providing the offline token

## Example Usage

### Basic Single Node Cluster

```hcl
resource "oai_cluster" "example" {
  name                 = "example-cluster"
  openshift_version    = "4.14"
  pull_secret         = var.pull_secret
  cpu_architecture    = "x86_64"
  control_plane_count = 1
  trigger_installation = true
}

resource "oai_infra_env" "example" {
  name                = "example-infra"
  pull_secret        = var.pull_secret
  cpu_architecture   = "x86_64"
  cluster_id         = oai_cluster.example.id
  ssh_authorized_key = var.ssh_public_key
}
```

### Three Node Cluster with Custom Networking

```hcl
resource "oai_cluster" "example" {
  name                     = "example-cluster"
  openshift_version        = "4.14"
  pull_secret             = var.pull_secret
  cpu_architecture        = "x86_64"
  control_plane_count     = 3
  base_dns_domain         = "example.com"
  cluster_network_cidr    = "10.128.0.0/14"
  service_network_cidr    = "172.30.0.0/16"
  api_vips               = ["192.168.1.100"]
  ingress_vips           = ["192.168.1.101"]
  trigger_installation   = true
}
```

## Resources

- [`oai_cluster`](resources/cluster.md) - Manages OpenShift cluster definitions and installations
- [`oai_infra_env`](resources/infra_env.md) - Manages infrastructure environments and discovery ISOs
- [`oai_host`](resources/host.md) - Manages discovered hosts and role assignments
- [`oai_manifest`](resources/manifest.md) - Manages custom cluster manifests

## Data Sources

- [`oai_openshift_versions`](data-sources/openshift_versions.md) - Available OpenShift versions
- [`oai_supported_operators`](data-sources/supported_operators.md) - Supported OLM operators
- [`oai_operator_bundles`](data-sources/operator_bundles.md) - Available operator bundles
- [`oai_support_levels`](data-sources/support_levels.md) - Feature support matrix

## Installation Workflow

The typical workflow for deploying an OpenShift cluster involves:

1. **Cluster Definition** - Create a cluster resource with required configuration
2. **Infrastructure Environment** - Set up discovery ISO generation
3. **Host Discovery** - Boot target hosts from the generated ISO
4. **Installation** - Terraform automatically triggers installation once hosts are ready
5. **Post-Installation** - Apply custom manifests and configuration

Refer to the [examples directory](https://github.com/benemon/terraform-provider-openshift-assisted-installer/tree/main/examples) for complete configuration examples.