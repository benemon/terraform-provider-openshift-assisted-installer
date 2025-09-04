package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ClusterValidationsDataSource{}

func NewClusterValidationsDataSource() datasource.DataSource {
	return &ClusterValidationsDataSource{}
}

// ClusterValidationsDataSource defines the data source implementation.
type ClusterValidationsDataSource struct {
	client *client.Client
}

// ClusterValidationModel represents a single validation result.
type ClusterValidationModel struct {
	ID              types.String `tfsdk:"id"`
	Status          types.String `tfsdk:"status"`
	Message         types.String `tfsdk:"message"`
	ValidationID    types.String `tfsdk:"validation_id"`
	ValidationName  types.String `tfsdk:"validation_name"`
	ValidationGroup types.String `tfsdk:"validation_group"`
	ValidationType  types.String `tfsdk:"validation_type"`
	Category        types.String `tfsdk:"category"`
}

// ClusterValidationsDataSourceModel describes the data source data model.
type ClusterValidationsDataSourceModel struct {
	ID               types.String                 `tfsdk:"id"`
	ClusterID        types.String                 `tfsdk:"cluster_id"`
	ValidationTypes  []types.String               `tfsdk:"validation_types"`
	StatusFilter     []types.String               `tfsdk:"status_filter"`
	ValidationNames  []types.String               `tfsdk:"validation_names"`
	Categories       []types.String               `tfsdk:"categories"`
	Validations      []ClusterValidationModel     `tfsdk:"validations"`
}

func (d *ClusterValidationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_validations"
}

func (d *ClusterValidationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Retrieves cluster-level validation information from the OpenShift Assisted Installer API. This data source provides pre-installation validation results that can be used to determine cluster readiness and troubleshoot configuration issues.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for this data source instance",
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the cluster to retrieve validations for",
				Required:            true,
			},
			"validation_types": schema.ListAttribute{
				MarkdownDescription: "Filter validations by type: 'blocking' or 'non-blocking'. If not specified, all validations are returned.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"status_filter": schema.ListAttribute{
				MarkdownDescription: "Filter validations by status: 'success', 'failure', 'pending', 'disabled'. If not specified, all statuses are returned.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"validation_names": schema.ListAttribute{
				MarkdownDescription: "Filter validations by specific validation IDs. If not specified, all validations are returned.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"categories": schema.ListAttribute{
				MarkdownDescription: "Filter validations by category: 'network', 'hardware', 'operators', 'cluster', 'platform', 'storage'. If not specified, all categories are returned.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"validations": schema.ListNestedAttribute{
				MarkdownDescription: "List of cluster validation results matching the filter criteria",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Validation identifier",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "Validation status: success, failure, pending, disabled",
							Computed:            true,
						},
						"message": schema.StringAttribute{
							MarkdownDescription: "Human-readable validation message",
							Computed:            true,
						},
						"validation_id": schema.StringAttribute{
							MarkdownDescription: "Specific validation identifier (if available)",
							Computed:            true,
						},
						"validation_name": schema.StringAttribute{
							MarkdownDescription: "Human-readable validation name (if available)",
							Computed:            true,
						},
						"validation_group": schema.StringAttribute{
							MarkdownDescription: "Validation group (if available)",
							Computed:            true,
						},
						"validation_type": schema.StringAttribute{
							MarkdownDescription: "Whether this validation is blocking or non-blocking",
							Computed:            true,
						},
						"category": schema.StringAttribute{
							MarkdownDescription: "Validation category: network, hardware, operators, cluster, platform, storage",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *ClusterValidationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ClusterValidationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterValidationsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get cluster validations from API
	clusterValidations, err := d.client.GetClusterValidations(ctx, data.ClusterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read cluster validations, got error: %s", err),
		)
		return
	}

	// Convert validation types filter to strings for comparison
	var validationTypesFilter []string
	if len(data.ValidationTypes) > 0 {
		for _, vt := range data.ValidationTypes {
			if !vt.IsNull() && !vt.IsUnknown() {
				validationTypesFilter = append(validationTypesFilter, vt.ValueString())
			}
		}
	}

	// Convert status filter to strings for comparison
	var statusFilter []string
	if len(data.StatusFilter) > 0 {
		for _, status := range data.StatusFilter {
			if !status.IsNull() && !status.IsUnknown() {
				statusFilter = append(statusFilter, status.ValueString())
			}
		}
	}

	// Convert validation names filter to strings for comparison
	var validationNamesFilter []string
	if len(data.ValidationNames) > 0 {
		for _, name := range data.ValidationNames {
			if !name.IsNull() && !name.IsUnknown() {
				validationNamesFilter = append(validationNamesFilter, name.ValueString())
			}
		}
	}

	// Convert categories filter to strings for comparison
	var categoriesFilter []string
	if len(data.Categories) > 0 {
		for _, category := range data.Categories {
			if !category.IsNull() && !category.IsUnknown() {
				categoriesFilter = append(categoriesFilter, category.ValueString())
			}
		}
	}

	// Process validations and apply filters
	var filteredValidations []ClusterValidationModel
	for groupName, validationsGroup := range clusterValidations.ValidationsInfo {
		for _, validation := range validationsGroup {
			// Determine validation type (blocking/non-blocking)
			validationType := "non-blocking"
			validationID := validation.ValidationID
			if validationID == "" {
				validationID = validation.ID
			}
			if models.IsBlockingValidation(validationID) {
				validationType = "blocking"
			}

			// Apply validation type filter
			if len(validationTypesFilter) > 0 {
				found := false
				for _, filterType := range validationTypesFilter {
					if strings.EqualFold(validationType, filterType) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Apply status filter
			if len(statusFilter) > 0 {
				found := false
				for _, filterStatus := range statusFilter {
					if strings.EqualFold(validation.Status, filterStatus) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Apply validation names filter
			if len(validationNamesFilter) > 0 {
				found := false
				for _, filterName := range validationNamesFilter {
					if validationID == filterName || validation.ID == filterName {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Apply categories filter
			if len(categoriesFilter) > 0 {
				category := string(models.GetValidationCategory(validationID))
				found := false
				for _, filterCategory := range categoriesFilter {
					if strings.EqualFold(category, filterCategory) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Build validation model
			validationModel := ClusterValidationModel{
				ID:              types.StringValue(validation.ID),
				Status:          types.StringValue(validation.Status),
				Message:         types.StringValue(validation.Message),
				ValidationID:    types.StringValue(validationID),
				ValidationName:  types.StringValue(validation.ValidationName),
				ValidationGroup: types.StringValue(groupName),
				ValidationType:  types.StringValue(validationType),
				Category:        types.StringValue(string(models.GetValidationCategory(validationID))),
			}

			filteredValidations = append(filteredValidations, validationModel)
		}
	}

	// Update the model with filtered results
	data.ID = types.StringValue(fmt.Sprintf("cluster-validations-%s", data.ClusterID.ValueString()))
	data.Validations = filteredValidations

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}