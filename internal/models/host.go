package models

import (
	"time"
)

type Host struct {
	Kind                          string                           `json:"kind"`
	ID                            string                           `json:"id"`
	Href                          string                           `json:"href"`
	ClusterID                     string                           `json:"cluster_id,omitempty"`
	InfraEnvID                    string                           `json:"infra_env_id"`
	Status                        string                           `json:"status"`
	StatusInfo                    string                           `json:"status_info"`
	Progress                      *Progress                        `json:"progress,omitempty"`
	CreatedAt                     time.Time                        `json:"created_at,omitempty"`
	UpdatedAt                     time.Time                        `json:"updated_at,omitempty"`
	RequestedHostname             string                           `json:"requested_hostname,omitempty"`
	HostName                      string                           `json:"host_name,omitempty"`
	Role                          string                           `json:"role,omitempty"`
	DisksSelectedConfig           []DiskConfig                     `json:"disks_selected_config,omitempty"`
	DisksSkipFormatting           []DiskSkipFormatting             `json:"disks_skip_formatting,omitempty"`
	MachineConfigPoolName         string                           `json:"machine_config_pool_name,omitempty"`
	IgnitionEndpointToken         string                           `json:"ignition_endpoint_token,omitempty"`
	IgnitionEndpointHTTPHeaders   []IgnitionEndpointHTTPHeader     `json:"ignition_endpoint_http_headers,omitempty"`
	NodeLabels                    []NodeLabel                      `json:"node_labels,omitempty"`
}

type Progress struct {
	CurrentStage        string `json:"current_stage,omitempty"`
	ProgressInfo        string `json:"progress_info,omitempty"`
	StageStartedAt      time.Time `json:"stage_started_at,omitempty"`
	StageUpdatedAt      time.Time `json:"stage_updated_at,omitempty"`
}

type DiskConfig struct {
	ID   string `json:"id"`
	Role string `json:"role,omitempty"`
}

type DiskSkipFormatting struct {
	DiskID string `json:"disk_id"`
}

type IgnitionEndpointHTTPHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type NodeLabel struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type HostUpdateParams struct {
	RequestedHostname            *string                        `json:"requested_hostname,omitempty"`
	HostName                     *string                        `json:"host_name,omitempty"`
	Role                         *string                        `json:"role,omitempty"`
	DisksSelectedConfig          []DiskConfig                   `json:"disks_selected_config,omitempty"`
	DisksSkipFormatting          []DiskSkipFormatting           `json:"disks_skip_formatting,omitempty"`
	MachineConfigPoolName        *string                        `json:"machine_config_pool_name,omitempty"`
	IgnitionEndpointToken        *string                        `json:"ignition_endpoint_token,omitempty"`
	IgnitionEndpointHTTPHeaders  []IgnitionEndpointHTTPHeader   `json:"ignition_endpoint_http_headers,omitempty"`
	NodeLabels                   []NodeLabel                    `json:"node_labels,omitempty"`
}

type BindHostParams struct {
	ClusterID string `json:"cluster_id"`
}