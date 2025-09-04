package provider

import (
	"context"
	"fmt"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ClusterDataSource{}

func NewClusterDataSource() datasource.DataSource {
	return &ClusterDataSource{}
}

// ClusterDataSource defines the data source implementation.
type ClusterDataSource struct {
	client *client.Client
}

// ClusterDataSourceModel describes the data source data model.
type ClusterDataSourceModel struct {
	// Required fields
	ID                    types.String   `tfsdk:"id"`
	Kind                  types.String   `tfsdk:"kind"`
	Href                  types.String   `tfsdk:"href"`
	Status                types.String   `tfsdk:"status"`
	StatusInfo            types.String   `tfsdk:"status_info"`
	
	// Core cluster info
	Name                  types.String   `tfsdk:"name"`
	UserName              types.String   `tfsdk:"user_name"`
	OrgID                 types.String   `tfsdk:"org_id"`
	EmailDomain           types.String   `tfsdk:"email_domain"`
	OpenshiftVersion      types.String   `tfsdk:"openshift_version"`
	OCPReleaseImage       types.String   `tfsdk:"ocp_release_image"`
	OpenshiftClusterID    types.String   `tfsdk:"openshift_cluster_id"`
	BaseDNSDomain         types.String   `tfsdk:"base_dns_domain"`
	CPUArchitecture       types.String   `tfsdk:"cpu_architecture"`
	
	// Network configuration
	ClusterNetworkCIDR    types.String   `tfsdk:"cluster_network_cidr"`
	ClusterNetworkHostPrefix types.Int64 `tfsdk:"cluster_network_host_prefix"`
	ServiceNetworkCIDR    types.String   `tfsdk:"service_network_cidr"`
	MachineNetworkCIDR    types.String   `tfsdk:"machine_network_cidr"`
	APIVips               []ClusterAPIVipModel  `tfsdk:"api_vips"`
	APIVipDNSName         types.String   `tfsdk:"api_vip_dns_name"`
	IngressVips           []ClusterAPIVipModel  `tfsdk:"ingress_vips"`
	NetworkType           types.String   `tfsdk:"network_type"`
	ClusterNetworks       types.List     `tfsdk:"cluster_networks"`
	ServiceNetworks       types.List     `tfsdk:"service_networks"`
	MachineNetworks       types.List     `tfsdk:"machine_networks"`
	
	// Host configuration
	ControlPlaneCount     types.Int64    `tfsdk:"control_plane_count"`
	HighAvailabilityMode  types.String   `tfsdk:"high_availability_mode"`
	SchedulableMasters    types.Bool     `tfsdk:"schedulable_masters"`
	SchedulableMastersForced types.Bool  `tfsdk:"schedulable_masters_forced_true"`
	
	// Host counts
	TotalHostCount        types.Int64    `tfsdk:"total_host_count"`
	ReadyHostCount        types.Int64    `tfsdk:"ready_host_count"`
	EnabledHostCount      types.Int64    `tfsdk:"enabled_host_count"`
	
	// Security & Access
	SSHPublicKey          types.String   `tfsdk:"ssh_public_key"`
	PullSecretSet         types.Bool     `tfsdk:"pull_secret_set"`
	
	// Proxy configuration
	HTTPProxy             types.String   `tfsdk:"http_proxy"`
	HTTPSProxy            types.String   `tfsdk:"https_proxy"`
	NoProxy               types.String   `tfsdk:"no_proxy"`
	
	// Timestamps
	StatusUpdatedAt       types.String   `tfsdk:"status_updated_at"`
	InstallStartedAt      types.String   `tfsdk:"install_started_at"`
	InstallCompletedAt    types.String   `tfsdk:"install_completed_at"`
	CreatedAt             types.String   `tfsdk:"created_at"`
	UpdatedAt             types.String   `tfsdk:"updated_at"`
	
	// Progress and validation
	Progress              types.Object   `tfsdk:"progress"`
	ValidationsInfo       types.String   `tfsdk:"validations_info"`
	
	// Logs
	LogsInfo              types.Object   `tfsdk:"logs_info"`
	ControllerLogsCollectedAt types.String `tfsdk:"controller_logs_collected_at"`
	ControllerLogsStartedAt   types.String `tfsdk:"controller_logs_started_at"`
	
	// Advanced configuration
	InstallConfigOverrides types.String  `tfsdk:"install_config_overrides"`
	DiskEncryption        types.Object   `tfsdk:"disk_encryption"`
	VipDhcpAllocation     types.Bool     `tfsdk:"vip_dhcp_allocation"`
	UserManagedNetworking types.Bool     `tfsdk:"user_managed_networking"`
	AdditionalNTPSource   types.String   `tfsdk:"additional_ntp_source"`
	Hyperthreading        types.String   `tfsdk:"hyperthreading"`
	
	// System info
	Platform              types.Object   `tfsdk:"platform"`
	ImageInfo             types.Object   `tfsdk:"image_info"`
	IgnitionEndpoint      types.Object   `tfsdk:"ignition_endpoint"`
	LoadBalancer          types.Object   `tfsdk:"load_balancer"`
	
	// Connectivity and networking details  
	ConnectivityMajorityGroups types.String `tfsdk:"connectivity_majority_groups"`
	IPCollisions          types.String   `tfsdk:"ip_collisions"`
	HostNetworks          types.List     `tfsdk:"host_networks"`
	
	// Validation overrides
	IgnoredHostValidations    types.String `tfsdk:"ignored_host_validations"`
	IgnoredClusterValidations types.String `tfsdk:"ignored_cluster_validations"`
	
	// Operators and features
	MonitoredOperators    types.List     `tfsdk:"monitored_operators"`
	FeatureUsage          types.String   `tfsdk:"feature_usage"`
	AMSSubscriptionID     types.String   `tfsdk:"ams_subscription_id"`
	
	// Day-2 and import
	Imported              types.Bool     `tfsdk:"imported"`
	Tags                  types.String   `tfsdk:"tags"`
	LastInstallationPreparation types.Object `tfsdk:"last_installation_preparation"`
	OrgSoftTimeoutsEnabled types.Bool    `tfsdk:"org_soft_timeouts_enabled"`
	
	// Legacy support
	HostsCount            types.Int64    `tfsdk:"hosts_count"`
	ReadyHostsCount       types.Int64    `tfsdk:"ready_hosts_count"`
	MastersCount          types.Int64    `tfsdk:"masters_count"`
	WorkersCount          types.Int64    `tfsdk:"workers_count"`
	InstallationStartedAt types.String   `tfsdk:"installation_started_at"`
	InstallationCompletedAt types.String `tfsdk:"installation_completed_at"`
}

// ClusterAPIVipModel represents API/Ingress VIP configuration for data source
type ClusterAPIVipModel struct {
	IP           types.String `tfsdk:"ip"`
	ClusterID    types.String `tfsdk:"cluster_id"`
	Verification types.String `tfsdk:"verification"`
}

func (d *ClusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (d *ClusterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Retrieves information about an OpenShift cluster managed by the Assisted Service. This data source provides comprehensive cluster details including status, configuration, network settings, and host counts.",

		Attributes: map[string]schema.Attribute{
			// Required fields
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the cluster (UUID)",
				Required:            true,
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "Indicates the type of this object ('Cluster' or 'AddHostsCluster')",
				Computed:            true,
			},
			"href": schema.StringAttribute{
				MarkdownDescription: "Self link to this cluster",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current cluster status (insufficient, ready, error, preparing-for-installation, pending-for-input, installing, finalizing, installed, adding-hosts, cancelled, installing-pending-user-action)",
				Computed:            true,
			},
			"status_info": schema.StringAttribute{
				MarkdownDescription: "Additional information pertaining to the status of the cluster",
				Computed:            true,
			},
			
			// Core cluster info
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the OpenShift cluster",
				Computed:            true,
			},
			"user_name": schema.StringAttribute{
				MarkdownDescription: "Username associated with the cluster",
				Computed:            true,
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: "Organization ID",
				Computed:            true,
			},
			"email_domain": schema.StringAttribute{
				MarkdownDescription: "Email domain",
				Computed:            true,
			},
			"openshift_version": schema.StringAttribute{
				MarkdownDescription: "Version of the OpenShift cluster",
				Computed:            true,
			},
			"ocp_release_image": schema.StringAttribute{
				MarkdownDescription: "OpenShift release image URI",
				Computed:            true,
			},
			"openshift_cluster_id": schema.StringAttribute{
				MarkdownDescription: "Cluster ID on OCP system (UUID)",
				Computed:            true,
			},
			"base_dns_domain": schema.StringAttribute{
				MarkdownDescription: "Base domain of the cluster. All DNS records must be sub-domains of this base and include the cluster name",
				Computed:            true,
			},
			"cpu_architecture": schema.StringAttribute{
				MarkdownDescription: "The CPU architecture of the image (x86_64, aarch64, arm64, ppc64le, s390x, multi)",
				Computed:            true,
			},
			
			// Network configuration
			"cluster_network_cidr": schema.StringAttribute{
				MarkdownDescription: "IP address block from which Pod IPs are allocated",
				Computed:            true,
			},
			"cluster_network_host_prefix": schema.Int64Attribute{
				MarkdownDescription: "The subnet prefix length to assign to each individual node",
				Computed:            true,
			},
			"service_network_cidr": schema.StringAttribute{
				MarkdownDescription: "The IP address pool to use for service IP addresses",
				Computed:            true,
			},
			"machine_network_cidr": schema.StringAttribute{
				MarkdownDescription: "A CIDR that all hosts belonging to the cluster should have interfaces with IP addresses that belong to this CIDR",
				Computed:            true,
			},
			"api_vip_dns_name": schema.StringAttribute{
				MarkdownDescription: "The domain name used to reach the OpenShift cluster API",
				Computed:            true,
			},
			"network_type": schema.StringAttribute{
				MarkdownDescription: "The desired network type used (OpenShiftSDN, OVNKubernetes)",
				Computed:            true,
			},
			"cluster_networks": schema.ListAttribute{
				MarkdownDescription: "Cluster networks that are associated with this cluster",
				Computed:            true,
				ElementType:         types.ObjectType{},
			},
			"service_networks": schema.ListAttribute{
				MarkdownDescription: "Service networks that are associated with this cluster",
				Computed:            true,
				ElementType:         types.ObjectType{},
			},
			"machine_networks": schema.ListAttribute{
				MarkdownDescription: "Machine networks that are associated with this cluster",
				Computed:            true,
				ElementType:         types.ObjectType{},
			},
			
			// VIP Configuration
			"api_vips": schema.ListNestedAttribute{
				MarkdownDescription: "The virtual IPs used to reach the OpenShift cluster's API",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							MarkdownDescription: "Virtual IP address",
							Computed:            true,
						},
						"cluster_id": schema.StringAttribute{
							MarkdownDescription: "Cluster ID this VIP belongs to",
							Computed:            true,
						},
						"verification": schema.StringAttribute{
							MarkdownDescription: "VIP verification status",
							Computed:            true,
						},
					},
				},
			},
			"ingress_vips": schema.ListNestedAttribute{
				MarkdownDescription: "The virtual IPs used for cluster ingress traffic",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							MarkdownDescription: "Virtual IP address",
							Computed:            true,
						},
						"cluster_id": schema.StringAttribute{
							MarkdownDescription: "Cluster ID this VIP belongs to",
							Computed:            true,
						},
						"verification": schema.StringAttribute{
							MarkdownDescription: "VIP verification status",
							Computed:            true,
						},
					},
				},
			},
			"vip_dhcp_allocation": schema.BoolAttribute{
				MarkdownDescription: "Indicate if virtual IP DHCP allocation mode is enabled",
				Computed:            true,
			},
			
			// Host configuration and counts
			"control_plane_count": schema.Int64Attribute{
				MarkdownDescription: "Specifies the required number of control plane nodes that should be part of the cluster",
				Computed:            true,
			},
			"high_availability_mode": schema.StringAttribute{
				MarkdownDescription: "Guaranteed availability of the installed cluster ('Full', 'None') - DEPRECATED, use control_plane_count instead",
				Computed:            true,
			},
			"schedulable_masters": schema.BoolAttribute{
				MarkdownDescription: "Schedule workloads on masters",
				Computed:            true,
			},
			"schedulable_masters_forced_true": schema.BoolAttribute{
				MarkdownDescription: "Indicates if schedule workloads on masters will be enabled regardless the value of 'schedulable_masters' property",
				Computed:            true,
			},
			"total_host_count": schema.Int64Attribute{
				MarkdownDescription: "All hosts associated to this cluster",
				Computed:            true,
			},
			"ready_host_count": schema.Int64Attribute{
				MarkdownDescription: "Hosts associated to this cluster that are in 'known' state",
				Computed:            true,
			},
			"enabled_host_count": schema.Int64Attribute{
				MarkdownDescription: "Hosts associated to this cluster that are not in 'disabled' state",
				Computed:            true,
			},
			
			// Security & Access
			"ssh_public_key": schema.StringAttribute{
				MarkdownDescription: "SSH public key for debugging OpenShift nodes",
				Computed:            true,
			},
			"pull_secret_set": schema.BoolAttribute{
				MarkdownDescription: "True if the pull secret has been added to the cluster",
				Computed:            true,
			},
			
			// Proxy configuration
			"http_proxy": schema.StringAttribute{
				MarkdownDescription: "A proxy URL to use for creating HTTP connections outside the cluster",
				Computed:            true,
			},
			"https_proxy": schema.StringAttribute{
				MarkdownDescription: "A proxy URL to use for creating HTTPS connections outside the cluster",
				Computed:            true,
			},
			"no_proxy": schema.StringAttribute{
				MarkdownDescription: "A comma-separated list of destination domain names, domains, IP addresses, or other network CIDRs to exclude from proxying",
				Computed:            true,
			},
			
			// Timestamps
			"status_updated_at": schema.StringAttribute{
				MarkdownDescription: "The last time that the cluster status was updated",
				Computed:            true,
			},
			"install_started_at": schema.StringAttribute{
				MarkdownDescription: "The time that this cluster started installation",
				Computed:            true,
			},
			"install_completed_at": schema.StringAttribute{
				MarkdownDescription: "The time that this cluster completed installation",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The time that this cluster was created",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The last time that this cluster was updated",
				Computed:            true,
			},
			
			// Progress and validation
			"progress": schema.SingleNestedAttribute{
				MarkdownDescription: "Installation progress percentages of the cluster",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"current_stage": schema.StringAttribute{
						MarkdownDescription: "Current installation stage",
						Computed:            true,
					},
					"installation_percentage": schema.Int64Attribute{
						MarkdownDescription: "Installation progress percentage",
						Computed:            true,
					},
					"stage_started_at": schema.StringAttribute{
						MarkdownDescription: "Time when current stage started",
						Computed:            true,
					},
					"stage_updated_at": schema.StringAttribute{
						MarkdownDescription: "Time when current stage was last updated",
						Computed:            true,
					},
				},
			},
			"validations_info": schema.StringAttribute{
				MarkdownDescription: "JSON-formatted string containing the validation results for each validation id grouped by category",
				Computed:            true,
			},
			
			// Advanced configuration
			"install_config_overrides": schema.StringAttribute{
				MarkdownDescription: "JSON-formatted string containing the user overrides for the install-config.yaml file",
				Computed:            true,
			},
			"disk_encryption": schema.SingleNestedAttribute{
				MarkdownDescription: "Information regarding hosts' installation disks encryption",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enable_on": schema.StringAttribute{
						MarkdownDescription: "Enable disk encryption on (none, all, masters, workers)",
						Computed:            true,
					},
					"mode": schema.StringAttribute{
						MarkdownDescription: "Disk encryption mode (tpmv2, tang)",
						Computed:            true,
					},
				},
			},
			"user_managed_networking": schema.BoolAttribute{
				MarkdownDescription: "Indicate if the networking is managed by the user (DEPRECATED)",
				Computed:            true,
			},
			"additional_ntp_source": schema.StringAttribute{
				MarkdownDescription: "A comma-separated list of NTP sources (name or IP) going to be added to all the hosts",
				Computed:            true,
			},
			"hyperthreading": schema.StringAttribute{
				MarkdownDescription: "Enable/disable hyperthreading on master nodes, worker nodes, or a combination of them",
				Computed:            true,
			},
			
			// System objects
			"platform": schema.SingleNestedAttribute{
				MarkdownDescription: "Platform configuration",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Platform type (baremetal, vsphere, etc.)",
						Computed:            true,
					},
				},
			},
			"image_info": schema.SingleNestedAttribute{
				MarkdownDescription: "Image information",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"created_at": schema.StringAttribute{
						MarkdownDescription: "Image creation timestamp",
						Computed:            true,
					},
					"expires_at": schema.StringAttribute{
						MarkdownDescription: "Image expiration timestamp",
						Computed:            true,
					},
					"download_url": schema.StringAttribute{
						MarkdownDescription: "Image download URL",
						Computed:            true,
					},
					"size_bytes": schema.Int64Attribute{
						MarkdownDescription: "Image size in bytes",
						Computed:            true,
					},
				},
			},
			"ignition_endpoint": schema.SingleNestedAttribute{
				MarkdownDescription: "Explicit ignition endpoint overrides the default ignition endpoint",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: "Ignition endpoint URL",
						Computed:            true,
					},
					"ca_certificate": schema.StringAttribute{
						MarkdownDescription: "CA certificate for ignition endpoint",
						Computed:            true,
					},
				},
			},
			"load_balancer": schema.SingleNestedAttribute{
				MarkdownDescription: "Load balancer configuration",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Load balancer type",
						Computed:            true,
					},
				},
			},
			
			// Logs
			"logs_info": schema.SingleNestedAttribute{
				MarkdownDescription: "The progress of log collection or empty if logs are not applicable",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"state": schema.StringAttribute{
						MarkdownDescription: "Log collection state",
						Computed:            true,
					},
					"state_info": schema.StringAttribute{
						MarkdownDescription: "Additional info about log collection state",
						Computed:            true,
					},
				},
			},
			"controller_logs_collected_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when controller logs were collected",
				Computed:            true,
			},
			"controller_logs_started_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when controller log collection started",
				Computed:            true,
			},
			
			// Connectivity and networking details
			"connectivity_majority_groups": schema.StringAttribute{
				MarkdownDescription: "JSON formatted string containing the majority groups for connectivity checks",
				Computed:            true,
			},
			"ip_collisions": schema.StringAttribute{
				MarkdownDescription: "JSON formatted string containing ip collisions detected in the cluster",
				Computed:            true,
			},
			"host_networks": schema.ListAttribute{
				MarkdownDescription: "List of host networks to be filled during query",
				Computed:            true,
				ElementType:         types.ObjectType{},
			},
			
			// Validation overrides
			"ignored_host_validations": schema.StringAttribute{
				MarkdownDescription: "JSON formatted string containing a list of host validations to be ignored",
				Computed:            true,
			},
			"ignored_cluster_validations": schema.StringAttribute{
				MarkdownDescription: "JSON formatted string containing a list of cluster validations to be ignored",
				Computed:            true,
			},
			
			// Operators and features
			"monitored_operators": schema.ListAttribute{
				MarkdownDescription: "Operators that are associated with this cluster",
				Computed:            true,
				ElementType:         types.ObjectType{},
			},
			"feature_usage": schema.StringAttribute{
				MarkdownDescription: "JSON-formatted string containing the usage information by feature name",
				Computed:            true,
			},
			"ams_subscription_id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the AMS subscription in OCM",
				Computed:            true,
			},
			
			// Day-2 and import
			"imported": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether this cluster is an imported day-2 cluster or a regular cluster",
				Computed:            true,
			},
			"tags": schema.StringAttribute{
				MarkdownDescription: "A comma-separated list of tags that are associated to the cluster",
				Computed:            true,
			},
			"last_installation_preparation": schema.SingleNestedAttribute{
				MarkdownDescription: "Last installation preparation information",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"reason": schema.StringAttribute{
						MarkdownDescription: "Reason for last installation preparation",
						Computed:            true,
					},
					"status": schema.StringAttribute{
						MarkdownDescription: "Status of last installation preparation",
						Computed:            true,
					},
				},
			},
			"org_soft_timeouts_enabled": schema.BoolAttribute{
				MarkdownDescription: "Indication if organization soft timeouts is enabled for the cluster",
				Computed:            true,
			},
			
			// Legacy support (backwards compatibility)
			"hosts_count": schema.Int64Attribute{
				MarkdownDescription: "Total number of hosts in the cluster (legacy field, use total_host_count)",
				Computed:            true,
			},
			"ready_hosts_count": schema.Int64Attribute{
				MarkdownDescription: "Number of hosts that are ready for installation (legacy field, use ready_host_count)",
				Computed:            true,
			},
			"masters_count": schema.Int64Attribute{
				MarkdownDescription: "Number of master/control-plane hosts (derived from host roles)",
				Computed:            true,
			},
			"workers_count": schema.Int64Attribute{
				MarkdownDescription: "Number of worker hosts (derived from host roles)",
				Computed:            true,
			},
			"installation_started_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when installation was started (legacy field, use install_started_at)",
				Computed:            true,
			},
			"installation_completed_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when installation was completed (legacy field, use install_completed_at)",
				Computed:            true,
			},
		},
	}
}

func (d *ClusterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get cluster from API
	cluster, err := d.client.GetCluster(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read cluster, got error: %s", err),
		)
		return
	}

	// Map API response to data model
	data.Name = types.StringValue(cluster.Name)
	data.BaseDNSDomain = types.StringValue(cluster.BaseDNSDomain)
	data.OpenshiftVersion = types.StringValue(cluster.OpenshiftVersion)
	data.CPUArchitecture = types.StringValue(cluster.CPUArchitecture)
	data.Status = types.StringValue(cluster.Status)
	data.StatusInfo = types.StringValue(cluster.StatusInfo)
	data.Kind = types.StringValue(cluster.Kind)
	
	// Handle platform - construct nested object
	if cluster.Platform != nil && cluster.Platform.Type != "" {
		platformObj, diag := types.ObjectValue(
			map[string]attr.Type{
				"type": types.StringType,
			},
			map[string]attr.Value{
				"type": types.StringValue(cluster.Platform.Type),
			},
		)
		resp.Diagnostics.Append(diag...)
		data.Platform = platformObj
	}

	// Handle timestamps
	// Note: InstallStartedAt and InstallCompletedAt are not in the basic cluster model
	if !cluster.CreatedAt.IsZero() {
		data.CreatedAt = types.StringValue(cluster.CreatedAt.String())
	}
	if !cluster.UpdatedAt.IsZero() {
		data.UpdatedAt = types.StringValue(cluster.UpdatedAt.String())
	}

	// Handle API VIPs
	if cluster.APIVips != nil {
		var apiVips []ClusterAPIVipModel
		for _, vip := range cluster.APIVips {
			apiVips = append(apiVips, ClusterAPIVipModel{
				IP:           types.StringValue(vip.IP),
				ClusterID:    data.ID, // Use the cluster ID from the data
				Verification: types.StringValue(""), // Not available in basic cluster model
			})
		}
		data.APIVips = apiVips
	}

	// Handle Ingress VIPs  
	if cluster.IngressVips != nil {
		var ingressVips []ClusterAPIVipModel
		for _, vip := range cluster.IngressVips {
			ingressVips = append(ingressVips, ClusterAPIVipModel{
				IP:           types.StringValue(vip.IP),
				ClusterID:    data.ID, // Use the cluster ID from the data
				Verification: types.StringValue(""), // Not available in basic cluster model
			})
		}
		data.IngressVips = ingressVips
	}

	// Handle network configuration
	data.ClusterNetworkCIDR = types.StringValue(cluster.ClusterNetworkCIDR)
	data.ServiceNetworkCIDR = types.StringValue(cluster.ServiceNetworkCIDR)
	// MachineNetworkCIDR is not a direct field in the cluster model

	// Handle host counts
	data.HostsCount = types.Int64Value(int64(cluster.HostCount))
	// ReadyHostsCount, MastersCount, WorkersCount are not available in basic cluster model

	// Handle boolean flags
	data.UserManagedNetworking = types.BoolValue(cluster.UserManagedNetworking)
	data.VipDhcpAllocation = types.BoolValue(cluster.VipDHCPAllocation)
	data.SchedulableMasters = types.BoolValue(cluster.SchedulableMasters)

	// Control plane count (fallback to high availability mode if needed)
	if cluster.ControlPlaneCount > 0 {
		data.ControlPlaneCount = types.Int64Value(int64(cluster.ControlPlaneCount))
	} else if cluster.HighAvailabilityMode != "" {
		// Legacy fallback
		if cluster.HighAvailabilityMode == "Full" {
			data.ControlPlaneCount = types.Int64Value(3)
		} else {
			data.ControlPlaneCount = types.Int64Value(1)
		}
	}

	// Handle href
	data.Href = types.StringValue(cluster.Href)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}