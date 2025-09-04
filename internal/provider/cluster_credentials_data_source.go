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
var _ datasource.DataSource = &ClusterCredentialsDataSource{}

func NewClusterCredentialsDataSource() datasource.DataSource {
	return &ClusterCredentialsDataSource{}
}

// ClusterCredentialsDataSource defines the data source implementation.
type ClusterCredentialsDataSource struct {
	client *client.Client
}

// ClusterCredentialsDataSourceModel describes the data source data model.
type ClusterCredentialsDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	ClusterID  types.String `tfsdk:"cluster_id"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	ConsoleURL types.String `tfsdk:"console_url"`
}

func (d *ClusterCredentialsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_credentials"
}

func (d *ClusterCredentialsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Retrieves admin credentials for an installed OpenShift cluster. This data source can only be used after the cluster installation is complete.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for this data source instance",
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the cluster to retrieve credentials for",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Admin username for cluster access (typically 'kubeadmin')",
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Admin password for cluster access",
				Computed:            true,
				Sensitive:           true,
			},
			"console_url": schema.StringAttribute{
				MarkdownDescription: "URL of the OpenShift web console",
				Computed:            true,
			},
		},
	}
}

func (d *ClusterCredentialsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ClusterCredentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterCredentialsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get cluster credentials from API
	credentials, err := d.client.GetClusterCredentials(ctx, data.ClusterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read cluster credentials, got error: %s", err),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.ID = data.ClusterID // Use cluster_id as the unique identifier
	data.Username = types.StringValue(credentials.Username)
	data.Password = types.StringValue(credentials.Password)
	data.ConsoleURL = types.StringValue(credentials.ConsoleURL)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}