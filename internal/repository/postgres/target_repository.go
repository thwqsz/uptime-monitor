package postgres

import (
	"context"
	"database/sql"

	"github.com/thwqsz/uptime-monitor/internal/models"
)

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

func (r *PostgresTargetRepository) DeleteTarget(ctx context.Context, targetID, userID int64) (int64, error) {
	query := `DELETE FROM targets WHERE id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, query, targetID, userID)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (r *PostgresTargetRepository) GetTargetByID(ctx context.Context, targetID int64) (*models.Target, error) {
	query := `SELECT id, user_id, url, timeout, interval_time, created_at
				FROM targets 
				WHERE id = $1`
	target := &models.Target{}
	err := r.db.QueryRowContext(ctx, query, targetID).Scan(&target.ID, &target.UserID, &target.URL, &target.Timeout, &target.IntervalTime, &target.CreatedAt)
	if err != nil {
		return nil, err
	}
	return target, nil
}

func (r *PostgresTargetRepository) GetAllTargets(ctx context.Context) ([]*models.Target, error) {
	var targets []*models.Target
	query := `  SELECT id, user_id, url, timeout, interval_time, created_at
				FROM targets`

	targetRows, err := r.db.QueryContext(ctx, query)
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
