package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"digital-wallet/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WalletService struct {
	walletRepo      domain.WalletRepository
	transRepo       domain.TransactionRepository
	cacheRepo       domain.CacheRepository
	eventProducer   domain.EventProducer
}

func NewWalletService(
	wRepo domain.WalletRepository,
	tRepo domain.TransactionRepository,
	cRepo domain.CacheRepository,
	evt domain.EventProducer,
) *WalletService {
	return &WalletService{
		walletRepo:    wRepo,
		transRepo:     tRepo,
		cacheRepo:     cRepo,
		eventProducer: evt,
	}
}

func (s *WalletService) CreateWallet(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	wallet := &domain.Wallet{
		UserID:  userID,
		Balance: 0,
	}
	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (*domain.Wallet, error) {
	// 1. Check Cache
	cached, err := s.cacheRepo.GetWallet(ctx, walletID)
	if err == nil && cached != nil {
		return cached, nil
	}
	
	// 2. Fetch from DB
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// 3. Set Cache
	if err := s.cacheRepo.SetWallet(ctx, wallet); err != nil {
		log.Printf("failed to set cache for wallet %s: %v", walletID, err)
	}

	return wallet, nil
}

func (s *WalletService) TransferMoney(ctx context.Context, senderID, receiverID uuid.UUID, amount int64) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if senderID == receiverID {
		return nil, errors.New("cannot transfer to self")
	}

	var transaction *domain.Transaction

	// Transactional Block
	err := s.walletRepo.WithTx(ctx, func(tx *gorm.DB) error {
		// Deadlock Prevention: Sort Locks
		firstID, secondID := senderID, receiverID
		if firstID.String() > secondID.String() {
			firstID, secondID = receiverID, senderID
		}

		// Lock First Wallet
		w1, err := s.walletRepo.GetByIDWithLock(ctx, tx, firstID)
		if err != nil {
			return fmt.Errorf("failed to lock wallet %s: %w", firstID, err)
		}

		// Lock Second Wallet
		w2, err := s.walletRepo.GetByIDWithLock(ctx, tx, secondID)
		if err != nil {
			return fmt.Errorf("failed to lock wallet %s: %w", secondID, err)
		}

		// Identify Sender and Receiver
		var sender, receiver *domain.Wallet
		if w1.ID == senderID {
			sender = w1
			receiver = w2
		} else {
			sender = w2
			receiver = w1
		}

		// Logic Check
		if sender.Balance < amount {
			return errors.New("insufficient funds")
		}

		// Update Balances
		newSenderBal := sender.Balance - amount
		newReceiverBal := receiver.Balance + amount

		if err := s.walletRepo.UpdateBalance(ctx, tx, sender.ID, newSenderBal); err != nil {
			return err
		}
		if err := s.walletRepo.UpdateBalance(ctx, tx, receiver.ID, newReceiverBal); err != nil {
			return err
		}

		// Create Transaction Log
		transaction = &domain.Transaction{
			SenderID:   &sender.ID,
			ReceiverID: &receiver.ID,
			Amount:     amount,
			Type:       "TRANSFER",
		}
		if err := s.transRepo.Create(ctx, tx, transaction); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Post-Transaction Actions (Best Effort)
	
	// Invalidate Cache
	_ = s.cacheRepo.InvalidateWallet(ctx, senderID)
	_ = s.cacheRepo.InvalidateWallet(ctx, receiverID)

	// Publish Event
	event := domain.TransferEvent{
		TransactionID: transaction.ID,
		SenderID:      senderID,
		ReceiverID:    receiverID,
		Amount:        amount,
	}
	if err := s.eventProducer.PublishTransferEvent(ctx, event); err != nil {
		log.Printf("failed to publish transfer event: %v", err)
	}

	return transaction, nil
}
