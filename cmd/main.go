package main

import (
	"log"

	"github.com/vanchonlee/oncallkit/db"
	"github.com/vanchonlee/oncallkit/router"
	"github.com/vanchonlee/oncallkit/workers"
)

func main() {
	// Connect to DB and Redis
	pg := db.NewPostgres("postgres://slar:slar@localhost:5432/slar?sslmode=disable")
	redis := db.NewRedis("localhost:6379")

	// Start worker
	go workers.StartWorker(pg, redis)

	// Start API server
	r := router.NewGinRouter(pg, redis)
	log.Println("API server running at :8080")
	r.Run(":8080")
}
