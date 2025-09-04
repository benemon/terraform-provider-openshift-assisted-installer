package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestClusterCredentialsDataSource_Schema(t *testing.T) {
	ctx := context.Background()
	dataSource := NewClusterCredentialsDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", resp.Diagnostics)
	}

	// Check that required attributes exist
	attrs := resp.Schema.Attributes
	if _, ok := attrs["cluster_id"]; !ok {
		t.Error("cluster_id attribute is missing")
	}
	if _, ok := attrs["username"]; !ok {
		t.Error("username attribute is missing")
	}
	if _, ok := attrs["password"]; !ok {
		t.Error("password attribute is missing")
	}
	if _, ok := attrs["console_url"]; !ok {
		t.Error("console_url attribute is missing")
	}

	// Check that password is marked as sensitive
	if !attrs["password"].IsSensitive() {
		t.Error("password attribute should be marked as sensitive")
	}
}

func TestClusterCredentialsDataSource_Read(t *testing.T) {
	// Mock credentials response
	mockCredentials := models.Credentials{
		Username:   "kubeadmin",
		Password:   "secret123",
		ConsoleURL: "https://console-openshift-console.apps.test-cluster.example.com",
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		expectedPath := "/v2/clusters/test-cluster-id/credentials"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Check authentication header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("Missing Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockCredentials)
	}))
	defer server.Close()

	// Create test provider with mock server
	testProvider := &OAIProvider{}
	testProvider.Configure(context.Background(), provider.ConfigureRequest{}, &provider.ConfigureResponse{
		DataSourceData: client.NewClient(client.ClientConfig{
			BaseURL:      server.URL,
			OfflineToken: "test-token",
		}),
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"oai": providerserver.NewProtocol6WithError(testProvider),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "oai_cluster_credentials" "test" {
						cluster_id = "test-cluster-id"
					}
				`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.oai_cluster_credentials.test", "cluster_id", "test-cluster-id"),
					resource.TestCheckResourceAttr("data.oai_cluster_credentials.test", "username", "kubeadmin"),
					resource.TestCheckResourceAttr("data.oai_cluster_credentials.test", "password", "secret123"),
					resource.TestCheckResourceAttr("data.oai_cluster_credentials.test", "console_url", "https://console-openshift-console.apps.test-cluster.example.com"),
				),
			},
		},
	})
}

func TestClusterCredentialsDataSource_Read_Error(t *testing.T) {
	// Create mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Cluster not found"))
	}))
	defer server.Close()

	// Create test provider with mock server
	testProvider := &OAIProvider{}
	testProvider.Configure(context.Background(), provider.ConfigureRequest{}, &provider.ConfigureResponse{
		DataSourceData: client.NewClient(client.ClientConfig{
			BaseURL:      server.URL,
			OfflineToken: "test-token",
		}),
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"oai": providerserver.NewProtocol6WithError(testProvider),
		},
		Steps: []resource.TestStep{
			{
				Config: `
					data "oai_cluster_credentials" "test" {
						cluster_id = "nonexistent-cluster"
					}
				`,
				ExpectError: regexp.MustCompile("API request failed with status 404"),
			},
		},
	})
}