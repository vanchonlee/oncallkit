package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/vanchonlee/oncallkit/db"
	"github.com/vanchonlee/oncallkit/models"
)

type AlertManagerService struct {
	PG           *sql.DB
	AlertService *AlertService
}

func NewAlertManagerService(pg *sql.DB, alertService *AlertService) *AlertManagerService {
	return &AlertManagerService{
		PG:           pg,
		AlertService: alertService,
	}
}

// ProcessWebhook processes incoming AlertManager webhook and creates alerts
func (s *AlertManagerService) ProcessWebhook(webhook *models.AlertManagerWebhook) error {
	for _, amAlert := range webhook.Alerts {
		// Convert AlertManager alert to internal alert
		alert, err := s.convertToInternalAlert(webhook, &amAlert)
		if err != nil {
			return fmt.Errorf("failed to convert alert: %w", err)
		}

		// Handle different alert statuses
		switch amAlert.Status {
		case "firing":
			err = s.handleFiringAlert(alert, &amAlert)
		case "resolved":
			err = s.handleResolvedAlert(alert, &amAlert)
		default:
			continue // Skip unknown statuses
		}

		if err != nil {
			return fmt.Errorf("failed to handle alert status %s: %w", amAlert.Status, err)
		}
	}

	return nil
}

// convertToInternalAlert converts AlertManager alert to internal alert format
func (s *AlertManagerService) convertToInternalAlert(webhook *models.AlertManagerWebhook, amAlert *models.AlertManagerAlert) (*db.Alert, error) {
	// Generate alert ID based on fingerprint or labels
	alertID := amAlert.Fingerprint
	if alertID == "" {
		alertID = s.generateAlertID(amAlert.Labels)
	}

	// Extract severity from labels
	severity := s.extractSeverity(amAlert.Labels)

	// Create description from annotations
	description := s.createDescription(amAlert.Annotations, amAlert.Labels)

	alert := &db.Alert{
		ID:          alertID,
		Title:       amAlert.Labels["alertname"],
		Description: description,
		Severity:    severity,
		Status:      "new",
		Source:      "alertmanager",
		CreatedAt:   amAlert.StartsAt,
		UpdatedAt:   time.Now(),
	}

	return alert, nil
}

// handleFiringAlert handles firing alerts
func (s *AlertManagerService) handleFiringAlert(alert *db.Alert, amAlert *models.AlertManagerAlert) error {
	// Check if alert already exists
	var existingAlert db.Alert
	err := s.PG.QueryRow("SELECT id, status FROM alerts WHERE id = $1", alert.ID).Scan(&existingAlert.ID, &existingAlert.Status)

	if err == sql.ErrNoRows {
		// Create new alert
		alert.Status = "new"
		_, err = s.PG.Exec(`INSERT INTO alerts (id, title, description, status, created_at, updated_at, severity, source) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			alert.ID, alert.Title, alert.Description, alert.Status, alert.CreatedAt, alert.UpdatedAt, alert.Severity, alert.Source)
		return err
	} else if err != nil {
		return err
	}

	// Update existing alert if it was closed
	if existingAlert.Status == "closed" {
		now := time.Now()
		_, err = s.PG.Exec("UPDATE alerts SET status = 'new', updated_at = $1 WHERE id = $2", now, alert.ID)
		return err
	}

	return nil
}

// handleResolvedAlert handles resolved alerts
func (s *AlertManagerService) handleResolvedAlert(alert *db.Alert, amAlert *models.AlertManagerAlert) error {
	var existingAlert db.Alert
	err := s.PG.QueryRow("SELECT id, status FROM alerts WHERE id = $1", alert.ID).Scan(&existingAlert.ID, &existingAlert.Status)

	if err == sql.ErrNoRows {
		// Alert doesn't exist, create it as closed
		alert.Status = "closed"
		_, err = s.PG.Exec(`INSERT INTO alerts (id, title, description, status, created_at, updated_at, severity, source) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			alert.ID, alert.Title, alert.Description, alert.Status, alert.CreatedAt, alert.UpdatedAt, alert.Severity, alert.Source)
		return err
	} else if err != nil {
		return err
	}

	// Update existing alert to closed
	if existingAlert.Status != "closed" {
		now := time.Now()
		_, err = s.PG.Exec("UPDATE alerts SET status = 'closed', updated_at = $1 WHERE id = $2", now, alert.ID)
		return err
	}

	return nil
}

// Helper functions

func (s *AlertManagerService) generateAlertID(labels map[string]string) string {
	// Create a consistent ID from labels
	var parts []string
	for key, value := range labels {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	return fmt.Sprintf("am-%x", strings.Join(parts, ","))
}

func (s *AlertManagerService) extractSeverity(labels map[string]string) string {
	if severity, exists := labels["severity"]; exists {
		switch strings.ToLower(severity) {
		case "critical":
			return "critical"
		case "warning":
			return "warning"
		case "info":
			return "info"
		default:
			return "warning"
		}
	}
	return "warning"
}

func (s *AlertManagerService) createDescription(annotations, labels map[string]string) string {
	// Try to get description from annotations
	if summary, exists := annotations["summary"]; exists {
		return summary
	}
	if description, exists := annotations["description"]; exists {
		return description
	}

	// Fallback to creating description from labels
	if alertname, exists := labels["alertname"]; exists {
		if instance, exists := labels["instance"]; exists {
			return fmt.Sprintf("%s on %s", alertname, instance)
		}
		return alertname
	}

	return "Alert from AlertManager"
}
