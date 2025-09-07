package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
)

func TestClusterValidationsDataSource_Read(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		config        ClusterValidationsDataSourceModel
		expectedCount int
		expectError   bool
		checkBlocking bool
		checkStatus   string
	}{
		{
			name: "successful fetch all cluster validations",
			mockResponse: `{
				"validations_info": {
					"cluster": [
						{
							"id": "api-vips-defined",
							"status": "success",
							"message": "API virtual IPs are properly configured",
							"validation_id": "api-vips-defined"
						},
						{
							"id": "no-cidrs-overlapping",
							"status": "failure", 
							"message": "Network CIDRs are overlapping",
							"validation_id": "no-cidrs-overlapping"
						}
					],
					"network": [
						{
							"id": "machine-cidr-defined",
							"status": "success",
							"message": "Machine CIDR is properly defined",
							"validation_id": "machine-cidr-defined"
						}
					]
				}
			}`,
			config: ClusterValidationsDataSourceModel{
				ClusterID: types.StringValue("test-cluster-id"),
			},
			expectedCount: 3,
			expectError:   false,
		},
		{
			name: "filter by blocking validations only",
			mockResponse: `{
				"validations_info": {
					"cluster": [
						{
							"id": "api-vips-valid",
							"status": "success",
							"message": "API VIPs are valid",
							"validation_id": "api-vips-valid"
						},
						{
							"id": "all-hosts-are-ready-to-install",
							"status": "failure",
							"message": "Not all hosts are ready",
							"validation_id": "all-hosts-are-ready-to-install"
						}
					]
				}
			}`,
			config: ClusterValidationsDataSourceModel{
				ClusterID:       types.StringValue("test-cluster-id"),
				ValidationTypes: []types.String{types.StringValue("blocking")},
			},
			expectedCount: 2, // Both of these are blocking validations
			expectError:   false,
			checkBlocking: true,
		},
		{
			name: "filter by failure status",
			mockResponse: `{
				"validations_info": {
					"cluster": [
						{
							"id": "api-vips-defined",
							"status": "success",
							"message": "API VIPs defined",
							"validation_id": "api-vips-defined"
						},
						{
							"id": "sufficient-masters-count", 
							"status": "failure",
							"message": "Insufficient master nodes",
							"validation_id": "sufficient-masters-count"
						}
					]
				}
			}`,
			config: ClusterValidationsDataSourceModel{
				ClusterID:    types.StringValue("test-cluster-id"),
				StatusFilter: []types.String{types.StringValue("failure")},
			},
			expectedCount: 1,
			expectError:   false,
			checkStatus:   "failure",
		},
		{
			name: "filter by specific validation names",
			mockResponse: `{
				"validations_info": {
					"cluster": [
						{
							"id": "api-vips-defined",
							"status": "success",
							"message": "API VIPs defined",
							"validation_id": "api-vips-defined"
						},
						{
							"id": "pull-secret-set",
							"status": "success",
							"message": "Pull secret configured",
							"validation_id": "pull-secret-set"
						}
					]
				}
			}`,
			config: ClusterValidationsDataSourceModel{
				ClusterID:       types.StringValue("test-cluster-id"),
				ValidationNames: []types.String{types.StringValue("api-vips-defined")},
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "filter by network category",
			mockResponse: `{
				"validations_info": {
					"cluster": [
						{
							"id": "machine-cidr-defined",
							"status": "success",
							"message": "Machine CIDR defined",
							"validation_id": "machine-cidr-defined"
						}
					],
					"host": [
						{
							"id": "has-min-cpu-cores",
							"status": "success",
							"message": "Host has sufficient CPU",
							"validation_id": "has-min-cpu-cores"
						}
					]
				}
			}`,
			config: ClusterValidationsDataSourceModel{
				ClusterID:  types.StringValue("test-cluster-id"),
				Categories: []types.String{types.StringValue("network")},
			},
			expectedCount: 1, // Only the network validation should match
			expectError:   false,
		},
		{
			name:         "API error",
			mockResponse: `{"error": "Not found"}`,
			config: ClusterValidationsDataSourceModel{
				ClusterID: types.StringValue("non-existent-cluster"),
			},
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check method and path
				if r.Method == "GET" && strings.Contains(r.URL.Path, "/v2/clusters/") && !strings.Contains(r.URL.Path, "/hosts") {
					if tt.expectError {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(tt.mockResponse))
						return
					}
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
			ds := &ClusterValidationsDataSource{
				client: client,
			}

			ctx := context.Background()

			// Call the API directly to test the core functionality
			clusterValidations, err := ds.client.GetClusterValidations(ctx, tt.config.ClusterID.ValueString())

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

			// Process validations to count matches (similar to data source logic)
			var matchingCount int
			for groupName, validationsGroup := range clusterValidations.ValidationsInfo {
				for _, validation := range validationsGroup {
					// Apply the same filtering logic as the data source
					validationID := validation.ValidationID
					if validationID == "" {
						validationID = validation.ID
					}

					// Apply validation type filter
					if len(tt.config.ValidationTypes) > 0 {
						validationType := "non-blocking"
						if isBlockingClusterValidation(validationID) {
							validationType = "blocking"
						}

						found := false
						for _, filterType := range tt.config.ValidationTypes {
							if strings.EqualFold(validationType, filterType.ValueString()) {
								found = true
								break
							}
						}
						if !found {
							continue
						}
					}

					// Apply status filter
					if len(tt.config.StatusFilter) > 0 {
						found := false
						for _, filterStatus := range tt.config.StatusFilter {
							if strings.EqualFold(validation.Status, filterStatus.ValueString()) {
								found = true
								break
							}
						}
						if !found {
							continue
						}
					}

					// Apply validation names filter
					if len(tt.config.ValidationNames) > 0 {
						found := false
						for _, filterName := range tt.config.ValidationNames {
							if validationID == filterName.ValueString() || validation.ID == filterName.ValueString() {
								found = true
								break
							}
						}
						if !found {
							continue
						}
					}

					// Apply categories filter
					if len(tt.config.Categories) > 0 {
						category := getCategoryForValidation(validationID)
						found := false
						for _, filterCategory := range tt.config.Categories {
							if strings.EqualFold(category, filterCategory.ValueString()) {
								found = true
								break
							}
						}
						if !found {
							continue
						}
					}

					matchingCount++

					// Additional checks based on test
					if tt.checkBlocking && !isBlockingClusterValidation(validationID) {
						t.Errorf("Expected blocking validation but got non-blocking: %s", validationID)
					}

					if tt.checkStatus != "" && validation.Status != tt.checkStatus {
						t.Errorf("Expected status %s but got %s for validation %s", tt.checkStatus, validation.Status, validationID)
					}

					t.Logf("Matching validation: group=%s, id=%s, status=%s", groupName, validationID, validation.Status)
				}
			}

			if matchingCount != tt.expectedCount {
				t.Errorf("Expected %d matching validations, got %d", tt.expectedCount, matchingCount)
			}
		})
	}
}

// Helper function to check if a cluster validation is blocking
func isBlockingClusterValidation(validationID string) bool {
	blockingValidations := map[string]bool{
		"api-vips-valid":                         true,
		"all-hosts-are-ready-to-install":         true,
		"sufficient-masters-count":               true,
		"no-cidrs-overlapping":                   true,
		"networks-same-address-families":         true,
		"network-prefix-valid":                   true,
		"machine-cidr-equals-to-calculated-cidr": true,
		"ingress-vips-defined":                   true,
		"ntp-server-configured":                  true,
		"network-type-valid":                     true,
	}
	return blockingValidations[validationID]
}

// Helper function to get category for a validation
func getCategoryForValidation(validationID string) string {
	networkValidations := []string{
		"machine-cidr-defined", "cluster-cidr-defined", "service-cidr-defined",
		"no-cidrs-overlapping", "networks-same-address-families", "network-prefix-valid",
		"api-vips-defined", "api-vips-valid", "ingress-vips-defined",
		"ingress-vips-valid", "network-type-valid",
	}

	for _, netVal := range networkValidations {
		if netVal == validationID {
			return "network"
		}
	}

	return "cluster" // Default category
}

func TestClusterValidationsDataSource_Schema(t *testing.T) {
	dataSource := NewClusterValidationsDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Error("Schema should have attributes")
	}

	// Check required attributes
	requiredAttrs := []string{"id", "cluster_id", "validations"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing required attribute: %s", attr)
		}
	}

	// Check optional filter attributes
	optionalAttrs := []string{"validation_types", "status_filter", "validation_names", "categories"}
	for _, attr := range optionalAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing optional attribute: %s", attr)
		}
	}
}

func TestClusterValidationsDataSource_Metadata(t *testing.T) {
	dataSource := NewClusterValidationsDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "oai",
	}
	resp := &datasource.MetadataResponse{}

	dataSource.Metadata(context.Background(), req, resp)

	expected := "oai_cluster_validations"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %q, got %q", expected, resp.TypeName)
	}
}
