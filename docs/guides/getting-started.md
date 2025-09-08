---
page_title: "Getting Started with OpenShift Assisted Installer"
subcategory: "Getting Started"
---

# Getting Started with OpenShift Assisted Installer

This guide will walk you through deploying your first OpenShift cluster using the Terraform OpenShift Assisted Installer provider.

## Prerequisites

Before you begin, ensure you have:

- A Red Hat account with access to OpenShift
- An offline token from [console.redhat.com](https://console.redhat.com)
- Target hosts that meet OpenShift system requirements
- Terraform >= 1.0 installed

## Step 1: Obtain Authentication Token

1. Navigate to [console.redhat.com](https://console.redhat.com)
2. Click your profile menu and select "API Tokens"
3. Generate an offline token for the Assisted Service API
4. Copy the token for use in your configuration

## Step 2: Basic Provider Configuration

Create a `main.tf` file with the provider configuration:

```hcl
terraform {
  required_providers {
    oai = {
      source = "benemon/openshift-assisted-installer"
    }
  }
}

provider "oai" {
  offline_token = var.offline_token
}

variable "offline_token" {
  description = "Red Hat offline token"
  type        = string
  sensitive   = true
}

variable "pull_secret" {
  description = "Red Hat pull secret (JSON format)"
  type        = string
  sensitive   = true
}
```

## Step 3: Single Node Cluster Example

For your first deployment, we recommend starting with a single node cluster:

```hcl
# Get available OpenShift versions
data "oai_openshift_versions" "available" {
  only_latest = true
}

locals {
  # Select latest production version
  openshift_version = [
    for v in data.oai_openshift_versions.available.versions :
    v.version if v.support_level == "production"
  ][0]
}

# Create cluster definition
resource "oai_cluster" "single_node" {
  name                 = "my-first-cluster"
  openshift_version    = local.openshift_version
  pull_secret         = var.pull_secret
  cpu_architecture    = "x86_64"
  control_plane_count = 1
  base_dns_domain     = "example.com"
}

# Create infrastructure environment for host discovery
resource "oai_infra_env" "discovery" {
  name             = "discovery-iso"
  pull_secret     = var.pull_secret
  cpu_architecture = "x86_64"
  cluster_id      = oai_cluster.single_node.id
  ssh_authorized_key = file("~/.ssh/id_rsa.pub")
}

# Trigger cluster installation once hosts are ready
resource "oai_cluster_installation" "single_node" {
  cluster_id          = oai_cluster.single_node.id
  wait_for_hosts      = true
  expected_host_count = 1
  
  timeouts {
    create = "120m"
  }
}

# Output important information
output "cluster_id" {
  value = oai_cluster.single_node.id
}

output "iso_download_url" {
  value = oai_infra_env.discovery.download_url
}
```

## Step 4: Deploy the Infrastructure

1. Initialize Terraform:
   ```bash
   terraform init
   ```

2. Plan the deployment:
   ```bash
   terraform plan -var="offline_token=YOUR_TOKEN" -var="pull_secret=$(cat pull-secret.txt)"
   ```

3. Apply the configuration:
   ```bash
   terraform apply -var="offline_token=YOUR_TOKEN" -var="pull_secret=$(cat pull-secret.txt)"
   ```

## Step 5: Boot Your Host

1. Download the discovery ISO:
   ```bash
   curl -L -o discovery.iso "$(terraform output -raw iso_download_url)"
   ```

2. Create a bootable USB drive:
   ```bash
   # Linux
   sudo dd if=discovery.iso of=/dev/sdX bs=4M status=progress
   
   # macOS
   sudo dd if=discovery.iso of=/dev/diskX bs=4m
   ```

3. Boot your target host from the USB drive

## Step 6: Monitor Installation

The host will automatically:
1. Boot from the discovery ISO
2. Inventory hardware and register with the cluster
3. Wait for installation to be triggered automatically once host is ready
4. Install OpenShift (30-90 minutes)

Monitor progress:
```bash
# Check cluster status
terraform show | grep -A 5 "status"

# Or use the Red Hat console
echo "View cluster at: https://console.redhat.com/openshift/assisted-installer/clusters/$(terraform output -raw cluster_id)"
```

## Step 7: Access Your Cluster

Once installation completes (`status = "installed"`), you can access your cluster:

1. Download the kubeconfig from the Red Hat console
2. Set your KUBECONFIG environment variable
3. Access your cluster:
   ```bash
   oc get nodes
   oc get clusterversion
   ```

## Next Steps

Now that you have a basic cluster running, consider:

- [Deploying a multi-node cluster](multi-node-cluster.md)
- [Configuring custom networking](networking.md)
- [Adding operators and customisation](operators.md)
- [Managing hosts and scaling](host-management.md)

## Troubleshooting

### Host Not Discovered
- Verify network connectivity from the discovery environment
- Check hardware meets minimum requirements
- Review console logs during host boot

### Installation Stuck
- Check the status_info field for specific errors
- Verify DNS resolution for your base domain
- Ensure sufficient resources (CPU, memory, disk)

### Timeout Issues
- Increase timeout values in your configuration
- Check network connectivity to Red Hat services
- Verify pull secret is valid and not expired

For more detailed troubleshooting, see the individual resource documentation.