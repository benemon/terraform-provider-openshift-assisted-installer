package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &HostResource{}
var _ resource.ResourceWithImportState = &HostResource{}

func NewHostResource() resource.Resource {
	return &HostResource{}
}

// HostResource defines the resource implementation.
type HostResource struct {
	client *client.Client
}

// HostResourceModel describes the resource data model.
type HostResourceModel struct {
	ID                types.String         `tfsdk:"id"`
	InfraEnvID        types.String         `tfsdk:"infra_env_id"`
	ClusterID         types.String         `tfsdk:"cluster_id"`
	RequestedHostname types.String         `tfsdk:"requested_hostname"`
	Role              types.String         `tfsdk:"role"`
	
	// Computed fields
	Status            types.String         `tfsdk:"status"`
	StatusInfo        types.String         `tfsdk:"status_info"`
	Progress          *HostProgressModel   `tfsdk:"progress"`
	CreatedAt         types.String         `tfsdk:"created_at"`
	UpdatedAt         types.String         `tfsdk:"updated_at"`
}

type HostProgressModel struct {
	CurrentStage   types.String `tfsdk:"current_stage"`
	ProgressInfo   types.String `tfsdk:"progress_info"`
	StageStartedAt types.String `tfsdk:"stage_started_at"`
	StageUpdatedAt types.String `tfsdk:"stage_updated_at"`
}

func (r *HostResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (r *HostResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Host resource for managing OpenShift cluster hosts discovered through an infrastructure environment.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Host identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"infra_env_id": schema.StringAttribute{
				MarkdownDescription: "Infrastructure environment ID where this host was discovered.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "Cluster ID to bind this host to. If not specified, host remains unbound.",
				Optional:            true,
			},
			"requested_hostname": schema.StringAttribute{
				MarkdownDescription: "Requested hostname for this host. If not specified, a default will be assigned.",
				Optional:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role assignment for this host in the cluster.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("auto-assign"),
				Validators: []validator.String{
					stringvalidator.OneOf("master", "worker", "bootstrap", "auto-assign"),
				},
			},
			
			// Computed attributes
			"status": schema.StringAttribute{
				MarkdownDescription: "Current status of the host.",
				Computed:            true,
			},
			"status_info": schema.StringAttribute{
				MarkdownDescription: "Detailed status information for the host.",
				Computed:            true,
			},
			"progress": schema.SingleNestedAttribute{
				MarkdownDescription: "Installation progress information.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"current_stage": schema.StringAttribute{
						MarkdownDescription: "Current installation stage.",
						Computed:            true,
					},
					"progress_info": schema.StringAttribute{
						MarkdownDescription: "Detailed progress information.",
						Computed:            true,
					},
					"stage_started_at": schema.StringAttribute{
						MarkdownDescription: "When the current stage started.",
						Computed:            true,
					},
					"stage_updated_at": schema.StringAttribute{
						MarkdownDescription: "When the current stage was last updated.",
						Computed:            true,
					},
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "When the host was first discovered.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "When the host was last updated.",
				Computed:            true,
			},
		},
	}
}

func (r *HostResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *HostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HostResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Note: Hosts are discovered automatically when they boot from the ISO
	// The "create" operation here is really about configuring an existing discovered host
	// We first need to wait for a host to be discovered in the specified infra-env
	
	// For now, we require the host ID to be provided as an import operation
	// A full implementation would include polling for discovered hosts
	
	if data.ID.IsNull() || data.ID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Host ID Required",
			"Host resources cannot be created directly. Hosts are discovered when they boot from the infrastructure environment ISO. Use 'terraform import' to manage existing discovered hosts.",
		)
		return
	}

	// Get the host to verify it exists
	host, err := r.client.GetHost(ctx, data.InfraEnvID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading host", fmt.Sprintf("Could not read host %s: %s", data.ID.ValueString(), err))
		return
	}

	// Configure the host based on the plan
	if err := r.configureHost(ctx, &data, host); err != nil {
		resp.Diagnostics.AddError("Error configuring host", fmt.Sprintf("Could not configure host %s: %s", data.ID.ValueString(), err))
		return
	}

	// Read the updated host state
	updatedHost, err := r.client.GetHost(ctx, data.InfraEnvID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated host", fmt.Sprintf("Could not read updated host %s: %s", data.ID.ValueString(), err))
		return
	}

	// Update model with response data
	r.apiToTerraformModel(ctx, updatedHost, &data)

	tflog.Info(ctx, "Successfully configured host", map[string]any{
		"host_id":      data.ID.ValueString(),
		"infra_env_id": data.InfraEnvID.ValueString(),
		"cluster_id":   data.ClusterID.ValueString(),
		"role":         data.Role.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HostResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the host from the API
	host, err := r.client.GetHost(ctx, data.InfraEnvID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading host", fmt.Sprintf("Could not read host %s: %s", data.ID.ValueString(), err))
		return
	}

	// Update model with current API state
	r.apiToTerraformModel(ctx, host, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data HostResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get current host state
	host, err := r.client.GetHost(ctx, data.InfraEnvID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading host", fmt.Sprintf("Could not read host %s: %s", data.ID.ValueString(), err))
		return
	}

	// Configure the host based on the updated plan
	if err := r.configureHost(ctx, &data, host); err != nil {
		resp.Diagnostics.AddError("Error updating host", fmt.Sprintf("Could not update host %s: %s", data.ID.ValueString(), err))
		return
	}

	// Read the updated host state
	updatedHost, err := r.client.GetHost(ctx, data.InfraEnvID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated host", fmt.Sprintf("Could not read updated host %s: %s", data.ID.ValueString(), err))
		return
	}

	// Update model with response data
	r.apiToTerraformModel(ctx, updatedHost, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HostResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Unbind the host from its cluster if bound
	if !data.ClusterID.IsNull() && data.ClusterID.ValueString() != "" {
		tflog.Info(ctx, "Unbinding host from cluster", map[string]any{
			"host_id":      data.ID.ValueString(),
			"infra_env_id": data.InfraEnvID.ValueString(),
			"cluster_id":   data.ClusterID.ValueString(),
		})

		err := r.client.UnbindHost(ctx, data.InfraEnvID.ValueString(), data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error unbinding host", fmt.Sprintf("Could not unbind host %s from cluster: %s", data.ID.ValueString(), err))
			return
		}
	}

	tflog.Info(ctx, "Successfully unbound host", map[string]any{
		"host_id":      data.ID.ValueString(),
		"infra_env_id": data.InfraEnvID.ValueString(),
	})

	// Note: We don't actually delete the host from the infrastructure environment
	// as it represents a physical/virtual machine that may still be running
	// We just unbind it from any cluster association
}

func (r *HostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import state expects "infra_env_id/host_id" format
	idParts := len(req.ID)
	if idParts == 0 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: infra_env_id/host_id. Got: %q", req.ID),
		)
		return
	}

	// For simplicity, assume the ID is just the host ID and require infra_env_id to be set in config
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// Helper functions

func (r *HostResource) configureHost(ctx context.Context, data *HostResourceModel, currentHost *models.Host) error {
	var needsUpdate bool
	updateParams := models.HostUpdateParams{}

	// Check if hostname needs updating
	if !data.RequestedHostname.IsNull() {
		hostname := data.RequestedHostname.ValueString()
		if currentHost.RequestedHostname != hostname {
			updateParams.RequestedHostname = &hostname
			needsUpdate = true
		}
	}

	// Check if role needs updating
	if !data.Role.IsNull() {
		role := data.Role.ValueString()
		if currentHost.Role != role {
			updateParams.Role = &role
			needsUpdate = true
		}
	}

	// Update host configuration if needed
	if needsUpdate {
		tflog.Info(ctx, "Updating host configuration", map[string]any{
			"host_id":            data.ID.ValueString(),
			"requested_hostname": updateParams.RequestedHostname,
			"role":               updateParams.Role,
		})

		_, err := r.client.UpdateHost(ctx, data.InfraEnvID.ValueString(), data.ID.ValueString(), updateParams)
		if err != nil {
			return fmt.Errorf("failed to update host configuration: %w", err)
		}
	}

	// Handle cluster binding/unbinding
	desiredClusterID := ""
	if !data.ClusterID.IsNull() {
		desiredClusterID = data.ClusterID.ValueString()
	}

	if currentHost.ClusterID != desiredClusterID {
		if desiredClusterID != "" {
			// Bind host to cluster
			tflog.Info(ctx, "Binding host to cluster", map[string]any{
				"host_id":    data.ID.ValueString(),
				"cluster_id": desiredClusterID,
			})

			bindParams := models.BindHostParams{
				ClusterID: desiredClusterID,
			}

			err := r.client.BindHost(ctx, data.InfraEnvID.ValueString(), data.ID.ValueString(), bindParams)
			if err != nil {
				return fmt.Errorf("failed to bind host to cluster: %w", err)
			}
		} else if currentHost.ClusterID != "" {
			// Unbind host from cluster
			tflog.Info(ctx, "Unbinding host from cluster", map[string]any{
				"host_id":         data.ID.ValueString(),
				"current_cluster": currentHost.ClusterID,
			})

			err := r.client.UnbindHost(ctx, data.InfraEnvID.ValueString(), data.ID.ValueString())
			if err != nil {
				return fmt.Errorf("failed to unbind host from cluster: %w", err)
			}
		}
	}

	return nil
}

func (r *HostResource) apiToTerraformModel(ctx context.Context, host *models.Host, data *HostResourceModel) {
	data.ID = types.StringValue(host.ID)
	data.InfraEnvID = types.StringValue(host.InfraEnvID)
	data.Status = types.StringValue(host.Status)
	data.StatusInfo = types.StringValue(host.StatusInfo)
	
	if host.ClusterID != "" {
		data.ClusterID = types.StringValue(host.ClusterID)
	} else {
		data.ClusterID = types.StringNull()
	}
	
	if host.RequestedHostname != "" {
		data.RequestedHostname = types.StringValue(host.RequestedHostname)
	} else {
		data.RequestedHostname = types.StringNull()
	}
	
	if host.Role != "" {
		data.Role = types.StringValue(host.Role)
	} else {
		data.Role = types.StringValue("auto-assign")
	}
	
	// Convert progress information
	if host.Progress != nil {
		data.Progress = &HostProgressModel{
			CurrentStage: types.StringValue(host.Progress.CurrentStage),
			ProgressInfo: types.StringValue(host.Progress.ProgressInfo),
		}
		
		if !host.Progress.StageStartedAt.IsZero() {
			data.Progress.StageStartedAt = types.StringValue(host.Progress.StageStartedAt.Format("2006-01-02T15:04:05Z"))
		}
		
		if !host.Progress.StageUpdatedAt.IsZero() {
			data.Progress.StageUpdatedAt = types.StringValue(host.Progress.StageUpdatedAt.Format("2006-01-02T15:04:05Z"))
		}
	}
	
	if !host.CreatedAt.IsZero() {
		data.CreatedAt = types.StringValue(host.CreatedAt.Format("2006-01-02T15:04:05Z"))
	}
	
	if !host.UpdatedAt.IsZero() {
		data.UpdatedAt = types.StringValue(host.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	}
}