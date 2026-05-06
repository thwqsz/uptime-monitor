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
	"github.com/segmentio/kafka-go"
	"github.com/thwqsz/uptime-monitor/internal/api"
	"github.com/thwqsz/uptime-monitor/internal/auth"
	"github.com/thwqsz/uptime-monitor/internal/broker"
	"github.com/thwqsz/uptime-monitor/internal/cache"
	"github.com/thwqsz/uptime-monitor/internal/config"
	"github.com/thwqsz/uptime-monitor/internal/db"
	"github.com/thwqsz/uptime-monitor/internal/grpc/checkerpb"
	"github.com/thwqsz/uptime-monitor/internal/repository/postgres"
	"github.com/thwqsz/uptime-monitor/internal/service"
	"github.com/thwqsz/uptime-monitor/internal/worker"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Run(conf *config.Config, log *zap.Logger) error {
	log.Info("starting uptime_monitor")

	// подключение бд
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
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// инициализация кэша
	clientForCache, err := cache.InitClientRedis(rootCtx, conf.RedisAddr)
	if err != nil {
		log.Error("cache is not connected", zap.Error(err))
		return err
	}
	defer clientForCache.Close()
	cacheManager := cache.NewCache(clientForCache)

	logRepo := postgres.NewPostgresCheckLogRepository(database)
	checkResPro := service.NewCheckServiceForKafka(logRepo, cacheManager, log)

	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:     []string{"localhost:9092"},
			Topic:       "check_results",
			StartOffset: kafka.FirstOffset,
		})
	defer reader.Close()
	consumerGetResultFromChecker := broker.NewConsumer(log, reader, checkResPro)
	go consumerGetResultFromChecker.Run(rootCtx)

	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "check_tasks",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()
	producerSendTaskToChecker := broker.NewProducer(writer, log)

	loop := worker.NewLoop(repoTarget, producerSendTaskToChecker, conf.WorkerCount, log, rootCtx)
	go loop.Run()
	targetService := service.NewTargetService(repoTarget, loop)
	targetHandler := api.NewTargetHandler(targetService)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"POST", "OPTIONS", "DELETE", "GET"},
		AllowedHeaders:   []string{"Content-Type", "Accept", "Authorization"},
		MaxAge:           300,
		AllowCredentials: false,
	}))

	conn, err := grpc.NewClient("localhost:8088", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	checkClient := checkerpb.NewCheckerServiceClient(conn)
	manualCheckService := service.NewManualService(repoTarget, logRepo, checkClient)
	manualCheckHandler := api.NewCheckHandler(manualCheckService)

	r.Post("/auth/register", authHandler.RegisterHandler)
	r.Post("/auth/login", authHandler.LoginHandler)
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(conf.JWTSecret))
		r.Post("/targets", targetHandler.TargetCreateHandler)
		r.Get("/targets", targetHandler.TargetListHandler)
		r.Delete("/targets/{id}", targetHandler.DeleteTargetHandler)
		r.Get("/targets/{id}/check", manualCheckHandler.CheckHandler)
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
