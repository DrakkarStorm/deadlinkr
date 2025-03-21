package model

type LinkResult struct {
	SourceURL  string `json:"source_url"`
	TargetURL  string `json:"target_url"`
	Status     int    `json:"status"`
	Error      string `json:"error,omitempty"`
	IsExternal bool   `json:"is_external"`
}
