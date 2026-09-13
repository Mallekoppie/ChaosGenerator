package targethttp

// UseCase represents a test use case
type UseCase struct {
	ID          string
	Name        string
	Path        string
	Method      string
	Description string
}

// GetUseCases returns all available use cases
func GetUseCases() []UseCase {
	return []UseCase{
		{
			ID:          "uc1",
			Name:        "Health Check",
			Path:        "/health",
			Method:      "GET",
			Description: "Minimal overhead health check endpoint",
		},
		{
			ID:          "uc2",
			Name:        "Status Endpoint",
			Path:        "/api/v1/status",
			Method:      "GET",
			Description: "Status endpoint with lightweight JSON metadata",
		},
		{
			ID:          "uc3",
			Name:        "Simple Query",
			Path:        "/api/v1/users/",
			Method:      "GET",
			Description: "Simple query operation with small payload (1-2KB)",
		},
		{
			ID:          "uc4",
			Name:        "Create Resource",
			Path:        "/api/v1/users",
			Method:      "POST",
			Description: "Create resource operation with request/response (1-2KB)",
		},
		{
			ID:          "uc5",
			Name:        "List Collection",
			Path:        "/api/v1/users",
			Method:      "GET",
			Description: "List/collection query with pagination (50-100KB)",
		},
		{
			ID:          "uc6",
			Name:        "Bulk Update",
			Path:        "/api/v1/users/bulk",
			Method:      "PATCH",
			Description: "Bulk update operation with larger payloads (100-200KB)",
		},
		{
			ID:          "uc7",
			Name:        "Report Generation",
			Path:        "/api/v1/reports/generate",
			Method:      "POST",
			Description: "CPU-intensive report generation (500KB-1MB response)",
		},
		{
			ID:          "uc8",
			Name:        "File Upload",
			Path:        "/api/v1/documents/upload",
			Method:      "POST",
			Description: "Large request payload simulation (10-25MB)",
		},
		{
			ID:          "uc9",
			Name:        "Data Export",
			Path:        "/api/v1/transactions/export",
			Method:      "GET",
			Description: "Large response payload and streaming (25-50MB)",
		},
		{
			ID:          "uc10",
			Name:        "Network Saturation",
			Path:        "/api/v1/data/process",
			Method:      "POST",
			Description: "Network bandwidth saturation testing (100-250MB)",
		},
		{
			ID:          "uc11",
			Name:        "Memory Pressure",
			Path:        "/api/v1/analytics/stream",
			Method:      "POST",
			Description: "Long-running operations for memory leak detection (10-20KB)",
		},
	}
}
