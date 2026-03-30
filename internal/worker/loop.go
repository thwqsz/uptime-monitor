package worker

import (
	"context"
	"sync"
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
	source          GetAllTargeter
	targetCheck     Checker
	workerCount     int
	jobs            chan int64
	log             *zap.Logger
	targetCancelMap map[int64]context.CancelFunc
	mu              sync.Mutex
	ctx             context.Context
}

func NewLoop(source GetAllTargeter, targetCheck Checker, workerCount int, log *zap.Logger, ctx context.Context) *Loop {
	return &Loop{
		source:          source,
		targetCheck:     targetCheck,
		workerCount:     workerCount,
		jobs:            make(chan int64, workerCount),
		log:             log,
		targetCancelMap: make(map[int64]context.CancelFunc),
		ctx:             ctx,
	}
}

func (w *Loop) Run() {
	for i := 0; i < w.workerCount; i++ {
		go w.worker()
	}
	targets, err := w.source.GetAllTargets(w.ctx)
	if err != nil {
		w.log.Error("error with db", zap.Error(err))
		return
	}
	for _, x := range targets {
		w.StartTarget(x)
	}
	select {
	case <-w.ctx.Done():
		return
	}
}

func (w *Loop) StartTarget(target *models.Target) {
	w.mu.Lock()
	if w.targetCancelMap[target.ID] != nil {
		w.log.Error("target is already in work")
		w.mu.Unlock()
		return
	}
	ctxForTarget, cancel := context.WithCancel(w.ctx)
	w.targetCancelMap[target.ID] = cancel
	w.mu.Unlock()
	go w.targetScheduler(ctxForTarget, target)
}

func (w *Loop) StopTarget(targetID int64) {
	w.mu.Lock()
	if w.targetCancelMap[targetID] == nil {
		w.log.Error("target id not found")
		w.mu.Unlock()
		return
	}
	cancel := w.targetCancelMap[targetID]
	delete(w.targetCancelMap, targetID)
	w.mu.Unlock()
	cancel()
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

func (w *Loop) worker() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case job, ok := <-w.jobs:
			if !ok {
				return
			}
			_, err := w.targetCheck.CheckTarget(w.ctx, job)
			if err != nil {
				w.log.Error("error during check", zap.Int64("ID", job), zap.Error(err))
			}
		}
	}
}
