package provider

import (
	"context"
	"fmt"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ManifestDataSource{}

func NewManifestDataSource() datasource.DataSource {
	return &ManifestDataSource{}
}

// ManifestDataSource defines the data source implementation.
type ManifestDataSource struct {
	client *client.Client
}

// ManifestDataSourceModel describes the data source data model.
// All fields match exactly with Swagger manifest definition
type ManifestDataSourceModel struct {
	// Computed fields
	ID types.String `tfsdk:"id"`

	// Required fields for lookup
	ClusterID types.String `tfsdk:"cluster_id"`
	FileName  types.String `tfsdk:"file_name"`

	// Swagger manifest fields
	Folder         types.String `tfsdk:"folder"`
	ManifestSource types.String `tfsdk:"manifest_source"` // Missing critical field!

	// Content (may be base64 encoded)
	Content types.String `tfsdk:"content"`

	// Legacy fields (not in Swagger but keeping for backwards compatibility)
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *ManifestDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_manifest"
}

func (d *ManifestDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Retrieves information about a specific cluster manifest managed by the OpenShift Assisted Service. Manifests are YAML or JSON configuration files that customize the OpenShift cluster installation.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed identifier for this manifest (cluster_id/folder/file_name)",
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the cluster this manifest belongs to",
				Required:            true,
			},
			"file_name": schema.StringAttribute{
				MarkdownDescription: "The name of the manifest file (must end with .yaml, .yml, or .json)",
				Required:            true,
			},
			"folder": schema.StringAttribute{
				MarkdownDescription: "The folder that contains the files. Manifests can be placed in 'manifests' or 'openshift' directories.",
				Optional:            true,
				Computed:            true,
			},
			"manifest_source": schema.StringAttribute{
				MarkdownDescription: "Describes whether manifest is sourced from a user or created by the system ('user' or 'system')",
				Computed:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The decoded content of the manifest file (YAML or JSON format)",
				Computed:            true,
				Sensitive:           true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "When the manifest was created (RFC3339 format)",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "When the manifest was last updated (RFC3339 format)",
				Computed:            true,
			},
		},
	}
}

func (d *ManifestDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ManifestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ManifestDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set default folder if not specified
	folder := data.Folder.ValueString()
	if folder == "" {
		folder = "manifests"
		data.Folder = types.StringValue(folder)
	}

	// Set computed ID
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", data.ClusterID.ValueString(), folder, data.FileName.ValueString()))

	// Get manifests from API and find the specific one
	manifests, err := d.client.ListManifests(ctx, data.ClusterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to list manifests, got error: %s", err),
		)
		return
	}

	// Find the specific manifest
	var targetManifest *models.Manifest
	for _, manifest := range manifests {
		if manifest.FileName == data.FileName.ValueString() && manifest.Folder == folder {
			targetManifest = &manifest
			break
		}
	}

	if targetManifest == nil {
		resp.Diagnostics.AddError(
			"Manifest Not Found",
			fmt.Sprintf("Manifest with filename '%s' in folder '%s' not found in cluster '%s'",
				data.FileName.ValueString(), folder, data.ClusterID.ValueString()),
		)
		return
	}

	// Map API response to data model
	data.ManifestSource = types.StringValue(targetManifest.ManifestSource)

	// Download the manifest content
	content, err := d.client.DownloadManifestContent(ctx, data.ClusterID.ValueString(), data.FileName.ValueString(), folder)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to download manifest content, got error: %s", err),
		)
		return
	}

	// Set the content field
	data.Content = types.StringValue(content)

	// Note: CreatedAt, UpdatedAt are not available from the API
	data.CreatedAt = types.StringNull()
	data.UpdatedAt = types.StringNull()

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
