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
	rows, err := s.PG.Query(`SELECT id, title, description, status, created_at, updated_at, severity, source FROM alerts ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var alerts []db.Alert
	for rows.Next() {
		var a db.Alert
		rows.Scan(&a.ID, &a.Title, &a.Description, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.Severity, &a.Source)
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
	_, err := s.PG.Exec(`INSERT INTO alerts (id, title, description, status, created_at, updated_at, severity, source) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		alert.ID, alert.Title, alert.Description, alert.Status, alert.CreatedAt, alert.UpdatedAt, alert.Severity, alert.Source)
	if err != nil {
		return alert, err
	}
	b, _ := json.Marshal(alert)
	s.Redis.RPush(context.Background(), "alerts:queue", b)
	return alert, nil
}

func (s *AlertService) GetAlert(id string) (db.Alert, error) {
	var a db.Alert
	err := s.PG.QueryRow(`SELECT id, title, description, status, created_at, updated_at, severity, source FROM alerts WHERE id=$1`, id).
		Scan(&a.ID, &a.Title, &a.Description, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.Severity, &a.Source)
	return a, err
}

func (s *AlertService) AckAlert(id string) error {
	now := time.Now()
	_, err := s.PG.Exec(`UPDATE alerts SET status='acked', acked_at=$1, updated_at=$1 WHERE id=$2`, now, id)
	return err
}

func (s *AlertService) UnackAlert(id string) error {
	now := time.Now()
	_, err := s.PG.Exec(`UPDATE alerts SET status='new', updated_at=$1 WHERE id=$2`, now, id)
	return err
}

func (s *AlertService) CloseAlert(id string) error {
	now := time.Now()
	_, err := s.PG.Exec(`UPDATE alerts SET status='closed', updated_at=$1 WHERE id=$2`, now, id)
	return err
}
