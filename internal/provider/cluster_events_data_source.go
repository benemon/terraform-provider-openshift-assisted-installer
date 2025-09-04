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
var _ datasource.DataSource = &ClusterEventsDataSource{}

func NewClusterEventsDataSource() datasource.DataSource {
	return &ClusterEventsDataSource{}
}

// ClusterEventsDataSource defines the data source implementation.
type ClusterEventsDataSource struct {
	client *client.Client
}

// ClusterEventsDataSourceModel describes the data source data model.
type ClusterEventsDataSourceModel struct {
	ID           types.String              `tfsdk:"id"`
	ClusterID    types.String              `tfsdk:"cluster_id"`
	HostID       types.String              `tfsdk:"host_id"`
	InfraEnvID   types.String              `tfsdk:"infra_env_id"`
	Severities   types.List                `tfsdk:"severities"`
	Categories   types.List                `tfsdk:"categories"`
	Message      types.String              `tfsdk:"message"`
	Order        types.String              `tfsdk:"order"`
	Limit        types.Int64               `tfsdk:"limit"`
	Offset       types.Int64               `tfsdk:"offset"`
	ClusterLevel types.Bool                `tfsdk:"cluster_level"`
	Events       []EventModel              `tfsdk:"events"`
}

// EventModel represents a single event
type EventModel struct {
	Name        types.String `tfsdk:"name"`
	ClusterID   types.String `tfsdk:"cluster_id"`
	HostID      types.String `tfsdk:"host_id"`
	InfraEnvID  types.String `tfsdk:"infra_env_id"`
	Severity    types.String `tfsdk:"severity"`
	Category    types.String `tfsdk:"category"`
	Message     types.String `tfsdk:"message"`
	EventTime   types.String `tfsdk:"event_time"`
	RequestID   types.String `tfsdk:"request_id"`
	Props       types.String `tfsdk:"props"`
}

func (d *ClusterEventsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_events"
}

func (d *ClusterEventsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Retrieves events for a cluster, host, or infrastructure environment. Useful for monitoring installation progress and troubleshooting.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for this data source instance",
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "Filter events by cluster ID",
				Optional:            true,
			},
			"host_id": schema.StringAttribute{
				MarkdownDescription: "Filter events by host ID",
				Optional:            true,
			},
			"infra_env_id": schema.StringAttribute{
				MarkdownDescription: "Filter events by infrastructure environment ID",
				Optional:            true,
			},
			"severities": schema.ListAttribute{
				MarkdownDescription: "Filter by event severities (info, warning, error, critical)",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"categories": schema.ListAttribute{
				MarkdownDescription: "Filter by event categories (user, metrics)",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"message": schema.StringAttribute{
				MarkdownDescription: "Filter events by message pattern",
				Optional:            true,
			},
			"order": schema.StringAttribute{
				MarkdownDescription: "Order events by event_time (asc, desc)",
				Optional:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of events to return",
				Optional:            true,
			},
			"offset": schema.Int64Attribute{
				MarkdownDescription: "Number of events to skip",
				Optional:            true,
			},
			"cluster_level": schema.BoolAttribute{
				MarkdownDescription: "Include cluster-level events",
				Optional:            true,
			},
			"events": schema.ListNestedAttribute{
				MarkdownDescription: "List of events matching the filter criteria",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Event name",
							Computed:            true,
						},
						"cluster_id": schema.StringAttribute{
							MarkdownDescription: "Cluster ID associated with this event",
							Computed:            true,
						},
						"host_id": schema.StringAttribute{
							MarkdownDescription: "Host ID associated with this event",
							Computed:            true,
						},
						"infra_env_id": schema.StringAttribute{
							MarkdownDescription: "Infrastructure environment ID associated with this event",
							Computed:            true,
						},
						"severity": schema.StringAttribute{
							MarkdownDescription: "Event severity (info, warning, error, critical)",
							Computed:            true,
						},
						"category": schema.StringAttribute{
							MarkdownDescription: "Event category (user, metrics)",
							Computed:            true,
						},
						"message": schema.StringAttribute{
							MarkdownDescription: "Event message",
							Computed:            true,
						},
						"event_time": schema.StringAttribute{
							MarkdownDescription: "Timestamp when the event occurred",
							Computed:            true,
						},
						"request_id": schema.StringAttribute{
							MarkdownDescription: "Request ID that caused this event",
							Computed:            true,
						},
						"props": schema.StringAttribute{
							MarkdownDescription: "Additional event properties in JSON format",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *ClusterEventsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ClusterEventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterEventsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build query parameters
	params := make(map[string]string)
	
	if !data.HostID.IsNull() && !data.HostID.IsUnknown() {
		params["host_id"] = data.HostID.ValueString()
	}
	if !data.InfraEnvID.IsNull() && !data.InfraEnvID.IsUnknown() {
		params["infra_env_id"] = data.InfraEnvID.ValueString()
	}
	if !data.Message.IsNull() && !data.Message.IsUnknown() {
		params["message"] = data.Message.ValueString()
	}
	if !data.Order.IsNull() && !data.Order.IsUnknown() {
		params["order"] = data.Order.ValueString()
	}
	if !data.Limit.IsNull() && !data.Limit.IsUnknown() {
		params["limit"] = fmt.Sprintf("%d", data.Limit.ValueInt64())
	}
	if !data.Offset.IsNull() && !data.Offset.IsUnknown() {
		params["offset"] = fmt.Sprintf("%d", data.Offset.ValueInt64())
	}
	if !data.ClusterLevel.IsNull() && !data.ClusterLevel.IsUnknown() {
		params["cluster_level"] = fmt.Sprintf("%t", data.ClusterLevel.ValueBool())
	}

	// Handle list parameters (severities and categories)
	if !data.Severities.IsNull() && !data.Severities.IsUnknown() {
		var severities []string
		resp.Diagnostics.Append(data.Severities.ElementsAs(ctx, &severities, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, severity := range severities {
			params["severities"] = severity // Note: API might expect multiple values differently
		}
	}

	if !data.Categories.IsNull() && !data.Categories.IsUnknown() {
		var categories []string
		resp.Diagnostics.Append(data.Categories.ElementsAs(ctx, &categories, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, category := range categories {
			params["categories"] = category // Note: API might expect multiple values differently
		}
	}

	// Get cluster ID - could be from filter or required
	clusterID := ""
	if !data.ClusterID.IsNull() && !data.ClusterID.IsUnknown() {
		clusterID = data.ClusterID.ValueString()
	}

	// Get events from API
	eventsResp, err := d.client.GetClusterEvents(ctx, clusterID, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read cluster events, got error: %s", err),
		)
		return
	}

	// Map response to model
	events := make([]EventModel, len(eventsResp.Events))
	for i, event := range eventsResp.Events {
		events[i] = EventModel{
			Name:        types.StringValue(event.Name),
			ClusterID:   types.StringValue(event.ClusterID),
			HostID:      types.StringValue(event.HostID),
			InfraEnvID:  types.StringValue(event.InfraEnvID),
			Severity:    types.StringValue(event.Severity),
			Category:    types.StringValue(event.Category),
			Message:     types.StringValue(event.Message),
			EventTime:   types.StringValue(event.EventTime.Format("2006-01-02T15:04:05Z07:00")),
			RequestID:   types.StringValue(event.RequestID),
			Props:       types.StringValue(event.Props),
		}
	}

	// Set computed values
	data.ID = types.StringValue(fmt.Sprintf("events-%s", clusterID)) // Generate a unique ID
	data.Events = events

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}