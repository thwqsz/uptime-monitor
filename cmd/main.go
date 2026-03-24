package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/thwqsz/uptime-monitor/internal/api"
	"github.com/thwqsz/uptime-monitor/internal/checker"
	"github.com/thwqsz/uptime-monitor/internal/config"
	"github.com/thwqsz/uptime-monitor/internal/db"
	"github.com/thwqsz/uptime-monitor/internal/logger"
	"github.com/thwqsz/uptime-monitor/internal/repository"
	"github.com/thwqsz/uptime-monitor/internal/service"
	"github.com/thwqsz/uptime-monitor/internal/worker"
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

	userRepo := repository.NewUserRepository(database)
	authService := service.NewAuthService(userRepo, conf.JWTSecret)
	authHandler := api.NewAuthHandler(authService)

	repoTarget := repository.NewPostgresTargetRepository(database)
	targetService := service.NewTargetService(repoTarget)
	targetHandler := api.NewTargetHandler(targetService)

	// Подключаю чекер
	check := checker.NewHTTPChecker(http.Client{})
	logRepo := repository.NewPostgresCheckLogRepository(database)
	checkService := service.NewCheckService(logRepo, repoTarget, check)
	checkHandler := api.NewCheckHandler(checkService)

	// Подключаю воркер
	loop := worker.NewLoop(checkService, repoTarget, log)
	go loop.Run(context.Background())

	r := chi.NewRouter()

	r.Post("/auth/register", authHandler.RegisterHandler)
	r.Post("/auth/login", authHandler.LoginHandler)
	r.Group(func(r chi.Router) {
		r.Use(service.JWTMiddleware(conf.JWTSecret))
		r.Get("/test", testHandler)
		r.Post("/targets", targetHandler.TargetCreateHandler)
		r.Get("/targets", targetHandler.TargetListHandler)
		r.Delete("/targets/{id}", targetHandler.DeleteTargetHandler)
		r.Get("/targets/{id}/check", checkHandler.CheckHandler)
	})

	port := fmt.Sprintf(":%d", conf.Port)
	log.Info("http server started", zap.String("port", port))
	err = http.ListenAndServe(port, r)
	log.Fatal("failed to starts server", zap.Error(err))
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := service.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Write([]byte(fmt.Sprintf("%v", userID)))
}
