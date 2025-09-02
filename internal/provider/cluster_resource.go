package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
)

var _ resource.Resource = &ClusterResource{}
var _ resource.ResourceWithImportState = &ClusterResource{}

type OLMOperatorModel struct {
	Name       types.String `tfsdk:"name"`
	Properties types.String `tfsdk:"properties"`
}

func NewClusterResource() resource.Resource {
	return &ClusterResource{}
}

type ClusterResource struct {
	client *client.Client
}

type ClusterResourceModel struct {
	Timeouts                     timeouts.Value `tfsdk:"timeouts"`
	ID                           types.String   `tfsdk:"id"`
	Name                         types.String   `tfsdk:"name"`
	OpenshiftVersion             types.String   `tfsdk:"openshift_version"`
	PullSecret                   types.String   `tfsdk:"pull_secret"`
	CPUArchitecture              types.String   `tfsdk:"cpu_architecture"`
	BaseDNSDomain                types.String   `tfsdk:"base_dns_domain"`
	ClusterNetworkCIDR           types.String   `tfsdk:"cluster_network_cidr"`
	ClusterNetworkHostPrefix     types.Int64    `tfsdk:"cluster_network_host_prefix"`
	ServiceNetworkCIDR           types.String   `tfsdk:"service_network_cidr"`
	APIVips                      types.List     `tfsdk:"api_vips"`
	IngressVips                  types.List     `tfsdk:"ingress_vips"`
	SSHPublicKey                 types.String   `tfsdk:"ssh_public_key"`
	VipDHCPAllocation            types.Bool     `tfsdk:"vip_dhcp_allocation"`
	HTTPProxy                    types.String   `tfsdk:"http_proxy"`
	HTTPSProxy                   types.String   `tfsdk:"https_proxy"`
	NoProxy                      types.String   `tfsdk:"no_proxy"`
	UserManagedNetworking        types.Bool     `tfsdk:"user_managed_networking"`
	AdditionalNTPSource          types.String   `tfsdk:"additional_ntp_source"`
	Hyperthreading               types.String   `tfsdk:"hyperthreading"`
	ControlPlaneCount            types.Int64    `tfsdk:"control_plane_count"`
	HighAvailabilityMode         types.String   `tfsdk:"high_availability_mode"`
	NetworkType                  types.String   `tfsdk:"network_type"`
	OLMOperators                 types.List     `tfsdk:"olm_operators"`
	TriggerInstallation          types.Bool     `tfsdk:"trigger_installation"`
	Status                       types.String   `tfsdk:"status"`
	StatusInfo                   types.String   `tfsdk:"status_info"`
	InstallCompleted             types.Bool     `tfsdk:"install_completed"`
	Kind                         types.String   `tfsdk:"kind"`
	Href                         types.String   `tfsdk:"href"`
}

func (r *ClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *ClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenShift cluster using the Assisted Service API",

		Attributes: map[string]schema.Attribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
			"id": schema.StringAttribute{
				MarkdownDescription: "Cluster identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Cluster name",
				Required:            true,
			},
			"openshift_version": schema.StringAttribute{
				MarkdownDescription: "OpenShift version to install",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pull_secret": schema.StringAttribute{
				MarkdownDescription: "Pull secret from Red Hat",
				Required:            true,
				Sensitive:           true,
			},
			"base_dns_domain": schema.StringAttribute{
				MarkdownDescription: "Base DNS domain for the cluster",
				Optional:            true,
				Computed:            true,
			},
			"cluster_network_cidr": schema.StringAttribute{
				MarkdownDescription: "CIDR range for pod network",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("10.128.0.0/14"),
			},
			"cluster_network_host_prefix": schema.Int64Attribute{
				MarkdownDescription: "Host subnet prefix length for pod network",
				Optional:            true,
				Computed:            true,
			},
			"service_network_cidr": schema.StringAttribute{
				MarkdownDescription: "CIDR range for service network",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("172.30.0.0/16"),
			},
			"api_vips": schema.ListAttribute{
				MarkdownDescription: "Virtual IPs for API servers",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"ingress_vips": schema.ListAttribute{
				MarkdownDescription: "Virtual IPs for ingress",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"ssh_public_key": schema.StringAttribute{
				MarkdownDescription: "SSH public key for cluster access",
				Optional:            true,
			},
			"vip_dhcp_allocation": schema.BoolAttribute{
				MarkdownDescription: "Enable DHCP VIP allocation",
				Optional:            true,
				Computed:            true,
			},
			"http_proxy": schema.StringAttribute{
				MarkdownDescription: "HTTP proxy URL",
				Optional:            true,
			},
			"https_proxy": schema.StringAttribute{
				MarkdownDescription: "HTTPS proxy URL",
				Optional:            true,
			},
			"no_proxy": schema.StringAttribute{
				MarkdownDescription: "Comma-separated list of hosts to bypass proxy",
				Optional:            true,
			},
			"user_managed_networking": schema.BoolAttribute{
				MarkdownDescription: "Enable user-managed networking",
				Optional:            true,
				Computed:            true,
			},
			"additional_ntp_source": schema.StringAttribute{
				MarkdownDescription: "Additional NTP source",
				Optional:            true,
			},
			"hyperthreading": schema.StringAttribute{
				MarkdownDescription: "Hyperthreading configuration (Enabled/Disabled)",
				Optional:            true,
				Computed:            true,
			},
			"high_availability_mode": schema.StringAttribute{
				MarkdownDescription: "High availability mode (Full/None)",
				Optional:            true,
				Computed:            true,
			},
			"network_type": schema.StringAttribute{
				MarkdownDescription: "Network type (OpenShiftSDN/OVNKubernetes)",
				Optional:            true,
				Computed:            true,
			},
			"cpu_architecture": schema.StringAttribute{
				MarkdownDescription: "CPU architecture (x86_64/arm64/ppc64le/s390x/multi) - required field",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"control_plane_count": schema.Int64Attribute{
				MarkdownDescription: "Number of control plane nodes (1 for SNO, 3/4/5 for multi-node). Replaces high_availability_mode.",
				Optional:            true,
				Computed:            true,
			},
			"olm_operators": schema.ListNestedAttribute{
				MarkdownDescription: "OLM operators to install during cluster deployment",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Operator name",
							Required:            true,
						},
						"properties": schema.StringAttribute{
							MarkdownDescription: "Operator properties (JSON string)",
							Optional:            true,
						},
					},
				},
			},
			"trigger_installation": schema.BoolAttribute{
				MarkdownDescription: "Whether to trigger cluster installation when cluster is ready. Defaults to true.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current cluster status",
				Computed:            true,
			},
			"status_info": schema.StringAttribute{
				MarkdownDescription: "Detailed status information",
				Computed:            true,
			},
			"install_completed": schema.BoolAttribute{
				MarkdownDescription: "Whether cluster installation has completed",
				Computed:            true,
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "Cluster kind",
				Computed:            true,
			},
			"href": schema.StringAttribute{
				MarkdownDescription: "Cluster href",
				Computed:            true,
			},
		},
	}
}

func (r *ClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ClusterResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 90*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// Convert Terraform model to API model
	createParams := r.modelToCreateParams(data)

	tflog.Info(ctx, "Creating cluster", map[string]interface{}{
		"name":              createParams.Name,
		"openshift_version": createParams.OpenshiftVersion,
	})

	// Create cluster
	cluster, err := r.client.CreateCluster(ctx, createParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating cluster",
			fmt.Sprintf("Could not create cluster: %s", err),
		)
		return
	}

	// Update state with created cluster data
	r.updateModelFromCluster(&data, cluster)

	tflog.Info(ctx, "Cluster created successfully", map[string]interface{}{
		"id":     cluster.ID,
		"status": cluster.Status,
	})

	// Check if we should trigger installation
	triggerInstallation := true
	if !data.TriggerInstallation.IsNull() {
		triggerInstallation = data.TriggerInstallation.ValueBool()
	}

	if triggerInstallation {
		// Wait for cluster to be ready for installation, then trigger it
		err = r.waitForInstallationReadyAndTrigger(ctx, cluster.ID, createTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error during cluster installation",
				fmt.Sprintf("Cluster created but installation failed: %s", err),
			)
			// Still save the state so user can see the cluster was created
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}

		// Update cluster state after installation
		updatedCluster, err := r.client.GetCluster(ctx, cluster.ID)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Could not refresh cluster state after installation",
				fmt.Sprintf("Installation may have succeeded but could not refresh state: %s", err),
			)
		} else {
			r.updateModelFromCluster(&data, updatedCluster)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ClusterResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := data.ID.ValueString()
	cluster, err := r.client.GetCluster(ctx, clusterID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading cluster",
			fmt.Sprintf("Could not read cluster %s: %s", clusterID, err),
		)
		return
	}

	r.updateModelFromCluster(&data, cluster)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ClusterResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := data.Timeouts.Update(ctx, 30*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	clusterID := data.ID.ValueString()
	updateParams := r.modelToUpdateParams(data)

	tflog.Info(ctx, "Updating cluster", map[string]interface{}{
		"id": clusterID,
	})

	cluster, err := r.client.UpdateCluster(ctx, clusterID, updateParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating cluster",
			fmt.Sprintf("Could not update cluster %s: %s", clusterID, err),
		)
		return
	}

	r.updateModelFromCluster(&data, cluster)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ClusterResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := data.ID.ValueString()

	tflog.Info(ctx, "Deleting cluster", map[string]interface{}{
		"id": clusterID,
	})

	err := r.client.DeleteCluster(ctx, clusterID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting cluster",
			fmt.Sprintf("Could not delete cluster %s: %s", clusterID, err),
		)
		return
	}
}

func (r *ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ClusterResource) modelToCreateParams(data ClusterResourceModel) models.ClusterCreateParams {
	params := models.ClusterCreateParams{
		Name:             data.Name.ValueString(),
		OpenshiftVersion: data.OpenshiftVersion.ValueString(),
		PullSecret:       data.PullSecret.ValueString(),
	}

	if !data.BaseDNSDomain.IsNull() {
		params.BaseDNSDomain = data.BaseDNSDomain.ValueString()
	}
	if !data.ClusterNetworkCIDR.IsNull() {
		params.ClusterNetworkCIDR = data.ClusterNetworkCIDR.ValueString()
	}
	if !data.ClusterNetworkHostPrefix.IsNull() {
		params.ClusterNetworkHostPrefix = int(data.ClusterNetworkHostPrefix.ValueInt64())
	}
	if !data.ServiceNetworkCIDR.IsNull() {
		params.ServiceNetworkCIDR = data.ServiceNetworkCIDR.ValueString()
	}
	if !data.SSHPublicKey.IsNull() {
		params.SSHPublicKey = data.SSHPublicKey.ValueString()
	}
	if !data.VipDHCPAllocation.IsNull() {
		params.VipDHCPAllocation = data.VipDHCPAllocation.ValueBool()
	}
	if !data.HTTPProxy.IsNull() {
		params.HTTPProxy = data.HTTPProxy.ValueString()
	}
	if !data.HTTPSProxy.IsNull() {
		params.HTTPSProxy = data.HTTPSProxy.ValueString()
	}
	if !data.NoProxy.IsNull() {
		params.NoProxy = data.NoProxy.ValueString()
	}
	if !data.UserManagedNetworking.IsNull() {
		params.UserManagedNetworking = data.UserManagedNetworking.ValueBool()
	}
	if !data.AdditionalNTPSource.IsNull() {
		params.AdditionalNTPSource = data.AdditionalNTPSource.ValueString()
	}
	if !data.Hyperthreading.IsNull() {
		params.Hyperthreading = data.Hyperthreading.ValueString()
	}
	if !data.HighAvailabilityMode.IsNull() {
		params.HighAvailabilityMode = data.HighAvailabilityMode.ValueString()
	}
	if !data.NetworkType.IsNull() {
		params.NetworkType = data.NetworkType.ValueString()
	}
	if !data.CPUArchitecture.IsNull() {
		params.CPUArchitecture = data.CPUArchitecture.ValueString()
	}
	if !data.ControlPlaneCount.IsNull() {
		params.ControlPlaneCount = int(data.ControlPlaneCount.ValueInt64())
	}

	// Convert OLM operators
	if !data.OLMOperators.IsNull() {
		var operators []OLMOperatorModel
		data.OLMOperators.ElementsAs(context.Background(), &operators, false)
		params.OLMOperators = make([]models.OLMOperator, len(operators))
		for i, op := range operators {
			params.OLMOperators[i] = models.OLMOperator{
				Name:       op.Name.ValueString(),
				Properties: op.Properties.ValueString(),
			}
		}
	}

	// Convert API VIPs
	if !data.APIVips.IsNull() {
		elements := make([]types.String, 0, len(data.APIVips.Elements()))
		data.APIVips.ElementsAs(context.Background(), &elements, false)
		params.APIVips = make([]string, len(elements))
		for i, elem := range elements {
			params.APIVips[i] = elem.ValueString()
		}
	}

	// Convert Ingress VIPs
	if !data.IngressVips.IsNull() {
		elements := make([]types.String, 0, len(data.IngressVips.Elements()))
		data.IngressVips.ElementsAs(context.Background(), &elements, false)
		params.IngressVips = make([]string, len(elements))
		for i, elem := range elements {
			params.IngressVips[i] = elem.ValueString()
		}
	}

	return params
}

func (r *ClusterResource) modelToUpdateParams(data ClusterResourceModel) models.ClusterUpdateParams {
	params := models.ClusterUpdateParams{}

	if !data.Name.IsNull() {
		name := data.Name.ValueString()
		params.Name = &name
	}
	if !data.BaseDNSDomain.IsNull() {
		domain := data.BaseDNSDomain.ValueString()
		params.BaseDNSDomain = &domain
	}
	if !data.SSHPublicKey.IsNull() {
		key := data.SSHPublicKey.ValueString()
		params.SSHPublicKey = &key
	}
	if !data.HTTPProxy.IsNull() {
		proxy := data.HTTPProxy.ValueString()
		params.HTTPProxy = &proxy
	}
	if !data.HTTPSProxy.IsNull() {
		proxy := data.HTTPSProxy.ValueString()
		params.HTTPSProxy = &proxy
	}
	if !data.NoProxy.IsNull() {
		noProxy := data.NoProxy.ValueString()
		params.NoProxy = &noProxy
	}
	if !data.AdditionalNTPSource.IsNull() {
		ntp := data.AdditionalNTPSource.ValueString()
		params.AdditionalNTPSource = &ntp
	}
	if !data.PullSecret.IsNull() {
		secret := data.PullSecret.ValueString()
		params.PullSecret = &secret
	}

	return params
}

func (r *ClusterResource) updateModelFromCluster(data *ClusterResourceModel, cluster *models.Cluster) {
	data.ID = types.StringValue(cluster.ID)
	data.Name = types.StringValue(cluster.Name)
	data.OpenshiftVersion = types.StringValue(cluster.OpenshiftVersion)
	data.Status = types.StringValue(cluster.Status)
	data.StatusInfo = types.StringValue(cluster.StatusInfo)
	data.Kind = types.StringValue(cluster.Kind)
	data.Href = types.StringValue(cluster.Href)

	// Set install completed based on status
	data.InstallCompleted = types.BoolValue(cluster.Status == "installed")

	if cluster.BaseDNSDomain != "" {
		data.BaseDNSDomain = types.StringValue(cluster.BaseDNSDomain)
	}
	if cluster.ClusterNetworkCIDR != "" {
		data.ClusterNetworkCIDR = types.StringValue(cluster.ClusterNetworkCIDR)
	}
	if cluster.ClusterNetworkHostPrefix > 0 {
		data.ClusterNetworkHostPrefix = types.Int64Value(int64(cluster.ClusterNetworkHostPrefix))
	}
	if cluster.ServiceNetworkCIDR != "" {
		data.ServiceNetworkCIDR = types.StringValue(cluster.ServiceNetworkCIDR)
	}
	if cluster.SSHPublicKey != "" {
		data.SSHPublicKey = types.StringValue(cluster.SSHPublicKey)
	}
	if cluster.HTTPProxy != "" {
		data.HTTPProxy = types.StringValue(cluster.HTTPProxy)
	}
	if cluster.HTTPSProxy != "" {
		data.HTTPSProxy = types.StringValue(cluster.HTTPSProxy)
	}
	if cluster.NoProxy != "" {
		data.NoProxy = types.StringValue(cluster.NoProxy)
	}
	if cluster.AdditionalNTPSource != "" {
		data.AdditionalNTPSource = types.StringValue(cluster.AdditionalNTPSource)
	}
	if cluster.Hyperthreading != "" {
		data.Hyperthreading = types.StringValue(cluster.Hyperthreading)
	}

	data.VipDHCPAllocation = types.BoolValue(cluster.VipDHCPAllocation)
	data.UserManagedNetworking = types.BoolValue(cluster.UserManagedNetworking)
	
	if cluster.ControlPlaneCount > 0 {
		data.ControlPlaneCount = types.Int64Value(int64(cluster.ControlPlaneCount))
	}

	// Set CPU architecture
	if cluster.CPUArchitecture != "" {
		data.CPUArchitecture = types.StringValue(cluster.CPUArchitecture)
	}

	// Convert OLM operators
	if len(cluster.OLMOperators) > 0 {
		operators := make([]OLMOperatorModel, len(cluster.OLMOperators))
		for i, op := range cluster.OLMOperators {
			operators[i] = OLMOperatorModel{
				Name:       types.StringValue(op.Name),
				Properties: types.StringValue(op.Properties),
			}
		}
		listValue, _ := types.ListValueFrom(context.Background(), types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":       types.StringType,
				"properties": types.StringType,
			},
		}, operators)
		data.OLMOperators = listValue
	}

	// Convert API VIPs
	if len(cluster.APIVips) > 0 {
		elements := make([]types.String, len(cluster.APIVips))
		for i, vip := range cluster.APIVips {
			elements[i] = types.StringValue(vip)
		}
		listValue, _ := types.ListValueFrom(context.Background(), types.StringType, elements)
		data.APIVips = listValue
	}

	// Convert Ingress VIPs
	if len(cluster.IngressVips) > 0 {
		elements := make([]types.String, len(cluster.IngressVips))
		for i, vip := range cluster.IngressVips {
			elements[i] = types.StringValue(vip)
		}
		listValue, _ := types.ListValueFrom(context.Background(), types.StringType, elements)
		data.IngressVips = listValue
	}
}

// waitForInstallationReadyAndTrigger waits for cluster to be ready, then triggers installation and waits for completion
func (r *ClusterResource) waitForInstallationReadyAndTrigger(ctx context.Context, clusterID string, timeout time.Duration) error {
	// Create a context with the remaining timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	tflog.Info(ctx, "Waiting for cluster to be ready for installation", map[string]interface{}{
		"cluster_id": clusterID,
	})

	// Use half the timeout for waiting for ready state, and half for installation
	readyTimeout := timeout / 2
	if readyTimeout > 10*time.Minute {
		readyTimeout = 10 * time.Minute
	}

	// Wait for cluster to be in a state where it can be installed
	err := r.waitForClusterState(ctx, clusterID, []string{"ready"}, readyTimeout)
	if err != nil {
		return fmt.Errorf("cluster did not become ready for installation: %w", err)
	}

	tflog.Info(ctx, "Cluster is ready, triggering installation", map[string]interface{}{
		"cluster_id": clusterID,
	})

	// Trigger installation
	err = r.client.InstallCluster(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to trigger cluster installation: %w", err)
	}

	tflog.Info(ctx, "Installation triggered, waiting for completion", map[string]interface{}{
		"cluster_id": clusterID,
	})

	// Wait for installation to complete
	err = r.waitForClusterState(ctx, clusterID, []string{"installed"}, timeout)
	if err != nil {
		return fmt.Errorf("cluster installation did not complete successfully: %w", err)
	}

	tflog.Info(ctx, "Cluster installation completed successfully", map[string]interface{}{
		"cluster_id": clusterID,
	})

	return nil
}

// waitForClusterState polls the cluster until it reaches one of the target states or times out
func (r *ClusterResource) waitForClusterState(ctx context.Context, clusterID string, targetStates []string, pollTimeout time.Duration) error {
	pollCtx, cancel := context.WithTimeout(ctx, pollTimeout)
	defer cancel()

	// Use a shorter poll interval for short timeouts (like in tests)
	pollInterval := 30 * time.Second
	if pollTimeout < 10*time.Second {
		pollInterval = 10 * time.Millisecond
	} else if pollTimeout < 1*time.Minute {
		pollInterval = 100 * time.Millisecond
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Check immediately before waiting for the first tick
	cluster, err := r.client.GetCluster(pollCtx, clusterID)
	if err == nil {
		tflog.Debug(pollCtx, "Initial cluster state check", map[string]interface{}{
			"cluster_id":     clusterID,
			"current_state":  cluster.Status,
			"target_states":  targetStates,
			"status_info":    cluster.StatusInfo,
		})

		// Check if we've already reached a target state
		for _, targetState := range targetStates {
			if cluster.Status == targetState {
				return nil
			}
		}

		// Check for error states
		errorStates := []string{"error", "cancelled"}
		for _, errorState := range errorStates {
			if cluster.Status == errorState {
				return fmt.Errorf("cluster reached error state: %s - %s", cluster.Status, cluster.StatusInfo)
			}
		}
	}

	for {
		select {
		case <-pollCtx.Done():
			return fmt.Errorf("timeout waiting for cluster to reach states %v", targetStates)
		case <-ticker.C:
			cluster, err := r.client.GetCluster(pollCtx, clusterID)
			if err != nil {
				tflog.Warn(pollCtx, "Failed to get cluster status during polling", map[string]interface{}{
					"cluster_id": clusterID,
					"error":      err.Error(),
				})
				continue
			}

			tflog.Debug(pollCtx, "Polling cluster state", map[string]interface{}{
				"cluster_id":     clusterID,
				"current_state":  cluster.Status,
				"target_states":  targetStates,
				"status_info":    cluster.StatusInfo,
			})

			// Check if we've reached a target state
			for _, targetState := range targetStates {
				if cluster.Status == targetState {
					return nil
				}
			}

			// Check for error states that we should not continue polling for
			errorStates := []string{"error", "cancelled"}
			for _, errorState := range errorStates {
				if cluster.Status == errorState {
					return fmt.Errorf("cluster reached error state: %s - %s", cluster.Status, cluster.StatusInfo)
				}
			}
		}
	}
}