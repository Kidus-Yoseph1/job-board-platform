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

	db "github.com/kidus-yoseph1/job-board-platform/job-service/db/generated"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/cache"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/config"
	jobGrpc "github.com/kidus-yoseph1/job-board-platform/job-service/internal/grpc"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/handler"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/middleware"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/repository"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/routes"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/service"
	"github.com/kidus-yoseph1/job-board-platform/job-service/pkg/logger"
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
		log.Fatalw("failed to open database connection", "error", err)
	}
	defer dbConn.Close()

	// Configure database connection pool settings
	dbConn.SetConnMaxLifetime(5 * time.Minute)
	dbConn.SetMaxOpenConns(25)
	dbConn.SetMaxIdleConns(10)

	// Startup database ping with 5s timeout
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err := dbConn.PingContext(pingCtx); err != nil {
		log.Fatalw("failed to ping database", "error", err, "db", cfg.DBName)
	}
	log.Infow("successfully connected to database", "db", cfg.DBName, "host", cfg.DBHost, "port", cfg.DBPort)

	// Initialize Redis Cache
	redisCache, err := cache.NewRedisCache("redis://" + cfg.RedisAddr)
	if err != nil {
		log.Fatalw("failed to connect to redis cache", "error", err, "addr", cfg.RedisAddr)
	}
	log.Infow("successfully connected to redis cache", "addr", cfg.RedisAddr)

	// Initialize Repository & Queries Layer
	queries := db.New(dbConn)
	jobRepo := repository.NewJobRepo(queries)
	applicationRepo := repository.NewApplicationRepo(queries)
	userRepo := repository.NewUserRepo(queries)
	companyRepo := repository.NewCompanyRepo(queries)

	// Initialize Cache & Services Layer
	jobCache := cache.NewJobCache(redisCache)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	jobService := service.NewJobService(jobRepo, companyRepo, jobCache)
	applicationService := service.NewApplicationService(applicationRepo, jobRepo, companyRepo, redisCache)

	// Initialize Handlers (injecting logger dependency)
	authHandler := handler.NewAuthHandler(authService, log)
	jobHandler := handler.NewJobHandler(jobService, log)
	applicationHandler := handler.NewApplicationHandler(applicationService, log)

	// Initialize Gin engine & setup routes
	r := gin.Default()
	r.Use(middleware.CORS())

	routes.Setup(r, applicationHandler, jobHandler, authHandler, cfg.JWTSecret)

	// Start gRPC Server
	grpcServer, err := jobGrpc.StartServer(":"+cfg.GRPCPort, queries, log)
	if err != nil {
		log.Fatalw("gRPC server failed to start", "error", err)
	}

	// Configure and run HTTP Server with Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Infow("starting job-service HTTP server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalw("HTTP server failed to start", "error", err)
		}
	}()

	// Wait for OS signal/cancellation
	<-ctx.Done()
	log.Info("shutting down job-service gracefully...")

	// Gracefully shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorw("HTTP server forced to shutdown", "error", err)
	}

	// Gracefully stop gRPC server
	grpcServer.GracefulStop()
	log.Info("job-service exited cleanly")
}
