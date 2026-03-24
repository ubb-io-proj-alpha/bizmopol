package dto

// Pipeline DTOs
type CreatePipelineRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdatePipelineRequest struct {
	Name string `json:"name" binding:"required"`
}

type PipelineResponse struct {
	ID        uint          `json:"id"`
	Name      string        `json:"name"`
	Stages    []StageResponse `json:"stages"`
	CreatedAt string        `json:"created_at"`
}

// Stage DTOs
type CreateStageRequest struct {
	Name       string `json:"name" binding:"required"`
	PipelineID uint   `json:"pipeline_id" binding:"required"`
	Position   int    `json:"position"`
}

type UpdateStageRequest struct {
	Name string `json:"name" binding:"required"`
}

type StageResponse struct {
	ID        uint          `json:"id"`
	Name      string        `json:"name"`
	Position  int           `json:"position"`
	Leads     []LeadResponse `json:"leads"`
	CreatedAt string        `json:"created_at"`
}

type UpdateStagePositionRequest struct {
	Position int `json:"position" binding:"required"`
}

// Lead DTOs
type CreateLeadRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	StageID *uint  `json:"stage_id"`
}

type UpdateLeadRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type LeadResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	StageID   *uint  `json:"stage_id"`
	Position  int    `json:"position"`
	CreatedAt string `json:"created_at"`
}

type MoveLeadRequest struct {
	StageID  *uint `json:"stage_id"`
	Position int   `json:"position" binding:"required"`
}
