package provider

import (
	"context"
	"testing"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

func TestManifestDataSource_Schema(t *testing.T) {
	ds := NewManifestDataSource()

	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}

	ds.Schema(context.Background(), schemaReq, schemaResp)

	// Verify no diagnostics
	assert.False(t, schemaResp.Diagnostics.HasError())

	// Verify schema structure - essential Swagger-compliant fields
	schema := schemaResp.Schema
	assert.NotNil(t, schema.Attributes["id"])
	assert.NotNil(t, schema.Attributes["cluster_id"])
	assert.NotNil(t, schema.Attributes["file_name"])
	assert.NotNil(t, schema.Attributes["folder"])

	// Verify required fields
	clusterIdAttr := schema.Attributes["cluster_id"]
	assert.True(t, clusterIdAttr.IsRequired())

	fileNameAttr := schema.Attributes["file_name"]
	assert.True(t, fileNameAttr.IsRequired())
}

func TestManifestDataSource_Metadata(t *testing.T) {
	ds := NewManifestDataSource()

	metadataReq := datasource.MetadataRequest{
		ProviderTypeName: "openshift_assisted_installer",
	}
	metadataResp := &datasource.MetadataResponse{}

	ds.Metadata(context.Background(), metadataReq, metadataResp)

	assert.Equal(t, "openshift_assisted_installer_manifest", metadataResp.TypeName)
}

func TestManifestDataSource_Configure(t *testing.T) {
	ds := &ManifestDataSource{}

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
