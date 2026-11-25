package handler

import (
	"errors"
	"log"
	"net/http"
	"taskflow/internal/dto"
	"taskflow/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Service *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		Service: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
		})
		return
	}

	user, err := h.Service.Register(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAuthInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid registration data",
			})
		case errors.Is(err, service.ErrAuthEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"message": "Email already registered",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"user": user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
		})
		return
	}

	result, err := h.Service.Login(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAuthInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid email or password",
			})
		default:
    log.Printf("login error: %v", err)
    c.JSON(http.StatusInternalServerError, gin.H{
        "message": "Internal server error",
    })
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data": result,
	})
}