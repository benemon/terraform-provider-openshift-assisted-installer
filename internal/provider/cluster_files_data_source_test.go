package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestClusterFilesDataSource_Schema(t *testing.T) {
	ctx := context.Background()
	dataSource := NewClusterFilesDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", resp.Diagnostics)
	}

	// Check that required attributes exist
	attrs := resp.Schema.Attributes
	requiredAttrs := []string{"id", "cluster_id", "file_name", "content"}
	for _, attr := range requiredAttrs {
		if _, ok := attrs[attr]; !ok {
			t.Errorf("%s attribute is missing", attr)
		}
	}

	// Check that cluster_id and file_name are required
	if !attrs["cluster_id"].IsRequired() {
		t.Error("cluster_id should be required")
	}
	if !attrs["file_name"].IsRequired() {
		t.Error("file_name should be required")
	}

	// Check that content is computed
	if !attrs["content"].IsComputed() {
		t.Error("content should be computed")
	}
}

func TestClusterFilesDataSource_Configure(t *testing.T) {
	dataSource := NewClusterFilesDataSource().(*ClusterFilesDataSource)
	
	// Test with nil provider data
	req := datasource.ConfigureRequest{ProviderData: nil}
	resp := &datasource.ConfigureResponse{}
	
	dataSource.Configure(context.Background(), req, resp)
	
	if resp.Diagnostics.HasError() {
		t.Error("Configure should not error with nil provider data")
	}
	
	// Test with wrong provider data type
	req.ProviderData = "wrong-type"
	resp = &datasource.ConfigureResponse{}
	
	dataSource.Configure(context.Background(), req, resp)
	
	if !resp.Diagnostics.HasError() {
		t.Error("Configure should error with wrong provider data type")
	}
	
	// Test with correct provider data
	testClient := client.NewClient(client.ClientConfig{
		BaseURL: "http://test.example.com",
	})
	req.ProviderData = testClient
	resp = &datasource.ConfigureResponse{}
	
	dataSource.Configure(context.Background(), req, resp)
	
	if resp.Diagnostics.HasError() {
		t.Errorf("Configure should not error with correct provider data: %+v", resp.Diagnostics)
	}
}

func TestClusterFilesDataSource_Read(t *testing.T) {
	mockFileContent := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value`

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		expectedPath := "/v2/clusters/test-cluster-id/downloads/files"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("file_name") != "manifests" {
			t.Errorf("Expected file_name=manifests, got %s", query.Get("file_name"))
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte(mockFileContent))
	}))
	defer server.Close()

	// Create client and data source
	testClient := client.NewClient(client.ClientConfig{
		BaseURL:      server.URL,
		OfflineToken: "test-token",
	})

	dataSource := NewClusterFilesDataSource().(*ClusterFilesDataSource)
	dataSource.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: testClient,
	}, &datasource.ConfigureResponse{})

	// Verify the data source was configured
	if dataSource.client == nil {
		t.Error("Expected client to be set after Configure")
	}
}