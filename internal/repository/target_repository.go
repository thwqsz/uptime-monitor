package repository

import (
	"context"
	"database/sql"

	"github.com/thwqsz/uptime-monitor/internal/models"
)

type TargetRepository interface {
	CreateTarget(ctx context.Context, target *models.Target) error
}
type PostgresTargetRepository struct {
	db *sql.DB
}

func NewPostgresTargetRepository(db *sql.DB) *PostgresTargetRepository {
	return &PostgresTargetRepository{db: db}
}

func (r *PostgresTargetRepository) CreateTarget(ctx context.Context, target *models.Target) error {
	query := `INSERT INTO targets (user_id, url, interval_time, timeout)
	VALUES ($1, $2, $3, $4)
	RETURNING targets.id, targets.created_at
	`
	err := r.db.QueryRowContext(ctx, query, target.UserID, target.URL, target.IntervalTime, target.Timeout).Scan(&target.ID, &target.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}
