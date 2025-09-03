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
var _ resource.Resource = &InfraEnvResource{}
var _ resource.ResourceWithImportState = &InfraEnvResource{}

func NewInfraEnvResource() resource.Resource {
	return &InfraEnvResource{}
}

// InfraEnvResource defines the resource implementation.
type InfraEnvResource struct {
	client *client.Client
}

// InfraEnvResourceModel describes the resource data model.
type InfraEnvResourceModel struct {
	ID                   types.String                  `tfsdk:"id"`
	Name                 types.String                  `tfsdk:"name"`
	ClusterID            types.String                  `tfsdk:"cluster_id"`
	CPUArchitecture      types.String                  `tfsdk:"cpu_architecture"`
	PullSecret           types.String                  `tfsdk:"pull_secret"`
	SSHAuthorizedKey     types.String                  `tfsdk:"ssh_authorized_key"`
	ImageType            types.String                  `tfsdk:"image_type"`
	OpenShiftVersion     types.String                  `tfsdk:"openshift_version"`
	Proxy                *InfraEnvProxyModel           `tfsdk:"proxy"`
	StaticNetworkConfig  []InfraEnvStaticNetworkModel  `tfsdk:"static_network_config"`
	KernelArguments      []InfraEnvKernelArgumentModel `tfsdk:"kernel_arguments"`
	IgnitionConfigOverride types.String                `tfsdk:"ignition_config_override"`
	
	// Computed fields
	DownloadURL          types.String                  `tfsdk:"download_url"`
	ExpiresAt            types.String                  `tfsdk:"expires_at"`
	Type                 types.String                  `tfsdk:"type"`
}

type InfraEnvProxyModel struct {
	HTTPProxy  types.String `tfsdk:"http_proxy"`
	HTTPSProxy types.String `tfsdk:"https_proxy"`
	NoProxy    types.String `tfsdk:"no_proxy"`
}

type InfraEnvStaticNetworkModel struct {
	NetworkYAML types.String `tfsdk:"network_yaml"`
	MACInterfaceMap []InfraEnvMACInterfaceModel `tfsdk:"mac_interface_map"`
}

type InfraEnvMACInterfaceModel struct {
	MACAddress        types.String `tfsdk:"mac_address"`
	LogicalNICName    types.String `tfsdk:"logical_nic_name"`
}

type InfraEnvKernelArgumentModel struct {
	Operation types.String `tfsdk:"operation"`
	Value     types.String `tfsdk:"value"`
}

func (r *InfraEnvResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_infra_env"
}

func (r *InfraEnvResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Infrastructure environment resource for OpenShift cluster host discovery. Creates a discovery ISO that hosts can boot from to join the cluster.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Infrastructure environment identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the infrastructure environment.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "Cluster ID to associate this infrastructure environment with. If provided, discovered hosts will be automatically bound to this cluster.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cpu_architecture": schema.StringAttribute{
				MarkdownDescription: "CPU architecture for the infrastructure environment.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("x86_64", "aarch64", "arm64", "ppc64le", "s390x", "multi"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pull_secret": schema.StringAttribute{
				MarkdownDescription: "Red Hat pull secret for downloading OpenShift images.",
				Required:            true,
				Sensitive:           true,
			},
			"ssh_authorized_key": schema.StringAttribute{
				MarkdownDescription: "SSH public key for accessing discovered hosts.",
				Optional:            true,
				Sensitive:           true,
			},
			"image_type": schema.StringAttribute{
				MarkdownDescription: "Type of discovery image to generate.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("minimal-iso"),
				Validators: []validator.String{
					stringvalidator.OneOf("full-iso", "minimal-iso"),
				},
			},
			"openshift_version": schema.StringAttribute{
				MarkdownDescription: "OpenShift version for the infrastructure environment. If not specified, uses the cluster's version.",
				Optional:            true,
			},
			"proxy": schema.SingleNestedAttribute{
				MarkdownDescription: "Proxy configuration for hosts discovered through this infrastructure environment.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"http_proxy": schema.StringAttribute{
						MarkdownDescription: "HTTP proxy URL.",
						Optional:            true,
					},
					"https_proxy": schema.StringAttribute{
						MarkdownDescription: "HTTPS proxy URL.",
						Optional:            true,
					},
					"no_proxy": schema.StringAttribute{
						MarkdownDescription: "Comma-separated list of hosts/domains to exclude from proxy.",
						Optional:            true,
					},
				},
			},
			"static_network_config": schema.ListNestedAttribute{
				MarkdownDescription: "Static network configuration for discovered hosts.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"network_yaml": schema.StringAttribute{
							MarkdownDescription: "Network configuration in YAML format.",
							Required:            true,
						},
						"mac_interface_map": schema.ListNestedAttribute{
							MarkdownDescription: "Mapping between MAC addresses and logical interface names.",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"mac_address": schema.StringAttribute{
										MarkdownDescription: "MAC address of the interface.",
										Required:            true,
									},
									"logical_nic_name": schema.StringAttribute{
										MarkdownDescription: "Logical name for the interface.",
										Required:            true,
									},
								},
							},
						},
					},
				},
			},
			"kernel_arguments": schema.ListNestedAttribute{
				MarkdownDescription: "Kernel arguments to apply to discovered hosts.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"operation": schema.StringAttribute{
							MarkdownDescription: "Operation to perform with the kernel argument.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("append", "replace", "delete"),
							},
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Kernel argument value.",
							Required:            true,
						},
					},
				},
			},
			"ignition_config_override": schema.StringAttribute{
				MarkdownDescription: "Custom ignition configuration to override defaults.",
				Optional:            true,
			},
			
			// Computed attributes
			"download_url": schema.StringAttribute{
				MarkdownDescription: "URL to download the discovery ISO.",
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "Expiration time for the discovery ISO.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the infrastructure environment.",
				Computed:            true,
			},
		},
	}
}

func (r *InfraEnvResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InfraEnvResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InfraEnvResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	createParams := r.terraformToCreateAPIModel(ctx, &data)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating infrastructure environment", map[string]any{
		"name":             data.Name.ValueString(),
		"cpu_architecture": data.CPUArchitecture.ValueString(),
		"cluster_id":       data.ClusterID.ValueString(),
	})

	// Create the infrastructure environment
	infraEnv, err := r.client.CreateInfraEnv(ctx, *createParams)
	if err != nil {
		resp.Diagnostics.AddError("Error creating infrastructure environment", fmt.Sprintf("Could not create infrastructure environment: %s", err))
		return
	}

	// Update model with response data
	r.apiToTerraformModel(ctx, infraEnv, &data)

	tflog.Info(ctx, "Successfully created infrastructure environment", map[string]any{
		"infra_env_id": data.ID.ValueString(),
		"download_url": data.DownloadURL.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InfraEnvResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InfraEnvResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the infrastructure environment from the API
	infraEnv, err := r.client.GetInfraEnv(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading infrastructure environment", fmt.Sprintf("Could not read infrastructure environment %s: %s", data.ID.ValueString(), err))
		return
	}

	// Update model with current API state
	r.apiToTerraformModel(ctx, infraEnv, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InfraEnvResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InfraEnvResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	updateParams := r.terraformToUpdateAPIModel(ctx, &data)

	tflog.Info(ctx, "Updating infrastructure environment", map[string]any{
		"infra_env_id": data.ID.ValueString(),
		"name":         data.Name.ValueString(),
	})

	// Update the infrastructure environment
	infraEnv, err := r.client.UpdateInfraEnv(ctx, data.ID.ValueString(), *updateParams)
	if err != nil {
		resp.Diagnostics.AddError("Error updating infrastructure environment", fmt.Sprintf("Could not update infrastructure environment %s: %s", data.ID.ValueString(), err))
		return
	}

	// Update model with response data
	r.apiToTerraformModel(ctx, infraEnv, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InfraEnvResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InfraEnvResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting infrastructure environment", map[string]any{
		"infra_env_id": data.ID.ValueString(),
	})

	// Delete the infrastructure environment
	err := r.client.DeleteInfraEnv(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting infrastructure environment", fmt.Sprintf("Could not delete infrastructure environment %s: %s", data.ID.ValueString(), err))
		return
	}

	tflog.Info(ctx, "Successfully deleted infrastructure environment", map[string]any{
		"infra_env_id": data.ID.ValueString(),
	})
}

func (r *InfraEnvResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper functions to convert between Terraform and API models

func (r *InfraEnvResource) terraformToCreateAPIModel(ctx context.Context, data *InfraEnvResourceModel) *models.InfraEnvCreateParams {
	params := &models.InfraEnvCreateParams{
		Name:            data.Name.ValueString(),
		CPUArchitecture: data.CPUArchitecture.ValueString(),
		PullSecret:      data.PullSecret.ValueString(),
	}

	if !data.ClusterID.IsNull() {
		params.ClusterID = data.ClusterID.ValueString()
	}

	if !data.SSHAuthorizedKey.IsNull() {
		params.SSHAuthorizedKey = data.SSHAuthorizedKey.ValueString()
	}

	if !data.ImageType.IsNull() {
		params.ImageType = data.ImageType.ValueString()
	}

	if !data.OpenShiftVersion.IsNull() {
		params.OpenshiftVersion = data.OpenShiftVersion.ValueString()
	}

	if !data.IgnitionConfigOverride.IsNull() {
		params.IgnitionConfigOverride = data.IgnitionConfigOverride.ValueString()
	}

	// Convert proxy settings
	if data.Proxy != nil {
		params.Proxy = &models.Proxy{}
		if !data.Proxy.HTTPProxy.IsNull() {
			params.Proxy.HTTPProxy = data.Proxy.HTTPProxy.ValueString()
		}
		if !data.Proxy.HTTPSProxy.IsNull() {
			params.Proxy.HTTPSProxy = data.Proxy.HTTPSProxy.ValueString()
		}
		if !data.Proxy.NoProxy.IsNull() {
			params.Proxy.NoProxy = data.Proxy.NoProxy.ValueString()
		}
	}

	// Convert static network config
	if len(data.StaticNetworkConfig) > 0 {
		params.StaticNetworkConfig = make([]models.HostStaticNetworkConfig, len(data.StaticNetworkConfig))
		for i, config := range data.StaticNetworkConfig {
			params.StaticNetworkConfig[i] = models.HostStaticNetworkConfig{
				NetworkYAML: config.NetworkYAML.ValueString(),
			}
			
			if len(config.MACInterfaceMap) > 0 {
				params.StaticNetworkConfig[i].MACInterfaceMap = make([]models.MACInterfaceMapEntry, len(config.MACInterfaceMap))
				for j, macMap := range config.MACInterfaceMap {
					params.StaticNetworkConfig[i].MACInterfaceMap[j] = models.MACInterfaceMapEntry{
						MACAddress:     macMap.MACAddress.ValueString(),
						LogicalNICName: macMap.LogicalNICName.ValueString(),
					}
				}
			}
		}
	}

	// Convert kernel arguments
	if len(data.KernelArguments) > 0 {
		params.KernelArguments = make([]models.KernelArgument, len(data.KernelArguments))
		for i, arg := range data.KernelArguments {
			params.KernelArguments[i] = models.KernelArgument{
				Operation: arg.Operation.ValueString(),
				Value:     arg.Value.ValueString(),
			}
		}
	}

	return params
}

func (r *InfraEnvResource) terraformToUpdateAPIModel(ctx context.Context, data *InfraEnvResourceModel) *models.InfraEnvUpdateParams {
	params := &models.InfraEnvUpdateParams{}

	// For update, we need to use pointers for optional fields
	if !data.Name.IsNull() {
		name := data.Name.ValueString()
		params.Name = &name
	}

	if !data.PullSecret.IsNull() {
		pullSecret := data.PullSecret.ValueString()
		params.PullSecret = &pullSecret
	}

	if !data.SSHAuthorizedKey.IsNull() {
		sshKey := data.SSHAuthorizedKey.ValueString()
		params.SSHAuthorizedKey = &sshKey
	}

	if !data.ImageType.IsNull() {
		imageType := data.ImageType.ValueString()
		params.ImageType = &imageType
	}

	if !data.IgnitionConfigOverride.IsNull() {
		ignition := data.IgnitionConfigOverride.ValueString()
		params.IgnitionConfigOverride = &ignition
	}

	// Convert proxy settings
	if data.Proxy != nil {
		params.Proxy = &models.Proxy{}
		if !data.Proxy.HTTPProxy.IsNull() {
			params.Proxy.HTTPProxy = data.Proxy.HTTPProxy.ValueString()
		}
		if !data.Proxy.HTTPSProxy.IsNull() {
			params.Proxy.HTTPSProxy = data.Proxy.HTTPSProxy.ValueString()
		}
		if !data.Proxy.NoProxy.IsNull() {
			params.Proxy.NoProxy = data.Proxy.NoProxy.ValueString()
		}
	}

	// Convert static network config
	if len(data.StaticNetworkConfig) > 0 {
		params.StaticNetworkConfig = make([]models.HostStaticNetworkConfig, len(data.StaticNetworkConfig))
		for i, config := range data.StaticNetworkConfig {
			params.StaticNetworkConfig[i] = models.HostStaticNetworkConfig{
				NetworkYAML: config.NetworkYAML.ValueString(),
			}
			
			if len(config.MACInterfaceMap) > 0 {
				params.StaticNetworkConfig[i].MACInterfaceMap = make([]models.MACInterfaceMapEntry, len(config.MACInterfaceMap))
				for j, macMap := range config.MACInterfaceMap {
					params.StaticNetworkConfig[i].MACInterfaceMap[j] = models.MACInterfaceMapEntry{
						MACAddress:     macMap.MACAddress.ValueString(),
						LogicalNICName: macMap.LogicalNICName.ValueString(),
					}
				}
			}
		}
	}

	// Convert kernel arguments
	if len(data.KernelArguments) > 0 {
		params.KernelArguments = make([]models.KernelArgument, len(data.KernelArguments))
		for i, arg := range data.KernelArguments {
			params.KernelArguments[i] = models.KernelArgument{
				Operation: arg.Operation.ValueString(),
				Value:     arg.Value.ValueString(),
			}
		}
	}

	return params
}

func (r *InfraEnvResource) apiToTerraformModel(ctx context.Context, infraEnv *models.InfraEnv, data *InfraEnvResourceModel) {
	data.ID = types.StringValue(infraEnv.ID)
	data.Name = types.StringValue(infraEnv.Name)
	data.Type = types.StringValue(infraEnv.Type)
	data.CPUArchitecture = types.StringValue(infraEnv.CPUArchitecture)
	
	if infraEnv.ClusterID != "" {
		data.ClusterID = types.StringValue(infraEnv.ClusterID)
	}
	
	if infraEnv.OpenshiftVersion != "" {
		data.OpenShiftVersion = types.StringValue(infraEnv.OpenshiftVersion)
	}
	
	if infraEnv.SSHAuthorizedKey != "" {
		data.SSHAuthorizedKey = types.StringValue(infraEnv.SSHAuthorizedKey)
	}
	
	if infraEnv.DownloadURL != "" {
		data.DownloadURL = types.StringValue(infraEnv.DownloadURL)
	}
	
	if !infraEnv.ExpiresAt.IsZero() {
		data.ExpiresAt = types.StringValue(infraEnv.ExpiresAt.Format("2006-01-02T15:04:05Z"))
	}
}