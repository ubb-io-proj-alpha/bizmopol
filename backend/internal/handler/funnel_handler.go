package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "backend/internal/dto"
    "backend/internal/service"
)

type FunnelHandler struct {
    service service.FunnelService
}

func NewFunnelHandler(s service.FunnelService) *FunnelHandler {
    return &FunnelHandler{service: s}
}

func (h *FunnelHandler) CreateFunnel(c *gin.Context) {
    var input dto.FunnelCreateRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    result, err := h.service.CreateFunnel(c, input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusCreated, result)
}

func (h *FunnelHandler) ListFunnels(c *gin.Context) {
    result, err := h.service.ListFunnels(c)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusOK, result)
}

func (h *FunnelHandler) GetFunnel(c *gin.Context) {
    id := c.Param("id")
    result, err := h.service.GetFunnel(c, id)
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

func (h *FunnelHandler) UpdateFunnel(c *gin.Context) {
    id := c.Param("id")
    var input dto.FunnelUpdateRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    result, err := h.service.UpdateFunnel(c, id, input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusOK, result)
}

func (h *FunnelHandler) DeleteFunnel(c *gin.Context) {
    id := c.Param("id")
    if err := h.service.DeleteFunnel(c, id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusNoContent, nil)
}

func (h *FunnelHandler) CreatePage(c *gin.Context) {
    funnelId := c.Param("id")
    var input dto.PageCreateRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    result, err := h.service.CreatePage(c, funnelId, input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusCreated, result)
}

func (h *FunnelHandler) UpdatePage(c *gin.Context) {
    pageId := c.Param("pageId")
    var input dto.PageUpdateRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    result, err := h.service.UpdatePage(c, pageId, input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusOK, result)
}

func (h *FunnelHandler) DeletePage(c *gin.Context) {
    pageId := c.Param("pageId")
    if err := h.service.DeletePage(c, pageId); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusNoContent, nil)
}

func (h *FunnelHandler) Resolve(c *gin.Context) {
    host := c.Query("host")
    path := c.Query("path")
    ip := c.ClientIP()
    userAgent := c.Request.UserAgent()

    result, err := h.service.ResolvePage(c, host, path, ip, userAgent)
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

func (h *FunnelHandler) Submit(c *gin.Context) {
    var input dto.FunnelSubmissionRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    err := h.service.SubmitForm(c, input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusCreated, gin.H{"success": true})
}