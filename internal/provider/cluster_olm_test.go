package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
)

func TestClusterResource_OLMOperators_modelToCreateParams(t *testing.T) {
	resource := &ClusterResource{}

	tests := []struct {
		name     string
		model    ClusterResourceModel
		expected []models.OLMOperator
	}{
		{
			name: "no operators",
			model: ClusterResourceModel{
				Name:             StringValue("test-cluster"),
				OpenshiftVersion: StringValue("4.15.20"),
				PullSecret:       StringValue("pull-secret"),
				CPUArchitecture:  StringValue("x86_64"),
				OLMOperators:     types.ListNull(types.ObjectType{}),
			},
			expected: nil,
		},
		{
			name: "single operator without properties",
			model: ClusterResourceModel{
				Name:             StringValue("test-cluster"),
				OpenshiftVersion: StringValue("4.15.20"),
				PullSecret:       StringValue("pull-secret"),
				CPUArchitecture:  StringValue("x86_64"),
				OLMOperators:     createOLMOperatorsList([]OLMOperatorModel{
					{Name: StringValue("local-storage-operator"), Properties: types.StringNull()},
				}),
			},
			expected: []models.OLMOperator{
				{Name: "local-storage-operator", Properties: ""},
			},
		},
		{
			name: "multiple operators with properties",
			model: ClusterResourceModel{
				Name:             StringValue("test-cluster"),
				OpenshiftVersion: StringValue("4.15.20"),
				PullSecret:       StringValue("pull-secret"),
				CPUArchitecture:  StringValue("x86_64"),
				OLMOperators:     createOLMOperatorsList([]OLMOperatorModel{
					{Name: StringValue("local-storage-operator"), Properties: StringValue("{\"version\":\"4.15\"}")},
					{Name: StringValue("odf-operator"), Properties: StringValue("{\"version\":\"4.15\",\"namespace\":\"openshift-storage\"}")},
				}),
			},
			expected: []models.OLMOperator{
				{Name: "local-storage-operator", Properties: "{\"version\":\"4.15\"}"},
				{Name: "odf-operator", Properties: "{\"version\":\"4.15\",\"namespace\":\"openshift-storage\"}"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resource.modelToCreateParams(tt.model)

			if len(result.OLMOperators) != len(tt.expected) {
				t.Errorf("Expected %d operators, got %d", len(tt.expected), len(result.OLMOperators))
				return
			}

			for i, expected := range tt.expected {
				if result.OLMOperators[i].Name != expected.Name {
					t.Errorf("Operator %d: expected name %q, got %q", i, expected.Name, result.OLMOperators[i].Name)
				}
				if result.OLMOperators[i].Properties != expected.Properties {
					t.Errorf("Operator %d: expected properties %q, got %q", i, expected.Properties, result.OLMOperators[i].Properties)
				}
			}
		})
	}
}

func TestClusterResource_OLMOperators_updateModelFromCluster(t *testing.T) {
	resource := &ClusterResource{}

	tests := []struct {
		name     string
		cluster  *models.Cluster
		expected []OLMOperatorModel
	}{
		{
			name: "no operators",
			cluster: &models.Cluster{
				ID:           "test-cluster",
				Name:         "test-cluster",
				Status:       "ready",
				OLMOperators: []models.OLMOperator{},
			},
			expected: []OLMOperatorModel{},
		},
		{
			name: "single operator",
			cluster: &models.Cluster{
				ID:     "test-cluster",
				Name:   "test-cluster",
				Status: "ready",
				OLMOperators: []models.OLMOperator{
					{Name: "local-storage-operator", Properties: "{\"version\":\"4.15\"}"},
				},
			},
			expected: []OLMOperatorModel{
				{Name: StringValue("local-storage-operator"), Properties: StringValue("{\"version\":\"4.15\"}")},
			},
		},
		{
			name: "multiple operators",
			cluster: &models.Cluster{
				ID:     "test-cluster",
				Name:   "test-cluster",
				Status: "ready",
				OLMOperators: []models.OLMOperator{
					{Name: "local-storage-operator", Properties: "{\"version\":\"4.15\"}"},
					{Name: "odf-operator", Properties: "{\"version\":\"4.15\",\"namespace\":\"openshift-storage\"}"},
					{Name: "cnv-operator", Properties: ""},
				},
			},
			expected: []OLMOperatorModel{
				{Name: StringValue("local-storage-operator"), Properties: StringValue("{\"version\":\"4.15\"}")},
				{Name: StringValue("odf-operator"), Properties: StringValue("{\"version\":\"4.15\",\"namespace\":\"openshift-storage\"}")},
				{Name: StringValue("cnv-operator"), Properties: StringValue("")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data ClusterResourceModel
			resource.updateModelFromCluster(&data, tt.cluster)

			if len(tt.expected) == 0 {
				// If no operators expected, the list should be null or empty
				if !data.OLMOperators.IsNull() && !data.OLMOperators.IsUnknown() {
					elements := data.OLMOperators.Elements()
					if len(elements) > 0 {
						t.Errorf("Expected no operators, but got %d", len(elements))
					}
				}
				return
			}

			// Extract operators from the model
			if data.OLMOperators.IsNull() || data.OLMOperators.IsUnknown() {
				t.Errorf("Expected operators to be set, but got null/unknown")
				return
			}

			var actualOperators []OLMOperatorModel
			diags := data.OLMOperators.ElementsAs(context.Background(), &actualOperators, false)
			if diags.HasError() {
				t.Errorf("Failed to extract operators from model: %v", diags)
				return
			}

			if len(actualOperators) != len(tt.expected) {
				t.Errorf("Expected %d operators, got %d", len(tt.expected), len(actualOperators))
				return
			}

			for i, expected := range tt.expected {
				if !actualOperators[i].Name.Equal(expected.Name) {
					t.Errorf("Operator %d: expected name %q, got %q", i, expected.Name.ValueString(), actualOperators[i].Name.ValueString())
				}
				if !actualOperators[i].Properties.Equal(expected.Properties) {
					t.Errorf("Operator %d: expected properties %q, got %q", i, expected.Properties.ValueString(), actualOperators[i].Properties.ValueString())
				}
			}
		})
	}
}

func TestClusterResource_ModelToCreateParams_Basic(t *testing.T) {
	// Test basic model to create params conversion
	resource := &ClusterResource{}

	model := ClusterResourceModel{
		Name:             StringValue("test-cluster"),
		OpenshiftVersion: StringValue("4.15.20"),
		PullSecret:       StringValue("pull-secret"),
		CPUArchitecture:  StringValue("x86_64"),
	}

	params := resource.modelToCreateParams(model)

	// Verify basic fields are set correctly
	if params.Name != "test-cluster" {
		t.Errorf("Expected name 'test-cluster', got %q", params.Name)
	}
	if params.OpenshiftVersion != "4.15.20" {
		t.Errorf("Expected version '4.15.20', got %q", params.OpenshiftVersion)
	}
	if params.CPUArchitecture != "x86_64" {
		t.Errorf("Expected cpu_architecture to be 'x86_64', got %q", params.CPUArchitecture)
	}
}

// Helper function to create OLM operators list for testing
func createOLMOperatorsList(operators []OLMOperatorModel) types.List {
	if len(operators) == 0 {
		return types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":       types.StringType,
				"properties": types.StringType,
			},
		})
	}

	listValue, _ := types.ListValueFrom(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":       types.StringType,
			"properties": types.StringType,
		},
	}, operators)
	
	return listValue
}