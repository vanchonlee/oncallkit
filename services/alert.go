package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/vanchonlee/oncallkit/db"
)

type AlertService struct {
	PG    *sql.DB
	Redis *redis.Client
}

func NewAlertService(pg *sql.DB, redis *redis.Client) *AlertService {
	return &AlertService{PG: pg, Redis: redis}
}

func (s *AlertService) ListAlerts() ([]db.Alert, error) {
	rows, err := s.PG.Query(`SELECT id, title, description, status, created_at, updated_at, severity, source, assigned_to, assigned_at FROM alerts ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var alerts []db.Alert
	for rows.Next() {
		var a db.Alert
		var assignedTo sql.NullString
		var assignedAt sql.NullTime
		err := rows.Scan(&a.ID, &a.Title, &a.Description, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.Severity, &a.Source, &assignedTo, &assignedAt)
		if err != nil {
			continue
		}
		if assignedTo.Valid {
			a.AssignedTo = assignedTo.String
		}
		if assignedAt.Valid {
			a.AssignedAt = &assignedAt.Time
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

func (s *AlertService) CreateAlertFromRequest(c *gin.Context) (db.Alert, error) {
	var alert db.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		return alert, err
	}
	alert.ID = uuid.New().String()
	alert.Status = "new"
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()

	// Auto-assign to current on-call user
	userService := NewUserService(s.PG, s.Redis)
	onCallUser, err := userService.GetCurrentOnCallUser()
	if err == nil {
		alert.AssignedTo = onCallUser.ID
		now := time.Now()
		alert.AssignedAt = &now
	}

	_, err = s.PG.Exec(`INSERT INTO alerts (id, title, description, status, created_at, updated_at, severity, source, assigned_to, assigned_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		alert.ID, alert.Title, alert.Description, alert.Status, alert.CreatedAt, alert.UpdatedAt, alert.Severity, alert.Source, alert.AssignedTo, alert.AssignedAt)
	if err != nil {
		return alert, err
	}
	b, _ := json.Marshal(alert)
	s.Redis.RPush(context.Background(), "alerts:queue", b)
	return alert, nil
}

func (s *AlertService) GetAlert(id string) (db.Alert, error) {
	var a db.Alert
	var assignedTo sql.NullString
	var assignedAt sql.NullTime
	err := s.PG.QueryRow(`SELECT id, title, description, status, created_at, updated_at, severity, source, assigned_to, assigned_at FROM alerts WHERE id=$1`, id).
		Scan(&a.ID, &a.Title, &a.Description, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.Severity, &a.Source, &assignedTo, &assignedAt)
	if assignedTo.Valid {
		a.AssignedTo = assignedTo.String
	}
	if assignedAt.Valid {
		a.AssignedAt = &assignedAt.Time
	}
	return a, err
}

func (s *AlertService) AckAlert(id string) error {
	now := time.Now()
	_, err := s.PG.Exec(`UPDATE alerts SET status = 'acked', acked_at = $1, updated_at = $2 WHERE id = $3`, now, now, id)
	return err
}

func (s *AlertService) UnackAlert(id string) error {
	now := time.Now()
	_, err := s.PG.Exec(`UPDATE alerts SET status = 'new', acked_at = NULL, updated_at = $1 WHERE id = $2`, now, id)
	return err
}

func (s *AlertService) CloseAlert(id string) error {
	now := time.Now()
	_, err := s.PG.Exec(`UPDATE alerts SET status = 'closed', updated_at = $1 WHERE id = $2`, now, id)
	return err
}

func (s *AlertService) AssignAlertToUser(alertID, userID string) error {
	now := time.Now()
	_, err := s.PG.Exec(`UPDATE alerts SET assigned_to = $1, assigned_at = $2, updated_at = $3 WHERE id = $4`,
		userID, now, now, alertID)
	return err
}
