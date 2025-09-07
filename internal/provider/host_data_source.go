package provider

import (
	"context"
	"fmt"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &HostDataSource{}

func NewHostDataSource() datasource.DataSource {
	return &HostDataSource{}
}

// HostDataSource defines the data source implementation.
type HostDataSource struct {
	client *client.Client
}

// All inventory-related fields are JSON strings per Swagger spec, not parsed objects

// HostDataSourceModel describes the data source data model.
// All fields match exactly with Swagger host definition
type HostDataSourceModel struct {
	// Required fields per Swagger
	ID         types.String `tfsdk:"id"`
	Kind       types.String `tfsdk:"kind"`
	Href       types.String `tfsdk:"href"`
	Status     types.String `tfsdk:"status"`
	StatusInfo types.String `tfsdk:"status_info"`

	// Associations
	InfraEnvID types.String `tfsdk:"infra_env_id"`
	ClusterID  types.String `tfsdk:"cluster_id"`

	// Status and progress
	StatusUpdatedAt types.String `tfsdk:"status_updated_at"`
	Progress        types.Object `tfsdk:"progress"`
	StageStartedAt  types.String `tfsdk:"stage_started_at"`
	StageUpdatedAt  types.String `tfsdk:"stage_updated_at"`
	ProgressStages  types.List   `tfsdk:"progress_stages"`

	// Connectivity (JSON strings per Swagger)
	Connectivity       types.String `tfsdk:"connectivity"`
	APIVipConnectivity types.String `tfsdk:"api_vip_connectivity"`
	TangConnectivity   types.String `tfsdk:"tang_connectivity"`

	// Hardware inventory (JSON string per Swagger)
	Inventory     types.String `tfsdk:"inventory"`
	FreeAddresses types.String `tfsdk:"free_addresses"`
	NTPSources    types.String `tfsdk:"ntp_sources"`
	DisksInfo     types.String `tfsdk:"disks_info"`

	// Host role and configuration
	Role          types.String `tfsdk:"role"`
	SuggestedRole types.String `tfsdk:"suggested_role"`
	Bootstrap     types.Bool   `tfsdk:"bootstrap"`

	// Logs
	LogsInfo        types.Object `tfsdk:"logs_info"`
	LogsCollectedAt types.String `tfsdk:"logs_collected_at"`
	LogsStartedAt   types.String `tfsdk:"logs_started_at"`

	// Installation
	InstallerVersion     types.String `tfsdk:"installer_version"`
	InstallationDiskPath types.String `tfsdk:"installation_disk_path"`
	InstallationDiskID   types.String `tfsdk:"installation_disk_id"`

	// Timestamps
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	CheckedInAt  types.String `tfsdk:"checked_in_at"`
	RegisteredAt types.String `tfsdk:"registered_at"`

	// Host details
	UserName              types.String `tfsdk:"user_name"`
	RequestedHostname     types.String `tfsdk:"requested_hostname"`
	ConnectionTimedOut    types.Bool   `tfsdk:"connection_timed_out"`
	DiscoveryAgentVersion types.String `tfsdk:"discovery_agent_version"`
	MediaStatus           types.String `tfsdk:"media_status"`

	// Advanced configuration (JSON strings per Swagger)
	IgnitionConfigOverrides  types.String `tfsdk:"ignition_config_overrides"`
	InstallerArgs            types.String `tfsdk:"installer_args"`
	Timestamp                types.Int64  `tfsdk:"timestamp"`
	MachineConfigPoolName    types.String `tfsdk:"machine_config_pool_name"`
	ImagesStatus             types.String `tfsdk:"images_status"`
	DomainNameResolutions    types.String `tfsdk:"domain_name_resolutions"`
	IgnitionEndpointTokenSet types.Bool   `tfsdk:"ignition_endpoint_token_set"`
	NodeLabels               types.String `tfsdk:"node_labels"`
	DisksToBeFormatted       types.String `tfsdk:"disks_to_be_formatted"`
	SkipFormattingDisks      types.String `tfsdk:"skip_formatting_disks"`

	// Validation info (JSON string per Swagger)
	ValidationsInfo types.String `tfsdk:"validations_info"`
}

func (d *HostDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (d *HostDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Retrieves information about a host discovered through an infrastructure environment. Hosts are physical or virtual machines that boot from the discovery ISO and can be assigned roles in an OpenShift cluster.",

		Attributes: map[string]schema.Attribute{
			// Required fields per Swagger
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the host (UUID)",
				Required:            true,
			},
			"infra_env_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the infrastructure environment this host was discovered in",
				Required:            true,
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "Indicates the type of this object ('Host', 'AddToExistingClusterHost')",
				Computed:            true,
			},
			"href": schema.StringAttribute{
				MarkdownDescription: "Self link to this host",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current host status (discovering, known, disconnected, insufficient, disabled, preparing-for-installation, preparing-failed, preparing-successful, pending-for-input, installing, installing-in-progress, installing-pending-user-action, resetting-pending-user-action, installed, error, resetting, added-to-existing-cluster, cancelled, binding, unbinding, unbinding-pending-user-action, known-unbound, disconnected-unbound, insufficient-unbound, disabled-unbound, discovering-unbound, reclaiming, reclaiming-rebooting)",
				Computed:            true,
			},
			"status_info": schema.StringAttribute{
				MarkdownDescription: "Additional information about the host status",
				Computed:            true,
			},

			// Associations
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The cluster that this host is associated with",
				Computed:            true,
			},

			// Status and progress - matching Swagger exactly
			"validations_info": schema.StringAttribute{
				MarkdownDescription: "JSON-formatted string containing the validation results for each validation id grouped by category (network, hardware, etc.)",
				Computed:            true,
			},
			"status_updated_at": schema.StringAttribute{
				MarkdownDescription: "The last time that the host status was updated",
				Computed:            true,
			},
			"progress": schema.SingleNestedAttribute{
				MarkdownDescription: "Installation progress information",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"current_stage": schema.StringAttribute{
						MarkdownDescription: "Current installation stage",
						Computed:            true,
					},
					"progress_info": schema.StringAttribute{
						MarkdownDescription: "Detailed progress information",
						Computed:            true,
					},
					"installation_percentage": schema.Int64Attribute{
						MarkdownDescription: "Installation completion percentage",
						Computed:            true,
					},
					"stage_started_at": schema.StringAttribute{
						MarkdownDescription: "When the current stage started",
						Computed:            true,
					},
					"stage_updated_at": schema.StringAttribute{
						MarkdownDescription: "When the current stage was last updated",
						Computed:            true,
					},
				},
			},
			"stage_started_at": schema.StringAttribute{
				MarkdownDescription: "Time at which the current progress stage started",
				Computed:            true,
			},
			"stage_updated_at": schema.StringAttribute{
				MarkdownDescription: "Time at which the current progress stage was last updated",
				Computed:            true,
			},
			"progress_stages": schema.ListAttribute{
				MarkdownDescription: "List of progress stages",
				Computed:            true,
				ElementType:         types.ObjectType{},
			},

			// Connectivity (JSON strings per Swagger)
			"connectivity": schema.StringAttribute{
				MarkdownDescription: "JSON string containing connectivity information",
				Computed:            true,
			},
			"api_vip_connectivity": schema.StringAttribute{
				MarkdownDescription: "JSON string containing API VIP connectivity response",
				Computed:            true,
			},
			"tang_connectivity": schema.StringAttribute{
				MarkdownDescription: "JSON string containing Tang connectivity information",
				Computed:            true,
			},

			// Hardware inventory (JSON string per Swagger - NOT parsed objects)
			"inventory": schema.StringAttribute{
				MarkdownDescription: "JSON string containing hardware inventory information collected from the host",
				Computed:            true,
			},
			"free_addresses": schema.StringAttribute{
				MarkdownDescription: "JSON string containing list of free IP addresses available on this host",
				Computed:            true,
			},
			"ntp_sources": schema.StringAttribute{
				MarkdownDescription: "JSON string containing the configured NTP sources on the host",
				Computed:            true,
			},
			"disks_info": schema.StringAttribute{
				MarkdownDescription: "JSON string containing additional information about disks",
				Computed:            true,
			},

			// Host role and configuration
			"role": schema.StringAttribute{
				MarkdownDescription: "The role assigned to this host (master, worker, auto-assign)",
				Computed:            true,
			},
			"suggested_role": schema.StringAttribute{
				MarkdownDescription: "The suggested role for this host",
				Computed:            true,
			},
			"bootstrap": schema.BoolAttribute{
				MarkdownDescription: "Whether this host is the bootstrap node",
				Computed:            true,
			},

			// Logs
			"logs_info": schema.SingleNestedAttribute{
				MarkdownDescription: "The progress of log collection or empty if logs are not applicable",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"state": schema.StringAttribute{
						MarkdownDescription: "Log collection state",
						Computed:            true,
					},
					"state_info": schema.StringAttribute{
						MarkdownDescription: "Additional info about log collection state",
						Computed:            true,
					},
				},
			},
			"logs_collected_at": schema.StringAttribute{
				MarkdownDescription: "When logs were collected from this host",
				Computed:            true,
			},
			"logs_started_at": schema.StringAttribute{
				MarkdownDescription: "When log collection started",
				Computed:            true,
			},

			// Installation
			"installer_version": schema.StringAttribute{
				MarkdownDescription: "Installer version",
				Computed:            true,
			},
			"installation_disk_path": schema.StringAttribute{
				MarkdownDescription: "Contains the inventory disk path, used for backward compatibility with the old UI",
				Computed:            true,
			},
			"installation_disk_id": schema.StringAttribute{
				MarkdownDescription: "Contains the inventory disk id to install on",
				Computed:            true,
			},

			// Timestamps
			"created_at": schema.StringAttribute{
				MarkdownDescription: "When the host was created",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "When the host was last updated",
				Computed:            true,
			},
			"checked_in_at": schema.StringAttribute{
				MarkdownDescription: "The last time the host's agent communicated with the service",
				Computed:            true,
			},
			"registered_at": schema.StringAttribute{
				MarkdownDescription: "The last time the host's agent tried to register in the service",
				Computed:            true,
			},

			// Host details
			"user_name": schema.StringAttribute{
				MarkdownDescription: "Username for host access",
				Computed:            true,
			},
			"requested_hostname": schema.StringAttribute{
				MarkdownDescription: "Requested hostname for this host",
				Computed:            true,
			},
			"connection_timed_out": schema.BoolAttribute{
				MarkdownDescription: "Indicate that connection to assisted service was timed out when soft timeout is enabled",
				Computed:            true,
			},
			"discovery_agent_version": schema.StringAttribute{
				MarkdownDescription: "Discovery agent version",
				Computed:            true,
			},
			"media_status": schema.StringAttribute{
				MarkdownDescription: "Media status (connected, disconnected)",
				Computed:            true,
			},

			// Advanced configuration (JSON strings per Swagger)
			"ignition_config_overrides": schema.StringAttribute{
				MarkdownDescription: "JSON formatted string containing the user overrides for the host's pointer ignition",
				Computed:            true,
			},
			"installer_args": schema.StringAttribute{
				MarkdownDescription: "Additional arguments for the installer",
				Computed:            true,
			},
			"timestamp": schema.Int64Attribute{
				MarkdownDescription: "The time on the host as seconds since the Unix epoch",
				Computed:            true,
			},
			"machine_config_pool_name": schema.StringAttribute{
				MarkdownDescription: "Machine config pool name for this host",
				Computed:            true,
			},
			"images_status": schema.StringAttribute{
				MarkdownDescription: "JSON string containing array of image statuses",
				Computed:            true,
			},
			"domain_name_resolutions": schema.StringAttribute{
				MarkdownDescription: "JSON string containing the domain name resolution result",
				Computed:            true,
			},
			"ignition_endpoint_token_set": schema.BoolAttribute{
				MarkdownDescription: "True if the token to fetch the ignition from ignition_endpoint_url is set",
				Computed:            true,
			},
			"node_labels": schema.StringAttribute{
				MarkdownDescription: "JSON containing node's labels",
				Computed:            true,
			},
			"disks_to_be_formatted": schema.StringAttribute{
				MarkdownDescription: "A comma-separated list of disks that will be formatted once installation begins",
				Computed:            true,
			},
			"skip_formatting_disks": schema.StringAttribute{
				MarkdownDescription: "A comma-separated list of host disks that the service will avoid formatting",
				Computed:            true,
			},
		},
	}
}

func (d *HostDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *HostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HostDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get host from API
	host, err := d.client.GetHost(ctx, data.InfraEnvID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read host, got error: %s", err),
		)
		return
	}

	// Map basic API response to data model
	data.ClusterID = types.StringValue(host.ClusterID)
	data.Status = types.StringValue(host.Status)
	data.StatusInfo = types.StringValue(host.StatusInfo)
	data.Kind = types.StringValue(host.Kind)
	data.Href = types.StringValue(host.Href)
	data.Role = types.StringValue(host.Role)
	data.MachineConfigPoolName = types.StringValue(host.MachineConfigPoolName)

	// Handle timestamps
	if !host.CreatedAt.IsZero() {
		data.CreatedAt = types.StringValue(host.CreatedAt.String())
	}
	if !host.UpdatedAt.IsZero() {
		data.UpdatedAt = types.StringValue(host.UpdatedAt.String())
	}

	// Handle progress (this is a complex nested structure, simplified for now)
	if host.Progress != nil {
		currentStage := types.StringValue("")
		progressInfo := types.StringValue("")
		stageStartedAt := types.StringValue("")
		stageUpdatedAt := types.StringValue("")

		if host.Progress.CurrentStage != "" {
			currentStage = types.StringValue(host.Progress.CurrentStage)
		}
		if host.Progress.ProgressInfo != "" {
			progressInfo = types.StringValue(host.Progress.ProgressInfo)
		}
		if !host.Progress.StageStartedAt.IsZero() {
			stageStartedAt = types.StringValue(host.Progress.StageStartedAt.String())
		}
		if !host.Progress.StageUpdatedAt.IsZero() {
			stageUpdatedAt = types.StringValue(host.Progress.StageUpdatedAt.String())
		}

		progressObj, diag := types.ObjectValue(
			map[string]attr.Type{
				"current_stage":           types.StringType,
				"progress_info":           types.StringType,
				"installation_percentage": types.Int64Type,
				"stage_started_at":        types.StringType,
				"stage_updated_at":        types.StringType,
			},
			map[string]attr.Value{
				"current_stage":           currentStage,
				"progress_info":           progressInfo,
				"installation_percentage": types.Int64Value(0), // Not available in basic Progress model
				"stage_started_at":        stageStartedAt,
				"stage_updated_at":        stageUpdatedAt,
			},
		)
		resp.Diagnostics.Append(diag...)
		data.Progress = progressObj
	}

	// Note: FreeAddresses are not available in the basic Host model

	// Note: Inventory is not available in the basic Host model
	// This would need to be fetched from other API endpoints if needed

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
