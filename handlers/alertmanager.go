package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vanchonlee/oncallkit/models"
	"github.com/vanchonlee/oncallkit/services"
)

type AlertManagerHandler struct {
	Service *services.AlertManagerService
}

func NewAlertManagerHandler(service *services.AlertManagerService) *AlertManagerHandler {
	return &AlertManagerHandler{Service: service}
}

// ReceiveWebhook handles incoming AlertManager webhooks
func (h *AlertManagerHandler) ReceiveWebhook(c *gin.Context) {
	var webhook models.AlertManagerWebhook

	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload: " + err.Error()})
		return
	}

	// Process the webhook
	if err := h.Service.ProcessWebhook(&webhook); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Webhook processed successfully",
		"alerts_processed": len(webhook.Alerts),
	})
}

// GetWebhookInfo returns information about the webhook endpoint
func (h *AlertManagerHandler) GetWebhookInfo(c *gin.Context) {
	info := gin.H{
		"endpoint":           "/api/alertmanager/webhook",
		"method":             "POST",
		"description":        "Receives webhooks from Prometheus AlertManager",
		"supported_versions": []string{"4"},
		"example_config": gin.H{
			"alertmanager_yml": `
route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'slar-webhook'

receivers:
- name: 'slar-webhook'
  webhook_configs:
  - url: 'http://your-slar-api/api/alertmanager/webhook'
    send_resolved: true
`,
		},
	}

	c.JSON(http.StatusOK, info)
}
