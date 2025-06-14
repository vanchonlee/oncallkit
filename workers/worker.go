package workers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
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
