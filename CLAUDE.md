# Claude Code Guide: Terraform Provider for OpenShift Assisted Service

## Project Overview

This document provides guidance for using Claude Code to develop a Terraform provider for the OpenShift Assisted Service API. The provider will enable Infrastructure as Code management of OpenShift cluster deployments using the Assisted Installer service.

### Repository Structure
```
terraform-provider-openshift-assisted-installer/
├── swagger/
│   └── swagger.yaml        # OpenShift Assisted Service API specification
├── internal/
│   └── provider/          # Provider implementation
├── examples/              # Example Terraform configurations
├── docs/                  # Documentation
├── README.md
├── LICENSE
└── CLAUDE.md              # This file
```

## API Analysis Summary

### API Characteristics
- **Base URL**: `https://api.openshift.com/api/assisted-install`
- **API Version**: v2
- **Authentication**: Multiple methods (userAuth, agentAuth, urlAuth)
- **Format**: RESTful with JSON payloads
- **Specification**: OpenAPI 2.0 (Swagger)

### Core Resources Identified

1. **Clusters** (`/v2/clusters`)
   - Full CRUD operations
   - UUID-based identification
   - Complex state machine (insufficient → ready → installing → installed)
   - Action endpoints for installation triggers

2. **InfraEnvs** (`/v2/infra-envs`)
   - Manages discovery ISO generation
   - Can be cluster-bound or independent
   - Handles host discovery configuration

3. **Hosts** (`/v2/infra-envs/{id}/hosts`)
   - Self-registering via discovery agents
   - Binding/unbinding to clusters
   - State tracking through installation

4. **Manifests** (`/v2/clusters/{id}/manifests`)
   - Custom configuration files
   - Support for manifests and openshift folders
   - Base64 encoded content

## Implementation Tasks for Claude Code

### Phase 1: Project Setup

```bash
# Clone the scaffolding repository to a temporary location
git clone https://github.com/hashicorp/terraform-provider-scaffolding-framework.git /tmp/scaffolding

# Navigate to your existing repository
cd terraform-provider-openshift-assisted-installer

# Copy the scaffolding content (excluding .git, README, and LICENSE)
cp -r /tmp/scaffolding/internal .
cp -r /tmp/scaffolding/examples .
cp -r /tmp/scaffolding/docs .
cp /tmp/scaffolding/go.mod .
cp /tmp/scaffolding/go.sum .
cp /tmp/scaffolding/main.go .
cp /tmp/scaffolding/.gitignore .
cp /tmp/scaffolding/.goreleaser.yml .
cp /tmp/scaffolding/Makefile .
cp /tmp/scaffolding/terraform-registry-manifest.json .

# Clean up
rm -rf /tmp/scaffolding
```

**Claude Code Tasks:**
1. Update go.mod:
   - Change module to: `module github.com/benemon/terraform-provider-openshift-assisted-installer`
2. Update all import paths in .go files to use:
   - `github.com/benemon/terraform-provider-openshift-assisted-installer/internal/provider`
3. Rename the provider from "scaffolding" to "oai" in:
   - main.go (provider name)
   - All resource and data source registrations
   - terraform-registry-manifest.json
4. Update Makefile to build `terraform-provider-oai`

### Phase 2: Provider Configuration

Create the provider configuration with these fields:

```go
type OAIProviderModel struct {
    Endpoint types.String `tfsdk:"endpoint"`
    Token    types.String `tfsdk:"token"`
    Timeout  types.String `tfsdk:"timeout"`
}
```

**Claude Code Tasks:**
1. Implement provider configuration schema in `internal/provider/provider.go`
2. Create HTTP client with authentication headers
3. Add timeout and retry logic
4. Implement provider validation
5. Set default endpoint to `https://api.openshift.com/api/assisted-install`

### Phase 3: Core Resources Implementation

#### 3.1 Cluster Resource (`oai_cluster`)

**Resource Name:** `oai_cluster`

**Schema Requirements:**
- name (required, string)
- openshift_version (required, string)
- pull_secret (required, sensitive string)
- base_dns_domain (optional, string)
- api_vips (optional, list of strings)
- ingress_vips (optional, list of strings)
- platform (optional, object)
- control_plane_count (optional, int)

**State Management:**
- Track installation progress
- Handle state transitions
- Implement wait conditions

**Claude Code Tasks:**
1. Create `internal/provider/cluster_resource.go`
2. Define resource as `oai_cluster`
3. Implement CRUD operations mapping to API endpoints
4. Add state tracking logic for installation phases
5. Handle `/v2/clusters/{id}/actions/install` trigger
6. Implement timeout handling with 90-minute default

#### 3.2 InfraEnv Resource (`oai_infra_env`)

**Resource Name:** `oai_infra_env`

**Schema Requirements:**
- name (required, string)
- pull_secret (required, sensitive string)
- cluster_id (optional, string)
- ssh_authorized_key (optional, string)
- static_network_config (optional, list of objects)

**Claude Code Tasks:**
1. Create `internal/provider/infra_env_resource.go`
2. Define resource as `oai_infra_env`
3. Implement ISO generation logic
4. Handle cluster binding
5. Add download URL as computed attribute

#### 3.3 Manifest Resource (`oai_manifest`)

**Resource Name:** `oai_manifest`

**Schema Requirements:**
- cluster_id (required, string)
- folder (optional, string, default: "manifests")
- file_name (required, string)
- content (required, string)

**Claude Code Tasks:**
1. Create `internal/provider/manifest_resource.go`
2. Define resource as `oai_manifest`
3. Add base64 encoding/decoding
4. Handle content updates
5. Implement import functionality

### Phase 4: Data Sources

#### 4.1 OpenShift Versions Data Source (`oai_openshift_versions`)

**Data Source Name:** `oai_openshift_versions`

```go
// Returns available OpenShift versions
// Filter by version string or only_latest flag
```

**Claude Code Tasks:**
1. Create `internal/provider/openshift_versions_data_source.go`
2. Define data source as `oai_openshift_versions`
3. Implement filtering logic
4. Map to `/v2/openshift-versions` endpoint

#### 4.2 Supported Operators Data Source (`oai_supported_operators`)

**Data Source Name:** `oai_supported_operators`

```go
// Lists all supported operators
// Include operator properties and requirements
```

**Claude Code Tasks:**
1. Create `internal/provider/supported_operators_data_source.go`
2. Define data source as `oai_supported_operators`
3. Map to `/v2/supported-operators` endpoint
4. Include operator properties schema

### Phase 5: Complex Features

#### 5.1 Asynchronous Operation Handling

```go
// Example wait configuration
func waitForClusterReady(ctx context.Context, client *Client, clusterID string) error {
    stateConf := &retry.StateChangeConf{
        Pending: []string{"insufficient", "preparing-for-installation", "installing"},
        Target:  []string{"installed"},
        Refresh: clusterStateRefreshFunc(client, clusterID),
        Timeout: 90 * time.Minute,
        Delay:   30 * time.Second,
    }
    _, err := stateConf.WaitForStateContext(ctx)
    return err
}
```

**Claude Code Tasks:**
1. Implement state polling functions
2. Add configurable timeouts
3. Create retry logic for transient failures
4. Add progress tracking attributes

#### 5.2 Host Discovery Workflow

The host discovery workflow requires special handling:
1. Create InfraEnv (generates ISO)
2. Hosts boot from ISO and self-register
3. Provider waits for expected host count
4. Bind hosts to cluster
5. Trigger installation

**Claude Code Tasks:**
1. Implement host waiting logic
2. Add host binding operations
3. Create validation checks
4. Handle host state transitions

### Phase 6: Testing

#### 6.1 Unit Tests

**Claude Code Tasks:**
1. Create unit tests for schema validation
2. Mock API responses
3. Test state transitions
4. Validate error handling

#### 6.2 Acceptance Tests

```go
func TestAccAssistedCluster_basic(t *testing.T) {
    // Test basic cluster creation
}

func TestAccAssistedCluster_complete(t *testing.T) {
    // Test with all optional fields
}
```

**Claude Code Tasks:**
1. Create acceptance test framework
2. Implement resource tests
3. Add import tests
4. Create upgrade tests

### Phase 7: Documentation

**Claude Code Tasks:**
1. Generate resource documentation
2. Create usage examples
3. Document wait conditions
4. Add troubleshooting guide

## Code Generation Requests for Claude Code

### Request 1: Provider Setup
"Using the swagger.yaml file in the swagger/ directory, create the basic provider structure for terraform-provider-oai using the Terraform Plugin Framework. Set up the provider configuration with endpoint and token authentication. The default endpoint should be https://api.openshift.com/api/assisted-install"

### Request 2: Generate API Client from Swagger
"Parse the swagger/swagger.yaml file and generate Go structs for all the API models defined in the definitions section. Create a structured client package that maps to the API endpoints."

### Request 3: Cluster Resource
"Using the swagger.yaml specification, implement the cluster resource for the Assisted Service API. Look at the /v2/clusters endpoints and the cluster definition to create the full CRUD operations. Include state management for the installation process where the cluster goes through states: insufficient → ready → installing → installed."

### Request 4: Generate Resource Schemas
"Read the swagger/swagger.yaml file and generate Terraform resource schemas for oai_cluster, oai_infra_env, and oai_manifest based on the API definitions. Map the Swagger types to appropriate Terraform schema types."

### Request 5: Async Operations
"Based on the cluster states defined in swagger.yaml, add asynchronous operation handling for long-running tasks like cluster installation. Implement polling with exponential backoff and configurable timeouts."

### Request 6: Data Sources from Swagger
"Using the swagger.yaml file, implement data sources for oai_openshift_versions and oai_supported_operators. Map these to the /v2/openshift-versions and /v2/supported-operators endpoints respectively."

### Request 7: Testing
"Generate comprehensive acceptance tests for the cluster resource based on the API responses defined in swagger.yaml, including creation, updates, and import scenarios."

## Working with the Local Swagger Specification

### Using Claude Code with swagger.yaml

Since you have the swagger.yaml file locally at `swagger/swagger.yaml`, Claude Code can:

1. **Parse and analyze the specification** to understand all endpoints, request/response models, and data types
2. **Generate Go structs** directly from the Swagger definitions
3. **Create API client code** that matches the exact API structure
4. **Build resource schemas** that align with the API requirements
5. **Generate validation logic** based on API constraints

### Example Claude Code Commands

```bash
# Generate all API models from Swagger
claude-code "Parse swagger/swagger.yaml and generate Go structs for all definitions in internal/models/"

# Create client from API paths
claude-code "Using swagger/swagger.yaml, generate an API client in internal/client/ with methods for all /v2/clusters endpoints"

# Build resource schema from definition
claude-code "Read the cluster definition from swagger/swagger.yaml and create a Terraform resource schema in internal/provider/cluster_resource.go"

# Generate validation functions
claude-code "Extract all validation rules from swagger/swagger.yaml for the cluster-create-params and generate validation functions"
```

### Workflow with Local Swagger

1. **Initial Analysis**
   - Have Claude Code analyze the swagger.yaml to list all resources and their operations
   - Identify which endpoints map to Terraform resources vs data sources

2. **Model Generation**
   - Generate all API models from the definitions section
   - Create separate files for each major resource type

3. **Client Creation**
   - Build HTTP client methods for each endpoint
   - Include proper error handling and response parsing

4. **Resource Implementation**
   - Map Swagger definitions to Terraform schemas
   - Implement CRUD operations using the generated client

5. **Testing**
   - Generate test cases based on example responses in the Swagger spec
   - Create mock responses from the Swagger examples

### 1. State Machine Complexity
The cluster installation involves multiple states and can take 30-90 minutes. Implement robust state tracking and user feedback.

### 2. Host Discovery Pattern
Hosts self-register after booting from the discovery ISO. The provider must wait for hosts to appear rather than creating them directly.

### 3. Action Endpoints
Some operations use action endpoints (`/actions/install`, `/actions/reset`). Map these to appropriate Terraform lifecycle points.

### 4. Validation Handling
The API performs extensive validations. Surface these clearly to users and support ignoring specific validations when needed.

### 5. Import Support
Leverage the `/v2/clusters/import` endpoint to support importing existing clusters for brownfield scenarios.

## Development Workflow

1. Start with provider configuration and authentication
2. Implement cluster resource with basic fields
3. Add state management and waiting logic
4. Implement infra_env resource
5. Add manifest support
6. Create data sources
7. Implement comprehensive testing
8. Document all resources and examples

## Useful API Patterns

### Pagination
Some endpoints support pagination with `limit` and `offset` parameters.

### Filtering
List endpoints often support filtering by owner, cluster_id, or other fields.

### Event Monitoring
The `/v2/events` endpoint can be used for debugging and progress tracking.

### Validation Results
Resources include `validations_info` fields with detailed validation results.

## Success Criteria

- [ ] Provider can authenticate with the Assisted Service API
- [ ] Cluster resource supports full lifecycle management
- [ ] Installation progress is tracked and reported
- [ ] Timeouts are configurable and reasonable
- [ ] Import functionality works for existing clusters
- [ ] Comprehensive test coverage exists
- [ ] Documentation includes working examples
- [ ] Error messages are clear and actionable

## References

- [Terraform Plugin Framework Documentation](https://developer.hashicorp.com/terraform/plugin/framework)
- [OpenShift Assisted Service Repository](https://github.com/openshift/assisted-service)
- [Assisted Service API Interactive Documentation](https://api.openshift.com/?urls.primaryName=assisted-service%20service)
- [Assisted Service Swagger Specification](https://github.com/openshift/assisted-service/blob/master/swagger.yaml)
- [Terraform Provider Scaffolding](https://github.com/hashicorp/terraform-provider-scaffolding-framework)
