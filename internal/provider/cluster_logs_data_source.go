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
var _ datasource.DataSource = &ClusterLogsDataSource{}

func NewClusterLogsDataSource() datasource.DataSource {
	return &ClusterLogsDataSource{}
}

// ClusterLogsDataSource defines the data source implementation.
type ClusterLogsDataSource struct {
	client *client.Client
}

// ClusterLogsDataSourceModel describes the data source data model.
type ClusterLogsDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ClusterID types.String `tfsdk:"cluster_id"`
	LogsType  types.String `tfsdk:"logs_type"`
	HostID    types.String `tfsdk:"host_id"`
	Content   types.String `tfsdk:"content"`
}

func (d *ClusterLogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_logs"
}

func (d *ClusterLogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Downloads cluster logs for troubleshooting and analysis. Logs can be filtered by type and host.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for this data source instance",
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the cluster to download logs for",
				Required:            true,
			},
			"logs_type": schema.StringAttribute{
				MarkdownDescription: "Type of logs to download (e.g., controller, host, all)",
				Optional:            true,
			},
			"host_id": schema.StringAttribute{
				MarkdownDescription: "Specific host ID to download logs for",
				Optional:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "Raw log content as a string",
				Computed:            true,
			},
		},
	}
}

func (d *ClusterLogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ClusterLogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterLogsDataSourceModel

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
	if !data.HostID.IsNull() && !data.HostID.IsUnknown() {
		params["host_id"] = data.HostID.ValueString()
	}

	// Download logs from API
	logContent, err := d.client.DownloadClusterLogs(ctx, data.ClusterID.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to download cluster logs, got error: %s", err),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.ID = types.StringValue(fmt.Sprintf("logs-%s", data.ClusterID.ValueString()))
	data.Content = types.StringValue(string(logContent))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
