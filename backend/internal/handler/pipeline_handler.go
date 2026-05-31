package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "backend/internal/dto"
    "backend/internal/service"
)

type PipelineHandler struct {
    service service.PipelineService
}

func NewPipelineHandler(s service.PipelineService) *PipelineHandler {
    return &PipelineHandler{service: s}
}

func (h *PipelineHandler) List(c *gin.Context) {
    result, err := h.service.ListPipelines(c)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusOK, result)
}

func (h *PipelineHandler) Create(c *gin.Context) {
    var input dto.PipelineCreateRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    result, err := h.service.CreatePipeline(c, input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusCreated, result)
}

func (h *PipelineHandler) GetByID(c *gin.Context) {
    id := c.Param("id")
    result, err := h.service.GetPipeline(c, id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    if result == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, result)
}

func (h *PipelineHandler) Update(c *gin.Context) {
    id := c.Param("id")
    var input dto.PipelineUpdateRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    result, err := h.service.UpdatePipeline(c, id, input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    if result == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, result)
}

func (h *PipelineHandler) Delete(c *gin.Context) {
    id := c.Param("id")
    if err := h.service.DeletePipeline(c, id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusNoContent, nil)
}

func (h *PipelineHandler) CreateStage(c *gin.Context) {
    pipelineID := c.Param("id")
    var input dto.StageCreateRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    result, err := h.service.CreateStage(c, pipelineID, input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusCreated, result)
}

func (h *PipelineHandler) UpdateStage(c *gin.Context) {
    stageID := c.Param("stage_id")
    var input dto.StageUpdateRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    result, err := h.service.UpdateStage(c, stageID, input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    if result == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, result)
}

func (h *PipelineHandler) DeleteStage(c *gin.Context) {
    stageID := c.Param("stage_id")
    if err := h.service.DeleteStage(c, stageID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusNoContent, nil)
}

func (h *PipelineHandler) MoveContact(c *gin.Context) {
    pipelineID := c.Param("id")
    var input dto.MoveContactRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    userID, _ := c.Get("userID")
    uid, _ := userID.(string)
    if err := h.service.MoveContact(c, pipelineID, input, uid); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *PipelineHandler) GetKanban(c *gin.Context) {
    pipelineID := c.Param("id")
    result, err := h.service.GetKanban(c, pipelineID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusOK, result)
}