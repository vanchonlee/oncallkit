package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vanchonlee/oncallkit/services"
)

type UserHandler struct {
	Service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{Service: service}
}

// User CRUD endpoints
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.Service.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	user, err := h.Service.CreateUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.Service.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.Service.UpdateUser(id, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	err := h.Service.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// On-call endpoints
func (h *UserHandler) GetCurrentOnCallUser(c *gin.Context) {
	user, err := h.Service.GetCurrentOnCallUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no on-call user found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) CreateOnCallSchedule(c *gin.Context) {
	schedule, err := h.Service.CreateOnCallSchedule(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, schedule)
}

func (h *UserHandler) ListOnCallSchedules(c *gin.Context) {
	schedules, err := h.Service.ListOnCallSchedules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, schedules)
}
