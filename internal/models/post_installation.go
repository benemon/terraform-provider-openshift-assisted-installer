package models

import "time"

// Credentials represents cluster admin credentials
type Credentials struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	ConsoleURL string `json:"console_url"`
}

// Event represents a cluster or host event
type Event struct {
	Name        string    `json:"name,omitempty"`
	ClusterID   string    `json:"cluster_id,omitempty"`
	HostID      string    `json:"host_id,omitempty"`
	InfraEnvID  string    `json:"infra_env_id,omitempty"`
	Severity    string    `json:"severity"`
	Category    string    `json:"category,omitempty"`
	Message     string    `json:"message"`
	EventTime   time.Time `json:"event_time"`
	RequestID   string    `json:"request_id,omitempty"`
	Props       string    `json:"props,omitempty"`
}

// EventsResponse represents the response from the events API
type EventsResponse struct {
	Events []Event `json:"events,omitempty"`
}

// LogsState represents the state of log collection
type LogsState string

const (
	LogsStateRequested  LogsState = "requested"
	LogsStateCollecting LogsState = "collecting"
	LogsStateCompleted  LogsState = "completed"
	LogsStateTimeout    LogsState = "timeout"
	LogsStateEmpty      LogsState = ""
)