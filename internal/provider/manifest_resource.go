package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"

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
var _ resource.Resource = &ManifestResource{}
var _ resource.ResourceWithImportState = &ManifestResource{}

func NewManifestResource() resource.Resource {
	return &ManifestResource{}
}

// ManifestResource defines the resource implementation.
type ManifestResource struct {
	client *client.Client
}

// ManifestResourceModel describes the resource data model.
type ManifestResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ClusterID types.String `tfsdk:"cluster_id"`
	FileName  types.String `tfsdk:"file_name"`
	Folder    types.String `tfsdk:"folder"`
	Content   types.String `tfsdk:"content"`

	// Computed fields
	ManifestSource types.String `tfsdk:"manifest_source"`
}

func (r *ManifestResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_manifest"
}

func (r *ManifestResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manifest resource for managing custom Kubernetes manifests in OpenShift clusters. Manifests allow custom configuration of cluster components and additional resources.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Manifest identifier (computed from cluster_id/folder/file_name).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "Cluster ID to associate this manifest with.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_name": schema.StringAttribute{
				MarkdownDescription: "Name of the manifest file. Must have .yaml, .yml, or .json extension.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						// Must end with .yaml, .yml, or .json
						regexp.MustCompile(`\.(yaml|yml|json)$`),
						"file_name must end with .yaml, .yml, or .json",
					),
				},
			},
			"folder": schema.StringAttribute{
				MarkdownDescription: "Folder where the manifest will be stored. Use 'manifests' for user manifests or 'openshift' for cluster-level manifests.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("manifests"),
				Validators: []validator.String{
					stringvalidator.OneOf("manifests", "openshift"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "Content of the manifest in YAML or JSON format. The content will be automatically base64-encoded for the API.",
				Required:            true,
			},

			// Computed attributes
			"manifest_source": schema.StringAttribute{
				MarkdownDescription: "Source information for the manifest.",
				Computed:            true,
			},
		},
	}
}

func (r *ManifestResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ManifestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ManifestResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate and encode content
	encodedContent, err := r.encodeManifestContent(data.Content.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid manifest content", fmt.Sprintf("Could not encode manifest content: %s", err))
		return
	}

	// Create the manifest parameters
	createParams := models.CreateManifestParams{
		FileName: data.FileName.ValueString(),
		Folder:   data.Folder.ValueString(),
		Content:  encodedContent,
	}

	tflog.Info(ctx, "Creating manifest", map[string]any{
		"cluster_id": data.ClusterID.ValueString(),
		"file_name":  data.FileName.ValueString(),
		"folder":     data.Folder.ValueString(),
	})

	// Create the manifest
	err = r.client.CreateManifest(ctx, data.ClusterID.ValueString(), createParams)
	if err != nil {
		resp.Diagnostics.AddError("Error creating manifest", fmt.Sprintf("Could not create manifest: %s", err))
		return
	}

	// Set the computed ID (cluster_id/folder/file_name)
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", data.ClusterID.ValueString(), data.Folder.ValueString(), data.FileName.ValueString()))

	// Read back the manifest to get computed fields
	manifests, err := r.client.ListManifests(ctx, data.ClusterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading created manifest", fmt.Sprintf("Could not read created manifest: %s", err))
		return
	}

	// Find the created manifest
	var createdManifest *models.Manifest
	for _, manifest := range manifests {
		if manifest.FileName == data.FileName.ValueString() && manifest.Folder == data.Folder.ValueString() {
			createdManifest = &manifest
			break
		}
	}

	if createdManifest != nil {
		data.ManifestSource = types.StringValue(createdManifest.ManifestSource)
	}

	tflog.Info(ctx, "Successfully created manifest", map[string]any{
		"manifest_id": data.ID.ValueString(),
		"cluster_id":  data.ClusterID.ValueString(),
		"file_name":   data.FileName.ValueString(),
		"folder":      data.Folder.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ManifestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ManifestResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// List manifests for the cluster to find this one
	manifests, err := r.client.ListManifests(ctx, data.ClusterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading manifests", fmt.Sprintf("Could not read manifests for cluster %s: %s", data.ClusterID.ValueString(), err))
		return
	}

	// Find the specific manifest
	var foundManifest *models.Manifest
	for _, manifest := range manifests {
		if manifest.FileName == data.FileName.ValueString() && manifest.Folder == data.Folder.ValueString() {
			foundManifest = &manifest
			break
		}
	}

	if foundManifest == nil {
		// Manifest not found, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Update computed fields
	data.ManifestSource = types.StringValue(foundManifest.ManifestSource)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ManifestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ManifestResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate and encode content
	encodedContent, err := r.encodeManifestContent(data.Content.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid manifest content", fmt.Sprintf("Could not encode manifest content: %s", err))
		return
	}

	// Create the update parameters
	updateParams := models.UpdateManifestParams{
		FileName: data.FileName.ValueString(),
		Content:  encodedContent,
	}

	tflog.Info(ctx, "Updating manifest", map[string]any{
		"cluster_id": data.ClusterID.ValueString(),
		"file_name":  data.FileName.ValueString(),
		"folder":     data.Folder.ValueString(),
	})

	// Update the manifest
	err = r.client.UpdateManifest(ctx, data.ClusterID.ValueString(), updateParams)
	if err != nil {
		resp.Diagnostics.AddError("Error updating manifest", fmt.Sprintf("Could not update manifest: %s", err))
		return
	}

	// Read back the manifest to get updated computed fields
	manifests, err := r.client.ListManifests(ctx, data.ClusterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated manifest", fmt.Sprintf("Could not read updated manifest: %s", err))
		return
	}

	// Find the updated manifest
	var updatedManifest *models.Manifest
	for _, manifest := range manifests {
		if manifest.FileName == data.FileName.ValueString() && manifest.Folder == data.Folder.ValueString() {
			updatedManifest = &manifest
			break
		}
	}

	if updatedManifest != nil {
		data.ManifestSource = types.StringValue(updatedManifest.ManifestSource)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ManifestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ManifestResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting manifest", map[string]any{
		"cluster_id": data.ClusterID.ValueString(),
		"file_name":  data.FileName.ValueString(),
		"folder":     data.Folder.ValueString(),
	})

	// Delete the manifest
	err := r.client.DeleteManifest(ctx, data.ClusterID.ValueString(), data.Folder.ValueString(), data.FileName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting manifest", fmt.Sprintf("Could not delete manifest: %s", err))
		return
	}

	tflog.Info(ctx, "Successfully deleted manifest", map[string]any{
		"manifest_id": data.ID.ValueString(),
		"cluster_id":  data.ClusterID.ValueString(),
		"file_name":   data.FileName.ValueString(),
		"folder":      data.Folder.ValueString(),
	})
}

func (r *ManifestResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import state expects "cluster_id/folder/file_name" format
	// For simplicity, we'll use the ID as provided and parse it in the resource
	idParts := req.ID
	if idParts == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: cluster_id/folder/file_name. Got: %q", req.ID),
		)
		return
	}

	// Set the ID for now - the Read method will populate other fields
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// Helper functions

func (r *ManifestResource) encodeManifestContent(content string) (string, error) {
	// Validate that content is not empty
	if content == "" {
		return "", fmt.Errorf("manifest content cannot be empty")
	}

	// The API expects base64-encoded content
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	return encoded, nil
}
