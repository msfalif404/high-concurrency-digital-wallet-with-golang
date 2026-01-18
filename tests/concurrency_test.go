package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"digital-wallet/internal/domain"
	"digital-wallet/internal/repository"
	"digital-wallet/internal/service"
	"digital-wallet/pkg/postgres"
	"digital-wallet/pkg/rabbitmq"
	"digital-wallet/pkg/redis"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// NOTE: This test requires infrastructure (Postgres, Redis, RabbitMQ) to be running.
func TestConcurrentTransfers(t *testing.T) {
	_ = godotenv.Load("../.env") // Load .env file if present
	// Configuration (Test Defaults)
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

	// Setup Infrastructure
	db, err := postgres.NewConnection(pgDSN)
	if err != nil {
		t.Skipf("Skipping integration test (DB not available): %v", err)
	}
	rdb, err := redis.NewClient(redisAddr, "", 0)
	if err != nil {
		t.Skipf("Skipping integration test (Redis not available): %v", err)
	}
	mq, err := rabbitmq.NewConnection(rabbitURL)
	if err != nil {
		t.Skipf("Skipping integration test (RabbitMQ not available): %v", err)
	}
	defer mq.Close()

	// Setup Components
	walletRepo := repository.NewWalletRepository(db)
	transRepo := repository.NewTransactionRepository(db)
	cacheRepo := repository.NewCacheRepository(rdb)
	eventProducer := repository.NewEventProducer(mq)
	svc := service.NewWalletService(walletRepo, transRepo, cacheRepo, eventProducer)

	ctx := context.Background()

	// Create Test Wallets
	senderID := uuid.New()
	receiverID := uuid.New()

	// Mock User IDs
	user1 := uuid.New()
	user2 := uuid.New()

	// Direct DB Insert to set initial balance
	walletA := &domain.Wallet{
		ID:      senderID,
		UserID:  user1,
		Balance: 1000, // 10.00
	}
	walletB := &domain.Wallet{
		ID:      receiverID,
		UserID:  user2,
		Balance: 0,
	}

	// Cleanup old data if any (optional, purely for local dev re-runs with same IDs)
	db.Exec("DELETE FROM transactions WHERE sender_id = ? OR receiver_id = ?", senderID, receiverID)
	db.Delete(&domain.Wallet{}, "id = ?", senderID)
	db.Delete(&domain.Wallet{}, "id = ?", receiverID)

	if err := walletRepo.Create(ctx, walletA); err != nil {
		t.Fatalf("Failed to create wallet A: %v", err)
	}
	if err := walletRepo.Create(ctx, walletB); err != nil {
		t.Fatalf("Failed to create wallet B: %v", err)
	}

	// Run Concurrency Test
	// Scenario: 50 concurrent transfers of 1 cent from A to B.
	concurrentCount := 50
	amount := int64(1)

	var wg sync.WaitGroup
	wg.Add(concurrentCount)

	errChan := make(chan error, concurrentCount)

	log.Println("Starting concurrent transfers...")
	start := time.Now()

	for i := 0; i < concurrentCount; i++ {
		go func() {
			defer wg.Done()
			_, err := svc.TransferMoney(context.Background(), senderID, receiverID, amount)
			if err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)
	log.Printf("Finished transfers in %v", duration)
	close(errChan)

	// Check for errors
	for err := range errChan {
		t.Errorf("Error during transfer: %v", err)
	}

	// Verification
	finalA, err := walletRepo.GetByID(ctx, senderID)
	if err != nil {
		t.Fatalf("Failed to get final wallet A: %v", err)
	}
	finalB, err := walletRepo.GetByID(ctx, receiverID)
	if err != nil {
		t.Fatalf("Failed to get final wallet B: %v", err)
	}

	expectedA := int64(1000 - concurrentCount*1)
	expectedB := int64(0 + concurrentCount*1)

	if finalA.Balance != expectedA {
		t.Errorf("Expected Wallet A balance %d, got %d", expectedA, finalA.Balance)
	}
	if finalB.Balance != expectedB {
		t.Errorf("Expected Wallet B balance %d, got %d", expectedB, finalB.Balance)
	}

	fmt.Printf("Test Passed! Final Balances: A=%d, B=%d\n", finalA.Balance, finalB.Balance)
}
