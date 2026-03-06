package frontend

import (
	"database/sql"
	"errors"
	"io/fs"
	"net/http"

	"github.com/ByChanderZap/exile-tracker/cmd/web/static"
	"github.com/ByChanderZap/exile-tracker/cmd/web/templates"
	"github.com/ByChanderZap/exile-tracker/models"
	"github.com/ByChanderZap/exile-tracker/pobparser"
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
	staticFS, _ := fs.Sub(static.Files, ".")
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	router.Get("/", h.handleHomePage)
	router.Get("/search", h.handleSearchAccounts)
	// show characters by accound and search characters within an account
	router.Get("/accounts/{accountId}/characters", h.handleCharactersByAccount)
	router.Get("/accounts/{accountId}/characters/search", h.handleCharactersSearchByAccount)
	router.Get("/snapshots/{characterId}", h.handleLoadedSnapshotsByCharacter)
	router.Get("/snapshots/{characterId}/detail/{snapshotId}", h.handleSnapshotDetail)
}

func (h *Handler) handleHomePage(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.repository.GetAllAccounts()
	if err != nil {
		h.log.Error().Err(err).Msg("Query to get accounts failed")
		http.Error(w, "Filed to load accounts", http.StatusInternalServerError)
		return
	}

	templates.Main(accounts, utils.StringValue).Render(r.Context(), w)
}

func (h *Handler) handleSearchAccounts(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("q")

	var accounts []models.Account
	var err error

	if searchTerm == "" {
		accounts, err = h.repository.GetAllAccounts()
	} else {
		accounts, err = h.repository.SearchAccounts(searchTerm)
	}

	if err != nil {
		h.log.Error().Err(err).Msg("Query to search accounts failed")
		http.Error(w, "Failed to search accounts", http.StatusInternalServerError)
		return
	}

	templates.AccountsTable(accounts, utils.StringValue).Render(r.Context(), w)
}

func (h *Handler) handleCharactersByAccount(w http.ResponseWriter, r *http.Request) {
	accId := chi.URLParam(r, "accountId")

	cs, err := h.repository.GetCharactersByAccountId(accId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "No characters to show", http.StatusNotFound)
			return
		}
		h.log.Error().Err(err).Msg("Query to get all characters by account id failed")
		http.Error(w, "Failed to load characters", http.StatusInternalServerError)
		return
	}
	templates.CharactersByAccountId(cs, accId, utils.StringValue).Render(r.Context(), w)
}

func (h *Handler) handleCharactersSearchByAccount(w http.ResponseWriter, r *http.Request) {
	accId := chi.URLParam(r, "accountId")
	searchTerm := r.URL.Query().Get("q")

	var characters []models.Character
	var err error

	if searchTerm == "" {
		characters, err = h.repository.GetCharactersByAccountId(accId)
	} else {
		characters, err = h.repository.SearchCharactersInAccount(repository.SearchCharactersInAccountParams{
			AccountId: accId,
			Query:     searchTerm,
		})
	}

	if err != nil {
		h.log.Error().Err(err).Msg("Query to search characters in account failed")
		http.Error(w, "Failed to search characters by account", http.StatusInternalServerError)
		return
	}

	templates.CharactersTable(characters, utils.StringValue).Render(r.Context(), w)
}

func (h *Handler) handleSnapshotDetail(w http.ResponseWriter, r *http.Request) {
	snapshotId := chi.URLParam(r, "snapshotId")

	snap, err := h.repository.GetSnapshotByID(snapshotId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Snapshot not found", http.StatusNotFound)
			return
		}
		h.log.Error().Err(err).Msg("Failed to get snapshot")
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if snap.PobCode == "" {
		http.Error(w, "This snapshot has no PoB data to display", http.StatusNotFound)
		return
	}

	pob, err := pobparser.Decode(snap.PobCode)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to decode PoB data")
		http.Error(w, "Failed to decode build data", http.StatusInternalServerError)
		return
	}

	summary := pob.Summarize()
	createdAt := snap.CreatedAt.Format("Jan 2, 2006 15:04")

	templates.SnapshotDetail(summary, pob, snap.ExportString, snap.ID, createdAt).Render(r.Context(), w)
}

func (h *Handler) handleLoadedSnapshotsByCharacter(w http.ResponseWriter, r *http.Request) {
	cId := chi.URLParam(r, "characterId")

	snaps, err := h.repository.GetSnapshotsByCharacterWithExtras(repository.GetSnapshotsByCharacterWithExtras{
		CharacterId: cId,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "No snapshots found", http.StatusNotFound)
			return
		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	templates.SnapshotsPage(snaps, utils.StringValue).Render(r.Context(), w)
}
