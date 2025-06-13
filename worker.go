package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func StartWorker(pg *sql.DB, redis *redis.Client) {
	for {
		// Lấy alert từ queue
		res, err := redis.BLPop(context.Background(), 0, "alerts:queue").Result()
		if err != nil || len(res) < 2 {
			time.Sleep(time.Second)
			continue
		}
		var alert Alert
		json.Unmarshal([]byte(res[1]), &alert)
		log.Printf("Worker: processing alert %s", alert.ID)

		// Lock alert (set key với TTL)
		lockKey := "alerts:lock:" + alert.ID
		ok, _ := redis.SetNX(context.Background(), lockKey, "locked", 5*time.Minute).Result()
		if !ok {
			continue // Đã có worker khác xử lý
		}

		// Push FCM (mock)
		log.Printf("Push FCM for alert %s", alert.ID)
		// TODO: Gọi hàm push FCM thực tế ở đây

		// Chờ ACK (giả lập 5 phút)
		ackKey := "alerts:ack:" + alert.ID
		ack, _ := redis.BLPop(context.Background(), 5*60*time.Second, ackKey).Result()
		if ack == nil || len(ack) < 2 {
			// Escalate
			log.Printf("Escalate alert %s", alert.ID)
			pg.Exec(`UPDATE alerts SET status='escalated', updated_at=$1 WHERE id=$2`, time.Now(), alert.ID)
		} else {
			// Đã ACK
			log.Printf("Alert %s ACKed", alert.ID)
			pg.Exec(`UPDATE alerts SET status='acked', updated_at=$1 WHERE id=$2`, time.Now(), alert.ID)
		}
		redis.Del(context.Background(), lockKey)
	}
}
