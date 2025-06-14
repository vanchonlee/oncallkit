package router

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"github.com/vanchonlee/oncallkit/handlers"
	"github.com/vanchonlee/oncallkit/services"
)

func NewGinRouter(pg *sql.DB, redis *redis.Client) *gin.Engine {
	r := gin.Default()

	// Initialize services
	alertService := services.NewAlertService(pg, redis)
	userService := services.NewUserService(pg, redis)

	// Initialize handlers
	alertHandler := handlers.NewAlertHandler(alertService)
	userHandler := handlers.NewUserHandler(userService)

	// ALERTS
	r.GET("/alerts", alertHandler.ListAlerts)
	r.POST("/alerts", alertHandler.CreateAlert)
	r.GET("/alerts/:id", alertHandler.GetAlert)
	r.POST("/alerts/:id/ack", alertHandler.AckAlert)
	r.POST("/alerts/:id/unack", alertHandler.UnackAlert)
	r.POST("/alerts/:id/close", alertHandler.CloseAlert)

	// USERS
	r.GET("/users", userHandler.ListUsers)
	r.POST("/users", userHandler.CreateUser)
	r.GET("/users/:id", userHandler.GetUser)
	r.PUT("/users/:id", userHandler.UpdateUser)
	r.DELETE("/users/:id", userHandler.DeleteUser)

	// ON-CALL
	r.GET("/oncall/current", userHandler.GetCurrentOnCallUser)
	r.GET("/oncall/schedules", userHandler.ListOnCallSchedules)
	r.POST("/oncall/schedules", userHandler.CreateOnCallSchedule)

	// DASHBOARD, UPTIME, ...
	r.GET("/dashboard", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Dashboard endpoint - TODO implement"})
	})

	r.GET("/uptime", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Uptime endpoint - TODO implement"})
	})

	// TODO: notes, tags, log, recipients endpoints

	return r
}
