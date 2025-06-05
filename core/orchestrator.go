package core

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
)

type Orchestrator struct {
	workers []Worker
}

func NewOrchestrator(workers []Worker) *Orchestrator {
	return &Orchestrator{workers}
}

func (o *Orchestrator) Start(ctx context.Context) (*cron.Cron, error) {
	c := cron.New()

	for _, worker := range o.workers {
		_, err := c.AddFunc(worker.Schedule(), func() {
			if worker.Ready() {
				go worker.Execute()
			}
		})

		if err != nil {
			fmt.Println("Error adding cron job:", err)
			return nil, err
		}
	}

	c.Start()
	return c, nil
}
