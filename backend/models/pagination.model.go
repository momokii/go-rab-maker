package models

type TablePaginationDataInput struct {
	Search  string `json:"search"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
}

// table related
// TableConfig holds configuration for the reusable table component
type TableConfig struct {
	BaseURL           string // Base URL for HTMX requests
	Title             string // Table title
	SearchEnabled     bool   // Enable search functionality
	PaginationEnabled bool   // Enable pagination
	PerPageEnabled    bool   // Enable items per page selection
}

// PaginationInfo holds pagination related data
type PaginationInfo struct {
	CurrentPage  int
	TotalPages   int
	TotalItems   int
	ItemsPerPage int
}
