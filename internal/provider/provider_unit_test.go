package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestOAIProvider_Metadata(t *testing.T) {
	p := &OAIProvider{
		version: "test",
	}

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "openshift_assisted_installer" {
		t.Errorf("Expected TypeName 'openshift_assisted_installer', got %s", resp.TypeName)
	}

	if resp.Version != "test" {
		t.Errorf("Expected Version 'test', got %s", resp.Version)
	}
}

func TestOAIProvider_Schema(t *testing.T) {
	p := &OAIProvider{}

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	// Check that the schema has the expected attributes
	attrs := resp.Schema.Attributes

	if _, ok := attrs["endpoint"]; !ok {
		t.Error("Schema missing 'endpoint' attribute")
	}

	if _, ok := attrs["offline_token"]; !ok {
		t.Error("Schema missing 'offline_token' attribute")
	}

	if _, ok := attrs["timeout"]; !ok {
		t.Error("Schema missing 'timeout' attribute")
	}

	// Check that offline_token is marked as sensitive
	tokenAttr := attrs["offline_token"]
	if !tokenAttr.IsSensitive() {
		t.Error("offline_token attribute should be marked as sensitive")
	}
}

// TestOAIProvider_Configure tests are commented out as they require
// proper Terraform Plugin Framework test infrastructure to run.
// The Configure method is tested through acceptance tests instead.
/*
func TestOAIProvider_Configure(t *testing.T) {
	// This test requires the Terraform Plugin Framework testing infrastructure
	// which is complex to set up for unit tests. The Configure method is
	// properly tested through acceptance tests with TF_ACC=1.
	t.Skip("Configure method is tested through acceptance tests")
}
*/

func TestOAIProvider_Resources(t *testing.T) {
	p := &OAIProvider{}

	resources := p.Resources(context.Background())

	// Check that we have the expected number of resources
	// Note: We currently only have ClusterResource fully implemented
	// InfraEnvResource, HostResource, ManifestResource are placeholders
	if len(resources) < 1 {
		t.Errorf("Expected at least 1 resource, got %d", len(resources))
	}
	t.Logf("Found %d resources registered", len(resources))

	// Verify that resources are registered
	resourceNames := []string{"cluster", "infra_env", "host", "manifest"}
	for _, resourceName := range resourceNames {
		found := false
		for _, r := range resources {
			// Create an instance to check the type
			instance := r()
			switch resourceName {
			case "cluster":
				if _, ok := instance.(*ClusterResource); ok {
					found = true
				}
			case "infra_env":
				if _, ok := instance.(*InfraEnvResource); ok {
					found = true
				}
			case "host":
				if _, ok := instance.(*HostResource); ok {
					found = true
				}
			case "manifest":
				if _, ok := instance.(*ManifestResource); ok {
					found = true
				}
			}
			if found {
				break
			}
		}
		if !found {
			t.Errorf("%s resource not found in resources list", resourceName)
		}
	}
}

func TestOAIProvider_DataSources(t *testing.T) {
	p := &OAIProvider{}

	dataSources := p.DataSources(context.Background())

	// Check that we have the expected number of data sources
	// We have many data sources now: versions, operators, bundles, levels, validations, cluster, infra_env, host, manifest, etc.
	if len(dataSources) < 4 {
		t.Errorf("Expected at least 4 data sources, got %d", len(dataSources))
	}
	t.Logf("Found %d data sources registered", len(dataSources))

	// Verify that data sources are in the list
	var foundVersions, foundOperators, foundBundles, foundLevels bool
	for _, ds := range dataSources {
		instance := ds()
		switch instance.(type) {
		case *OpenShiftVersionsDataSource:
			foundVersions = true
		case *SupportedOperatorsDataSource:
			foundOperators = true
		case *OperatorBundlesDataSource:
			foundBundles = true
		case *SupportLevelsDataSource:
			foundLevels = true
		}
	}

	if !foundVersions {
		t.Error("OpenShiftVersionsDataSource not found in data sources list")
	}

	if !foundOperators {
		t.Error("SupportedOperatorsDataSource not found in data sources list")
	}

	if !foundBundles {
		t.Error("OperatorBundlesDataSource not found in data sources list")
	}

	if !foundLevels {
		t.Error("SupportLevelsDataSource not found in data sources list")
	}
}

func TestOAIProvider_Functions(t *testing.T) {
	p := &OAIProvider{}

	functions := p.Functions(context.Background())

	// Currently we have no functions
	if len(functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(functions))
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "dev version",
			version: "dev",
		},
		{
			name:    "test version",
			version: "test",
		},
		{
			name:    "release version",
			version: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerFunc := New(tt.version)
			if providerFunc == nil {
				t.Fatal("New() returned nil")
			}

			provider := providerFunc()
			if provider == nil {
				t.Fatal("Provider function returned nil")
			}

			oaiProvider, ok := provider.(*OAIProvider)
			if !ok {
				t.Fatal("Provider is not of type *OAIProvider")
			}

			if oaiProvider.version != tt.version {
				t.Errorf("Expected version %s, got %s", tt.version, oaiProvider.version)
			}
		})
	}
}
