package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/internal/dto"
	"backend/internal/service"
)

type CalendarHandler struct {
	svc service.CalendarService
}

func NewCalendarHandler(svc service.CalendarService) *CalendarHandler {
	return &CalendarHandler{svc: svc}
}

func (h *CalendarHandler) GetStats(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	stats, err := h.svc.GetStats(c, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *CalendarHandler) ListEvents(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	var q dto.EventQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.ListEvents(c, uid, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CalendarHandler) CreateEvent(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	var req dto.EventCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.CreateEvent(c, uid, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *CalendarHandler) GetEvent(c *gin.Context) {
	id := c.Param("id")
	result, err := h.svc.GetEvent(c, id)
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

func (h *CalendarHandler) UpdateEvent(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	id := c.Param("id")
	var req dto.EventUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.UpdateEvent(c, uid, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CalendarHandler) DeleteEvent(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	id := c.Param("id")
	if err := h.svc.DeleteEvent(c, uid, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *CalendarHandler) UpcomingEvents(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	result, err := h.svc.UpcomingEvents(c, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CalendarHandler) GetZoomAccount(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	result, err := h.svc.GetZoomAccount(c, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if result == nil {
		c.JSON(http.StatusOK, gin.H{"configured": false})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CalendarHandler) CreateZoomAccount(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	var req dto.ZoomAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.CreateZoomAccount(c, uid, req)
	if err != nil {
		if err.Error() == "zoom account already exists, update it instead" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *CalendarHandler) UpdateZoomAccount(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	var req dto.ZoomAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.UpdateZoomAccount(c, uid, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CalendarHandler) DeleteZoomAccount(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	if err := h.svc.DeleteZoomAccount(c, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *CalendarHandler) TestZoomConnection(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	result, err := h.svc.TestZoomConnection(c, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}
