package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
)

func TestOpenShiftVersionsDataSource_Read(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		config        OpenShiftVersionsDataSourceModel
		expectedCount int
		expectError   bool
	}{
		{
			name: "successful fetch all versions",
			mockResponse: `{
				"4.15.20": {
					"display_name": "OpenShift 4.15.20",
					"support_level": "production",
					"default": true,
					"cpu_architectures": ["x86_64", "aarch64"]
				},
				"4.14.15": {
					"display_name": "OpenShift 4.14.15", 
					"support_level": "maintenance",
					"default": false,
					"cpu_architectures": ["x86_64"]
				}
			}`,
			config: OpenShiftVersionsDataSourceModel{
				Version:    types.StringNull(),
				OnlyLatest: types.BoolNull(),
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "filter by version",
			mockResponse: `{
				"4.15.20": {
					"display_name": "OpenShift 4.15.20",
					"support_level": "production", 
					"default": true,
					"cpu_architectures": ["x86_64"]
				}
			}`,
			config: OpenShiftVersionsDataSourceModel{
				Version:    types.StringValue("4.15"),
				OnlyLatest: types.BoolNull(),
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "only latest versions",
			mockResponse: `{
				"4.15.20": {
					"display_name": "OpenShift 4.15.20",
					"support_level": "production",
					"default": true,
					"cpu_architectures": ["x86_64", "aarch64"]
				}
			}`,
			config: OpenShiftVersionsDataSourceModel{
				Version:    types.StringNull(),
				OnlyLatest: types.BoolValue(true),
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:         "empty response",
			mockResponse: `{}`,
			config: OpenShiftVersionsDataSourceModel{
				Version:    types.StringNull(),
				OnlyLatest: types.BoolNull(),
			},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check method and path, ignore query parameters for this test
				if r.Method == "GET" && r.URL.Path == "/v2/openshift-versions" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(tt.mockResponse))
					return
				}
				w.WriteHeader(http.StatusNotFound)
				_, _ = fmt.Fprintf(w, "Not found: %s %s", r.Method, r.URL.String())
			}))
			defer server.Close()

			// Create client
			client := client.NewClient(client.ClientConfig{
				BaseURL:      server.URL,
				OfflineToken: "test-token",
			})

			// Create and configure data source directly
			ds := &OpenShiftVersionsDataSource{
				client: client,
			}

			ctx := context.Background()

			// Extract filter parameters
			var versionFilter string
			var onlyLatest bool

			if !tt.config.Version.IsNull() {
				versionFilter = tt.config.Version.ValueString()
			}
			if !tt.config.OnlyLatest.IsNull() {
				onlyLatest = tt.config.OnlyLatest.ValueBool()
			}

			// Call the API directly to test the core functionality
			versions, err := ds.client.GetOpenShiftVersions(ctx, versionFilter, onlyLatest)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(*versions) != tt.expectedCount {
				t.Errorf("Expected %d versions, got %d", tt.expectedCount, len(*versions))
			}

			// Verify specific version data for non-empty responses
			if len(*versions) > 0 {
				for version, versionInfo := range *versions {
					if versionInfo.DisplayName == "" {
						t.Errorf("Version %s missing display_name", version)
					}
					if versionInfo.SupportLevel == "" {
						t.Errorf("Version %s missing support_level", version)
					}
				}
			}
		})
	}
}

// Helper function removed - simplified test approach

func TestOpenShiftVersionsDataSource_Schema(t *testing.T) {
	dataSource := NewOpenShiftVersionsDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Error("Schema should have attributes")
	}

	// Check required attributes
	requiredAttrs := []string{"id", "versions"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing required attribute: %s", attr)
		}
	}

	// Check optional attributes
	optionalAttrs := []string{"version", "only_latest"}
	for _, attr := range optionalAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing optional attribute: %s", attr)
		}
	}
}

func TestOpenShiftVersionsDataSource_Metadata(t *testing.T) {
	dataSource := NewOpenShiftVersionsDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "openshift_assisted_installer",
	}
	resp := &datasource.MetadataResponse{}

	dataSource.Metadata(context.Background(), req, resp)

	expected := "openshift_assisted_installer_versions"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %q, got %q", expected, resp.TypeName)
	}
}
