package worker

import (
	"context"
	"time"

	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/repository"
	"github.com/thwqsz/uptime-monitor/internal/service"
	"go.uber.org/zap"
)

type Loop struct {
	check       *service.CheckService
	targets     repository.TargetRepository
	log         *zap.Logger
	workerCount int
	jobs        chan Job
}

func NewLoop(check *service.CheckService, targets repository.TargetRepository, log *zap.Logger, workerCount int) *Loop {
	return &Loop{check: check, targets: targets, log: log, workerCount: workerCount, jobs: make(chan Job, workerCount)}
}

type Job struct {
	TargetID int64
	DoneCh   chan struct{}
}

func (w *Loop) Run(ctx context.Context) {
	for i := 0; i < w.workerCount; i++ {
		go w.worker(ctx)
	}
	for {
		if ctx.Err() != nil {
			return
		}
		newTarg, err := w.targets.GetAllTargets(ctx)
		if err != nil {
			w.log.Error("error with db", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(30 * time.Second):
				continue
			}
		}
		w.runBatch(ctx, newTarg)
		select {
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
			continue
		}
	}
}

func (w *Loop) runBatch(ctx context.Context, targets []*models.Target) {
	jobsCount := len(targets)
	doneCh := make(chan struct{}, jobsCount)
	for _, x := range targets {
		select {
		case w.jobs <- Job{
			TargetID: x.ID,
			DoneCh:   doneCh,
		}:
		case <-ctx.Done():
			return
		}
	}
	for i := 0; i < jobsCount; i++ {
		select {
		case <-ctx.Done():
			return
		case <-doneCh:
		}
	}
}

func (w *Loop) worker(ctx context.Context) {
	for {
		select {
		case job, ok := <-w.jobs:
			if !ok {
				return
			}
			_, err := w.check.CheckTarget(ctx, job.TargetID)
			if err != nil {
				w.log.Error("error during check", zap.Int64("ID", job.TargetID), zap.Error(err))
			}
			job.DoneCh <- struct{}{}
		case <-ctx.Done():
			return
		}
	}
}
