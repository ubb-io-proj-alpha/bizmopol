package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "backend/internal/dto"
    "backend/internal/service"
)

type CustomFieldHandler struct {
    service service.CustomFieldService
}

func NewCustomFieldHandler(s service.CustomFieldService) *CustomFieldHandler {
    return &CustomFieldHandler{service: s}
}

func (h *CustomFieldHandler) List(c *gin.Context) {
    result, err := h.service.List(c)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusOK, result)
}

func (h *CustomFieldHandler) Create(c *gin.Context) {
    var input dto.CustomFieldCreateRequest
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

func (h *CustomFieldHandler) Update(c *gin.Context) {
    id := c.Param("id")
    var input dto.CustomFieldUpdateRequest
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

func (h *CustomFieldHandler) Delete(c *gin.Context) {
    id := c.Param("id")
    if err := h.service.Delete(c, id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusNoContent, nil)
}
