# Validation Troubleshooting Example

This example demonstrates comprehensive validation troubleshooting using the `openshift_assisted_installer_cluster_validations` and `openshift_assisted_installer_host_validations` data sources. It provides detailed analysis and troubleshooting guidance for OpenShift cluster validation failures.

## Overview

The OpenShift Assisted Installer performs extensive validation checks on both cluster configuration and individual host readiness. These validations are categorized into:

- **Blocking vs Non-blocking**: Blocking validations prevent installation
- **Categories**: Network, hardware, operators, cluster, platform, storage
- **Scope**: Cluster-level vs host-level validations

## Usage

### Basic Validation Check

```bash
# Set required variables
export TF_VAR_cluster_id="your-cluster-uuid"

# Run validation analysis
terraform plan
terraform apply

# View validation status
terraform output validation_health_check
```

### Network Troubleshooting

```bash
# Get network-specific validation issues
terraform output network_troubleshooting
```

### Hardware Analysis

```bash
# Analyze hardware validation failures by host
terraform output hardware_troubleshooting
```

### Operator Requirements

```bash
# Check operator-specific requirements (LSO, ODF, CNV)
terraform output operator_requirements_analysis
```

### Specific Host Debugging

```bash
# Debug a specific problematic host
export TF_VAR_debug_host_id="problematic-host-uuid"
export TF_VAR_infra_env_id="your-infra-env-uuid"

terraform apply
terraform output specific_host_debug
```

## Validation Categories

### Cluster-Level Validations

| Category | Examples | Description |
|----------|----------|-------------|
| **Network** | API VIPs, Ingress VIPs, CIDR configuration | Cluster networking setup |
| **Cluster** | Pull secret, DNS domain, master count | Basic cluster configuration |
| **Operators** | Operator requirements, compatibility | OLM operator prerequisites |

### Host-Level Validations

| Category | Examples | Description |
|----------|----------|-------------|
| **Hardware** | CPU cores, memory, disk space | Host resource requirements |
| **Network** | Connectivity, DNS resolution, routes | Host network connectivity |
| **Platform** | VMware settings, BIOS configuration | Platform-specific requirements |
| **Storage** | Disk performance, formatting, multipath | Storage and disk requirements |
| **Operators** | Local storage, virtualization support | Host operator prerequisites |

## Common Validation Failures and Fixes

### Cluster-Level Issues

#### Network Configuration
- **API VIPs not defined/valid**: Configure proper API virtual IP addresses
- **Ingress VIPs not defined/valid**: Configure proper Ingress virtual IP addresses  
- **Machine CIDR issues**: Ensure machine CIDR includes all host IP addresses
- **Network overlaps**: Fix overlapping network CIDRs

#### Basic Configuration
- **Pull secret not set**: Configure Red Hat pull secret in cluster definition
- **DNS domain issues**: Set proper base DNS domain
- **Insufficient masters**: Ensure 3 or 5 master nodes for HA clusters

### Host-Level Issues

#### Hardware Problems
- **Insufficient CPU cores**: Add more CPU cores or change host roles
- **Insufficient memory**: Add more memory or adjust cluster requirements
- **No valid disks**: Ensure hosts have adequate disk space (100GB+ recommended)
- **Missing inventory**: Restart hosts or check discovery agent

#### Network Problems  
- **No default route**: Configure default gateway on hosts
- **DNS resolution fails**: Fix DNS configuration for cluster domains
- **Host connectivity**: Ensure hosts can communicate with each other
- **Wrong machine CIDR**: Move hosts to correct network or adjust CIDR

#### Platform-Specific Issues
- **VMware disk UUID**: Enable `disk.EnableUUID=true` in VM settings
- **CPU virtualization**: Enable VT-x/AMD-V in BIOS/VM settings
- **BIOS settings**: Configure proper boot order and hardware features

#### Storage Issues
- **Slow disk performance**: Use faster storage (SSD recommended)
- **Disk already formatted**: Clean/wipe disks before installation
- **Multipath issues**: Configure proper multipath settings

## Output Structure

The example provides several detailed outputs:

### `validation_health_check`
High-level summary of validation status with recommendations.

### `network_troubleshooting`
Detailed network validation analysis with cluster and host-level issues.

### `hardware_troubleshooting`
Hardware validation issues grouped by host with specific failure details.

### `operator_requirements_analysis`
Operator-specific validation status (LSO, ODF, CNV) with requirements.

### `platform_specific_analysis`
Platform-specific validation issues with troubleshooting tips.

### `storage_analysis`
Storage and disk validation issues with resolution guidance.

### `specific_host_debug`
Detailed validation information for a specific problematic host.

## Integration with Installation Workflow

This validation analysis can be integrated into your cluster deployment workflow:

1. **Pre-flight Checks**: Run validation analysis before attempting installation
2. **Issue Resolution**: Use detailed outputs to fix specific problems
3. **Conditional Installation**: Use validation status to control installation flow
4. **Monitoring**: Continuously monitor validation status during cluster preparation

## Example Integration

```hcl
# Use validation results to control installation
locals {
  validation_ready = (
    length(data.openshift_assisted_installer_cluster_validations.basic_readiness.validations) == 0 &&
    length([for v in data.openshift_assisted_installer_host_validations.hardware_check.validations : v if v.status == "failure"]) == 0
  )
}

resource "openshift_assisted_installer_cluster_installation" "conditional" {
  count      = local.validation_ready ? 1 : 0
  cluster_id = var.cluster_id
  
  depends_on = [
    data.openshift_assisted_installer_cluster_validations.basic_readiness,
    data.openshift_assisted_installer_host_validations.hardware_check
  ]
}
```

## Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `cluster_id` | UUID of the cluster to analyze | Yes |
| `infra_env_id` | UUID of infrastructure environment (for specific host debugging) | No |
| `debug_host_id` | UUID of specific host to debug | No |

## Requirements

- OpenShift Assisted Installer cluster in discovery phase
- Terraform >= 1.0
- `oai` provider with validation data sources support