package handler

import (
	"net/http"
	"backend/internal/dto"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

type TestHandler struct {
	service service.TestService
}

func NewTestHandler(service service.TestService) *TestHandler {
	return &TestHandler{service: service}
}

func (handler *TestHandler) Create(c *gin.Context) {
	var input dto.TestCreateRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	result, err := handler.service.Create(c, input)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (handler *TestHandler) GetAll(c *gin.Context) {
	result, err := handler.service.GetAll(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}