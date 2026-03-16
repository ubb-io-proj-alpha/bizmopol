package dto

import "time"

type ContactCreateRequest struct {
    Name    string `json:"name" binding:"required"`
    Email   string `json:"email"`
    Phone   string `json:"phone"`
    Company string `json:"company"`
    Status  string `json:"status"`
    Notes   string `json:"notes"`
}

type ContactUpdateRequest struct {
    Name    string `json:"name"`
    Email   string `json:"email"`
    Phone   string `json:"phone"`
    Company string `json:"company"`
    Status  string `json:"status"`
    Notes   string `json:"notes"`
}

type ContactResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Phone     string    `json:"phone"`
    Company   string    `json:"company"`
    Status    string    `json:"status"`
    Notes     string    `json:"notes"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type ContactListResponse struct {
    Data       []ContactResponse `json:"data"`
    Total      int64             `json:"total"`
    Page       int               `json:"page"`
    PageSize   int               `json:"page_size"`
    TotalPages int               `json:"total_pages"`
}

type ContactQuery struct {
    Search   string `form:"search"`
    Status   string `form:"status"`
    SortBy   string `form:"sort_by"`
    SortDir  string `form:"sort_dir"`
    Page     int    `form:"page"`
    PageSize int    `form:"page_size"`
}

type ContactHistoryResponse struct {
    ID          string    `json:"id"`
    ContactID   string    `json:"contact_id"`
    Action      string    `json:"action"`
    Description string    `json:"description"`
    UserID      string    `json:"user_id"`
    CreatedAt   time.Time `json:"created_at"`
}

type AddHistoryRequest struct {
    Action      string `json:"action" binding:"required"`
    Description string `json:"description"`
}
