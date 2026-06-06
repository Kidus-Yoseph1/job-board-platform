package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	db "github.com/kidus-yoseph1/job-board-platform/notification-service/db/generated"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/config"
	notificationGrpc "github.com/kidus-yoseph1/job-board-platform/notification-service/internal/grpc"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/handler"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/middleware"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/repository"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/service"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/subscriber"
	ws "github.com/kidus-yoseph1/job-board-platform/notification-service/internal/websocket"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/worker"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/pkg/logger"
)

func main() {
	// Load application config
	cfg := config.Load()

	logger.Init(os.Getenv("APP_ENV"))
	log := logger.Get()
	defer log.Sync()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Initialize Database Connection
	dbConn, err := sql.Open("postgres", cfg.GetDBConnString())
	if err != nil {
		log.Fatalw("failed to open database", "error", err)
	}
	defer dbConn.Close()

	// Connection pool settings
	dbConn.SetConnMaxLifetime(5 * time.Minute)
	dbConn.SetMaxOpenConns(25)
	dbConn.SetMaxIdleConns(10)

	// Startup ping with 5s timeout
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err := dbConn.PingContext(pingCtx); err != nil {
		log.Fatalw("failed to ping database", "error", err, "db", cfg.DBName)
	}
	log.Infow("connected to database", "db", cfg.DBName, "host", cfg.DBHost, "port", cfg.DBPort)

	// Initialize Redis Client
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	defer rdb.Close()

	// Startup ping with 5s timeout
	redisCtx, redisCancel := context.WithTimeout(ctx, 5*time.Second)
	defer redisCancel()
	if err := rdb.Ping(redisCtx).Err(); err != nil {
		log.Fatalw("failed to connect to redis", "error", err, "addr", cfg.RedisAddr)
	}
	log.Infow("connected to redis", "addr", cfg.RedisAddr)

	// Initialize WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()
	log.Info("websocket hub event loop started")

	// Initialize Database Queries & Repositories
	queries := db.New(dbConn)
	notificationRepo := repository.NewNotificationRepo(queries)
	emailJobRepo := repository.NewEmailJobRepo(queries)

	// Initialize gRPC client connection to job-service
	grpcClient, err := notificationGrpc.NewJobBoardClient(cfg.JobServiceGRPCAddr)
	if err != nil {
		log.Fatalw("failed to connect to job-service gRPC server", "error", err, "addr", cfg.JobServiceGRPCAddr)
	}
	defer grpcClient.Close()

	// Initialize Redis Subscriber
	sub := subscriber.NewRedisSubscriber(rdb, hub, notificationRepo, emailJobRepo, grpcClient)
	go sub.Start(ctx)

	// Initialize and Start Email Worker
	emailWorker := worker.NewEmailWorker(
		emailJobRepo,
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPFromEmail,
	)
	go emailWorker.Start(ctx)

	// Initialize Services & Handlers
	notificationService := service.NewNotificationService(notificationRepo, log)
	notificationHandler := handler.NewNotificationHandler(notificationService, log)
	websocketHandler := handler.NewWebsocketHandler(hub, log)

	// Initialize Gin engine and routes
	r := gin.Default()

	// Register global middleware
	r.Use(middleware.CORSMiddleware())

	// Protected endpoints (Guarded by JWT Auth)
	authGroup := r.Group("")
	authGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Real-time WebSocket connection endpoint
		authGroup.GET("/ws", websocketHandler.HandleWS)

		// Notification history APIs
		authGroup.GET("/notifications", notificationHandler.ListNotificationsByUserHandler)
		authGroup.PUT("/notifications/read", notificationHandler.MarkAllNotificationsReadHandler)
		authGroup.PUT("/notifications/read/:id", notificationHandler.MarkNotificationReadHandler)
	}

	// Configure and run HTTP Server with Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Start the server in a separate background goroutine
	go func() {
		log.Infow("starting notification-service server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalw("server failed to start", "error", err)
		}
	}()

	// Wait for the context (interrupted by OS signals via signal.NotifyContext) to finish
	<-ctx.Done()
	log.Info("shutting down notification-service server gracefully...")

	// Create a context with 5s timeout to let remaining requests finish
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorw("server forced to shutdown", "error", err)
	}

	log.Info("notification-service exited cleanly")
}
