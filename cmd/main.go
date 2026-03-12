package main

import (
	"github.com/thwqsz/uptime-monitor/internal/config"
	"github.com/thwqsz/uptime-monitor/internal/db"
	"github.com/thwqsz/uptime-monitor/internal/logger"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.New()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	conf, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}
	log.Info("starting uptime_monitor")

	database, err := db.InitDB(conf.DB)
	if err != nil {
		log.Fatal("failed to init db", zap.Error(err))
	}
	defer database.Close()
	log.Info("database is connected")

}
