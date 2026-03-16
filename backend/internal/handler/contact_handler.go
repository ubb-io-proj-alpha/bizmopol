package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "backend/internal/dto"
    "backend/internal/service"
)

type ContactHandler struct {
    service service.ContactService
}

func NewContactHandler(s service.ContactService) *ContactHandler {
    return &ContactHandler{service: s}
}

func (h *ContactHandler) Create(c *gin.Context) {
    var input dto.ContactCreateRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    result, err := h.service.Create(c, input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusCreated, result)
}

func (h *ContactHandler) GetByID(c *gin.Context) {
    id := c.Param("id")
    result, err := h.service.GetByID(c, id)
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

func (h *ContactHandler) Update(c *gin.Context) {
    id := c.Param("id")
    var input dto.ContactUpdateRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    result, err := h.service.Update(c, id, input)
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

func (h *ContactHandler) Delete(c *gin.Context) {
    id := c.Param("id")
    if err := h.service.Delete(c, id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusNoContent, nil)
}

func (h *ContactHandler) List(c *gin.Context) {
    var q dto.ContactQuery
    if err := c.ShouldBindQuery(&q); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    result, err := h.service.List(c, q)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusOK, result)
}

func (h *ContactHandler) ListHistory(c *gin.Context) {
    id := c.Param("id")
    result, err := h.service.ListHistory(c, id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusOK, result)
}

func (h *ContactHandler) AddHistory(c *gin.Context) {
    id := c.Param("id")
    var input dto.AddHistoryRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    userID, _ := c.Get("userID")
    uid, _ := userID.(string)
    result, err := h.service.AddHistory(c, id, uid, input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusCreated, result)
}
