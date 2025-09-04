package models

import (
	"time"
)

type Cluster struct {
	Kind                   string                   `json:"kind"`
	ID                     string                   `json:"id"`
	Href                   string                   `json:"href"`
	Name                   string                   `json:"name"`
	OpenshiftVersion       string                   `json:"openshift_version"`
	OCPReleaseImage        string                   `json:"ocp_release_image,omitempty"`
	OpenshiftClusterID     string                   `json:"openshift_cluster_id,omitempty"`
	BaseDNSDomain          string                   `json:"base_dns_domain,omitempty"`
	ClusterNetworkCIDR     string                   `json:"cluster_network_cidr,omitempty"`
	ClusterNetworkHostPrefix int                    `json:"cluster_network_host_prefix,omitempty"`
	ServiceNetworkCIDR     string                   `json:"service_network_cidr,omitempty"`
	ClusterNetworks        []ClusterNetwork         `json:"cluster_networks,omitempty"`
	ServiceNetworks        []ServiceNetwork         `json:"service_networks,omitempty"`
	MachineNetworks        []MachineNetwork         `json:"machine_networks,omitempty"`
	APIVips                []APIVip                 `json:"api_vips,omitempty"`
	IngressVips            []IngressVip             `json:"ingress_vips,omitempty"`
	PullSecret             string                   `json:"pull_secret"`
	SSHPublicKey           string                   `json:"ssh_public_key,omitempty"`
	VipDHCPAllocation      bool                     `json:"vip_dhcp_allocation,omitempty"`
	HTTPProxy              string                   `json:"http_proxy,omitempty"`
	HTTPSProxy             string                   `json:"https_proxy,omitempty"`
	NoProxy                string                   `json:"no_proxy,omitempty"`
	UserManagedNetworking  bool                     `json:"user_managed_networking,omitempty"`
	AdditionalNTPSource    string                   `json:"additional_ntp_source,omitempty"`
	Hyperthreading         string                   `json:"hyperthreading,omitempty"`
	Status                 string                   `json:"status"`
	StatusInfo             string                   `json:"status_info"`
	StatusUpdatedAt        time.Time                `json:"status_updated_at,omitempty"`
	CreatedAt              time.Time                `json:"created_at,omitempty"`
	UpdatedAt              time.Time                `json:"updated_at,omitempty"`
	Platform               *Platform                `json:"platform,omitempty"`
	LoadBalancer           *LoadBalancer            `json:"load_balancer,omitempty"`
	DiskEncryption         *DiskEncryption          `json:"disk_encryption,omitempty"`
	IgnitionEndpoint       *IgnitionEndpoint        `json:"ignition_endpoint,omitempty"`
	Tags                   string                   `json:"tags,omitempty"`
	OLMOperators           []OLMOperator            `json:"olm_operators,omitempty"`
	ControlPlaneCount      int                      `json:"control_plane_count,omitempty"`
	CPUArchitecture        string                   `json:"cpu_architecture,omitempty"`
	SchedulableMasters     bool                     `json:"schedulable_masters,omitempty"`
	HighAvailabilityMode   string                   `json:"high_availability_mode,omitempty"`
	NetworkType            string                   `json:"network_type,omitempty"`
	HostCount              int                      `json:"total_host_count,omitempty"`
}

type Platform struct {
	Type                string                  `json:"type,omitempty"`
	External            *ExternalPlatform       `json:"external,omitempty"`
	Baremetal           *BaremetalPlatform      `json:"baremetal,omitempty"`
	Nutanix             *NutanixPlatform        `json:"nutanix,omitempty"`
	VSphere             *VSpherePlatform        `json:"vsphere,omitempty"`
	OCI                 *OCIPlatform            `json:"oci,omitempty"`
}

type ExternalPlatform struct {
	PlatformName                  string `json:"platform_name,omitempty"`
	CloudControllerManager        string `json:"cloud_controller_manager,omitempty"`
}

type BaremetalPlatform struct {
	APIVips                       []string `json:"api_vips,omitempty"`
	IngressVips                   []string `json:"ingress_vips,omitempty"`
}

type NutanixPlatform struct {
	APIVips                       []string `json:"api_vips,omitempty"`
	IngressVips                   []string `json:"ingress_vips,omitempty"`
}

type VSpherePlatform struct {
	APIVips                       []string `json:"api_vips,omitempty"`
	IngressVips                   []string `json:"ingress_vips,omitempty"`
	VCenters                      []VCenter `json:"vcenters,omitempty"`
}

type VCenter struct {
	Server                        string `json:"server"`
	Username                      string `json:"username"`
	Password                      string `json:"password"`
	Datacenter                    string `json:"datacenter"`
	DefaultDatastore              string `json:"default_datastore"`
	Folder                        string `json:"folder,omitempty"`
	ResourcePool                  string `json:"resource_pool,omitempty"`
	Cluster                       string `json:"cluster,omitempty"`
	Network                       string `json:"network,omitempty"`
}

type OCIPlatform struct {
	APIVips                       []string `json:"api_vips,omitempty"`
	IngressVips                   []string `json:"ingress_vips,omitempty"`
}

type OLMOperator struct {
	Name                          string `json:"name"`
	Properties                    string `json:"properties,omitempty"`
}

type ClusterNetwork struct {
	CIDR       string `json:"cidr"`
	HostPrefix int    `json:"host_prefix,omitempty"`
}

type ServiceNetwork struct {
	CIDR string `json:"cidr"`
}

type MachineNetwork struct {
	CIDR string `json:"cidr"`
}

type APIVip struct {
	IP string `json:"ip"`
}

type IngressVip struct {
	IP string `json:"ip"`
}

type LoadBalancer struct {
	Type string `json:"type,omitempty"`
}

type DiskEncryption struct {
	EnableOn    string `json:"enable_on,omitempty"`
	Mode        string `json:"mode,omitempty"`
	TangServers string `json:"tang_servers,omitempty"`
}

type IgnitionEndpoint struct {
	URL       string `json:"url,omitempty"`
	CACertPEM string `json:"ca_cert_pem,omitempty"`
}

type ClusterCreateParams struct {
	Name                          string                   `json:"name"`
	OpenshiftVersion              string                   `json:"openshift_version"`
	OCPReleaseImage               string                   `json:"ocp_release_image,omitempty"`
	PullSecret                    string                   `json:"pull_secret"`
	BaseDNSDomain                 string                   `json:"base_dns_domain,omitempty"`
	ClusterNetworkCIDR            string                   `json:"cluster_network_cidr,omitempty"`
	ClusterNetworkHostPrefix      int                      `json:"cluster_network_host_prefix,omitempty"`
	ServiceNetworkCIDR            string                   `json:"service_network_cidr,omitempty"`
	ClusterNetworks               []ClusterNetwork         `json:"cluster_networks,omitempty"`
	ServiceNetworks               []ServiceNetwork         `json:"service_networks,omitempty"`
	MachineNetworks               []MachineNetwork         `json:"machine_networks,omitempty"`
	APIVips                       []APIVip                 `json:"api_vips,omitempty"`
	IngressVips                   []IngressVip             `json:"ingress_vips,omitempty"`
	SSHPublicKey                  string                   `json:"ssh_public_key,omitempty"`
	VipDHCPAllocation             bool                     `json:"vip_dhcp_allocation,omitempty"`
	HTTPProxy                     string                   `json:"http_proxy,omitempty"`
	HTTPSProxy                    string                   `json:"https_proxy,omitempty"`
	NoProxy                       string                   `json:"no_proxy,omitempty"`
	UserManagedNetworking         bool                     `json:"user_managed_networking,omitempty"`
	AdditionalNTPSource           string                   `json:"additional_ntp_source,omitempty"`
	Hyperthreading                string                   `json:"hyperthreading,omitempty"`
	Platform                      *Platform                `json:"platform,omitempty"`
	LoadBalancer                  *LoadBalancer            `json:"load_balancer,omitempty"`
	DiskEncryption                *DiskEncryption          `json:"disk_encryption,omitempty"`
	IgnitionEndpoint              *IgnitionEndpoint        `json:"ignition_endpoint,omitempty"`
	Tags                          string                   `json:"tags,omitempty"`
	OLMOperators                  []OLMOperator            `json:"olm_operators,omitempty"`
	HighAvailabilityMode          string                   `json:"high_availability_mode,omitempty"`
	NetworkType                   string                   `json:"network_type,omitempty"`
	CPUArchitecture               string                   `json:"cpu_architecture,omitempty"`
	ControlPlaneCount             int                      `json:"control_plane_count,omitempty"`
	SchedulableMasters            *bool                    `json:"schedulable_masters,omitempty"`
}

type ClusterUpdateParams struct {
	Name                          *string                  `json:"name,omitempty"`
	BaseDNSDomain                 *string                  `json:"base_dns_domain,omitempty"`
	ClusterNetworkCIDR            *string                  `json:"cluster_network_cidr,omitempty"`
	ClusterNetworkHostPrefix      *int                     `json:"cluster_network_host_prefix,omitempty"`
	ServiceNetworkCIDR            *string                  `json:"service_network_cidr,omitempty"`
	ClusterNetworks               []ClusterNetwork         `json:"cluster_networks,omitempty"`
	ServiceNetworks               []ServiceNetwork         `json:"service_networks,omitempty"`
	MachineNetworks               []MachineNetwork         `json:"machine_networks,omitempty"`
	APIVips                       []APIVip                 `json:"api_vips,omitempty"`
	IngressVips                   []IngressVip             `json:"ingress_vips,omitempty"`
	SSHPublicKey                  *string                  `json:"ssh_public_key,omitempty"`
	VipDHCPAllocation             *bool                    `json:"vip_dhcp_allocation,omitempty"`
	HTTPProxy                     *string                  `json:"http_proxy,omitempty"`
	HTTPSProxy                    *string                  `json:"https_proxy,omitempty"`
	NoProxy                       *string                  `json:"no_proxy,omitempty"`
	UserManagedNetworking         *bool                    `json:"user_managed_networking,omitempty"`
	AdditionalNTPSource           *string                  `json:"additional_ntp_source,omitempty"`
	Hyperthreading                *string                  `json:"hyperthreading,omitempty"`
	Platform                      *Platform                `json:"platform,omitempty"`
	LoadBalancer                  *LoadBalancer            `json:"load_balancer,omitempty"`
	DiskEncryption                *DiskEncryption          `json:"disk_encryption,omitempty"`
	IgnitionEndpoint              *IgnitionEndpoint        `json:"ignition_endpoint,omitempty"`
	Tags                          *string                  `json:"tags,omitempty"`
	OLMOperators                  []OLMOperator            `json:"olm_operators,omitempty"`
	PullSecret                    *string                  `json:"pull_secret,omitempty"`
	ControlPlaneCount             *int                     `json:"control_plane_count,omitempty"`
	SchedulableMasters            *bool                    `json:"schedulable_masters,omitempty"`
}