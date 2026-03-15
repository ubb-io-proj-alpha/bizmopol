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
    Data  []ContactResponse `json:"data"`
    Total int64             `json:"total"`
}

type ContactQuery struct {
    Search  string `form:"search"`
    Status  string `form:"status"`
    SortBy  string `form:"sort_by"`
    SortDir string `form:"sort_dir"`
}
