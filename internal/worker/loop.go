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
	check   *service.CheckService
	targets repository.TargetRepository
	log     *zap.Logger
}

func NewLoop(check *service.CheckService, targets repository.TargetRepository, log *zap.Logger) *Loop {
	return &Loop{check: check, targets: targets, log: log}
}

func (w *Loop) Run(ctx context.Context) {
	for {
		allTargets, err := w.targets.GetAllTargets(ctx)
		if err != nil {
			w.log.Error("failed to get targets", zap.Error(err))
			time.Sleep(30 * time.Second)
			continue
		}
		w.checkAllTargets(ctx, allTargets)
		time.Sleep(time.Second * 30)
	}
}

func (w *Loop) checkAllTargets(ctx context.Context, allTargets []*models.Target) {
	for _, x := range allTargets {
		_, err := w.check.CheckTarget(ctx, x.ID)
		if err != nil {
			w.log.Error("error during check", zap.String("url", x.URL), zap.Error(err))
		}
	}
}
