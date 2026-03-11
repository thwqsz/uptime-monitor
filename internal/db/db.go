package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/thwqsz/uptime-monitor/internal/config"
)

func InitDB(cfg config.DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("port=%v host=%v dbname=%v user=%v password=%v", cfg.Port, cfg.Host, cfg.Name, cfg.User, cfg.Password)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
