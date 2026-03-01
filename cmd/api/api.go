package api

import (
	"context"
	"net/http"

	"github.com/ByChanderZap/exile-tracker/repository"
	"github.com/ByChanderZap/exile-tracker/services/accounts"
	"github.com/ByChanderZap/exile-tracker/services/characters"
	"github.com/ByChanderZap/exile-tracker/services/frontend"
	"github.com/ByChanderZap/exile-tracker/services/pobsnapshots"
	"github.com/ByChanderZap/exile-tracker/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type APIServer struct {
	addr       string
	server     *http.Server
	repository *repository.Repository
	log        zerolog.Logger
}

func NewAPIServer(addr string, db *repository.Repository) *APIServer {
	utils.BaseLogger.Info().Msg(addr)
	return &APIServer{
		addr:       addr,
		repository: db,
		log:        utils.ChildLogger("api"),
	}
}

func (s *APIServer) Start() error {

	router := chi.NewRouter()

	// Global middleware — applies to all routes (API + frontend)
	router.Use(middleware.Recoverer)
	router.Use(utils.ZerologMiddleware(s.log))

	v1Router := chi.NewRouter()
	frontendRouter := chi.NewRouter()

	// character endpoints
	cHandler := characters.NewHandler(s.repository, s.log)
	cHandler.RegisterRoutes(v1Router)

	// accounts endpoints
	aHandler := accounts.NewHandler(s.repository, s.log)
	aHandler.RegisterRoutes(v1Router)

	// pobsnapshots endpoints
	poeHandler := pobsnapshots.NewHandler(s.repository)
	poeHandler.RegisterRoutes(v1Router)

	// frontend endpoints
	fHandler := frontend.NewHandler(s.repository, s.log)
	fHandler.RegisterRoutes(frontendRouter)

	router.Mount("/api/v1", v1Router)
	router.Mount("/", frontendRouter)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: router,
	}

	s.log.Info().Msgf("Starting API server, listening on %s", s.addr)
	return s.server.ListenAndServe()
}

func (s *APIServer) Stop(ctx context.Context) error {
	log := utils.ChildLogger("api")
	log.Info().Msg("Stopping API server")
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
