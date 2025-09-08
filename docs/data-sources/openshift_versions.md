---
page_title: "Data Source: oai_openshift_versions"
subcategory: "General Information"
---

# oai_openshift_versions Data Source

Retrieves available OpenShift versions from the Assisted Service API. Use this data source to discover supported OpenShift versions for cluster deployment.

## Example Usage

### List All Versions

```hcl
data "oai_openshift_versions" "available" {}

output "all_versions" {
  value = data.oai_openshift_versions.available.versions
}
```

### Filter by Version Pattern

```hcl
data "oai_openshift_versions" "stable_4_14" {
  version_filter = "4.16"
}

output "openshift_4_14_versions" {
  value = data.oai_openshift_versions.stable_4_14.versions
}
```

### Get Latest Versions Only

```hcl
data "oai_openshift_versions" "latest" {
  only_latest = true
}

output "latest_versions" {
  value = data.oai_openshift_versions.latest.versions
}
```

### Filter by Architecture

```hcl
data "oai_openshift_versions" "arm64" {
  cpu_architecture = "arm64"
}

output "arm64_versions" {
  value = data.oai_openshift_versions.arm64.versions
}
```

## Argument Reference

### Optional Arguments

- `version_filter` (String) - Filter versions by pattern. Supports partial matching (e.g., "4.16" matches "4.16.1", "4.16.2", etc.).
- `only_latest` (Boolean) - Return only the latest version for each major.minor release. Default: false.
- `cpu_architecture` (String) - Filter versions by CPU architecture. Valid values: `x86_64`, `arm64`, `ppc64le`, `s390x`.

## Attribute Reference

The following attributes are exported:

- `versions` (List of Object) - List of available OpenShift versions. Each version object contains:
  - `version` (String) - Full version string (e.g., "4.16.1")
  - `display_name` (String) - Human-readable version name
  - `support_level` (String) - Support level for this version. Values: `production`, `maintenance`, `beta`, `dev-preview`
  - `default` (Boolean) - Whether this is the default version for new clusters
  - `cpu_architectures` (List of String) - List of supported CPU architectures
  - `release_image` (String) - Container image reference for this OpenShift release

## Version Support Levels

### production
- Fully supported for production workloads
- Recommended for production clusters
- Receives full support and security updates

### maintenance
- Previously production versions in maintenance mode
- Suitable for existing clusters
- Limited to critical security updates

### beta
- Release candidate versions
- Suitable for testing and evaluation
- May have known issues or limitations

### dev-preview
- Early access developer previews
- Not suitable for production use
- Used for feature preview and testing

## Practical Examples

### Select Latest Stable Version

```hcl
data "oai_openshift_versions" "production" {
  only_latest = true
}

locals {
  latest_production = [
    for v in data.oai_openshift_versions.production.versions :
    v if v.support_level == "production"
  ][0]
}

resource "oai_cluster" "example" {
  openshift_version = local.latest_production.version
  # ... other configuration
}
```

### Version Compatibility Check

```hcl
data "oai_openshift_versions" "available" {}

locals {
  compatible_versions = [
    for v in data.oai_openshift_versions.available.versions :
    v if contains(v.cpu_architectures, var.target_architecture)
  ]
}

output "compatible_versions" {
  description = "Versions compatible with target architecture"
  value       = local.compatible_versions
}
```

### Multi-Architecture Deployment

```hcl
data "oai_openshift_versions" "multi_arch" {
  cpu_architecture = "multi"
}

resource "oai_cluster" "heterogeneous" {
  openshift_version = data.oai_openshift_versions.multi_arch.versions[0].version
  cpu_architecture  = "multi"
  # ... other configuration
}
```

### Development vs Production Environments

```hcl
# Production environment - use stable versions only
data "oai_openshift_versions" "prod" {
  only_latest = true
}

locals {
  prod_version = [
    for v in data.oai_openshift_versions.prod.versions :
    v if v.support_level == "production"
  ][0].version
}

# Development environment - allow beta versions
data "oai_openshift_versions" "dev" {
  version_filter = "4.17"  # Latest development branch
}

locals {
  dev_version = data.oai_openshift_versions.dev.versions[0].version
}

resource "oai_cluster" "production" {
  openshift_version = local.prod_version
  # ... other configuration
}

resource "oai_cluster" "development" {
  openshift_version = local.dev_version
  # ... other configuration
}
```

## Version Selection Best Practices

### Production Clusters
- Always use `support_level = "production"` versions
- Consider using `only_latest = true` to get the most recent stable release
- Test new versions in non-production environments first

### Development/Testing
- Beta versions are acceptable for feature testing
- Use specific version filters to test particular features
- Consider multi-architecture versions for heterogeneous environments

### Long-Term Support
- Monitor version support levels as they change over time
- Plan upgrades before versions transition to maintenance mode
- Subscribe to OpenShift lifecycle announcements

## Integration with Other Resources

### Version Validation

```hcl
data "oai_openshift_versions" "available" {}

locals {
  is_valid_version = contains([
    for v in data.oai_openshift_versions.available.versions : v.version
  ], var.requested_version)
}

resource "oai_cluster" "validated" {
  count = local.is_valid_version ? 1 : 0
  
  openshift_version = var.requested_version
  # ... other configuration
}
```

### Automatic Version Updates

```hcl
data "oai_openshift_versions" "latest" {
  version_filter = "4.16"  # Stay within major.minor family
  only_latest    = true
}

resource "oai_cluster" "auto_update" {
  # Automatically uses latest 4.16.x version
  openshift_version = data.oai_openshift_versions.latest.versions[0].version
  # ... other configuration
}
```

This approach allows clusters to automatically receive patch version updates while maintaining compatibility within a major.minor release family.