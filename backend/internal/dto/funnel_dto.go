package dto

type FunnelCreateRequest struct {
    Name         string `json:"name" binding:"required"`
    Subdomain    string `json:"subdomain"`
    CustomDomain string `json:"custom_domain"`
}

type FunnelUpdateRequest struct {
    Name         string `json:"name"`
    Subdomain    string `json:"subdomain"`
    CustomDomain string `json:"custom_domain"`
}

type PageCreateRequest struct {
    Name      string `json:"name" binding:"required"`
    Path      string `json:"path" binding:"required"`
    Structure string `json:"structure"`
}

type PageUpdateRequest struct {
    Name      string `json:"name"`
    Path      string `json:"path"`
    Structure string `json:"structure"`
}

type FunnelSubmissionRequest struct {
    FunnelID string `json:"funnel_id" binding:"required"`
    PageID   string `json:"page_id" binding:"required"`
    Name     string `json:"name" binding:"required"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
}