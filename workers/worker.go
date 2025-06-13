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
			continue // Already processed by another worker
		}

		// Push FCM (mock)
		log.Printf("Push FCM for alert %s", alert.ID)
		// TODO: Call FCM push function here

		// Wait for ACK (mock 5 minutes)
		ackKey := "alerts:ack:" + alert.ID
		ack, _ := redis.BLPop(context.Background(), 5*60*time.Second, ackKey).Result()
		if ack == nil || len(ack) < 2 {
			// Escalate
			log.Printf("Escalate alert %s", alert.ID)
			pg.Exec(`UPDATE alerts SET status='escalated', updated_at=$1 WHERE id=$2`, time.Now(), alert.ID)
		} else {
			// ACKed
			log.Printf("Alert %s ACKed", alert.ID)
			pg.Exec(`UPDATE alerts SET status='acked', updated_at=$1 WHERE id=$2`, time.Now(), alert.ID)
		}
		redis.Del(context.Background(), lockKey)
	}
}
