package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/thwqsz/uptime-monitor/internal/api"
	"github.com/thwqsz/uptime-monitor/internal/auth"
	"github.com/thwqsz/uptime-monitor/internal/checker"
	"github.com/thwqsz/uptime-monitor/internal/config"
	"github.com/thwqsz/uptime-monitor/internal/db"
	"github.com/thwqsz/uptime-monitor/internal/repository/postgres"
	"github.com/thwqsz/uptime-monitor/internal/service"
	"github.com/thwqsz/uptime-monitor/internal/worker"
	"go.uber.org/zap"
)

func Run(conf *config.Config, log *zap.Logger) error {
	log.Info("starting uptime_monitor")

	database, err := db.InitDB(conf.DB)
	if err != nil {
		return err
	}
	defer database.Close()
	log.Info("database is connected")

	userRepo := postgres.NewUserRepository(database)
	authService := service.NewAuthService(userRepo, conf.JWTSecret)
	authHandler := api.NewAuthHandler(authService, log)

	repoTarget := postgres.NewPostgresTargetRepository(database)
	check := checker.NewHTTPChecker(&http.Client{})
	rootCtx, cancel := context.WithCancel(context.Background())
	logRepo := postgres.NewPostgresCheckLogRepository(database)
	checkService := service.NewCheckService(logRepo, repoTarget, check)
	loop := worker.NewLoop(repoTarget, checkService, conf.WorkerCount, log, rootCtx)
	go loop.Run()
	targetService := service.NewTargetService(repoTarget, loop)
	targetHandler := api.NewTargetHandler(targetService)
	checkHandler := api.NewCheckHandler(checkService)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"POST", "OPTIONS", "DELETE", "GET"},
		AllowedHeaders:   []string{"Content-Type", "Accept", "Authorization"},
		MaxAge:           300,
		AllowCredentials: false,
	}))

	r.Post("/auth/register", authHandler.RegisterHandler)
	r.Post("/auth/login", authHandler.LoginHandler)
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(conf.JWTSecret))
		r.Post("/targets", targetHandler.TargetCreateHandler)
		r.Get("/targets", targetHandler.TargetListHandler)
		r.Delete("/targets/{id}", targetHandler.DeleteTargetHandler)
		r.Get("/targets/{id}/check", checkHandler.CheckHandler)
	})

	port := fmt.Sprintf(":%d", conf.Port)
	log.Info("http server started", zap.String("port", port))
	server := &http.Server{
		Addr:    port,
		Handler: r,
	}

	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("error during ListenAndServe", zap.Error(err))
		}
	}()

	chOS := make(chan os.Signal, 1)
	signal.Notify(chOS, syscall.SIGINT, syscall.SIGTERM)
	<-chOS
	log.Info("got stop signal")
	cancel()
	ctxForShutDown, cancelCtxForShutDown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtxForShutDown()
	if err = server.Shutdown(ctxForShutDown); err != nil {
		log.Error("server was closed by force, timeout is out", zap.Error(err))
	}
	return nil
}
