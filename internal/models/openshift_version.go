package models

type OpenshiftVersions map[string]OpenshiftVersion

type OpenshiftVersion struct {
	DisplayName      string   `json:"display_name"`
	SupportLevel     string   `json:"support_level"`
	Default          bool     `json:"default,omitempty"`
	CPUArchitectures []string `json:"cpu_architectures,omitempty"`
}
