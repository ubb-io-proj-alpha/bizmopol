package handler

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"backend/internal/dto"
	"backend/internal/service"
)

type DocumentHandler struct {
	svc service.DocumentService
}

func NewDocumentHandler(svc service.DocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	var req dto.DocumentUploadRequest
	req.Title = c.PostForm("title")
	req.ContactID = c.PostForm("contact_id")

	if req.Title == "" {
		req.Title = header.Filename
	}

	doc, err := h.svc.Upload(c, uid, req, file, header)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, doc)
}

func (h *DocumentHandler) Get(c *gin.Context) {
	id := c.Param("id")
	doc, err := h.svc.Get(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if doc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (h *DocumentHandler) List(c *gin.Context) {
	var q dto.DocumentQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	docs, total, err := h.svc.List(c, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"documents": docs, "total": total})
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	if err := h.svc.Delete(c, id, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *DocumentHandler) Download(c *gin.Context) {
	id := c.Param("id")
	mode := c.DefaultQuery("mode", "signed")
	filePath, fileName, err := h.svc.GetFilePath(c, id, mode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	c.File(filePath)
}

func (h *DocumentHandler) View(c *gin.Context) {
	id := c.Param("id")
	mode := c.DefaultQuery("mode", "signed")
	filePath, _, err := h.svc.GetFilePath(c, id, mode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ext := filepath.Ext(filePath)
	if ext == ".pdf" {
		c.Header("Content-Type", "application/pdf")
	}
	c.Header("Content-Disposition", "inline")
	c.File(filePath)
}

func (h *DocumentHandler) Sign(c *gin.Context) {
	docID := c.Param("id")
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	email, _ := c.Get("userEmail")
	userEmail, _ := email.(string)

	var req dto.DocumentSignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SignerEmail == "" {
		req.SignerEmail = userEmail
	}

	sig, err := h.svc.Sign(c, docID, uid, userEmail, c.ClientIP(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sig)
}

func (h *DocumentHandler) SignatureImage(c *gin.Context) {
	sigID := c.Param("sig_id")
	imgPath, err := h.svc.GetSignatureImage(c, sigID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Header("Content-Type", "image/png")
	c.File(imgPath)
}

func (h *DocumentHandler) AuditLog(c *gin.Context) {
	docID := c.Param("id")
	logs, err := h.svc.GetAuditLog(c, docID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, logs)
}
