package main

import (
	"context"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"os/signal"
	"personal-homepage-service/config"
	"personal-homepage-service/core"
	"personal-homepage-service/workers/shipments"
	"syscall"
)

func main() {
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})

	if err != nil {
		log.Fatal(err)
	}

	orchestrator := core.NewOrchestrator([]core.Worker{
		shipments.NewWorker(db),
	})

	c, err := orchestrator.Start(context.Background())
	defer c.Stop()

	if err != nil {
		log.Fatal(err)
	}

	// Wait for termination signal to exit gracefully
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
