package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"digital-wallet/internal/handler"
	"digital-wallet/internal/repository"
	"digital-wallet/internal/service"
	"digital-wallet/internal/worker"
	"digital-wallet/pkg/postgres"
	"digital-wallet/pkg/rabbitmq"
	"digital-wallet/pkg/redis"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults/environment variables")
	} else {
		log.Println(".env file loaded successfully")
	}

	// Configuration (Defaults)
	pgDSN := os.Getenv("DATABASE_URL")
	if pgDSN == "" {
		pgDSN = "host=localhost user=postgres password=postgres dbname=wallet_db port=5432 sslmode=disable"
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	// Infrastructure

	db, err := postgres.NewConnection(pgDSN)
	if err != nil {
		log.Fatalf("Postgres init failed: %v", err)
	}

	rdb, err := redis.NewClient(redisAddr, "", 0)
	if err != nil {
		log.Fatalf("Redis init failed: %v", err)
	}

	mq, err := rabbitmq.NewConnection(rabbitURL)
	if err != nil {
		log.Fatalf("RabbitMQ init failed: %v", err)
	}
	defer mq.Close()

	// Repositories
	walletRepo := repository.NewWalletRepository(db)
	transRepo := repository.NewTransactionRepository(db)
	cacheRepo := repository.NewCacheRepository(rdb)
	eventProducer := repository.NewEventProducer(mq)

	// Service
	svc := service.NewWalletService(walletRepo, transRepo, cacheRepo, eventProducer)

	// Worker
	w := worker.NewWorker(mq)
	w.Start()

	// HTTP Handler & Server
	h := handler.NewHandler(svc)
	mux := handler.NewRouter(h)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start Server
	go func() {
		log.Println("Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
