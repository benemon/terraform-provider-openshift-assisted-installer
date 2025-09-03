package provider

import (
	"context"
	"fmt"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SupportLevelsDataSource{}

func NewSupportLevelsDataSource() datasource.DataSource {
	return &SupportLevelsDataSource{}
}

// SupportLevelsDataSource defines the data source implementation.
type SupportLevelsDataSource struct {
	client *client.Client
}

// SupportLevelsDataSourceModel describes the data source data model.
type SupportLevelsDataSourceModel struct {
	ID                 types.String         `tfsdk:"id"`
	OpenShiftVersion   types.String         `tfsdk:"openshift_version"`
	CPUArchitecture    types.String         `tfsdk:"cpu_architecture"`
	PlatformType       types.String         `tfsdk:"platform_type"`
	Features           map[string]string    `tfsdk:"features"`
	Architectures      map[string]string    `tfsdk:"architectures"`
}

func (d *SupportLevelsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_support_levels"
}

func (d *SupportLevelsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Support levels data source provides information about feature and architecture support levels for OpenShift versions.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Data source identifier.",
			},
			"openshift_version": schema.StringAttribute{
				MarkdownDescription: "Version of the OpenShift cluster (required).",
				Required:            true,
			},
			"cpu_architecture": schema.StringAttribute{
				MarkdownDescription: "CPU architecture filter (optional). Examples: x86_64, arm64, ppc64le, s390x.",
				Optional:            true,
			},
			"platform_type": schema.StringAttribute{
				MarkdownDescription: "Platform type filter (optional). Examples: baremetal, nutanix, vsphere.",
				Optional:            true,
			},
			"features": schema.MapAttribute{
				MarkdownDescription: "Map of feature names to their support levels (supported, tech-preview, dev-preview, unsupported).",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"architectures": schema.MapAttribute{
				MarkdownDescription: "Map of CPU architectures to their support levels for the specified OpenShift version.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *SupportLevelsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SupportLevelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SupportLevelsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	openshiftVersion := data.OpenShiftVersion.ValueString()
	cpuArchitecture := ""
	platformType := ""

	if !data.CPUArchitecture.IsNull() {
		cpuArchitecture = data.CPUArchitecture.ValueString()
	}

	if !data.PlatformType.IsNull() {
		platformType = data.PlatformType.ValueString()
	}

	tflog.Info(ctx, "Fetching support levels", map[string]any{
		"data_source":        "oai_support_levels",
		"openshift_version":  openshiftVersion,
		"cpu_architecture":   cpuArchitecture,
		"platform_type":      platformType,
	})

	// Fetch feature support levels
	features, err := d.client.GetSupportedFeatures(ctx, openshiftVersion, cpuArchitecture, platformType)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching supported features", fmt.Sprintf("Could not read supported features: %s", err))
		return
	}

	// Fetch architecture support levels  
	architectures, err := d.client.GetSupportedArchitectures(ctx, openshiftVersion)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching supported architectures", fmt.Sprintf("Could not read supported architectures: %s", err))
		return
	}

	// Convert to Terraform model
	data.ID = types.StringValue(fmt.Sprintf("support_levels_%s", openshiftVersion))
	data.Features = *features
	data.Architectures = *architectures

	tflog.Info(ctx, "Successfully fetched support levels", map[string]any{
		"feature_count":      len(data.Features),
		"architecture_count": len(data.Architectures),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}