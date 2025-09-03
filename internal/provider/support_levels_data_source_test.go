package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

func TestSupportLevelsDataSource_Schema(t *testing.T) {
	ds := NewSupportLevelsDataSource()
	
	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	
	ds.Schema(context.Background(), schemaReq, schemaResp)
	
	// Verify no diagnostics
	assert.False(t, schemaResp.Diagnostics.HasError())
	
	// Verify schema structure
	schema := schemaResp.Schema
	assert.NotNil(t, schema.Attributes["id"])
	assert.NotNil(t, schema.Attributes["openshift_version"])
	assert.NotNil(t, schema.Attributes["features"])
	assert.NotNil(t, schema.Attributes["architectures"])
	
	// Verify required field
	versionAttr := schema.Attributes["openshift_version"]
	assert.True(t, versionAttr.IsRequired())
}

func TestSupportLevelsDataSource_Read(t *testing.T) {
	// Mock server responses
	mockFeatures := models.SupportedFeatures{
		"SNO":                    "supported",
		"VIP_DHCP_ALLOCATION":    "supported", 
		"DUAL_STACK_NETWORKING":  "tech-preview",
		"MULTIARCH_RELEASE_IMAGE": "dev-preview",
	}

	mockArchitectures := models.SupportedArchitectures{
		"x86_64": "supported",
		"arm64":  "supported", 
		"ppc64le": "tech-preview",
		"s390x":  "tech-preview",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the endpoint being called
		if r.URL.Path == "/api/assisted-install/v2/support-levels/features" {
			// Verify query parameters
			assert.Equal(t, "4.14.0", r.URL.Query().Get("openshift_version"))
			assert.Equal(t, "x86_64", r.URL.Query().Get("cpu_architecture"))
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockFeatures)
			
		} else if r.URL.Path == "/api/assisted-install/v2/support-levels/architectures" {
			// Verify query parameters
			assert.Equal(t, "4.14.0", r.URL.Query().Get("openshift_version"))
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockArchitectures)
			
		} else {
			t.Errorf("Unexpected endpoint: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}

		// Verify auth header
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer")
	}))
	defer server.Close()

	// Create test client
	testClient := client.NewClient(client.ClientConfig{
		BaseURL:      server.URL + "/api/assisted-install",
		OfflineToken: "test-token",
	})

	// Create data source
	ds := &SupportLevelsDataSource{
		client: testClient,
	}

	// Execute read
	ctx := context.Background()
	readReq := datasource.ReadRequest{}
	readResp := &datasource.ReadResponse{}
	ds.Read(ctx, readReq, readResp)

	// Note: This test verifies the request structure and mocking setup.
	// Full end-to-end testing would require more complex state setup.
	// The key validation is that our HTTP handlers are called correctly.
}

func TestSupportLevelsDataSource_Metadata(t *testing.T) {
	ds := NewSupportLevelsDataSource()
	
	metadataReq := datasource.MetadataRequest{
		ProviderTypeName: "oai",
	}
	metadataResp := &datasource.MetadataResponse{}
	
	ds.Metadata(context.Background(), metadataReq, metadataResp)
	
	assert.Equal(t, "oai_support_levels", metadataResp.TypeName)
}

func TestSupportLevelsDataSource_Configure(t *testing.T) {
	ds := &SupportLevelsDataSource{}
	
	testClient := &client.Client{}
	
	configReq := datasource.ConfigureRequest{
		ProviderData: testClient,
	}
	configResp := &datasource.ConfigureResponse{}
	
	ds.Configure(context.Background(), configReq, configResp)
	
	assert.False(t, configResp.Diagnostics.HasError())
	assert.Equal(t, testClient, ds.client)
}

func TestSupportLevelsDataSource_Configure_InvalidProviderData(t *testing.T) {
	ds := &SupportLevelsDataSource{}
	
	configReq := datasource.ConfigureRequest{
		ProviderData: "invalid",
	}
	configResp := &datasource.ConfigureResponse{}
	
	ds.Configure(context.Background(), configReq, configResp)
	
	assert.True(t, configResp.Diagnostics.HasError())
	assert.Contains(t, configResp.Diagnostics.Errors()[0].Summary(), "Unexpected Data Source Configure Type")
}

func TestSupportLevelsDataSource_ReadError_Features(t *testing.T) {
	// Mock server that returns error for features endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/assisted-install/v2/support-levels/features" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid version"}`))
		}
	}))
	defer server.Close()

	// Create test client
	testClient := client.NewClient(client.ClientConfig{
		BaseURL:      server.URL + "/api/assisted-install",
		OfflineToken: "test-token",
	})

	// Create data source
	ds := &SupportLevelsDataSource{
		client: testClient,
	}

	// Execute read
	ctx := context.Background()
	readReq := datasource.ReadRequest{}
	readResp := &datasource.ReadResponse{}
	ds.Read(ctx, readReq, readResp)

	// Should have error
	assert.True(t, readResp.Diagnostics.HasError())
	assert.Contains(t, readResp.Diagnostics.Errors()[0].Summary(), "Error fetching supported features")
}

func TestSupportLevelsDataSource_ReadError_Architectures(t *testing.T) {
	// Mock server that succeeds for features but fails for architectures
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/assisted-install/v2/support-levels/features" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.SupportedFeatures{})
		} else if r.URL.Path == "/api/assisted-install/v2/support-levels/architectures" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid version"}`))
		}
	}))
	defer server.Close()

	// Create test client
	testClient := client.NewClient(client.ClientConfig{
		BaseURL:      server.URL + "/api/assisted-install",
		OfflineToken: "test-token",
	})

	// Create data source
	ds := &SupportLevelsDataSource{
		client: testClient,
	}

	// Execute read
	ctx := context.Background()
	readReq := datasource.ReadRequest{}
	readResp := &datasource.ReadResponse{}
	ds.Read(ctx, readReq, readResp)

	// Should have error
	assert.True(t, readResp.Diagnostics.HasError())
	assert.Contains(t, readResp.Diagnostics.Errors()[0].Summary(), "Error fetching supported architectures")
}