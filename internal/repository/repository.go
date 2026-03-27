package repository

import (
	"context"

	"github.com/thwqsz/uptime-monitor/internal/models"
)

type CheckLogRepository interface {
	CreateCheckLog(ctx context.Context, log *models.CheckLog) error
}

type TargetRepository interface {
	CreateTarget(ctx context.Context, target *models.Target) error
	GetTargetsByUserID(ctx context.Context, userID int64) ([]*models.Target, error)
	DeleteTarget(ctx context.Context, targetID, userID int64) (int64, error)
	GetTargetByID(ctx context.Context, targetID int64) (*models.Target, error)
	GetAllTargets(ctx context.Context) ([]*models.Target, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}
