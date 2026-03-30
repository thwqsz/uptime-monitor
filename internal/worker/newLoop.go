package worker

import (
	"context"
	"time"

	"github.com/thwqsz/uptime-monitor/internal/models"
	"go.uber.org/zap"
)

type GetAllTargeter interface {
	GetAllTargets(ctx context.Context) ([]*models.Target, error)
}

type Checker interface {
	CheckTarget(ctx context.Context, targetID int64) (*models.CheckLog, error)
}

type Loop struct {
	source      GetAllTargeter
	targetCheck Checker
	workerCount int
	jobs        chan int64
	log         *zap.Logger
}

func NewLoop(source GetAllTargeter, targetCheck Checker, workerCount int, log *zap.Logger) *Loop {
	return &Loop{
		source:      source,
		targetCheck: targetCheck,
		workerCount: workerCount,
		jobs:        make(chan int64, workerCount),
		log:         log,
	}
}

func (w *Loop) Run(ctx context.Context) {
	for i := 0; i < w.workerCount; i++ {
		go w.worker(ctx)
	}
	targets, err := w.source.GetAllTargets(ctx)
	if err != nil {
		w.log.Error("error with db", zap.Error(err))
		return
	}
	for _, x := range targets {
		go w.targetScheduler(ctx, x)
	}
	select {
	case <-ctx.Done():
		return
	}
}

func (w *Loop) targetScheduler(ctx context.Context, target *models.Target) {
	ticker := time.NewTicker(time.Duration(target.IntervalTime) * time.Second)
	defer ticker.Stop()
	select {
	case <-ctx.Done():
		return
	case w.jobs <- target.ID:
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			select {
			case <-ctx.Done():
				return
			case w.jobs <- target.ID:
			}
		}
	}
}

func (w *Loop) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-w.jobs:
			if !ok {
				return
			}
			_, err := w.targetCheck.CheckTarget(ctx, job)
			if err != nil {
				w.log.Error("error during check", zap.Int64("ID", job), zap.Error(err))
			}
		}
	}
}
