package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
)

var _ datasource.DataSource = &OpenShiftVersionsDataSource{}

func NewOpenShiftVersionsDataSource() datasource.DataSource {
	return &OpenShiftVersionsDataSource{}
}

type OpenShiftVersionsDataSource struct {
	client *client.Client
}

type OpenShiftVersionsDataSourceModel struct {
	ID         types.String                        `tfsdk:"id"`
	Version    types.String                        `tfsdk:"version"`
	OnlyLatest types.Bool                          `tfsdk:"only_latest"`
	Versions   []OpenShiftVersionModel             `tfsdk:"versions"`
}

type OpenShiftVersionModel struct {
	Version          types.String   `tfsdk:"version"`
	DisplayName      types.String   `tfsdk:"display_name"`
	SupportLevel     types.String   `tfsdk:"support_level"`
	Default          types.Bool     `tfsdk:"default"`
	CPUArchitectures types.List     `tfsdk:"cpu_architectures"`
}

func (d *OpenShiftVersionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_openshift_versions"
}

func (d *OpenShiftVersionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches available OpenShift versions from the Assisted Service API",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Filter by specific version pattern (e.g., '4.15', '4.15.20')",
				Optional:            true,
			},
			"only_latest": schema.BoolAttribute{
				MarkdownDescription: "Return only the latest versions",
				Optional:            true,
			},
			"versions": schema.ListNestedAttribute{
				MarkdownDescription: "List of available OpenShift versions",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"version": schema.StringAttribute{
							MarkdownDescription: "Version string (e.g., '4.15.20')",
							Computed:            true,
						},
						"display_name": schema.StringAttribute{
							MarkdownDescription: "Human-readable version name",
							Computed:            true,
						},
						"support_level": schema.StringAttribute{
							MarkdownDescription: "Support level (production, maintenance, dev-preview, etc.)",
							Computed:            true,
						},
						"default": schema.BoolAttribute{
							MarkdownDescription: "Whether this is the default version",
							Computed:            true,
						},
						"cpu_architectures": schema.ListAttribute{
							MarkdownDescription: "Supported CPU architectures",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *OpenShiftVersionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OpenShiftVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OpenShiftVersionsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract filter parameters
	var versionFilter string
	var onlyLatest bool
	
	if !data.Version.IsNull() {
		versionFilter = data.Version.ValueString()
	}
	if !data.OnlyLatest.IsNull() {
		onlyLatest = data.OnlyLatest.ValueBool()
	}

	tflog.Info(ctx, "Fetching OpenShift versions", map[string]interface{}{
		"version_filter": versionFilter,
		"only_latest":    onlyLatest,
	})

	// Call the API
	versions, err := d.client.GetOpenShiftVersions(ctx, versionFilter, onlyLatest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching OpenShift versions",
			fmt.Sprintf("Could not read OpenShift versions: %s", err),
		)
		return
	}

	// Convert to model
	data.Versions = make([]OpenShiftVersionModel, 0, len(*versions))
	for version, versionInfo := range *versions {
		// Convert CPU architectures to terraform list
		var archList types.List
		if len(versionInfo.CPUArchitectures) > 0 {
			archElements := make([]types.String, len(versionInfo.CPUArchitectures))
			for i, arch := range versionInfo.CPUArchitectures {
				archElements[i] = types.StringValue(arch)
			}
			var diags = make([]interface{}, 0)
			archList, _ = types.ListValueFrom(ctx, types.StringType, archElements)
			if len(diags) > 0 {
				resp.Diagnostics.AddWarning(
					"Could not convert CPU architectures",
					fmt.Sprintf("Failed to convert CPU architectures for version %s", version),
				)
			}
		} else {
			archList = types.ListNull(types.StringType)
		}

		data.Versions = append(data.Versions, OpenShiftVersionModel{
			Version:          types.StringValue(version),
			DisplayName:      types.StringValue(versionInfo.DisplayName),
			SupportLevel:     types.StringValue(versionInfo.SupportLevel),
			Default:          types.BoolValue(versionInfo.Default),
			CPUArchitectures: archList,
		})
	}

	// Set ID for the data source
	if versionFilter != "" {
		data.ID = types.StringValue(fmt.Sprintf("openshift_versions_%s", versionFilter))
	} else if onlyLatest {
		data.ID = types.StringValue("openshift_versions_latest")
	} else {
		data.ID = types.StringValue("openshift_versions_all")
	}

	tflog.Info(ctx, "Successfully fetched OpenShift versions", map[string]interface{}{
		"count": len(data.Versions),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}