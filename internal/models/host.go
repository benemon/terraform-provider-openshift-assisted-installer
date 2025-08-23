package models

import (
	"time"
)

type Host struct {
	Kind        string    `json:"kind"`
	ID          string    `json:"id"`
	Href        string    `json:"href"`
	ClusterID   string    `json:"cluster_id,omitempty"`
	InfraEnvID  string    `json:"infra_env_id"`
	Status      string    `json:"status"`
	StatusInfo  string    `json:"status_info"`
	Progress    *Progress `json:"progress,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	RequestedHostname string `json:"requested_hostname,omitempty"`
	Role              string `json:"role,omitempty"`
}

type Progress struct {
	CurrentStage        string `json:"current_stage,omitempty"`
	ProgressInfo        string `json:"progress_info,omitempty"`
	StageStartedAt      time.Time `json:"stage_started_at,omitempty"`
	StageUpdatedAt      time.Time `json:"stage_updated_at,omitempty"`
}

type HostUpdateParams struct {
	RequestedHostname *string `json:"requested_hostname,omitempty"`
	Role              *string `json:"role,omitempty"`
}

type BindHostParams struct {
	ClusterID string `json:"cluster_id"`
}