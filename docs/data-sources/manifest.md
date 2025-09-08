---
page_title: "Data Source: oai_manifest"
subcategory: "Custom Configuration"
---

# oai_manifest Data Source

Retrieves manifest files associated with a cluster from the Assisted Service API.

## Example Usage

### Read Specific Manifest

```hcl
data "oai_manifest" "custom_config" {
  cluster_id = oai_cluster.example.id
  file_name  = "99-custom-config.yaml"
}

output "manifest_content" {
  value = data.oai_manifest.custom_config.content
}
```

### Read Manifest from OpenShift Folder

```hcl
data "oai_manifest" "machineconfig" {
  cluster_id = oai_cluster.example.id
  file_name  = "99-worker-ssh.yaml"
  folder     = "openshift"
}

output "machineconfig_content" {
  value = data.oai_manifest.machineconfig.content
}
```

## Argument Reference

* `cluster_id` - (Required) The ID of the cluster containing the manifest.
* `file_name` - (Required) The name of the manifest file.
* `folder` - (Optional) The folder containing the manifest. Valid values: `manifests` (default), `openshift`.

## Attribute Reference

* `id` - The manifest identifier (cluster_id/folder/file_name).
* `cluster_id` - The cluster ID.
* `file_name` - The manifest file name.
* `folder` - The folder containing the manifest.
* `content` - The manifest content (base64 decoded).
* `created_at` - Creation timestamp.
* `updated_at` - Last update timestamp.