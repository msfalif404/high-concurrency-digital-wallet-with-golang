package domain

import "errors"

var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrInvalidAmount       = errors.New("amount must be positive")
	ErrSelfTransfer        = errors.New("cannot transfer money to yourself")
	ErrInternalServerError = errors.New("internal server error")
)
