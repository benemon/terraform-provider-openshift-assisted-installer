package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
)

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
				Name:              StringValue("test-cluster"),
				OpenshiftVersion:  StringValue("4.15.20"),
				PullSecret:        StringValue("pull-secret"),
				CPUArchitecture:   StringValue("x86_64"),
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
