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
var _ datasource.DataSource = &OperatorBundlesDataSource{}

func NewOperatorBundlesDataSource() datasource.DataSource {
	return &OperatorBundlesDataSource{}
}

// OperatorBundlesDataSource defines the data source implementation.
type OperatorBundlesDataSource struct {
	client *client.Client
}

// OperatorBundlesDataSourceModel describes the data source data model.
type OperatorBundlesDataSourceModel struct {
	ID      types.String          `tfsdk:"id"`
	Bundles []OperatorBundleModel `tfsdk:"bundles"`
}

type OperatorBundleModel struct {
	ID        types.String `tfsdk:"id"`
	Title     types.String `tfsdk:"title"`
	Operators types.List   `tfsdk:"operators"`
}

func (d *OperatorBundlesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_operator_bundles"
}

func (d *OperatorBundlesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Operator bundles data source provides information about available operator bundles for OpenShift clusters.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Data source identifier.",
			},
			"bundles": schema.ListNestedAttribute{
				MarkdownDescription: "List of available operator bundles.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Bundle identifier (e.g., 'virtualization', 'openshift-ai-nvidia').",
							Computed:            true,
						},
						"title": schema.StringAttribute{
							MarkdownDescription: "Bundle title.",
							Computed:            true,
						},
						"operators": schema.ListAttribute{
							MarkdownDescription: "List of operator names included in this bundle.",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *OperatorBundlesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OperatorBundlesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OperatorBundlesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Fetching operator bundles", map[string]any{
		"data_source": "oai_operator_bundles",
	})

	bundles, err := d.client.GetOperatorBundles(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching operator bundles", fmt.Sprintf("Could not read operator bundles: %s", err))
		return
	}

	// Convert to Terraform model
	data.ID = types.StringValue("operator_bundles_all")
	data.Bundles = make([]OperatorBundleModel, len(*bundles))

	for i, bundle := range *bundles {
		// Convert operator names to Terraform list
		operatorElements := make([]types.String, len(bundle.Operators))
		for j, operatorName := range bundle.Operators {
			operatorElements[j] = types.StringValue(operatorName)
		}

		operatorList, listDiags := types.ListValueFrom(ctx, types.StringType, operatorElements)
		if listDiags.HasError() {
			resp.Diagnostics.Append(listDiags...)
			return
		}

		data.Bundles[i] = OperatorBundleModel{
			ID:        types.StringValue(bundle.ID),
			Title:     types.StringValue(bundle.Title),
			Operators: operatorList,
		}
	}

	tflog.Info(ctx, "Successfully fetched operator bundles", map[string]any{
		"bundle_count": len(data.Bundles),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
