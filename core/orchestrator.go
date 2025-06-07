package core

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"time"
)

type Orchestrator struct {
	logger  *zap.Logger
	workers []Worker
}

func NewOrchestrator(logger *zap.Logger, workers []Worker) *Orchestrator {
	return &Orchestrator{logger, workers}
}

func (o *Orchestrator) Start(ctx context.Context) (*cron.Cron, error) {
	c := cron.New()

	for _, worker := range o.workers {
		_, err := c.AddFunc(worker.Schedule(), func() {
			now := time.Now()
			if worker.Ready(now) {
				go worker.Execute()
			}
		})

		if err != nil {
			o.logger.Error("Error adding cron job",
				zap.String("worker", fmt.Sprintf("%T", worker)),
				zap.String("details", err.Error()),
			)
			return nil, err
		}

		o.logger.Info("Worker started",
			zap.String("worker", fmt.Sprintf("%T", worker)),
			zap.String("schedule", worker.Schedule()),
		)
	}

	c.Start()
	return c, nil
}
