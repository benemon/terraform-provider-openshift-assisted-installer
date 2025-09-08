package provider

import (
	"context"
	"testing"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

func TestInfraEnvDataSource_Schema(t *testing.T) {
	ds := NewInfraEnvDataSource()

	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}

	ds.Schema(context.Background(), schemaReq, schemaResp)

	// Verify no diagnostics
	assert.False(t, schemaResp.Diagnostics.HasError())

	// Verify schema structure - essential Swagger-compliant fields
	schema := schemaResp.Schema
	assert.NotNil(t, schema.Attributes["id"])
	assert.NotNil(t, schema.Attributes["name"])
	assert.NotNil(t, schema.Attributes["openshift_version"])
	assert.NotNil(t, schema.Attributes["download_url"])
	assert.NotNil(t, schema.Attributes["type"])
	assert.NotNil(t, schema.Attributes["cluster_id"])
	assert.NotNil(t, schema.Attributes["cpu_architecture"])
	assert.NotNil(t, schema.Attributes["ssh_authorized_key"])

	// Verify required field
	idAttr := schema.Attributes["id"]
	assert.True(t, idAttr.IsRequired())
}

func TestInfraEnvDataSource_Metadata(t *testing.T) {
	ds := NewInfraEnvDataSource()

	metadataReq := datasource.MetadataRequest{
		ProviderTypeName: "openshift_assisted_installer",
	}
	metadataResp := &datasource.MetadataResponse{}

	ds.Metadata(context.Background(), metadataReq, metadataResp)

	assert.Equal(t, "openshift_assisted_installer_infra_env", metadataResp.TypeName)
}

func TestInfraEnvDataSource_Configure(t *testing.T) {
	ds := &InfraEnvDataSource{}

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
