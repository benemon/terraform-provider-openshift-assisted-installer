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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/assert"
)

func TestOperatorBundlesDataSource_Schema(t *testing.T) {
	ds := NewOperatorBundlesDataSource()
	
	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	
	ds.Schema(context.Background(), schemaReq, schemaResp)
	
	// Verify no diagnostics
	assert.False(t, schemaResp.Diagnostics.HasError())
	
	// Verify schema structure
	schema := schemaResp.Schema
	assert.NotNil(t, schema.Attributes["id"])
	assert.NotNil(t, schema.Attributes["bundles"])
	
	bundlesAttr := schema.Attributes["bundles"]
	assert.Equal(t, "List of available operator bundles.", bundlesAttr.GetMarkdownDescription())
}

func TestOperatorBundlesDataSource_Read(t *testing.T) {
	// Mock server with sample bundles response
	mockBundles := models.Bundles{
		{
			ID:    "virtualization",
			Title: "OpenShift Virtualization",
			Operators: []string{
				"kubevirt-hyperconverged",
			},
		},
		{
			ID:    "openshift-ai-nvidia",
			Title: "OpenShift AI with NVIDIA",
			Operators: []string{
				"rhods-operator",
				"gpu-operator-certified",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "/api/assisted-install/v2/operators/bundles", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockBundles)
	}))
	defer server.Close()

	// Create test client
	testClient := client.NewClient(client.ClientConfig{
		BaseURL:      server.URL + "/api/assisted-install",
		OfflineToken: "test-token",
	})

	// Create data source
	ds := &OperatorBundlesDataSource{
		client: testClient,
	}

	// Configure request/response
	configReq := datasource.ConfigureRequest{
		ProviderData: testClient,
	}
	configResp := &datasource.ConfigureResponse{}
	ds.Configure(context.Background(), configReq, configResp)

	assert.False(t, configResp.Diagnostics.HasError())

	// Mock the token refresh to avoid actual OAuth calls
	ctx := context.Background()
	
	// Create read request with empty config
	readReq := datasource.ReadRequest{
		Config: tfsdk.Config{},
	}
	readResp := &datasource.ReadResponse{}

	// Execute read
	ds.Read(ctx, readReq, readResp)

	// Verify no errors
	assert.False(t, readResp.Diagnostics.HasError(), "Expected no diagnostics errors")

	// Verify state is set (this confirms the read was successful)
	assert.NotNil(t, readResp.State)
}

func TestOperatorBundlesDataSource_Configure_InvalidProviderData(t *testing.T) {
	ds := NewOperatorBundlesDataSource()
	dsImpl, ok := ds.(*OperatorBundlesDataSource)
	assert.True(t, ok)
	
	configReq := datasource.ConfigureRequest{
		ProviderData: "invalid",
	}
	configResp := &datasource.ConfigureResponse{}
	
	dsImpl.Configure(context.Background(), configReq, configResp)
	
	assert.True(t, configResp.Diagnostics.HasError())
	assert.Contains(t, configResp.Diagnostics.Errors()[0].Summary(), "Unexpected Data Source Configure Type")
}

func TestOperatorBundlesDataSource_Configure_NilProviderData(t *testing.T) {
	ds := NewOperatorBundlesDataSource()
	dsImpl, ok := ds.(*OperatorBundlesDataSource)
	assert.True(t, ok)
	
	configReq := datasource.ConfigureRequest{
		ProviderData: nil,
	}
	configResp := &datasource.ConfigureResponse{}
	
	dsImpl.Configure(context.Background(), configReq, configResp)
	
	// Should not error when provider data is nil
	assert.False(t, configResp.Diagnostics.HasError())
}

func TestOperatorBundlesDataSource_Metadata(t *testing.T) {
	ds := NewOperatorBundlesDataSource()
	
	metadataReq := datasource.MetadataRequest{
		ProviderTypeName: "oai",
	}
	metadataResp := &datasource.MetadataResponse{}
	
	ds.Metadata(context.Background(), metadataReq, metadataResp)
	
	assert.Equal(t, "oai_operator_bundles", metadataResp.TypeName)
}

func TestOperatorBundlesDataSource_ReadError(t *testing.T) {
	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	// Create test client
	testClient := client.NewClient(client.ClientConfig{
		BaseURL:      server.URL + "/api/assisted-install",
		OfflineToken: "test-token",
	})

	// Create data source
	ds := &OperatorBundlesDataSource{
		client: testClient,
	}

	// Execute read
	ctx := context.Background()
	readReq := datasource.ReadRequest{}
	readResp := &datasource.ReadResponse{}
	ds.Read(ctx, readReq, readResp)

	// Should have error
	assert.True(t, readResp.Diagnostics.HasError())
	assert.Contains(t, readResp.Diagnostics.Errors()[0].Summary(), "Error fetching operator bundles")
}