package models

type Manifest struct {
	Folder         string `json:"folder"`
	FileName       string `json:"file_name"`
	ManifestSource string `json:"manifest_source,omitempty"`
}

type CreateManifestParams struct {
	Folder   string `json:"folder,omitempty"`
	FileName string `json:"file_name"`
	Content  string `json:"content"`
}

type UpdateManifestParams struct {
	FileName string `json:"file_name"`
	Content  string `json:"content"`
}
