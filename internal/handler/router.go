package handler

import (
	"net/http"
)

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /wallets", h.CreateWallet)
	mux.HandleFunc("GET /wallets/{id}", h.GetBalance)
	mux.HandleFunc("POST /transfers", h.Transfer)

	return mux
}
