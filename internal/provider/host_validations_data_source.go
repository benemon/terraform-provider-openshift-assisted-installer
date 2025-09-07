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
var _ datasource.DataSource = &HostValidationsDataSource{}

func NewHostValidationsDataSource() datasource.DataSource {
	return &HostValidationsDataSource{}
}

// HostValidationsDataSource defines the data source implementation.
type HostValidationsDataSource struct {
	client *client.Client
}

// HostValidationModel represents a single host validation result.
type HostValidationModel struct {
	ID              types.String `tfsdk:"id"`
	HostID          types.String `tfsdk:"host_id"`
	Status          types.String `tfsdk:"status"`
	Message         types.String `tfsdk:"message"`
	ValidationID    types.String `tfsdk:"validation_id"`
	ValidationName  types.String `tfsdk:"validation_name"`
	ValidationGroup types.String `tfsdk:"validation_group"`
	ValidationType  types.String `tfsdk:"validation_type"`
	Category        types.String `tfsdk:"category"`
}

// HostValidationsDataSourceModel describes the data source data model.
type HostValidationsDataSourceModel struct {
	ID              types.String          `tfsdk:"id"`
	ClusterID       types.String          `tfsdk:"cluster_id"`
	HostID          types.String          `tfsdk:"host_id"`
	InfraEnvID      types.String          `tfsdk:"infra_env_id"`
	ValidationTypes []types.String        `tfsdk:"validation_types"`
	StatusFilter    []types.String        `tfsdk:"status_filter"`
	ValidationNames []types.String        `tfsdk:"validation_names"`
	Categories      []types.String        `tfsdk:"categories"`
	Validations     []HostValidationModel `tfsdk:"validations"`
}

func (d *HostValidationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host_validations"
}

func (d *HostValidationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Retrieves host-level validation information from the OpenShift Assisted Installer API. This data source provides pre-installation validation results for individual hosts that can be used to determine host readiness and troubleshoot hardware, network, or configuration issues.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for this data source instance",
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the cluster to retrieve host validations for (when checking all hosts in a cluster)",
				Optional:            true,
			},
			"host_id": schema.StringAttribute{
				MarkdownDescription: "The ID of a specific host to retrieve validations for (requires infra_env_id)",
				Optional:            true,
			},
			"infra_env_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the infrastructure environment (required when host_id is specified)",
				Optional:            true,
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
				MarkdownDescription: "List of host validation results matching the filter criteria",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Validation identifier",
							Computed:            true,
						},
						"host_id": schema.StringAttribute{
							MarkdownDescription: "ID of the host this validation applies to",
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

func (d *HostValidationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *HostValidationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HostValidationsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate configuration - must specify either cluster_id or (infra_env_id + host_id)
	hasClusterID := !data.ClusterID.IsNull() && !data.ClusterID.IsUnknown() && data.ClusterID.ValueString() != ""
	hasHostID := !data.HostID.IsNull() && !data.HostID.IsUnknown() && data.HostID.ValueString() != ""
	hasInfraEnvID := !data.InfraEnvID.IsNull() && !data.InfraEnvID.IsUnknown() && data.InfraEnvID.ValueString() != ""

	if !hasClusterID && (!hasHostID || !hasInfraEnvID) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Must specify either 'cluster_id' to get validations for all hosts in a cluster, or both 'infra_env_id' and 'host_id' to get validations for a specific host",
		)
		return
	}

	if hasClusterID && (hasHostID || hasInfraEnvID) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Cannot specify both 'cluster_id' and ('infra_env_id' or 'host_id'). Use 'cluster_id' for all hosts in a cluster, or 'infra_env_id' + 'host_id' for a specific host",
		)
		return
	}

	var hostValidations *models.HostsValidationResponse
	var singleHostValidation *models.HostValidationResponse
	var err error

	if hasClusterID {
		// Get validations for all hosts in cluster
		hostValidations, err = d.client.GetHostValidations(ctx, data.ClusterID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read host validations for cluster %s, got error: %s", data.ClusterID.ValueString(), err),
			)
			return
		}
	} else {
		// Get validations for a specific host
		singleHostValidation, err = d.client.GetSingleHostValidations(ctx, data.InfraEnvID.ValueString(), data.HostID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read host validations for host %s in infra-env %s, got error: %s", data.HostID.ValueString(), data.InfraEnvID.ValueString(), err),
			)
			return
		}

		// Convert single host validation to hosts list format
		hostValidations = &models.HostsValidationResponse{
			Hosts: []models.HostValidationResponse{*singleHostValidation},
		}
	}

	// Convert filter arrays to string slices for comparison
	var validationTypesFilter []string
	if len(data.ValidationTypes) > 0 {
		for _, vt := range data.ValidationTypes {
			if !vt.IsNull() && !vt.IsUnknown() {
				validationTypesFilter = append(validationTypesFilter, vt.ValueString())
			}
		}
	}

	var statusFilter []string
	if len(data.StatusFilter) > 0 {
		for _, status := range data.StatusFilter {
			if !status.IsNull() && !status.IsUnknown() {
				statusFilter = append(statusFilter, status.ValueString())
			}
		}
	}

	var validationNamesFilter []string
	if len(data.ValidationNames) > 0 {
		for _, name := range data.ValidationNames {
			if !name.IsNull() && !name.IsUnknown() {
				validationNamesFilter = append(validationNamesFilter, name.ValueString())
			}
		}
	}

	var categoriesFilter []string
	if len(data.Categories) > 0 {
		for _, category := range data.Categories {
			if !category.IsNull() && !category.IsUnknown() {
				categoriesFilter = append(categoriesFilter, category.ValueString())
			}
		}
	}

	// Process host validations and apply filters
	var filteredValidations []HostValidationModel
	for _, host := range hostValidations.Hosts {
		for groupName, validationsGroup := range host.ValidationsInfo {
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
				validationModel := HostValidationModel{
					ID:              types.StringValue(validation.ID),
					HostID:          types.StringValue(host.ID),
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
	}

	// Update the model with filtered results
	if hasClusterID {
		data.ID = types.StringValue(fmt.Sprintf("host-validations-cluster-%s", data.ClusterID.ValueString()))
	} else {
		data.ID = types.StringValue(fmt.Sprintf("host-validations-host-%s", data.HostID.ValueString()))
	}
	data.Validations = filteredValidations

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
