---
page_title: "Resource: oai_manifest"
subcategory: "Custom Configuration"
---

# oai_manifest Resource

Manages custom Kubernetes manifests for OpenShift clusters. Manifests allow you to customise cluster configuration and deploy additional resources during cluster installation.

## Example Usage

### Basic Manifest

```hcl
resource "oai_manifest" "custom_config" {
  cluster_id = oai_cluster.example.id
  file_name  = "custom-config.yaml"
  folder     = "manifests"
  
  content = yamlencode({
    apiVersion = "v1"
    kind       = "ConfigMap"
    metadata = {
      name      = "custom-config"
      namespace = "openshift-config"
    }
    data = {
      custom_setting = "enabled"
    }
  })
}
```

### Machine Configuration

```hcl
resource "oai_manifest" "machine_config" {
  cluster_id = oai_cluster.example.id
  file_name  = "custom-machine-config.yaml"
  folder     = "openshift"  # System-level configuration
  
  content = yamlencode({
    apiVersion = "machineconfiguration.openshift.io/v1"
    kind       = "MachineConfig"
    metadata = {
      labels = {
        "machineconfiguration.openshift.io/role" = "worker"
      }
      name = "99-custom-worker-config"
    }
    spec = {
      config = {
        ignition = {
          version = "3.2.0"
        }
        systemd = {
          units = [{
            name     = "custom-service.service"
            enabled  = true
            contents = "[Unit]\nDescription=Custom Service\n[Service]\nExecStart=/usr/bin/echo 'Custom service started'\n[Install]\nWantedBy=multi-user.target"
          }]
        }
      }
    }
  })
}
```

### Multi-Document YAML

```hcl
resource "oai_manifest" "multiple_resources" {
  cluster_id = oai_cluster.example.id
  file_name  = "monitoring-config.yaml"
  folder     = "manifests"
  
  content = <<-EOT
    ---
    apiVersion: v1
    kind: Namespace
    metadata:
      name: custom-monitoring
      labels:
        name: custom-monitoring
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: prometheus-custom
      namespace: custom-monitoring
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: prometheus-custom
      template:
        metadata:
          labels:
            app: prometheus-custom
        spec:
          containers:
          - name: prometheus
            image: prom/prometheus:latest
            ports:
            - containerPort: 9090
  EOT
}
```

## Argument Reference

### Required Arguments

- `cluster_id` (String) - ID of the cluster to associate this manifest with.
- `file_name` (String) - Name of the manifest file. Must have `.yaml`, `.yml`, or `.json` extension.
- `content` (String) - Content of the manifest in YAML or JSON format.

### Optional Arguments

- `folder` (String) - Folder where the manifest will be stored. Valid values: `manifests` (user manifests), `openshift` (cluster-level manifests). Default: `manifests`.

## Argument Details

### File Name Requirements

The `file_name` must end with one of the supported extensions:
- `.yaml` - YAML format (recommended)
- `.yml` - YAML format (alternative)
- `.json` - JSON format

### Folder Types

**manifests** (Default):
- User-provided manifests for applications and custom resources
- Applied during cluster installation after core components are running
- Suitable for applications, custom resources, and configuration

**openshift**:
- System-level manifests for cluster configuration
- Applied early in the installation process
- Suitable for machine configurations, operator configurations, and core system settings

### Content Format

The `content` attribute accepts either YAML or JSON format. The content is automatically validated and base64-encoded for transmission to the API.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

### Computed Attributes

- `id` (String) - Unique identifier of the manifest (format: `cluster_id/folder/file_name`).
- `manifest_source` (String) - Source information for the manifest.

## Import

Manifests can be imported using the format `cluster_id/folder/file_name`:

```shell
terraform import oai_manifest.example 550e8400-e29b-41d4-a716-446655440000/manifests/custom-config.yaml
```

## Content Management

### Using yamlencode()

For complex YAML structures, use Terraform's `yamlencode()` function:

```hcl
resource "oai_manifest" "complex_config" {
  content = yamlencode({
    apiVersion = "v1"
    kind       = "Secret"
    metadata = {
      name      = "database-config"
      namespace = "default"
    }
    type = "Opaque"
    data = {
      username = base64encode("admin")
      password = base64encode(var.database_password)
    }
  })
}
```

### Using Heredoc Syntax

For multi-document YAML or when preserving exact formatting:

```hcl
resource "oai_manifest" "helm_chart" {
  content = <<-EOT
    ---
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: my-service-account
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: my-cluster-role
    rules:
    - apiGroups: [""]
      resources: ["pods"]
      verbs: ["get", "list", "watch"]
  EOT
}
```

### Loading from Files

To load manifest content from external files:

```hcl
resource "oai_manifest" "from_file" {
  content = file("${path.module}/manifests/application.yaml")
}
```

## Common Use Cases

### Cluster Configuration

Apply cluster-wide configuration changes:

```hcl
resource "oai_manifest" "cluster_config" {
  folder = "openshift"
  content = yamlencode({
    apiVersion = "config.openshift.io/v1"
    kind       = "Image"
    metadata = {
      name = "cluster"
    }
    spec = {
      registrySources = {
        allowedRegistries = [
          "registry.redhat.io",
          "quay.io",
          "docker.io"
        ]
      }
    }
  })
}
```

### Operator Installation

Install additional operators:

```hcl
resource "oai_manifest" "operator_subscription" {
  content = yamlencode({
    apiVersion = "operators.coreos.com/v1alpha1"
    kind       = "Subscription"
    metadata = {
      name      = "elasticsearch-operator"
      namespace = "openshift-operators-redhat"
    }
    spec = {
      channel = "stable"
      installPlanApproval = "Automatic"
      name    = "elasticsearch-operator"
      source  = "redhat-operators"
      sourceNamespace = "openshift-marketplace"
    }
  })
}
```

### Storage Configuration

Configure storage classes and persistent volumes:

```hcl
resource "oai_manifest" "storage_class" {
  content = yamlencode({
    apiVersion = "storage.k8s.io/v1"
    kind       = "StorageClass"
    metadata = {
      name = "fast-ssd"
      annotations = {
        "storageclass.kubernetes.io/is-default-class" = "false"
      }
    }
    provisioner = "kubernetes.io/no-provisioner"
    volumeBindingMode = "WaitForFirstConsumer"
  })
}
```

## Validation

The provider automatically validates manifest content:

- **Syntax Validation** - YAML/JSON syntax is checked
- **Kubernetes API Validation** - Basic API object structure is verified
- **File Extension Validation** - File name must have supported extension

Invalid manifests will cause Terraform operations to fail with descriptive error messages.

## Timing and Dependencies

Manifests are applied during cluster installation after the core Kubernetes API becomes available. They are processed in alphabetical order by file name within each folder.

To ensure proper ordering when dependencies exist between manifests, use prefixed file names:

```hcl
resource "oai_manifest" "namespace" {
  file_name = "01-namespace.yaml"
  # Creates namespace first
}

resource "oai_manifest" "application" {
  file_name = "02-application.yaml"
  # References namespace created above
}
```

## Examples

See the [examples directory](https://github.com/benemon/terraform-provider-openshift-assisted-installer/tree/main/examples) for complete configuration examples including various manifest types and use cases.