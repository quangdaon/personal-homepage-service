package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"personal-homepage-service/config"
	"personal-homepage-service/workers/shipments"
)

func main() {
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})

	if err != nil {
		log.Fatal(err)
	}

	shipmentWorker := shipments.NewWorker(db)

	shipmentWorker.Execute()
}
