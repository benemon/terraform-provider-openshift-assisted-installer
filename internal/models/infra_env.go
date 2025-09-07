package models

import (
	"time"
)

type InfraEnv struct {
	Kind                  string    `json:"kind"`
	ID                    string    `json:"id"`
	Href                  string    `json:"href"`
	Name                  string    `json:"name"`
	OpenshiftVersion      string    `json:"openshift_version"`
	CPUArchitecture       string    `json:"cpu_architecture,omitempty"`
	ClusterID             string    `json:"cluster_id,omitempty"`
	SSHAuthorizedKey      string    `json:"ssh_authorized_key,omitempty"`
	PullSecretSet         bool      `json:"pull_secret_set"`
	StaticNetworkConfig   string    `json:"static_network_config,omitempty"`
	AdditionalNTPSources  string    `json:"additional_ntp_sources,omitempty"`
	AdditionalTrustBundle string    `json:"additional_trust_bundle,omitempty"`
	Proxy                 *Proxy    `json:"proxy,omitempty"`
	Type                  string    `json:"type"`
	CreatedAt             time.Time `json:"created_at,omitempty"`
	UpdatedAt             time.Time `json:"updated_at,omitempty"`
	DownloadURL           string    `json:"download_url,omitempty"`
	ExpiresAt             time.Time `json:"expires_at,omitempty"`
	SizeBytes             int64     `json:"size_bytes,omitempty"`
}

type Proxy struct {
	HTTPProxy  string `json:"http_proxy,omitempty"`
	HTTPSProxy string `json:"https_proxy,omitempty"`
	NoProxy    string `json:"no_proxy,omitempty"`
}

type InfraEnvCreateParams struct {
	Name                   string                    `json:"name"`
	PullSecret             string                    `json:"pull_secret"`
	OpenshiftVersion       string                    `json:"openshift_version,omitempty"`
	CPUArchitecture        string                    `json:"cpu_architecture,omitempty"`
	ClusterID              string                    `json:"cluster_id,omitempty"`
	SSHAuthorizedKey       string                    `json:"ssh_authorized_key,omitempty"`
	StaticNetworkConfig    []HostStaticNetworkConfig `json:"static_network_config,omitempty"`
	AdditionalNTPSources   string                    `json:"additional_ntp_sources,omitempty"`
	AdditionalTrustBundle  string                    `json:"additional_trust_bundle,omitempty"`
	Proxy                  *Proxy                    `json:"proxy,omitempty"`
	IgnitionConfigOverride string                    `json:"ignition_config_override,omitempty"`
	ImageType              string                    `json:"image_type,omitempty"`
	KernelArguments        []KernelArgument          `json:"kernel_arguments,omitempty"`
}

type InfraEnvUpdateParams struct {
	Name                   *string                   `json:"name,omitempty"`
	PullSecret             *string                   `json:"pull_secret,omitempty"`
	SSHAuthorizedKey       *string                   `json:"ssh_authorized_key,omitempty"`
	StaticNetworkConfig    []HostStaticNetworkConfig `json:"static_network_config,omitempty"`
	AdditionalNTPSources   *string                   `json:"additional_ntp_sources,omitempty"`
	AdditionalTrustBundle  *string                   `json:"additional_trust_bundle,omitempty"`
	Proxy                  *Proxy                    `json:"proxy,omitempty"`
	IgnitionConfigOverride *string                   `json:"ignition_config_override,omitempty"`
	ImageType              *string                   `json:"image_type,omitempty"`
	KernelArguments        []KernelArgument          `json:"kernel_arguments,omitempty"`
}

// HostStaticNetworkConfig represents static network configuration for a host
type HostStaticNetworkConfig struct {
	NetworkYAML     string                 `json:"network_yaml"`
	MACInterfaceMap []MACInterfaceMapEntry `json:"mac_interface_map,omitempty"`
}

// MACInterfaceMapEntry maps MAC addresses to logical interface names
type MACInterfaceMapEntry struct {
	MACAddress     string `json:"mac_address"`
	LogicalNICName string `json:"logical_nic_name"`
}

// KernelArgument represents a kernel argument modification
type KernelArgument struct {
	Operation string `json:"operation"` // append, replace, delete
	Value     string `json:"value"`
}
