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
	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
)

func TestHostValidationsDataSource_Read(t *testing.T) {
	tests := []struct {
		name            string
		mockResponse    string
		config          HostValidationsDataSourceModel
		expectedCount   int
		expectError     bool
		checkBlocking   bool
		checkStatus     string
		checkHostCount  int
	}{
		{
			name: "successful fetch all host validations for cluster",
			mockResponse: `[
				{
					"id": "host-1",
					"validations_info": {
						"hardware": [
							{
								"id": "has-min-cpu-cores",
								"status": "success",
								"message": "Host has sufficient CPU cores",
								"validation_id": "has-min-cpu-cores"
							},
							{
								"id": "has-min-memory",
								"status": "failure",
								"message": "Insufficient memory",
								"validation_id": "has-min-memory"
							}
						],
						"network": [
							{
								"id": "has-default-route",
								"status": "success",
								"message": "Host has default route configured",
								"validation_id": "has-default-route"
							}
						]
					}
				},
				{
					"id": "host-2", 
					"validations_info": {
						"hardware": [
							{
								"id": "has-min-cpu-cores",
								"status": "success",
								"message": "Host has sufficient CPU cores",
								"validation_id": "has-min-cpu-cores"
							}
						]
					}
				}
			]`,
			config: HostValidationsDataSourceModel{
				ClusterID: types.StringValue("test-cluster-id"),
			},
			expectedCount:  4, // 3 validations from host-1 + 1 from host-2
			expectError:    false,
			checkHostCount: 2,
		},
		{
			name: "filter by blocking validations only",
			mockResponse: `[
				{
					"id": "host-1",
					"validations_info": {
						"hardware": [
							{
								"id": "has-cpu-cores-for-role",
								"status": "success",
								"message": "CPU cores sufficient for role",
								"validation_id": "has-cpu-cores-for-role"
							}
						],
						"network": [
							{
								"id": "has-default-route",
								"status": "failure",
								"message": "No default route",
								"validation_id": "has-default-route"
							}
						]
					}
				}
			]`,
			config: HostValidationsDataSourceModel{
				ClusterID:       types.StringValue("test-cluster-id"),
				ValidationTypes: []types.String{types.StringValue("blocking")},
			},
			expectedCount: 2, // Both validations are blocking
			expectError:   false,
			checkBlocking: true,
		},
		{
			name: "filter by failure status",
			mockResponse: `[
				{
					"id": "host-1",
					"validations_info": {
						"hardware": [
							{
								"id": "has-min-memory",
								"status": "success",
								"message": "Sufficient memory",
								"validation_id": "has-min-memory"
							},
							{
								"id": "has-min-cpu-cores",
								"status": "failure",
								"message": "Insufficient CPU cores",
								"validation_id": "has-min-cpu-cores"
							}
						]
					}
				}
			]`,
			config: HostValidationsDataSourceModel{
				ClusterID:    types.StringValue("test-cluster-id"),
				StatusFilter: []types.String{types.StringValue("failure")},
			},
			expectedCount: 1,
			expectError:   false,
			checkStatus:   "failure",
		},
		{
			name: "successful fetch single host validations",
			mockResponse: `{
				"id": "specific-host-id",
				"validations_info": {
					"hardware": [
						{
							"id": "has-inventory",
							"status": "success",
							"message": "Host inventory available",
							"validation_id": "has-inventory"
						}
					]
				}
			}`,
			config: HostValidationsDataSourceModel{
				InfraEnvID: types.StringValue("test-infra-env-id"),
				HostID:     types.StringValue("specific-host-id"),
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "filter by specific validation names",
			mockResponse: `[
				{
					"id": "host-1",
					"validations_info": {
						"hardware": [
							{
								"id": "has-min-cpu-cores",
								"status": "success",
								"message": "CPU sufficient",
								"validation_id": "has-min-cpu-cores"
							},
							{
								"id": "has-min-memory",
								"status": "success", 
								"message": "Memory sufficient",
								"validation_id": "has-min-memory"
							}
						]
					}
				}
			]`,
			config: HostValidationsDataSourceModel{
				ClusterID:       types.StringValue("test-cluster-id"),
				ValidationNames: []types.String{types.StringValue("has-min-cpu-cores")},
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "filter by hardware category",
			mockResponse: `[
				{
					"id": "host-1",
					"validations_info": {
						"hardware": [
							{
								"id": "has-min-cpu-cores",
								"status": "success",
								"message": "CPU sufficient",
								"validation_id": "has-min-cpu-cores"
							}
						],
						"network": [
							{
								"id": "has-default-route",
								"status": "success",
								"message": "Default route configured",
								"validation_id": "has-default-route"
							}
						]
					}
				}
			]`,
			config: HostValidationsDataSourceModel{
				ClusterID:  types.StringValue("test-cluster-id"),
				Categories: []types.String{types.StringValue("hardware")},
			},
			expectedCount: 1, // Only the hardware validation should match
			expectError:   false,
		},
		{
			name:         "API error for cluster hosts",
			mockResponse: `{"error": "Not found"}`,
			config: HostValidationsDataSourceModel{
				ClusterID: types.StringValue("non-existent-cluster"),
			},
			expectedCount: 0,
			expectError:   true,
		},
		{
			name:         "API error for single host",
			mockResponse: `{"error": "Not found"}`,
			config: HostValidationsDataSourceModel{
				InfraEnvID: types.StringValue("non-existent-infra-env"),
				HostID:     types.StringValue("non-existent-host"),
			},
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check different endpoints based on configuration
				var validPath bool
				if !tt.config.ClusterID.IsNull() {
					// Cluster hosts endpoint: /v2/clusters/{id}/hosts
					validPath = r.Method == "GET" && strings.Contains(r.URL.Path, "/v2/clusters/") && strings.Contains(r.URL.Path, "/hosts")
				} else {
					// Single host endpoint: /v2/infra-envs/{id}/hosts/{hostId}  
					validPath = r.Method == "GET" && strings.Contains(r.URL.Path, "/v2/infra-envs/") && strings.Contains(r.URL.Path, "/hosts/")
				}

				if validPath {
					if tt.expectError {
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(tt.mockResponse))
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.mockResponse))
					return
				}
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(fmt.Sprintf("Not found: %s %s", r.Method, r.URL.String())))
			}))
			defer server.Close()

			// Create client
			client := client.NewClient(client.ClientConfig{
				BaseURL:      server.URL,
				OfflineToken: "test-token",
			})

			// Create and configure data source directly
			ds := &HostValidationsDataSource{
				client: client,
			}

			ctx := context.Background()

			var err error
			var hostValidations *models.HostsValidationResponse
			var singleHostValidation *models.HostValidationResponse

			// Call appropriate client method based on configuration
			if !tt.config.ClusterID.IsNull() {
				hostValidations, err = ds.client.GetHostValidations(ctx, tt.config.ClusterID.ValueString())
			} else {
				singleHostValidation, err = ds.client.GetSingleHostValidations(ctx, tt.config.InfraEnvID.ValueString(), tt.config.HostID.ValueString())
				if err == nil {
					// Convert single host to hosts list format
					hostValidations = &models.HostsValidationResponse{
						Hosts: []models.HostValidationResponse{*singleHostValidation},
					}
				}
			}

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

			// Verify host count if specified
			if tt.checkHostCount > 0 {
				if len(hostValidations.Hosts) != tt.checkHostCount {
					t.Errorf("Expected %d hosts, got %d", tt.checkHostCount, len(hostValidations.Hosts))
				}
			}

			// Process validations to count matches (similar to data source logic)
			var matchingCount int
			for _, host := range hostValidations.Hosts {
				for groupName, validationsGroup := range host.ValidationsInfo {
					for _, validation := range validationsGroup {
						validationID := validation.ValidationID
						if validationID == "" {
							validationID = validation.ID
						}

						// Apply validation type filter
						if len(tt.config.ValidationTypes) > 0 {
							validationType := "non-blocking"
							if isBlockingHostValidation(validationID) {
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
							category := getCategoryForHostValidation(validationID)
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
						if tt.checkBlocking && !isBlockingHostValidation(validationID) {
							t.Errorf("Expected blocking validation but got non-blocking: %s", validationID)
						}

						if tt.checkStatus != "" && validation.Status != tt.checkStatus {
							t.Errorf("Expected status %s but got %s for validation %s", tt.checkStatus, validation.Status, validationID)
						}

						t.Logf("Matching validation: host=%s, group=%s, id=%s, status=%s", host.ID, groupName, validationID, validation.Status)
					}
				}
			}

			if matchingCount != tt.expectedCount {
				t.Errorf("Expected %d matching validations, got %d", tt.expectedCount, matchingCount)
			}
		})
	}
}

// Helper function to check if a host validation is blocking
func isBlockingHostValidation(validationID string) bool {
	blockingValidations := map[string]bool{
		"has-cpu-cores-for-role":              true,
		"has-memory-for-role":                 true,
		"ignition-downloadable":               true,
		"belongs-to-majority-group":           true,
		"valid-platform-network-settings":    true,
		"sufficient-installation-diskspeed":   true,
		"sufficient-network-latency":          true,
		"sufficient-packet-loss":              true,
		"has-default-route":                   true,
		"api-domain-name-resolved-correctly":  true,
		"api-int-domain-name-resolved-correctly": true,
		"apps-domain-name-resolved-correctly": true,
		"dns-wildcard-not-configured":         true,
		"non-overlapping-subnets":             true,
		"hostname-unique":                     true,
		"hostname-valid":                      true,
		"belongs-to-machine-cidr":             true,
		"lso-requirements-satisfied":          true,
		"odf-requirements-satisfied":          true,
		"cnv-requirements-satisfied":          true,
		"lvm-requirements-satisfied":          true,
		"compatible-agent":                    true,
		"no-skip-installation-disk":           true,
		"no-skip-missing-disk":                true,
		"media-connected":                     true,
	}
	return blockingValidations[validationID]
}

// Helper function to get category for a host validation
func getCategoryForHostValidation(validationID string) string {
	networkValidations := []string{
		"has-default-route", "api-domain-name-resolved-correctly",
		"api-int-domain-name-resolved-correctly", "apps-domain-name-resolved-correctly",
		"non-overlapping-subnets", "belongs-to-machine-cidr",
		"sufficient-network-latency", "sufficient-packet-loss", "mtu-valid",
	}
	
	hardwareValidations := []string{
		"has-min-cpu-cores", "has-min-memory", "has-min-valid-disks",
		"has-cpu-cores-for-role", "has-memory-for-role", "connected", "has-inventory",
	}
	
	operatorValidations := []string{
		"lso-requirements-satisfied", "odf-requirements-satisfied",
		"cnv-requirements-satisfied", "lvm-requirements-satisfied",
	}
	
	storageValidations := []string{
		"sufficient-installation-diskspeed", "no-skip-installation-disk",
		"no-skip-missing-disk", "disk-encryption-requirements-satisfied",
	}
	
	platformValidations := []string{
		"compatible-with-cluster-platform", "valid-platform-network-settings",
		"vsphere-disk-uuid-enabled", "compatible-agent",
	}
	
	for _, netVal := range networkValidations {
		if netVal == validationID {
			return "network"
		}
	}
	
	for _, hwVal := range hardwareValidations {
		if hwVal == validationID {
			return "hardware"
		}
	}
	
	for _, opVal := range operatorValidations {
		if opVal == validationID {
			return "operators"
		}
	}
	
	for _, storageVal := range storageValidations {
		if storageVal == validationID {
			return "storage"
		}
	}
	
	for _, platformVal := range platformValidations {
		if platformVal == validationID {
			return "platform"
		}
	}
	
	return "cluster" // Default category
}

func TestHostValidationsDataSource_Schema(t *testing.T) {
	dataSource := NewHostValidationsDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Error("Schema should have attributes")
	}

	// Check required attributes
	requiredAttrs := []string{"id", "validations"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing required attribute: %s", attr)
		}
	}

	// Check optional configuration attributes
	optionalAttrs := []string{"cluster_id", "host_id", "infra_env_id", "validation_types", "status_filter", "validation_names", "categories"}
	for _, attr := range optionalAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing optional attribute: %s", attr)
		}
	}
}

func TestHostValidationsDataSource_Metadata(t *testing.T) {
	dataSource := NewHostValidationsDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "oai",
	}
	resp := &datasource.MetadataResponse{}

	dataSource.Metadata(context.Background(), req, resp)

	expected := "oai_host_validations"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %q, got %q", expected, resp.TypeName)
	}
}