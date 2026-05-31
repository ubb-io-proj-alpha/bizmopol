package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/internal/dto"
	"backend/internal/service"
)

type CommunicationHandler struct {
	svc service.CommunicationService
}

func NewCommunicationHandler(svc service.CommunicationService) *CommunicationHandler {
	return &CommunicationHandler{svc: svc}
}

func (h *CommunicationHandler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetStats(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *CommunicationHandler) SyncInbox(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	result, err := h.svc.SyncInbox(c, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CommunicationHandler) ListThreads(c *gin.Context) {
	var q dto.ThreadQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.ListThreads(c, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CommunicationHandler) GetThread(c *gin.Context) {
	id := c.Param("id")
	result, err := h.svc.GetThread(c, id)
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

func (h *CommunicationHandler) SendEmail(c *gin.Context) {
	var req dto.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	result, err := h.svc.SendEmail(c, uid, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *CommunicationHandler) ReplyToThread(c *gin.Context) {
	threadID := c.Param("id")
	var req dto.ReplyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	result, err := h.svc.ReplyToThread(c, uid, threadID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *CommunicationHandler) DeleteThread(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteThread(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *CommunicationHandler) MarkRead(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.MarkThreadRead(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *CommunicationHandler) ArchiveThread(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.ArchiveThread(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *CommunicationHandler) UpdateThreadStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateThreadStatus(c, id, body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *CommunicationHandler) StarMessage(c *gin.Context) {
	msgID := c.Param("id")
	var body struct {
		Starred bool `json:"starred"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.StarMessage(c, msgID, body.Starred); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *CommunicationHandler) SendBulkEmail(c *gin.Context) {
	var req dto.BulkEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	result, err := h.svc.SendBulkEmail(c, uid, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *CommunicationHandler) ListBulkJobs(c *gin.Context) {
	result, err := h.svc.ListBulkJobs(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CommunicationHandler) GetBulkJob(c *gin.Context) {
	id := c.Param("id")
	result, err := h.svc.GetBulkJob(c, id)
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

func (h *CommunicationHandler) ListTemplates(c *gin.Context) {
	result, err := h.svc.ListTemplates(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CommunicationHandler) CreateTemplate(c *gin.Context) {
	var req dto.TemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.CreateTemplate(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *CommunicationHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	var req dto.TemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.UpdateTemplate(c, id, req)
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

func (h *CommunicationHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteTemplate(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *CommunicationHandler) ListSignatures(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	result, err := h.svc.ListSignatures(c, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CommunicationHandler) CreateSignature(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	var req dto.SignatureCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.CreateSignature(c, uid, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *CommunicationHandler) UpdateSignature(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	var req dto.SignatureUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.UpdateSignature(c, id, uid, req)
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

func (h *CommunicationHandler) DeleteSignature(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteSignature(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *CommunicationHandler) GetContactThreads(c *gin.Context) {
	contactID := c.Param("id")
	result, err := h.svc.GetContactThreads(c, contactID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CommunicationHandler) GetEmailAccount(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	result, err := h.svc.GetEmailAccount(c, uid)
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

func (h *CommunicationHandler) CreateEmailAccount(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	var req dto.EmailAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.CreateEmailAccount(c, uid, req)
	if err != nil {
		if err.Error() == "email account already exists, update it instead" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *CommunicationHandler) UpdateEmailAccount(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	var req dto.EmailAccountUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.UpdateEmailAccount(c, uid, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CommunicationHandler) DeleteEmailAccount(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	if err := h.svc.DeleteEmailAccount(c, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *CommunicationHandler) TestConnection(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	result, err := h.svc.TestConnection(c, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}
