package repository

import (
	"context"

	"digital-wallet/internal/domain"
	"gorm.io/gorm"
)

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, tx *gorm.DB, transaction *domain.Transaction) error {
	conn := r.db
	if tx != nil {
		conn = tx
	}
	return conn.WithContext(ctx).Create(transaction).Error
}
