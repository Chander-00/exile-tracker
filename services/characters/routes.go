package characters

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	apimodels "github.com/ByChanderZap/exile-tracker/models/api"
	models "github.com/ByChanderZap/exile-tracker/models/api"
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
	router.Get("/characters", h.handleGetAllCharacters)
	router.Get("/characters/{id}", h.handleGetCharacterByID)
	router.Get("/characters/account/{accountId}", h.handleGetCharactersByAccount)
	router.Post("/characters", h.handleCreateCharacter)
	router.Put("/characters/{id}", h.handleUpdateCharacter)
	router.Patch("/characters/{id}/kill", h.handleKillCharacter)
	router.Post("/characters/to-fetch", h.handleAddCharactersToFetch)
	router.Get("/characters/to-fetch", h.handleGetAllCharactersToFetch)
	// router.Delete("/characters/{id}", h.handleDeleteCharacter)
}

func (h *Handler) handleGetAllCharacters(w http.ResponseWriter, r *http.Request) {
	characters, err := h.repository.GetAllCharacters()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, characters)
}

func (h *Handler) handleGetCharacterByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	character, err := h.repository.GetCharacterByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"Message": "Character not found",
			})
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, character)
}

func (h *Handler) handleGetCharactersByAccount(w http.ResponseWriter, r *http.Request) {
	accountId := chi.URLParam(r, "accountId")
	characters, err := h.repository.GetCharactersByAccountId(accountId)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, characters)
}

func (h *Handler) handleCreateCharacter(w http.ResponseWriter, r *http.Request) {
	var payload apimodels.CreateCharacterInput
	if err := utils.ParseJson(r, &payload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err)
		return
	}
	if err := utils.ValidatePayload(&payload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err)
		return
	}
	err := h.repository.CreateCharacter(payload.AccountId, payload.CharacterName, payload.CurrentLeague)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Character created",
	})
}

func (h *Handler) handleUpdateCharacter(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var payload apimodels.UpdateCharacterInput
	if err := utils.ParseJson(r, &payload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err)
		return
	}
	if err := utils.ValidatePayload(&payload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	c, err := h.repository.GetCharacterByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithError(w, http.StatusNotFound, fmt.Errorf("the character that you are trying to update does not exists"))
			return
		}
		utils.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	err = h.repository.UpdateCharacter(repository.UpdateCharacterParams{
		ID:            c.ID,
		CharacterName: payload.CharacterName,
		CurrentLeague: payload.CurrentLeague,
		Died:          c.Died,
		UpdatedAt:     time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Character updated",
	})
}

func (h *Handler) handleKillCharacter(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.repository.KillCharacter(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Character killed",
	})
}

func (h *Handler) handleAddCharactersToFetch(w http.ResponseWriter, r *http.Request) {
	var payload models.AddCharacterToFetchInput
	if err := utils.ParseJson(r, &payload); err != nil {
		h.log.Error().Err(err).Msg("Error decoding add characters to fetch input")
		utils.RespondWithError(w, http.StatusBadRequest, err)
		return
	}
	if err := utils.ValidatePayload(&payload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	err := h.repository.AddCharacterToFetch(repository.AddCharactersToFetchParams{
		CharacterId: payload.CharacterId,
	})
	if err != nil {
		h.log.Error().Err(err).Msg("Error while trying to add character to fetch")
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Errorf("couldnt add character to fetch"))
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Message": fmt.Sprintf("Character %s is now tracked", payload.CharacterId),
	})
}

func (h *Handler) handleGetAllCharactersToFetch(w http.ResponseWriter, r *http.Request) {
	ctf, err := h.repository.GetCharactersToFetch()
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err)
		return
	}
	utils.WriteJSON(w, http.StatusAccepted, ctf)
}

// func (h *Handler) handleDeleteCharacter(w http.ResponseWriter, r *http.Request) {
// 	id := chi.URLParam(r, "id")
// 	err := h.repository.DeleteCharacter(id)
// 	if err != nil {
// 		utils.RespondWithError(w, http.StatusInternalServerError, err)
// 		return
// 	}

// 	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
// 		"message": "Character deleted",
// 	})
// }
