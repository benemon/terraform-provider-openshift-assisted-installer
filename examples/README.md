# OpenShift Assisted Installer Terraform Provider Examples

This directory contains practical examples for using the OpenShift Assisted Installer Terraform Provider.

## Prerequisites

Before using any examples:

1. **Set up authentication:**
   ```bash
   export OFFLINE_TOKEN="your_red_hat_offline_token"
   ```
   Get your offline token from: https://console.redhat.com/openshift/token

2. **Prepare pull secret:**
   ```bash
   # Download your pull secret from https://console.redhat.com/openshift/install/pull-secret
   # Save it as pull-secret.json in the same directory as your .tf files
   ```

3. **SSH key for host access:**
   ```bash
   # Ensure you have an SSH public key
   ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa
   ```

## Examples Overview

### 📊 Data Sources (`data_sources/comprehensive_data_sources.tf`)
**Purpose:** Explore available OpenShift versions, operators, bundles, and support levels

**Use cases:**
- Research what OpenShift versions are available
- Discover supported operators and bundles
- Check feature support levels for different platforms
- Validate API connectivity and authentication

**Run with:**
```bash
terraform init
terraform plan -out=data.plan
terraform apply data.plan
```

### 🔧 Single Node OpenShift (`sno/single_node_cluster.tf`)
**Purpose:** Minimal cluster for edge computing, development, or resource-constrained environments

**Configuration:**
- 1 control plane node (also runs workloads)
- Minimal operator set
- Basic networking
- Suitable for development/testing

**Hardware Requirements:**
- 1 machine: 4+ vCPUs, 16+ GB RAM, 120+ GB storage

### 📦 Compact Cluster (`3no/compact_cluster.tf`)
**Purpose:** Small production cluster with control plane nodes running workloads

**Configuration:**
- 3 control plane nodes (schedulable)
- No dedicated worker nodes
- Storage and data foundation operators
- Load balancing with VIPs

**Hardware Requirements:**
- 3 machines: 4+ vCPUs, 16+ GB RAM, 120+ GB storage each

### 🏢 Full Production Cluster (`standard/full_cluster.tf`)
**Purpose:** Enterprise-grade cluster with dedicated control plane and worker nodes

**Configuration:**
- 3 dedicated control plane nodes
- 2+ dedicated worker nodes  
- Comprehensive operator suite (AI/ML, virtualization, service mesh)
- Advanced monitoring and authentication
- Production-grade storage and networking

**Hardware Requirements:**
- 3 control nodes: 4+ vCPUs, 16+ GB RAM, 120+ GB storage each
- 2+ worker nodes: 8+ vCPUs, 32+ GB RAM, 200+ GB storage each

### ✅ Validation & Troubleshooting (`validation-workflow/`, `validation-check/`, `validation-troubleshooting/`)
**Purpose:** Pre-installation validation, conditional deployments, and comprehensive troubleshooting

**Key Features:**
- **Pre-flight checks** before cluster installation
- **Conditional resource creation** based on validation status  
- **Detailed troubleshooting** with categorized failure analysis
- **Operator requirements** validation (LSO, ODF, CNV, LVM)
- **Network, hardware, and platform** validation analysis

**Three validation examples:**
1. **`validation-workflow/`** - Complete validation-driven cluster deployment
2. **`validation-check/`** - Simple validation checking patterns
3. **`validation-troubleshooting/`** - Comprehensive validation analysis and troubleshooting

**Use cases:**
- Verify cluster readiness before installation
- Troubleshoot configuration and hardware issues
- Implement validation-based conditional deployments
- Monitor validation status during cluster preparation

## Common Usage Pattern

1. **Initialize Terraform:**
   ```bash
   terraform init
   ```

2. **Plan deployment:**
   ```bash
   terraform plan -out=cluster.plan
   ```

3. **Apply configuration:**
   ```bash
   terraform apply cluster.plan
   ```

4. **Download discovery ISO:**
   - Check outputs for `download_url`
   - Download ISO to boot your machines

5. **Boot machines:**
   - Boot required number of machines from ISO
   - Wait for discovery and validation

6. **Monitor installation:**
   - Installation starts automatically when requirements are met
   - Access cluster console when complete

7. **Clean up (when done):**
   ```bash
   terraform destroy
   ```

## Network Configuration

All examples assume:
- **Machine Network:** `192.168.1.0/24`
- **API VIP:** `192.168.1.100`
- **Ingress VIP:** `192.168.1.101`
- **Base Domain:** `example.com`

**Customize these values** in the examples for your environment.

## DNS Requirements

For multi-node clusters, configure DNS:
```
# A records
api.cluster-name.example.com        → 192.168.1.100
*.apps.cluster-name.example.com     → 192.168.1.101

# Or use /etc/hosts for testing
192.168.1.100  api.cluster-name.example.com
192.168.1.101  console-openshift-console.apps.cluster-name.example.com
192.168.1.101  oauth-openshift.apps.cluster-name.example.com
```

## Troubleshooting

### Authentication Issues
- Verify `OFFLINE_TOKEN` is set correctly
- Check token expiration (tokens expire after ~30 days)
- Ensure pull-secret.json is valid JSON

### Host Discovery Issues
- Verify machines boot from ISO successfully
- Check network connectivity from booted machines
- Ensure sufficient hardware resources
- Review validation errors in web console

### Installation Failures
- Check cluster validation requirements
- Verify VIP addresses are available and not in use
- Ensure DNS configuration for multi-node clusters
- Monitor installation logs

### Validation Troubleshooting
Use the validation examples to diagnose issues:

```bash
# Basic validation check
cd validation-check/
terraform apply -var="cluster_id=your-cluster-uuid"
terraform output validation_summary

# Comprehensive troubleshooting
cd validation-troubleshooting/
terraform apply -var="cluster_id=your-cluster-uuid"
terraform output network_troubleshooting
terraform output hardware_troubleshooting
```

**Common validation failures:**
- **Network issues:** VIP conflicts, DNS resolution, machine CIDR problems
- **Hardware problems:** Insufficient CPU/memory, missing inventory, slow disks
- **Operator requirements:** Storage devices, CPU features, platform compatibility
- **Platform issues:** VMware settings, BIOS configuration, boot order

## Getting Help

- **Provider Issues:** https://github.com/benemon/terraform-provider-openshift-assisted-installer/issues
- **OpenShift Documentation:** https://docs.openshift.com/container-platform/
- **Red Hat Support:** https://access.redhat.com/support/

## Security Note

The examples include basic authentication with default credentials. **Always change default passwords and configure proper identity providers for production use.**