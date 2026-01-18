package handler

import (
	"encoding/json"
	"net/http"

	"digital-wallet/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler struct {
	svc *service.WalletService
	validator *validator.Validate
}

func NewHandler(svc *service.WalletService) *Handler {
	return &Handler{
		svc: svc,
		validator: validator.New(),
	}
}

// Responses
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func respondError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Code: code, Message: msg})
}

func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// Request Models
type CreateWalletReq struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

type TransferReq struct {
	SenderID   string `json:"sender_id" validate:"required,uuid"`
	ReceiverID string `json:"receiver_id" validate:"required,uuid"`
	Amount     int64  `json:"amount" validate:"required,gt=0"`
}

// Handlers

func (h *Handler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var req CreateWalletReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	uid, _ := uuid.Parse(req.UserID)
	wallet, err := h.svc.CreateWallet(r.Context(), uid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, wallet)
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "Missing wallet ID")
		return
	}

	walletID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid wallet ID format")
		return
	}

	wallet, err := h.svc.GetBalance(r.Context(), walletID)
	if err != nil {
		if err.Error() == "record not found" { // Simplified error check
			respondError(w, http.StatusNotFound, "Wallet not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, wallet)
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req TransferReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	senderID, _ := uuid.Parse(req.SenderID)
	receiverID, _ := uuid.Parse(req.ReceiverID)

	tx, err := h.svc.TransferMoney(r.Context(), senderID, receiverID, req.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, tx)
}
