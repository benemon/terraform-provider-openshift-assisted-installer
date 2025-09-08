package provider

import (
	"context"
	"testing"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

func TestClusterDataSource_Schema(t *testing.T) {
	ds := NewClusterDataSource()

	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}

	ds.Schema(context.Background(), schemaReq, schemaResp)

	// Verify no diagnostics
	assert.False(t, schemaResp.Diagnostics.HasError())

	// Verify schema structure - essential Swagger-compliant fields
	schema := schemaResp.Schema
	assert.NotNil(t, schema.Attributes["id"])
	assert.NotNil(t, schema.Attributes["name"])
	assert.NotNil(t, schema.Attributes["status"])
	assert.NotNil(t, schema.Attributes["openshift_version"])
	assert.NotNil(t, schema.Attributes["platform"])
	assert.NotNil(t, schema.Attributes["base_dns_domain"])
	assert.NotNil(t, schema.Attributes["cluster_network_cidr"])
	assert.NotNil(t, schema.Attributes["service_network_cidr"])
	assert.NotNil(t, schema.Attributes["api_vips"])
	assert.NotNil(t, schema.Attributes["ingress_vips"])

	// Verify required field
	idAttr := schema.Attributes["id"]
	assert.True(t, idAttr.IsRequired())
}

func TestClusterDataSource_Metadata(t *testing.T) {
	ds := NewClusterDataSource()

	metadataReq := datasource.MetadataRequest{
		ProviderTypeName: "openshift_assisted_installer",
	}
	metadataResp := &datasource.MetadataResponse{}

	ds.Metadata(context.Background(), metadataReq, metadataResp)

	assert.Equal(t, "openshift_assisted_installer_cluster", metadataResp.TypeName)
}

func TestClusterDataSource_Configure(t *testing.T) {
	ds := &ClusterDataSource{}

	// Test with valid client
	testClient := client.NewClient(client.ClientConfig{
		BaseURL:      "https://api.example.com",
		OfflineToken: "test-token",
	})

	configReq := datasource.ConfigureRequest{
		ProviderData: testClient,
	}
	configResp := &datasource.ConfigureResponse{}

	ds.Configure(context.Background(), configReq, configResp)

	assert.False(t, configResp.Diagnostics.HasError())
	assert.Equal(t, testClient, ds.client)
}

func TestClusterDataSource_ConfigureError(t *testing.T) {
	ds := &ClusterDataSource{}

	// Test with invalid client type
	configReq := datasource.ConfigureRequest{
		ProviderData: "invalid-client",
	}
	configResp := &datasource.ConfigureResponse{}

	ds.Configure(context.Background(), configReq, configResp)

	assert.True(t, configResp.Diagnostics.HasError())
	assert.Nil(t, ds.client)
}
