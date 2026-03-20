package repository

import (
	"context"
	"database/sql"

	"github.com/thwqsz/uptime-monitor/internal/models"
)

type TargetRepository interface {
	CreateTarget(ctx context.Context, target *models.Target) error
	GetTargetsByUserID(ctx context.Context, userID int64) ([]*models.Target, error)
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

func (r *PostgresTargetRepository) GetTargetsByUserID(ctx context.Context, userID int64) ([]*models.Target, error) {
	var targets []*models.Target
	query := `  SELECT id, user_id, url, timeout, interval_time, created_at
				FROM targets 
				WHERE user_id = $1`

	targetRows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer targetRows.Close()

	for targetRows.Next() {
		target := &models.Target{}
		err := targetRows.Scan(&target.ID, &target.UserID, &target.URL, &target.Timeout, &target.IntervalTime, &target.CreatedAt)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	if err = targetRows.Err(); err != nil {
		return nil, err
	}
	return targets, nil
}
