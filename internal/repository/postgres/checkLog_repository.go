package postgres

import (
	"context"
	"database/sql"

	"github.com/thwqsz/uptime-monitor/internal/models"
)

type PostgresCheckLogRepository struct {
	db *sql.DB
}

func NewPostgresCheckLogRepository(db *sql.DB) *PostgresCheckLogRepository {
	return &PostgresCheckLogRepository{db: db}
}

func (r *PostgresCheckLogRepository) CreateCheckLog(ctx context.Context, log *models.CheckLog) error {
	query := `INSERT INTO check_logs (status, status_code, error_msg, response_time_ms, target_id, checked_at) 
			  VALUES ($1, $2, $3, $4, $5, $6)
			  RETURNING check_logs.id
			  `
	err := r.db.QueryRowContext(ctx, query, log.Status, log.StatusCode, log.ErrorMsg, log.ResponseTimeMs, log.TargetID, log.CheckedAt).Scan(&log.ID)
	if err != nil {
		return err
	}
	return nil
}
