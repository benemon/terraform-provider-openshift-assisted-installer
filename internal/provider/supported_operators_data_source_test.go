package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
)

func TestSupportedOperatorsDataSource_Read(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   string
		expectedCount  int
		expectError    bool
	}{
		{
			name: "successful fetch operators",
			mockResponse: `[
				"local-storage-operator",
				"odf-operator", 
				"cnv-operator",
				"lvm-operator"
			]`,
			expectedCount: 4,
			expectError:   false,
		},
		{
			name:          "empty response",
			mockResponse:  `[]`,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "single operator",
			mockResponse: `[
				"local-storage-operator"
			]`,
			expectedCount: 1,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" && r.URL.Path == "/v2/supported-operators" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.mockResponse))
					return
				}
				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			// Create client
			client := client.NewClient(client.ClientConfig{
				BaseURL: server.URL,
				Token:   "test-token",
			})

			// Create and configure data source directly
			ds := &SupportedOperatorsDataSource{
				client: client,
			}

			ctx := context.Background()

			// Call the API directly to test the core functionality
			operators, err := ds.client.GetSupportedOperators(ctx)

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

			if len(operators) != tt.expectedCount {
				t.Errorf("Expected %d operators, got %d", tt.expectedCount, len(operators))
			}

			// Verify operator names for non-empty responses
			if len(operators) > 0 {
				for _, operator := range operators {
					if operator == "" {
						t.Errorf("Found empty operator name")
					}
				}

				// Check for expected operators in the successful case
				if tt.name == "successful fetch operators" {
					expectedOperators := map[string]bool{
						"local-storage-operator": false,
						"odf-operator":           false,
						"cnv-operator":           false,
						"lvm-operator":           false,
					}
					
					for _, operator := range operators {
						if _, exists := expectedOperators[operator]; exists {
							expectedOperators[operator] = true
						}
					}
					
					for operator, found := range expectedOperators {
						if !found {
							t.Errorf("Expected operator %s not found", operator)
						}
					}
				}
			}
		})
	}
}

func TestSupportedOperatorsDataSource_Schema(t *testing.T) {
	dataSource := NewSupportedOperatorsDataSource()
	
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}
	
	dataSource.Schema(context.Background(), req, resp)
	
	if resp.Schema.Attributes == nil {
		t.Error("Schema should have attributes")
	}
	
	// Check required attributes
	requiredAttrs := []string{"id", "operators"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing required attribute: %s", attr)
		}
	}
}

func TestSupportedOperatorsDataSource_Metadata(t *testing.T) {
	dataSource := NewSupportedOperatorsDataSource()
	
	req := datasource.MetadataRequest{
		ProviderTypeName: "oai",
	}
	resp := &datasource.MetadataResponse{}
	
	dataSource.Metadata(context.Background(), req, resp)
	
	expected := "oai_supported_operators"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %q, got %q", expected, resp.TypeName)
	}
}

func TestSupportedOperatorsDataSource_Configure_Error(t *testing.T) {
	dataSource := NewSupportedOperatorsDataSource()
	
	// Test with wrong provider data type
	req := datasource.ConfigureRequest{
		ProviderData: "invalid-type",
	}
	resp := &datasource.ConfigureResponse{}
	
	// Cast to the concrete type to test Configure method
	if configurable, ok := dataSource.(*SupportedOperatorsDataSource); ok {
		configurable.Configure(context.Background(), req, resp)
	}
	
	if !resp.Diagnostics.HasError() {
		t.Error("Expected configuration error with invalid provider data type")
	}
}

func TestSupportedOperatorsDataSource_Configure_Nil(t *testing.T) {
	dataSource := NewSupportedOperatorsDataSource()
	
	// Test with nil provider data (should not error)
	req := datasource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &datasource.ConfigureResponse{}
	
	// Cast to the concrete type to test Configure method
	if configurable, ok := dataSource.(*SupportedOperatorsDataSource); ok {
		configurable.Configure(context.Background(), req, resp)
	}
	
	if resp.Diagnostics.HasError() {
		t.Error("Should not error with nil provider data")
	}
}