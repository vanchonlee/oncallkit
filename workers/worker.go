package workers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/vanchonlee/oncallkit/db"
)

func StartWorker(pg *sql.DB, redis *redis.Client) {
	log.Println("Worker started, waiting for alerts...")
	for {
		// Get alert from queue
		res, err := redis.BLPop(context.Background(), 0, "alerts:queue").Result()
		if err != nil || len(res) < 2 {
			time.Sleep(time.Second)
			continue
		}
		var alert db.Alert
		json.Unmarshal([]byte(res[1]), &alert)
		log.Printf("Worker: processing alert %s", alert.ID)

		// Lock alert (set key with TTL)
		lockKey := "alerts:lock:" + alert.ID
		ok, _ := redis.SetNX(context.Background(), lockKey, "locked", 5*time.Minute).Result()
		if !ok {
			log.Printf("Alert %s already locked by another worker", alert.ID)
			continue // Already processed by another worker
		}

		// Push FCM immediately (mock)
		log.Printf("Push FCM for alert %s", alert.ID)
		// TODO: Call FCM push function here

		// Handle ACK/escalation in separate goroutine
		go handleAlertAck(pg, redis, alert)
	}
}

func handleAlertAck(pg *sql.DB, redis *redis.Client, alert db.Alert) {
	defer func() {
		// Release lock when done
		lockKey := "alerts:lock:" + alert.ID
		redis.Del(context.Background(), lockKey)
		log.Printf("Worker: released lock for alert %s", alert.ID)
	}()

	// Set escalation timer using Redis TTL (5 minutes)
	escalationKey := "alerts:escalation:" + alert.ID
	err := redis.Set(context.Background(), escalationKey, "pending", 5*time.Minute).Err()
	if err != nil {
		log.Printf("Worker: failed to set escalation timer for alert %s: %v", alert.ID, err)
		return
	}
	log.Printf("Worker: set escalation timer (5min) for alert %s", alert.ID)

	ackKey := "alerts:ack:" + alert.ID

	// Poll for ACK or escalation timeout
	for {
		// Check if ACK received
		ackResult, err := redis.Get(context.Background(), ackKey).Result()
		if err == nil {
			log.Printf("Worker: alert %s acknowledged by %s", alert.ID, ackResult)
			// Clean up escalation timer
			redis.Del(context.Background(), escalationKey)
			return
		}

		// Check if escalation timer expired
		exists, err := redis.Exists(context.Background(), escalationKey).Result()
		if err != nil {
			log.Printf("Worker: error checking escalation timer for alert %s: %v", alert.ID, err)
			time.Sleep(10 * time.Second)
			continue
		}

		if exists == 0 {
			// Escalation timer expired, escalate alert
			log.Printf("Worker: escalating alert %s (no ACK after 5 minutes)", alert.ID)

			// Update alert status to escalated in DB
			_, err := pg.Exec("UPDATE alerts SET status = 'escalated', updated_at = NOW() WHERE id = $1", alert.ID)
			if err != nil {
				log.Printf("Worker: failed to update alert %s status to escalated: %v", alert.ID, err)
			}

			// TODO: Send escalation notification (email, SMS, etc.)
			log.Printf("Worker: sent escalation notification for alert %s", alert.ID)
			return
		}

		// Sleep before next check (poll every 10 seconds)
		time.Sleep(10 * time.Second)
	}
}

func StartUptimeWorker(pg *sql.DB, redis *redis.Client) {
	log.Println("Uptime worker started, monitoring services...")

	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get active services from database
			services, err := getActiveServices(pg)
			if err != nil {
				log.Printf("Uptime worker: failed to get services from database: %v", err)
				continue
			}

			for _, service := range services {
				go checkServiceUptime(pg, redis, service.Name, service.URL)
			}
		}
	}
}

func getActiveServices(pg *sql.DB) ([]db.Service, error) {
	rows, err := pg.Query(`
		SELECT id, name, url, type, method, interval_seconds, timeout_seconds, expected_status 
		FROM services 
		WHERE is_active = true AND is_enabled = true
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []db.Service
	for rows.Next() {
		var service db.Service
		err := rows.Scan(
			&service.ID,
			&service.Name,
			&service.URL,
			&service.Type,
			&service.Method,
			&service.Interval,
			&service.Timeout,
			&service.ExpectedStatus,
		)
		if err != nil {
			log.Printf("Error scanning service: %v", err)
			continue
		}
		services = append(services, service)
	}

	return services, nil
}

func checkServiceUptime(pg *sql.DB, redis *redis.Client, serviceName, url string) {
	start := time.Now()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	duration := time.Since(start)

	isUp := err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 400
	if resp != nil {
		resp.Body.Close()
	}

	status := "up"
	if !isUp {
		status = "down"
		log.Printf("Uptime worker: %s is DOWN (error: %v)", serviceName, err)

		// Generate alert for downtime
		alert := db.Alert{
			ID:          generateAlertID(),
			Title:       "Service Down: " + serviceName,
			Description: "Service " + serviceName + " is down",
			Status:      "open",
			Severity:    "critical",
			Source:      "uptime-monitor",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Add alert to queue
		alertJSON, _ := json.Marshal(alert)
		redis.LPush(context.Background(), "alerts:queue", alertJSON)
		log.Printf("Uptime worker: queued downtime alert for %s", serviceName)
	} else {
		log.Printf("Uptime worker: %s is UP (response time: %v)", serviceName, duration)
	}

	// Store uptime check result in Redis
	uptimeKey := "uptime:" + serviceName
	uptimeData := map[string]interface{}{
		"status":        status,
		"response_time": duration.Milliseconds(),
		"checked_at":    time.Now().Unix(),
		"url":           url,
	}
	uptimeJSON, _ := json.Marshal(uptimeData)
	redis.Set(context.Background(), uptimeKey, uptimeJSON, 5*time.Minute)
}

func generateAlertID() string {
	return time.Now().Format("20060102150405") + "-uptime"
}
