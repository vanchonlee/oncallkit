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

func (s *AlertService) ListAlerts() ([]db.AlertResponse, error) {
	query := `
		SELECT 
			a.id, a.title, a.description, a.status, a.created_at, a.updated_at, 
			a.severity, a.source, a.assigned_to, a.assigned_at,
			u.name, u.email
		FROM alerts a
		LEFT JOIN users u ON a.assigned_to = u.id
		ORDER BY a.created_at DESC 
		LIMIT 100
	`

	rows, err := s.PG.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []db.AlertResponse
	for rows.Next() {
		var a db.AlertResponse
		var assignedTo sql.NullString
		var assignedAt sql.NullTime
		var userName sql.NullString
		var userEmail sql.NullString

		err := rows.Scan(
			&a.ID, &a.Title, &a.Description, &a.Status, &a.CreatedAt, &a.UpdatedAt,
			&a.Severity, &a.Source, &assignedTo, &assignedAt,
			&userName, &userEmail,
		)
		if err != nil {
			continue
		}

		if assignedTo.Valid {
			a.AssignedTo = assignedTo.String
		}
		if assignedAt.Valid {
			a.AssignedAt = &assignedAt.Time
		}
		if userName.Valid {
			a.AssignedToName = userName.String
		}
		if userEmail.Valid {
			a.AssignedToEmail = userEmail.String
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

func (s *AlertService) GetAlert(id string) (db.AlertResponse, error) {
	var a db.AlertResponse
	var assignedTo sql.NullString
	var assignedAt sql.NullTime
	var userName sql.NullString
	var userEmail sql.NullString

	query := `
		SELECT 
			a.id, a.title, a.description, a.status, a.created_at, a.updated_at, 
			a.severity, a.source, a.assigned_to, a.assigned_at,
			u.name, u.email
		FROM alerts a
		LEFT JOIN users u ON a.assigned_to = u.id
		WHERE a.id = $1
	`

	err := s.PG.QueryRow(query, id).Scan(
		&a.ID, &a.Title, &a.Description, &a.Status, &a.CreatedAt, &a.UpdatedAt,
		&a.Severity, &a.Source, &assignedTo, &assignedAt,
		&userName, &userEmail,
	)

	if assignedTo.Valid {
		a.AssignedTo = assignedTo.String
	}
	if assignedAt.Valid {
		a.AssignedAt = &assignedAt.Time
	}
	if userName.Valid {
		a.AssignedToName = userName.String
	}
	if userEmail.Valid {
		a.AssignedToEmail = userEmail.String
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
