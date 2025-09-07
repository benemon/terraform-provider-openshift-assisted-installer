package provider

import (
	"context"
	"fmt"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ClusterFilesDataSource{}

func NewClusterFilesDataSource() datasource.DataSource {
	return &ClusterFilesDataSource{}
}

// ClusterFilesDataSource defines the data source implementation.
type ClusterFilesDataSource struct {
	client *client.Client
}

// ClusterFilesDataSourceModel describes the data source data model.
type ClusterFilesDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ClusterID types.String `tfsdk:"cluster_id"`
	FileName  types.String `tfsdk:"file_name"`
	LogsType  types.String `tfsdk:"logs_type"`
	Content   types.String `tfsdk:"content"`
}

func (d *ClusterFilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_files"
}

func (d *ClusterFilesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Downloads specific cluster installation files such as ignition configs, manifests, and install configuration. Available files: bootstrap.ign, master.ign, worker.ign, metadata.json, install-config.yaml, logs, manifests.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for this data source instance",
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the cluster to download files for",
				Required:            true,
			},
			"file_name": schema.StringAttribute{
				MarkdownDescription: "Name of the file to download (bootstrap.ign, master.ign, worker.ign, metadata.json, install-config.yaml, logs, manifests)",
				Required:            true,
			},
			"logs_type": schema.StringAttribute{
				MarkdownDescription: "Type of logs when file_name is 'logs' (controller, host, etc.)",
				Optional:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "Raw file content as a string",
				Computed:            true,
			},
		},
	}
}

func (d *ClusterFilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ClusterFilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterFilesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build query parameters
	params := make(map[string]string)

	if !data.LogsType.IsNull() && !data.LogsType.IsUnknown() {
		params["logs_type"] = data.LogsType.ValueString()
	}

	// Download file from API
	fileContent, err := d.client.DownloadClusterFiles(ctx, data.ClusterID.ValueString(), data.FileName.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to download cluster file '%s', got error: %s", data.FileName.ValueString(), err),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.ID = types.StringValue(fmt.Sprintf("file-%s-%s", data.ClusterID.ValueString(), data.FileName.ValueString()))
	data.Content = types.StringValue(string(fileContent))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
