package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

func NewGinRouter(pg *sql.DB, redis *redis.Client) *gin.Engine {
	r := gin.Default()

	// ALERTS
	r.GET("/alerts", func(c *gin.Context) {
		rows, err := pg.Query(`SELECT id, title, description, status, created_at, updated_at, severity, source FROM alerts ORDER BY created_at DESC LIMIT 100`)
		if err != nil {
			c.JSON(500, gin.H{"error": "db error"})
			return
		}
		defer rows.Close()
		var alerts []Alert
		for rows.Next() {
			var a Alert
			rows.Scan(&a.ID, &a.Title, &a.Description, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.Severity, &a.Source)
			alerts = append(alerts, a)
		}
		c.JSON(200, alerts)
	})

	r.POST("/alerts", func(c *gin.Context) {
		var alert Alert
		if err := c.ShouldBindJSON(&alert); err != nil {
			c.JSON(400, gin.H{"error": "bad request"})
			return
		}
		alert.ID = uuid.New().String()
		alert.Status = "new"
		alert.CreatedAt = time.Now()
		alert.UpdatedAt = time.Now()
		_, err := pg.Exec(`INSERT INTO alerts (id, title, description, status, created_at, updated_at, severity, source) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			alert.ID, alert.Title, alert.Description, alert.Status, alert.CreatedAt, alert.UpdatedAt, alert.Severity, alert.Source)
		if err != nil {
			c.JSON(500, gin.H{"error": "db error"})
			return
		}
		b, _ := json.Marshal(alert)
		redis.RPush(context.Background(), "alerts:queue", b)
		c.JSON(201, alert)
	})

	r.GET("/alerts/:id", func(c *gin.Context) {
		id := c.Param("id")
		var a Alert
		err := pg.QueryRow(`SELECT id, title, description, status, created_at, updated_at, severity, source FROM alerts WHERE id=$1`, id).
			Scan(&a.ID, &a.Title, &a.Description, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.Severity, &a.Source)
		if err != nil {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		c.JSON(200, a)
	})

	r.POST("/alerts/:id/ack", func(c *gin.Context) {
		id := c.Param("id")
		now := time.Now()
		_, err := pg.Exec(`UPDATE alerts SET status='acked', acked_at=$1, updated_at=$1 WHERE id=$2`, now, id)
		if err != nil {
			c.JSON(500, gin.H{"error": "db error"})
			return
		}
		c.Status(200)
	})

	r.POST("/alerts/:id/unack", func(c *gin.Context) {
		id := c.Param("id")
		now := time.Now()
		_, err := pg.Exec(`UPDATE alerts SET status='new', updated_at=$1 WHERE id=$2`, now, id)
		if err != nil {
			c.JSON(500, gin.H{"error": "db error"})
			return
		}
		c.Status(200)
	})

	r.POST("/alerts/:id/close", func(c *gin.Context) {
		id := c.Param("id")
		now := time.Now()
		_, err := pg.Exec(`UPDATE alerts SET status='closed', updated_at=$1 WHERE id=$2`, now, id)
		if err != nil {
			c.JSON(500, gin.H{"error": "db error"})
			return
		}
		c.Status(200)
	})

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
