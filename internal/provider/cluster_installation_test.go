package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
)

func TestClusterResource_waitForClusterState(t *testing.T) {
	tests := []struct {
		name           string
		targetStates   []string
		mockResponses  []string
		expectError    bool
		expectedError  string
		pollTimeout    time.Duration
	}{
		{
			name:         "successful state transition",
			targetStates: []string{"ready"},
			mockResponses: []string{
				`{"id": "test-cluster", "status": "insufficient", "status_info": "Waiting for hosts"}`,
				`{"id": "test-cluster", "status": "ready", "status_info": "Cluster is ready for installation"}`,
			},
			expectError: false,
			pollTimeout: 2 * time.Minute,
		},
		{
			name:         "reaches error state",
			targetStates: []string{"ready"},
			mockResponses: []string{
				`{"id": "test-cluster", "status": "insufficient", "status_info": "Waiting for hosts"}`,
				`{"id": "test-cluster", "status": "error", "status_info": "Validation failed"}`,
			},
			expectError:   true,
			expectedError: "cluster reached error state: error - Validation failed",
			pollTimeout:   2 * time.Minute,
		},
		{
			name:         "timeout before reaching target state",
			targetStates: []string{"ready"},
			mockResponses: []string{
				`{"id": "test-cluster", "status": "insufficient", "status_info": "Still waiting"}`,
				`{"id": "test-cluster", "status": "insufficient", "status_info": "Still waiting"}`,
				`{"id": "test-cluster", "status": "insufficient", "status_info": "Still waiting"}`,
			},
			expectError:   true,
			expectedError: "timeout waiting for cluster to reach states [ready]",
			pollTimeout:   100 * time.Millisecond, // Short timeout to trigger quickly
		},
		{
			name:         "multiple target states - reaches second one",
			targetStates: []string{"ready", "installed"},
			mockResponses: []string{
				`{"id": "test-cluster", "status": "insufficient", "status_info": "Waiting"}`,
				`{"id": "test-cluster", "status": "installing", "status_info": "Installing"}`,
				`{"id": "test-cluster", "status": "installed", "status_info": "Installation complete"}`,
			},
			expectError: false,
			pollTimeout: 2 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responseIndex := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" && strings.Contains(r.URL.Path, "/clusters/test-cluster") {
					if responseIndex < len(tt.mockResponses) {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(tt.mockResponses[responseIndex]))
						responseIndex++
					} else {
						// Repeat last response for timeout scenarios
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(tt.mockResponses[len(tt.mockResponses)-1]))
					}
					return
				}
				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			client := client.NewClient(client.ClientConfig{
				BaseURL: server.URL,
				Token:   "test-token",
			})

			resource := &ClusterResource{client: client}
			ctx := context.Background()

			err := resource.waitForClusterState(ctx, "test-cluster", tt.targetStates, tt.pollTimeout)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestClusterResource_waitForInstallationReadyAndTrigger(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  map[string][]string // URL path -> responses
		expectError    bool
		expectedError  string
	}{
		{
			name: "successful installation workflow",
			mockResponses: map[string][]string{
				"/v2/clusters/test-cluster": {
					`{"id": "test-cluster", "status": "insufficient", "status_info": "Waiting for hosts"}`,
					`{"id": "test-cluster", "status": "ready", "status_info": "Ready for installation"}`,
					`{"id": "test-cluster", "status": "installing", "status_info": "Installation in progress"}`,
					`{"id": "test-cluster", "status": "installed", "status_info": "Installation complete"}`,
				},
			},
			expectError: false,
		},
		{
			name: "installation trigger fails",
			mockResponses: map[string][]string{
				"/v2/clusters/test-cluster": {
					`{"id": "test-cluster", "status": "ready", "status_info": "Ready for installation"}`,
				},
			},
			expectError:   true,
			expectedError: "failed to trigger cluster installation",
		},
		{
			name: "cluster never becomes ready",
			mockResponses: map[string][]string{
				"/v2/clusters/test-cluster": {
					`{"id": "test-cluster", "status": "insufficient", "status_info": "Missing hosts"}`,
					`{"id": "test-cluster", "status": "error", "status_info": "Validation failed"}`,
				},
			},
			expectError:   true,
			expectedError: "cluster did not become ready for installation",
		},
		{
			name: "installation fails after trigger",
			mockResponses: map[string][]string{
				"/v2/clusters/test-cluster": {
					`{"id": "test-cluster", "status": "ready", "status_info": "Ready for installation"}`,
					`{"id": "test-cluster", "status": "installing", "status_info": "Installation in progress"}`,
					`{"id": "test-cluster", "status": "error", "status_info": "Installation failed"}`,
				},
			},
			expectError:   true,
			expectedError: "cluster installation did not complete successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responseCounters := make(map[string]int)
			installTriggered := false

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				path := r.URL.Path

				if r.Method == "GET" && path == "/v2/clusters/test-cluster" {
					responses, exists := tt.mockResponses[path]
					if !exists {
						w.WriteHeader(http.StatusNotFound)
						return
					}

					counter := responseCounters[path]
					if counter < len(responses) {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(responses[counter]))
						responseCounters[path]++
					} else {
						// Repeat last response
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(responses[len(responses)-1]))
					}
					return
				}

				if r.Method == "POST" && path == "/v2/clusters/test-cluster/actions/install" {
					installTriggered = true
					if tt.name == "installation trigger fails" {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(`{"error": "installation failed"}`))
						return
					}
					w.WriteHeader(http.StatusOK)
					return
				}

				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			client := client.NewClient(client.ClientConfig{
				BaseURL: server.URL,
				Token:   "test-token",
			})

			resource := &ClusterResource{client: client}
			ctx := context.Background()

			err := resource.waitForInstallationReadyAndTrigger(ctx, "test-cluster", 30*time.Second)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if !installTriggered {
					t.Errorf("Expected installation to be triggered but it wasn't")
				}
			}
		})
	}
}

func TestClusterResource_modelToCreateParams_NewFields(t *testing.T) {
	resource := &ClusterResource{}

	tests := []struct {
		name     string
		model    ClusterResourceModel
		expected models.ClusterCreateParams
	}{
		{
			name: "with new required fields",
			model: ClusterResourceModel{
				Name:             StringValue("test-cluster"),
				OpenshiftVersion: StringValue("4.15.20"),
				PullSecret:       StringValue("pull-secret"),
				CPUArchitecture:  StringValue("x86_64"),
				ControlPlaneCount: Int64Value(3),
			},
			expected: models.ClusterCreateParams{
				Name:              "test-cluster",
				OpenshiftVersion:  "4.15.20",
				PullSecret:        "pull-secret",
				CPUArchitecture:   "x86_64",
				ControlPlaneCount: 3,
			},
		},
		{
			name: "with SNO configuration",
			model: ClusterResourceModel{
				Name:              StringValue("sno-cluster"),
				OpenshiftVersion:  StringValue("4.15.20"),
				PullSecret:        StringValue("pull-secret"),
				CPUArchitecture:   StringValue("x86_64"),
				ControlPlaneCount: Int64Value(1),
			},
			expected: models.ClusterCreateParams{
				Name:              "sno-cluster",
				OpenshiftVersion:  "4.15.20",
				PullSecret:        "pull-secret",
				CPUArchitecture:   "x86_64",
				ControlPlaneCount: 1,
			},
		},
		{
			name: "with multi-arch",
			model: ClusterResourceModel{
				Name:              StringValue("multi-arch-cluster"),
				OpenshiftVersion:  StringValue("4.15.20"),
				PullSecret:        StringValue("pull-secret"),
				CPUArchitecture:   StringValue("multi"),
				ControlPlaneCount: Int64Value(3),
			},
			expected: models.ClusterCreateParams{
				Name:              "multi-arch-cluster",
				OpenshiftVersion:  "4.15.20",
				PullSecret:        "pull-secret",
				CPUArchitecture:   "multi",
				ControlPlaneCount: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resource.modelToCreateParams(tt.model)

			if result.Name != tt.expected.Name {
				t.Errorf("Expected name %q, got %q", tt.expected.Name, result.Name)
			}
			if result.CPUArchitecture != tt.expected.CPUArchitecture {
				t.Errorf("Expected cpu_architecture %q, got %q", tt.expected.CPUArchitecture, result.CPUArchitecture)
			}
			if result.ControlPlaneCount != tt.expected.ControlPlaneCount {
				t.Errorf("Expected control_plane_count %d, got %d", tt.expected.ControlPlaneCount, result.ControlPlaneCount)
			}
		})
	}
}

// Helper functions for creating test values
func StringValue(s string) types.String {
	return types.StringValue(s)
}

func Int64Value(i int64) types.Int64 {
	return types.Int64Value(i)
}