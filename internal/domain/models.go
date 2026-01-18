package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Wallet represents a user's digital wallet.
type Wallet struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Balance   int64     `gorm:"not null;default:0;check:balance >= 0" json:"balance"` // Stored in cents, must be >= 0
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Transaction represents a money transfer history.
type Transaction struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SenderID   *uuid.UUID `gorm:"type:uuid" json:"sender_id,omitempty"`   // Nullable for deposits
	ReceiverID *uuid.UUID `gorm:"type:uuid" json:"receiver_id,omitempty"` // Nullable for withdrawals
	Amount     int64     `gorm:"not null" json:"amount"`
	Type       string    `gorm:"not null" json:"type"` // "TRANSFER", "DEPOSIT", "WITHDRAWAL"
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TransferEvent is the payload sent to RabbitMQ.
type TransferEvent struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	SenderID      uuid.UUID `json:"sender_id"`
	ReceiverID    uuid.UUID `json:"receiver_id"`
	Amount        int64     `json:"amount"`
}

// Repository Interfaces

type WalletRepository interface {
	Create(ctx context.Context, wallet *Wallet) error
	GetByID(ctx context.Context, id uuid.UUID) (*Wallet, error)
	GetByIDWithLock(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*Wallet, error)
	UpdateBalance(ctx context.Context, tx *gorm.DB, id uuid.UUID, newBalance int64) error
	WithTx(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *gorm.DB, transaction *Transaction) error
}

type CacheRepository interface {
	GetWallet(ctx context.Context, walletID uuid.UUID) (*Wallet, error)
	SetWallet(ctx context.Context, wallet *Wallet) error
	InvalidateWallet(ctx context.Context, walletID uuid.UUID) error
}

type EventProducer interface {
	PublishTransferEvent(ctx context.Context, event TransferEvent) error
}
