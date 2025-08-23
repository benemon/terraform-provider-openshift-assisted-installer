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

var _ datasource.DataSource = &SupportedOperatorsDataSource{}

func NewSupportedOperatorsDataSource() datasource.DataSource {
	return &SupportedOperatorsDataSource{}
}

type SupportedOperatorsDataSource struct {
	client *client.Client
}

type SupportedOperatorsDataSourceModel struct {
	ID        types.String   `tfsdk:"id"`
	Operators types.List     `tfsdk:"operators"`
}

func (d *SupportedOperatorsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_supported_operators"
}

func (d *SupportedOperatorsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches list of supported operators from the Assisted Service API",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"operators": schema.ListAttribute{
				MarkdownDescription: "List of supported operator names",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *SupportedOperatorsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SupportedOperatorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SupportedOperatorsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Fetching supported operators")

	// Call the API
	operators, err := d.client.GetSupportedOperators(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching supported operators",
			fmt.Sprintf("Could not read supported operators: %s", err),
		)
		return
	}

	// Convert to terraform list
	if len(operators) > 0 {
		operatorElements := make([]types.String, len(operators))
		for i, operator := range operators {
			operatorElements[i] = types.StringValue(operator)
		}
		
		operatorList, diags := types.ListValueFrom(ctx, types.StringType, operatorElements)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		
		data.Operators = operatorList
	} else {
		data.Operators = types.ListNull(types.StringType)
	}

	// Set ID for the data source
	data.ID = types.StringValue("supported_operators")

	tflog.Info(ctx, "Successfully fetched supported operators", map[string]interface{}{
		"count": len(operators),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}