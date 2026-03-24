package handler

import (
	"net/http"
	"strconv"

	"backend/internal/dto"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

type PipelineHandler struct {
	service *service.PipelineService
}

func NewPipelineHandler(service *service.PipelineService) *PipelineHandler {
	return &PipelineHandler{service: service}
}

// Pipeline endpoints
func (h *PipelineHandler) CreatePipeline(c *gin.Context) {
	var req dto.CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Extract userID from JWT token
	userID := uint(1) // Placeholder

	pipeline, err := h.service.CreatePipeline(req.Name, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pipeline)
}

func (h *PipelineHandler) GetPipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	pipeline, err := h.service.GetPipeline(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pipeline not found"})
		return
	}

	c.JSON(http.StatusOK, pipeline)
}

func (h *PipelineHandler) GetUserPipelines(c *gin.Context) {
	// TODO: Extract userID from JWT token
	userID := uint(1) // Placeholder

	pipelines, err := h.service.GetUserPipelines(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pipelines)
}

func (h *PipelineHandler) DeletePipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = h.service.DeletePipeline(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pipeline deleted"})
}

func (h *PipelineHandler) UpdatePipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req dto.UpdatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdatePipeline(uint(id), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pipeline updated"})
}

// Stage endpoints
func (h *PipelineHandler) CreateStage(c *gin.Context) {
	var req dto.CreateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stage, err := h.service.CreateStage(req.Name, req.PipelineID, req.Position)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return with leads array
	response := dto.StageResponse{
		ID:       stage.ID,
		Name:     stage.Name,
		Position: stage.Position,
		Leads:    []dto.LeadResponse{},
	}
	c.JSON(http.StatusCreated, response)
}

func (h *PipelineHandler) UpdateStagePosition(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("stageId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Stage ID"})
		return
	}

	var req dto.UpdateStagePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateStagePosition(uint(id), req.Position)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stage position updated"})
}

func (h *PipelineHandler) DeleteStage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("stageId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = h.service.DeleteStage(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stage deleted"})
}

func (h *PipelineHandler) UpdateStage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("stageId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Stage ID"})
		return
	}

	var req dto.UpdateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateStage(uint(id), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stage updated"})
}

// Lead endpoints
func (h *PipelineHandler) CreateLead(c *gin.Context) {
	var req dto.CreateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lead, err := h.service.CreateLead(req.Name, req.Email, req.Phone, req.StageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, lead)
}

func (h *PipelineHandler) MoveLeadToStage(c *gin.Context) {
	leadID, err := strconv.ParseUint(c.Param("leadId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Lead ID"})
		return
	}

	var req dto.MoveLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.MoveLeadToStage(uint(leadID), req.StageID, req.Position)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lead moved"})
}

func (h *PipelineHandler) DeleteLead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("leadId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = h.service.DeleteLead(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lead deleted"})
}

func (h *PipelineHandler) UpdateLead(c *gin.Context) {
	leadID, err := strconv.ParseUint(c.Param("leadId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Lead ID"})
		return
	}

	var req dto.UpdateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateLead(uint(leadID), req.Name, req.Email, req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lead updated"})
}
