# Terraform Provider for OpenShift Assisted Installer

This Terraform provider enables Infrastructure as Code management of OpenShift clusters using the [OpenShift Assisted Service API](https://api.openshift.com/api/assisted-install).

## Features

- **Cluster Management**: Create, update, and manage OpenShift clusters
- **Infrastructure Environments**: Manage discovery ISOs for host bootstrapping 
- **Custom Manifests**: Apply custom configuration manifests to clusters
- **Data Sources**: Query supported OpenShift versions and operators
- **Asynchronous Operations**: Handle long-running cluster installations with configurable timeouts

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

The provider supports authentication via:

1. **Provider configuration**:
   ```hcl
   provider "oai" {
     token = "your-api-token"
   }
   ```

2. **Environment variable**:
   ```bash
   export OAI_TOKEN="your-api-token"
   ```

## Usage

### Basic Cluster Creation

```hcl
provider "oai" {
  endpoint = "https://api.openshift.com/api/assisted-install"
  token    = var.oai_token
}

resource "oai_cluster" "example" {
  name              = "my-cluster"
  openshift_version = "4.15.20"
  pull_secret       = var.pull_secret
  
  # Network configuration
  base_dns_domain   = "example.com" 
  api_vips         = ["192.168.1.100"]
  ingress_vips     = ["192.168.1.101"]
  
  # SSH access
  ssh_public_key = file("~/.ssh/id_rsa.pub")
  
  timeouts {
    create = "90m"
  }
}
```

### Infrastructure Environment for Host Discovery

```hcl
resource "oai_infra_env" "example" {
  name         = "my-infra-env"
  cluster_id   = oai_cluster.example.id
  pull_secret  = var.pull_secret
  
  ssh_authorized_key = file("~/.ssh/id_rsa.pub")
  
  # Static network configuration (optional)
  static_network_config = jsonencode([
    {
      dns_resolver = {
        config = {
          server = ["192.168.1.1"]
        }
      }
    }
  ])
}
```

### Custom Manifests

```hcl
resource "oai_manifest" "example" {
  cluster_id = oai_cluster.example.id
  folder     = "manifests"
  file_name  = "custom-config.yaml"
  
  content = base64encode(templatefile("${path.module}/manifests/custom-config.yaml", {
    cluster_name = oai_cluster.example.name
  }))
}
```

## Resources

- **`oai_cluster`** - OpenShift cluster management
- **`oai_infra_env`** - Infrastructure environment for host discovery
- **`oai_manifest`** - Custom cluster manifests

## Data Sources

- **`oai_openshift_versions`** - Available OpenShift versions
- **`oai_supported_operators`** - Supported OLM operators

## Configuration Reference

### Provider Configuration

| Argument   | Description                    | Default                                       |
|------------|--------------------------------|-----------------------------------------------|
| `endpoint` | Assisted Service API endpoint  | `https://api.openshift.com/api/assisted-install` |
| `token`    | Authentication token           | Required                                      |
| `timeout`  | Default request timeout        | `30s`                                         |

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
