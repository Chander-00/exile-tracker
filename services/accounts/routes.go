package accounts

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	apimodels "github.com/ByChanderZap/exile-tracker/models/api"
	"github.com/ByChanderZap/exile-tracker/repository"
	"github.com/ByChanderZap/exile-tracker/utils"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type Handler struct {
	repository *repository.Repository
	log        zerolog.Logger
}

func NewHandler(db *repository.Repository, logger zerolog.Logger) *Handler {
	return &Handler{
		repository: db,
		log:        logger,
	}
}

func (h *Handler) RegisterRoutes(router *chi.Mux) {
	router.Get("/accounts", h.handleGetAllAccounts)
	router.Get("/accounts/{id}", h.handleGetAccountByID)
	router.Post("/accounts", h.handleCreateAccount)
	router.Put("/accounts/{id}", h.handleUpdateAccount)
}

func (h *Handler) handleGetAllAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.repository.GetAllAccounts()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, accounts)
}

func (h *Handler) handleGetAccountByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	account, err := h.repository.GetAccountByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithError(w, http.StatusNotFound, fmt.Errorf("account not found"))
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, account)
}

func (h *Handler) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var payload apimodels.CreateAccountInput
	if err := utils.ParseJson(r, &payload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err)
		return
	}
	if err := utils.ValidatePayload(&payload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	var player string
	if payload.Player != nil {
		player = *payload.Player
	}
	err := h.repository.CreateAccount(payload.AccountName, player)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Account created",
	})
}

func (h *Handler) handleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var payload apimodels.UpdateAccountInput
	if err := utils.ParseJson(r, &payload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err)
		return
	}
	if err := utils.ValidatePayload(&payload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	acc, err := h.repository.GetAccountByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithError(w, http.StatusNotFound, fmt.Errorf("the character that you are trying to update does not exists"))
			return
		}
		utils.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	var player string
	if payload.Player != nil {
		player = *payload.Player
	}

	err = h.repository.UpdateAccount(repository.UpdateAccountParams{
		ID:          acc.ID,
		AccountName: payload.AccountName,
		Player:      player,
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Account updated",
	})
}
