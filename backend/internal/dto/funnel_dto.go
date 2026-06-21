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
    Name        string `json:"name" binding:"required"`
    Path        string `json:"path" binding:"required"`
    Structure   string `json:"structure"`
    HtmlContent string `json:"html_content"`
    CssContent  string `json:"css_content"`
}

type PageUpdateRequest struct {
    Name        string `json:"name"`
    Path        string `json:"path"`
    Structure   string `json:"structure"`
    HtmlContent string `json:"html_content"`
    CssContent  string `json:"css_content"`
}

type FunnelSubmissionRequest struct {
    FunnelID string `json:"funnel_id" binding:"required"`
    PageID   string `json:"page_id" binding:"required"`
    Name     string `json:"name" binding:"required"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
}

type FunnelListItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Subdomain    string `json:"subdomain"`
	CustomDomain string `json:"custom_domain"`
	PagesCount   int    `json:"pages_count"`
}

type FunnelListResponse struct {
	Data       []FunnelListItem `json:"data"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
}

type PaginationQuery struct {
    Page  int `form:"page" binding:"omitempty,min=1"`
    Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}