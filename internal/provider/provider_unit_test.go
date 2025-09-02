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

	if resp.TypeName != "oai" {
		t.Errorf("Expected TypeName 'oai', got %s", resp.TypeName)
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

	if _, ok := attrs["token"]; !ok {
		t.Error("Schema missing 'token' attribute")
	}

	if _, ok := attrs["timeout"]; !ok {
		t.Error("Schema missing 'timeout' attribute")
	}

	// Check that token is marked as sensitive
	tokenAttr := attrs["token"]
	if !tokenAttr.IsSensitive() {
		t.Error("Token attribute should be marked as sensitive")
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
	expectedResources := 1 // Currently only ClusterResource
	if len(resources) != expectedResources {
		t.Errorf("Expected %d resources, got %d", expectedResources, len(resources))
	}

	// Verify that NewClusterResource is in the list
	found := false
	for _, r := range resources {
		// Create an instance to check the type
		instance := r()
		if _, ok := instance.(*ClusterResource); ok {
			found = true
			break
		}
	}

	if !found {
		t.Error("ClusterResource not found in resources list")
	}
}

func TestOAIProvider_DataSources(t *testing.T) {
	p := &OAIProvider{}

	dataSources := p.DataSources(context.Background())

	// Check that we have the expected number of data sources
	expectedDataSources := 2 // OpenShiftVersionsDataSource and SupportedOperatorsDataSource
	if len(dataSources) != expectedDataSources {
		t.Errorf("Expected %d data sources, got %d", expectedDataSources, len(dataSources))
	}

	// Verify that both data sources are in the list
	var foundVersions, foundOperators bool
	for _, ds := range dataSources {
		instance := ds()
		switch instance.(type) {
		case *OpenShiftVersionsDataSource:
			foundVersions = true
		case *SupportedOperatorsDataSource:
			foundOperators = true
		}
	}

	if !foundVersions {
		t.Error("OpenShiftVersionsDataSource not found in data sources list")
	}
	
	if !foundOperators {
		t.Error("SupportedOperatorsDataSource not found in data sources list")
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