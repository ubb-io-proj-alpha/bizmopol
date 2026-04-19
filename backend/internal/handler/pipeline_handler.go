package handler

import (
	"errors"
	"net/http"
	"strconv"

	"backend/internal/dto"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PipelineHandler struct {
	service *service.PipelineService
}

func NewPipelineHandler(service *service.PipelineService) *PipelineHandler {
	return &PipelineHandler{service: service}
}

func getAuthenticatedUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authenticated user"})
		return "", false
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return "", false
	}

	return userIDStr, true
}

func parseUintParam(c *gin.Context, param string, errorMessage string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(param), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		return 0, false
	}

	return uint(id), true
}

func handlePipelineError(c *gin.Context, err error, notFoundMessage string) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": notFoundMessage})
	case errors.Is(err, service.ErrStageIDRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stage ID is required"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// Pipeline endpoints
func (h *PipelineHandler) CreatePipeline(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pipeline, err := h.service.CreatePipeline(req.Name, userID)
	if err != nil {
		handlePipelineError(c, err, "Pipeline not found")
		return
	}

	c.JSON(http.StatusCreated, pipeline)
}

func (h *PipelineHandler) GetPipeline(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	id, ok := parseUintParam(c, "id", "Invalid ID")
	if !ok {
		return
	}

	pipeline, err := h.service.GetPipeline(id, userID)
	if err != nil {
		handlePipelineError(c, err, "Pipeline not found")
		return
	}

	c.JSON(http.StatusOK, pipeline)
}

func (h *PipelineHandler) GetUserPipelines(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	pipelines, err := h.service.GetUserPipelines(userID)
	if err != nil {
		handlePipelineError(c, err, "Pipelines not found")
		return
	}

	c.JSON(http.StatusOK, pipelines)
}

func (h *PipelineHandler) DeletePipeline(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	id, ok := parseUintParam(c, "id", "Invalid ID")
	if !ok {
		return
	}

	err := h.service.DeletePipeline(id, userID)
	if err != nil {
		handlePipelineError(c, err, "Pipeline not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pipeline deleted"})
}

func (h *PipelineHandler) UpdatePipeline(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	id, ok := parseUintParam(c, "id", "Invalid ID")
	if !ok {
		return
	}

	var req dto.UpdatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdatePipeline(id, userID, req.Name)
	if err != nil {
		handlePipelineError(c, err, "Pipeline not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pipeline updated"})
}

// Stage endpoints
func (h *PipelineHandler) CreateStage(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	pipelineID, ok := parseUintParam(c, "id", "Invalid pipeline ID")
	if !ok {
		return
	}

	var req dto.CreateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stage, err := h.service.CreateStage(req.Name, pipelineID, userID, req.Position)
	if err != nil {
		handlePipelineError(c, err, "Pipeline not found")
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
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	pipelineID, ok := parseUintParam(c, "id", "Invalid pipeline ID")
	if !ok {
		return
	}

	stageID, ok := parseUintParam(c, "stageId", "Invalid Stage ID")
	if !ok {
		return
	}

	var req dto.UpdateStagePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateStagePosition(stageID, pipelineID, userID, req.Position)
	if err != nil {
		handlePipelineError(c, err, "Stage not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stage position updated"})
}

func (h *PipelineHandler) DeleteStage(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	pipelineID, ok := parseUintParam(c, "id", "Invalid pipeline ID")
	if !ok {
		return
	}

	stageID, ok := parseUintParam(c, "stageId", "Invalid Stage ID")
	if !ok {
		return
	}

	err := h.service.DeleteStage(stageID, pipelineID, userID)
	if err != nil {
		handlePipelineError(c, err, "Stage not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stage deleted"})
}

func (h *PipelineHandler) UpdateStage(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	pipelineID, ok := parseUintParam(c, "id", "Invalid pipeline ID")
	if !ok {
		return
	}

	stageID, ok := parseUintParam(c, "stageId", "Invalid Stage ID")
	if !ok {
		return
	}

	var req dto.UpdateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateStage(stageID, pipelineID, userID, req.Name)
	if err != nil {
		handlePipelineError(c, err, "Stage not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stage updated"})
}

// Lead endpoints
func (h *PipelineHandler) CreateLead(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	pipelineID, ok := parseUintParam(c, "id", "Invalid pipeline ID")
	if !ok {
		return
	}

	var req dto.CreateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lead, err := h.service.CreateLead(req.Name, req.Email, req.Phone, pipelineID, req.StageID, userID)
	if err != nil {
		handlePipelineError(c, err, "Lead or stage not found")
		return
	}

	c.JSON(http.StatusCreated, lead)
}

func (h *PipelineHandler) MoveLeadToStage(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	pipelineID, ok := parseUintParam(c, "id", "Invalid pipeline ID")
	if !ok {
		return
	}

	leadID, ok := parseUintParam(c, "leadId", "Invalid Lead ID")
	if !ok {
		return
	}

	var req dto.MoveLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.MoveLeadToStage(leadID, pipelineID, userID, req.StageID, req.Position)
	if err != nil {
		handlePipelineError(c, err, "Lead or stage not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lead moved"})
}

func (h *PipelineHandler) DeleteLead(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	pipelineID, ok := parseUintParam(c, "id", "Invalid pipeline ID")
	if !ok {
		return
	}

	leadID, ok := parseUintParam(c, "leadId", "Invalid Lead ID")
	if !ok {
		return
	}

	err := h.service.DeleteLead(leadID, pipelineID, userID)
	if err != nil {
		handlePipelineError(c, err, "Lead not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lead deleted"})
}

func (h *PipelineHandler) UpdateLead(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	pipelineID, ok := parseUintParam(c, "id", "Invalid pipeline ID")
	if !ok {
		return
	}

	leadID, ok := parseUintParam(c, "leadId", "Invalid Lead ID")
	if !ok {
		return
	}

	var req dto.UpdateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateLead(leadID, pipelineID, userID, req.Name, req.Email, req.Phone)
	if err != nil {
		handlePipelineError(c, err, "Lead not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lead updated"})
}
