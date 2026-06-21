package handler

import (
    "fmt"
    "net/http"
    "strings"


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
	q := dto.PaginationQuery{Page: 1, Limit: 9}

	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nieprawidłowe parametry paginacji"})
		return
	}

	listItems, totalCount, err := h.service.ListFunnels(c, q.Page, q.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	response := dto.FunnelListResponse{
		Data:       listItems,
		TotalCount: totalCount,
		Page:       q.Page,
		Limit:      q.Limit,
	}

	c.JSON(http.StatusOK, response)
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

func (h *FunnelHandler) ServeLivePage(c *gin.Context) {
    host := c.Request.Host
    if strings.Contains(host, ":") {
        host = strings.Split(host, ":")[0]
    }

    path := c.Request.URL.Path
    if path == "" {
        path = "/"
    } else if !strings.HasPrefix(path, "/") {
        path = "/" + path
    }

    userAgent := c.Request.UserAgent()

    result, err := h.service.ResolvePage(c, host, path, userAgent)
    if err != nil {
        c.String(http.StatusInternalServerError, "Błąd serwera (500)")
        return
    }
    if result == nil {
        c.String(http.StatusNotFound, "Strona nie została znaleziona (404)")
        return
    }

    fullHTML := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="pl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>%s</style>
</head>
<body>
    %s
</body>
</html>`, result.Name, result.CSSContent, result.HTMLContent)

    c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fullHTML))
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