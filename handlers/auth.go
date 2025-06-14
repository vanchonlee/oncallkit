package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vanchonlee/oncallkit/services"
)

type AuthHandler struct {
	Service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{Service: service}
}

// Login endpoint
func (h *AuthHandler) Login(c *gin.Context) {
	response, err := h.Service.Login(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// ChangePassword endpoint
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.Service.ChangePassword(req.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// SetupAdmin endpoint - creates admin user if not exists
func (h *AuthHandler) SetupAdmin(c *gin.Context) {
	err := h.Service.SetupAdminUser()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "Admin user setup completed",
		"email":    "admin@slar.com",
		"password": "admin123",
		"note":     "Please change the default password after first login",
	})
}
