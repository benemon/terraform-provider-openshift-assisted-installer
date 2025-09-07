package models

// SupportedFeatures represents the support levels for features
type SupportedFeatures map[string]string

// SupportedFeaturesResponse represents the API response wrapper for support levels
type SupportedFeaturesResponse struct {
	Features SupportedFeatures `json:"features"`
}

// SupportedArchitecturesResponse represents the API response wrapper for architecture support
type SupportedArchitecturesResponse struct {
	Architectures SupportedArchitectures `json:"architectures"`
}

// DetailedSupportedFeatures represents detailed feature support information
type DetailedSupportedFeatures map[string]DetailedFeature

// DetailedFeature represents detailed information about a feature
type DetailedFeature struct {
	SupportLevel      string                 `json:"support_level"`
	Incompatibilities []string               `json:"incompatibilities,omitempty"`
	Dependencies      []string               `json:"dependencies,omitempty"`
	Properties        map[string]interface{} `json:"properties,omitempty"`
}

// SupportedArchitectures represents architecture support levels
type SupportedArchitectures map[string]string
