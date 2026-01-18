package postgres

import (
	"fmt"
	"log"
	
	"digital-wallet/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate schema
	err = db.AutoMigrate(&domain.Wallet{}, &domain.Transaction{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Connected to PostgreSQL")
	return db, nil
}
