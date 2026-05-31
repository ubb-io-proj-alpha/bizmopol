package dto

import "time"

type PipelineCreateRequest struct {
    Name        string               `json:"name" binding:"required"`
    Description string               `json:"description"`
    Stages      []StageCreateRequest `json:"stages"`
}

type PipelineUpdateRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type StageCreateRequest struct {
    Name      string `json:"name" binding:"required"`
    Color     string `json:"color"`
    SortOrder int    `json:"sort_order"`
}

type StageUpdateRequest struct {
    Name      string `json:"name"`
    Color     string `json:"color"`
    SortOrder int    `json:"sort_order"`
}

type StageResponse struct {
    ID         string    `json:"id"`
    PipelineID string    `json:"pipeline_id"`
    Name       string    `json:"name"`
    Color      string    `json:"color"`
    SortOrder  int       `json:"sort_order"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

type PipelineResponse struct {
    ID          string          `json:"id"`
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Stages      []StageResponse `json:"stages"`
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
}

type MoveContactRequest struct {
    ContactID string `json:"contact_id" binding:"required"`
    StageID   string `json:"stage_id" binding:"required"`
}

type KanbanContactResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Phone     string    `json:"phone"`
    Company   string    `json:"company"`
    Status    string    `json:"status"`
    LeadScore int       `json:"lead_score"`
    StageID   string    `json:"stage_id"`
    CreatedAt time.Time `json:"created_at"`
}

type KanbanColumnResponse struct {
    Stage    StageResponse           `json:"stage"`
    Contacts []KanbanContactResponse `json:"contacts"`
}

type KanbanResponse struct {
    Pipeline PipelineResponse       `json:"pipeline"`
    Columns  []KanbanColumnResponse `json:"columns"`
}