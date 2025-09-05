---
page_title: "Data Source: oai_cluster_credentials"
subcategory: "Cluster Management"
---

# oai_cluster_credentials Data Source

Retrieves cluster credentials from the Assisted Service API after installation completes.

## Example Usage

### Get Cluster Credentials

```hcl
data "oai_cluster_credentials" "admin" {
  cluster_id = oai_cluster.example.id
}

output "cluster_access" {
  value = {
    console_url = data.oai_cluster_credentials.admin.console_url
    username    = data.oai_cluster_credentials.admin.username
  }
  sensitive = true
}

output "kubeconfig_secret" {
  value     = data.oai_cluster_credentials.admin.password
  sensitive = true
}
```

### Store Credentials in External System

```hcl
data "oai_cluster_credentials" "creds" {
  cluster_id = oai_cluster.production.id
}

resource "vault_generic_secret" "cluster_creds" {
  path = "secret/openshift/production"
  
  data_json = jsonencode({
    console_url = data.oai_cluster_credentials.creds.console_url
    username    = data.oai_cluster_credentials.creds.username
    password    = data.oai_cluster_credentials.creds.password
  })
}
```

## Argument Reference

* `cluster_id` - (Required) The ID of the installed cluster.

## Attribute Reference

* `id` - The data source ID (same as cluster_id).
* `cluster_id` - The cluster ID.
* `username` - The admin username (typically "kubeadmin").
* `password` - The admin password (sensitive).
* `console_url` - The OpenShift web console URL.

**Note:** Credentials are only available after the cluster installation completes successfully.