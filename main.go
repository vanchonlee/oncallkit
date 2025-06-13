package main

import (
	"log"
)

func main() {
	// Kết nối DB và Redis
	pg := NewPostgres("localhost:5432")
	redis := NewRedis("localhost:6379")

	// Khởi động worker (chạy goroutine)
	go StartWorker(pg, redis)

	// Khởi động API server với Gin
	r := NewGinRouter(pg, redis)
	log.Println("API server running at :8080")
	r.Run(":8080")
}
