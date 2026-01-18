package repository

import (
	"context"

	"digital-wallet/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) domain.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(ctx context.Context, wallet *domain.Wallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

func (r *walletRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Wallet, error) {
	var wallet domain.Wallet
	if err := r.db.WithContext(ctx).First(&wallet, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetByIDWithLock(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*domain.Wallet, error) {
	var wallet domain.Wallet
	conn := r.db
	if tx != nil {
		conn = tx
	}
	if err := conn.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&wallet, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) UpdateBalance(ctx context.Context, tx *gorm.DB, id uuid.UUID, newBalance int64) error {
	conn := r.db
	if tx != nil {
		conn = tx
	}
	// GORM updates fields with non-zero values by default, but 0 balance is valid.
	// So we specifically select the column or use map.
	return conn.WithContext(ctx).Model(&domain.Wallet{}).Where("id = ?", id).Update("balance", newBalance).Error
}

func (r *walletRepository) WithTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
