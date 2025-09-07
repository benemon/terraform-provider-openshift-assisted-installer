package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestClusterLogsDataSource_Schema(t *testing.T) {
	ctx := context.Background()
	dataSource := NewClusterLogsDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", resp.Diagnostics)
	}

	// Check that required attributes exist
	attrs := resp.Schema.Attributes
	requiredAttrs := []string{"id", "cluster_id", "content"}
	for _, attr := range requiredAttrs {
		if _, ok := attrs[attr]; !ok {
			t.Errorf("%s attribute is missing", attr)
		}
	}

	optionalAttrs := []string{"logs_type", "host_id"}
	for _, attr := range optionalAttrs {
		if _, ok := attrs[attr]; !ok {
			t.Errorf("%s attribute is missing", attr)
		}
	}
}

func TestClusterLogsDataSource_Configure(t *testing.T) {
	dataSource := NewClusterLogsDataSource().(*ClusterLogsDataSource)

	// Test with nil provider data
	req := datasource.ConfigureRequest{ProviderData: nil}
	resp := &datasource.ConfigureResponse{}

	dataSource.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Error("Configure should not error with nil provider data")
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

func TestClusterLogsDataSource_Read(t *testing.T) {
	mockLogContent := "2023-01-01 12:00:00 INFO: Cluster installation started\n2023-01-01 12:05:00 INFO: Host validation completed\n"

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		expectedPath := "/v2/clusters/test-cluster-id/logs"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(mockLogContent))
	}))
	defer server.Close()

	// Create client and data source
	testClient := client.NewClient(client.ClientConfig{
		BaseURL:      server.URL,
		OfflineToken: "test-token",
	})

	dataSource := NewClusterLogsDataSource().(*ClusterLogsDataSource)
	dataSource.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: testClient,
	}, &datasource.ConfigureResponse{})

	// Verify the data source was configured
	if dataSource.client == nil {
		t.Error("Expected client to be set after Configure")
	}
}
