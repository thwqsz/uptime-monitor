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

//func main() {
//	log, err := logger.New()
//	if err != nil {
//		panic(err)
//	}
//	defer log.Sync()
//	writer := &kafka.Writer{
//		Addr:     kafka.TCP("localhost:9092"),
//		Topic:    "check_tasks",
//		Balancer: &kafka.LeastBytes{},
//	}
//	defer writer.Close()
//	target := &models.Target{
//		ID:      1,
//		URL:     "https://google.com",
//		Timeout: 2,
//	}
//	prod := kafkaprod.NewProducer(writer, log)
//	if _, err := prod.SendTask(context.Background(), target); err != nil {
//		panic(err)
//	}
//}
