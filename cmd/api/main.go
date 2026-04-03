package main

import (
	"github.com/thwqsz/uptime-monitor/internal/app"
	"github.com/thwqsz/uptime-monitor/internal/config"
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
	err = app.Run(&conf, log)
	if err != nil {
		log.Fatal("error during run", zap.Error(err))
	}
}
