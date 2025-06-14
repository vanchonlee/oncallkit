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
	uptimeService := services.NewUptimeService(pg, redis)
	alertManagerService := services.NewAlertManagerService(pg, alertService)
	authService := services.NewAuthService(pg, redis)
	apiKeyService := services.NewAPIKeyService(pg)

	// Initialize handlers
	alertHandler := handlers.NewAlertHandler(alertService)
	userHandler := handlers.NewUserHandler(userService)
	uptimeHandler := handlers.NewUptimeHandler(uptimeService)
	alertManagerHandler := handlers.NewAlertManagerHandler(alertManagerService)
	authHandler := handlers.NewAuthHandler(authService)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService, alertService, userService)

	// Initialize middleware
	authMiddleware := handlers.NewAuthMiddleware(authService.JWTService)

	// AUTHENTICATION
	r.POST("/auth/login", authHandler.Login)
	r.POST("/auth/change-password", authHandler.ChangePassword)
	r.POST("/auth/setup-admin", authHandler.SetupAdmin)

	// ALERTS
	r.GET("/alerts", alertHandler.ListAlerts)
	r.POST("/alerts", alertHandler.CreateAlert)
	r.GET("/alerts/:id", alertHandler.GetAlert)
	r.POST("/alerts/:id/ack", alertHandler.AckAlert)
	r.POST("/alerts/:id/unack", alertHandler.UnackAlert)
	r.POST("/alerts/:id/close", alertHandler.CloseAlert)

	// ALERTMANAGER INTEGRATION
	r.POST("/alertmanager/webhook", alertManagerHandler.ReceiveWebhook)
	r.GET("/alertmanager/info", alertManagerHandler.GetWebhookInfo)

	// API KEY MANAGEMENT (requires JWT authentication)
	apiKeyRoutes := r.Group("/api-keys")
	apiKeyRoutes.Use(authMiddleware.JWTAuthMiddleware())
	{
		apiKeyRoutes.POST("", apiKeyHandler.CreateAPIKey)
		apiKeyRoutes.GET("", apiKeyHandler.ListAPIKeys)
		apiKeyRoutes.GET("/:id", apiKeyHandler.GetAPIKey)
		apiKeyRoutes.PUT("/:id", apiKeyHandler.UpdateAPIKey)
		apiKeyRoutes.DELETE("/:id", apiKeyHandler.DeleteAPIKey)
		apiKeyRoutes.POST("/:id/regenerate", apiKeyHandler.RegenerateAPIKey)
		apiKeyRoutes.GET("/stats", apiKeyHandler.GetAPIKeyStats)
	}

	// WEBHOOK ENDPOINTS (uses API key authentication)
	r.POST("/alert/webhook", apiKeyHandler.WebhookAlert)

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

	// UPTIME MONITORING
	r.GET("/uptime", uptimeHandler.GetUptimeDashboard)
	r.GET("/uptime/services", uptimeHandler.ListServices)
	r.POST("/uptime/services", uptimeHandler.CreateService)
	r.GET("/uptime/services/:id", uptimeHandler.GetService)
	r.GET("/uptime/services/:id/stats", uptimeHandler.GetServiceStats)
	r.GET("/uptime/services/:id/history", uptimeHandler.GetServiceHistory)

	// DASHBOARD
	r.GET("/dashboard", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Dashboard endpoint - TODO implement"})
	})

	// TODO: notes, tags, log, recipients endpoints

	return r
}
