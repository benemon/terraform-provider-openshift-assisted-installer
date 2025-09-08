---
page_title: "Data Source: openshift_assisted_installer_cluster_files"
subcategory: "Cluster Management"
---

# openshift_assisted_installer_cluster_files Data Source

Lists and retrieves files associated with a cluster, including manifests, discovery ignition, and installation files.

## Example Usage

### List All Cluster Files

```hcl
data "openshift_assisted_installer_cluster_files" "all" {
  cluster_id = openshift_assisted_installer_cluster.example.id
}

output "manifest_files" {
  value = [
    for file in data.openshift_assisted_installer_cluster_files.all.files :
    file.file_name if file.folder == "manifests"
  ]
}
```

### Filter by File Type

```hcl
data "openshift_assisted_installer_cluster_files" "ignition" {
  cluster_id = openshift_assisted_installer_cluster.example.id
  file_type  = "ignition"
}

output "ignition_files" {
  value = data.openshift_assisted_installer_cluster_files.ignition.files
}
```

### Get Specific Folder Files

```hcl
data "openshift_assisted_installer_cluster_files" "openshift" {
  cluster_id = openshift_assisted_installer_cluster.example.id
  folder     = "openshift"
}

output "openshift_manifests" {
  value = {
    for file in data.openshift_assisted_installer_cluster_files.openshift.files :
    file.file_name => file.size
  }
}
```

## Argument Reference

* `cluster_id` - (Required) The cluster ID to retrieve files for.
* `file_type` - (Optional) Filter by file type. Valid values:
  * `manifests` - Custom manifest files
  * `ignition` - Ignition configuration files
  * `logs` - Log files
* `folder` - (Optional) Filter by folder. Valid values: `manifests`, `openshift`.

## Attribute Reference

* `id` - The data source ID (same as cluster_id).
* `cluster_id` - The cluster ID.
* `files` - List of files with the following attributes:
  * `file_name` - Name of the file.
  * `folder` - Folder containing the file.
  * `size` - File size in bytes.
  * `download_url` - URL to download the file.
  * `created_at` - When the file was created.
  * `updated_at` - When the file was last updated.