package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "backend/internal/dto"
    "backend/internal/service"
)

type AuthHandler struct {
    service service.AuthService
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
    return &AuthHandler{service: s}
}

func (h *AuthHandler) Register(c *gin.Context) {
    var input dto.RegisterRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    result, err := h.service.Register(c, input)
    if err != nil {
        if err == service.ErrUserExists {
            c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }

    c.JSON(http.StatusCreated, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
    var input dto.LoginRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    result, err := h.service.Login(c, input)
    if err != nil {
        if err == service.ErrInvalidCredentials {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }

    c.JSON(http.StatusOK, result)
}
