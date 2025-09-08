package provider

import (
	"context"
	"testing"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

func TestHostDataSource_Schema(t *testing.T) {
	ds := NewHostDataSource()

	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}

	ds.Schema(context.Background(), schemaReq, schemaResp)

	// Verify no diagnostics
	assert.False(t, schemaResp.Diagnostics.HasError())

	// Verify schema structure - essential Swagger-compliant fields
	schema := schemaResp.Schema
	assert.NotNil(t, schema.Attributes["id"])
	assert.NotNil(t, schema.Attributes["infra_env_id"])
	assert.NotNil(t, schema.Attributes["cluster_id"])
	assert.NotNil(t, schema.Attributes["status"])
	assert.NotNil(t, schema.Attributes["role"])
	assert.NotNil(t, schema.Attributes["progress"])
	assert.NotNil(t, schema.Attributes["inventory"])
	assert.NotNil(t, schema.Attributes["requested_hostname"])

	// Verify required fields
	idAttr := schema.Attributes["id"]
	assert.True(t, idAttr.IsRequired())

	infraEnvIdAttr := schema.Attributes["infra_env_id"]
	assert.True(t, infraEnvIdAttr.IsRequired())
}

func TestHostDataSource_Metadata(t *testing.T) {
	ds := NewHostDataSource()

	metadataReq := datasource.MetadataRequest{
		ProviderTypeName: "openshift_assisted_installer",
	}
	metadataResp := &datasource.MetadataResponse{}

	ds.Metadata(context.Background(), metadataReq, metadataResp)

	assert.Equal(t, "openshift_assisted_installer_host", metadataResp.TypeName)
}

func TestHostDataSource_Configure(t *testing.T) {
	ds := &HostDataSource{}

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
