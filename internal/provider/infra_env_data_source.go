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
var _ datasource.DataSource = &InfraEnvDataSource{}

func NewInfraEnvDataSource() datasource.DataSource {
	return &InfraEnvDataSource{}
}

// InfraEnvDataSource defines the data source implementation.
type InfraEnvDataSource struct {
	client *client.Client
}

// InfraEnvDataSourceModel describes the data source data model.
// All fields match exactly with Swagger infra-env definition
type InfraEnvDataSourceModel struct {
	// Required fields per Swagger
	ID        types.String `tfsdk:"id"`
	Kind      types.String `tfsdk:"kind"`
	Href      types.String `tfsdk:"href"`
	Name      types.String `tfsdk:"name"`
	Type      types.Object `tfsdk:"type"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`

	// Core infra-env info
	OpenshiftVersion types.String `tfsdk:"openshift_version"`
	UserName         types.String `tfsdk:"user_name"`
	OrgID            types.String `tfsdk:"org_id"`
	EmailDomain      types.String `tfsdk:"email_domain"`

	// Proxy configuration
	Proxy types.Object `tfsdk:"proxy"`

	// Network and NTP configuration
	AdditionalNTPSources types.String `tfsdk:"additional_ntp_sources"`
	SSHAuthorizedKey     types.String `tfsdk:"ssh_authorized_key"`
	StaticNetworkConfig  types.String `tfsdk:"static_network_config"`

	// Security
	PullSecretSet         types.Bool   `tfsdk:"pull_secret_set"`
	AdditionalTrustBundle types.String `tfsdk:"additional_trust_bundle"`

	// Ignition configuration
	IgnitionConfigOverride types.String `tfsdk:"ignition_config_override"`

	// Cluster association
	ClusterID types.String `tfsdk:"cluster_id"`

	// Image details
	SizeBytes        types.Int64  `tfsdk:"size_bytes"`
	DownloadURL      types.String `tfsdk:"download_url"`
	GeneratorVersion types.String `tfsdk:"generator_version"`
	ExpiresAt        types.String `tfsdk:"expires_at"`
	CPUArchitecture  types.String `tfsdk:"cpu_architecture"`
	KernelArguments  types.String `tfsdk:"kernel_arguments"`
}

func (d *InfraEnvDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_infra_env"
}

func (d *InfraEnvDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Retrieves information about an infrastructure environment managed by the OpenShift Assisted Service. Infrastructure environments generate discovery ISOs that hosts boot from to join a cluster.",

		Attributes: map[string]schema.Attribute{
			// Required fields per Swagger
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the object (UUID)",
				Required:            true,
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "Indicates the type of this object ('InfraEnv')",
				Computed:            true,
			},
			"href": schema.StringAttribute{
				MarkdownDescription: "Self link",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the infra-env",
				Computed:            true,
			},
			"type": schema.SingleNestedAttribute{
				MarkdownDescription: "Image type configuration",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Type of image (full-iso, minimal-iso)",
						Computed:            true,
					},
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when infra-env was created",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The last time that this infra-env was updated",
				Computed:            true,
			},

			// Core infra-env info
			"openshift_version": schema.StringAttribute{
				MarkdownDescription: "Version of the OpenShift cluster (used to infer the RHCOS version)",
				Computed:            true,
			},
			"user_name": schema.StringAttribute{
				MarkdownDescription: "Username associated with the infra-env",
				Computed:            true,
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: "Organization ID",
				Computed:            true,
			},
			"email_domain": schema.StringAttribute{
				MarkdownDescription: "Email domain",
				Computed:            true,
			},

			// Proxy configuration
			"proxy": schema.SingleNestedAttribute{
				MarkdownDescription: "HTTP/HTTPS proxy configuration",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"http_proxy": schema.StringAttribute{
						MarkdownDescription: "A proxy URL to use for creating HTTP connections outside the cluster",
						Computed:            true,
					},
					"https_proxy": schema.StringAttribute{
						MarkdownDescription: "A proxy URL to use for creating HTTPS connections outside the cluster",
						Computed:            true,
					},
					"no_proxy": schema.StringAttribute{
						MarkdownDescription: "An \"*\" or a comma-separated list of destination domain names, domains, IP addresses, or other network CIDRs to exclude from proxying",
						Computed:            true,
					},
				},
			},

			// Network and NTP configuration
			"additional_ntp_sources": schema.StringAttribute{
				MarkdownDescription: "A comma-separated list of NTP sources (name or IP) going to be added to all the hosts",
				Computed:            true,
			},
			"ssh_authorized_key": schema.StringAttribute{
				MarkdownDescription: "SSH public key for debugging the installation",
				Computed:            true,
			},
			"static_network_config": schema.StringAttribute{
				MarkdownDescription: "Static network configuration string in the format expected by discovery ignition generation",
				Computed:            true,
			},

			// Security
			"pull_secret_set": schema.BoolAttribute{
				MarkdownDescription: "True if the pull secret has been added to the cluster",
				Computed:            true,
			},
			"additional_trust_bundle": schema.StringAttribute{
				MarkdownDescription: "PEM-encoded X.509 certificate bundle. Hosts discovered by this infra-env will trust the certificates in this bundle",
				Computed:            true,
			},

			// Ignition configuration
			"ignition_config_override": schema.StringAttribute{
				MarkdownDescription: "JSON formatted string containing the user overrides for the initial ignition config",
				Computed:            true,
			},

			// Cluster association
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "If set, all hosts that register will be associated with the specified cluster",
				Computed:            true,
			},

			// Image details
			"size_bytes": schema.Int64Attribute{
				MarkdownDescription: "Image size in bytes",
				Computed:            true,
			},
			"download_url": schema.StringAttribute{
				MarkdownDescription: "URL to download the discovery ISO",
				Computed:            true,
			},
			"generator_version": schema.StringAttribute{
				MarkdownDescription: "Image generator version",
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "When the discovery ISO expires",
				Computed:            true,
			},
			"cpu_architecture": schema.StringAttribute{
				MarkdownDescription: "The CPU architecture of the image (x86_64, aarch64, arm64, ppc64le, s390x)",
				Computed:            true,
			},
			"kernel_arguments": schema.StringAttribute{
				MarkdownDescription: "JSON formatted string array representing the discovery image kernel arguments",
				Computed:            true,
			},
		},
	}
}

func (d *InfraEnvDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InfraEnvDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InfraEnvDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get infrastructure environment from API
	infraEnv, err := d.client.GetInfraEnv(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read infrastructure environment, got error: %s", err),
		)
		return
	}

	// Map API response to data model
	data.Name = types.StringValue(infraEnv.Name)
	data.ClusterID = types.StringValue(infraEnv.ClusterID)
	data.CPUArchitecture = types.StringValue(infraEnv.CPUArchitecture)
	data.OpenshiftVersion = types.StringValue(infraEnv.OpenshiftVersion)
	data.DownloadURL = types.StringValue(infraEnv.DownloadURL)
	// Handle type - construct nested object
	if infraEnv.Type != "" {
		typeObj, diag := types.ObjectValue(
			map[string]attr.Type{
				"type": types.StringType,
			},
			map[string]attr.Value{
				"type": types.StringValue(infraEnv.Type),
			},
		)
		resp.Diagnostics.Append(diag...)
		data.Type = typeObj
	}
	// ImageType is not in the basic InfraEnv model
	data.Kind = types.StringValue(infraEnv.Kind)
	data.Href = types.StringValue(infraEnv.Href)

	// Handle timestamps
	if !infraEnv.ExpiresAt.IsZero() {
		data.ExpiresAt = types.StringValue(infraEnv.ExpiresAt.String())
	}
	// GeneratedAt is not in the basic InfraEnv model
	if !infraEnv.CreatedAt.IsZero() {
		data.CreatedAt = types.StringValue(infraEnv.CreatedAt.String())
	}
	if !infraEnv.UpdatedAt.IsZero() {
		data.UpdatedAt = types.StringValue(infraEnv.UpdatedAt.String())
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
