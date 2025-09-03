package models

// Bundle represents an operator bundle per Swagger specification
type Bundle struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Operators []string `json:"operators"`
}

// Bundles represents a list of bundles
type Bundles []Bundle