package ssh

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/rs/zerolog"

	"github.com/ByChanderZap/exile-tracker/repository"
	"github.com/ByChanderZap/exile-tracker/tui"
	"github.com/ByChanderZap/exile-tracker/utils"
)

type SSHServer struct {
	server      *ssh.Server
	repo        *repository.Repository
	addr        string
	hostKeyPath string
	log         zerolog.Logger
}

func NewSSHServer(addr, hostKeyPath string, repo *repository.Repository) *SSHServer {
	return &SSHServer{
		repo:        repo,
		addr:        addr,
		hostKeyPath: hostKeyPath,
		log:         utils.ChildLogger("ssh"),
	}
}

func (s *SSHServer) Start() error {
	srv, err := wish.NewServer(
		ssh.AllocatePty(),
		wish.WithAddress(s.addr),
		wish.WithHostKeyPath(s.hostKeyPath),
		wish.WithPublicKeyAuth(PublicKeyHandler),
		wish.WithMiddleware(
			bm.Middleware(s.teaHandler()),
			lm.Middleware(),
		),
	)
	if err != nil {
		return fmt.Errorf("could not create SSH server: %w", err)
	}

	s.server = srv
	s.log.Info().Str("addr", s.addr).Msg("Starting SSH server")

	if err := srv.ListenAndServe(); err != nil && err != ssh.ErrServerClosed {
		return fmt.Errorf("SSH server error: %w", err)
	}
	return nil
}

func (s *SSHServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	s.log.Info().Msg("Shutting down SSH server")
	return s.server.Shutdown(ctx)
}

func (s *SSHServer) teaHandler() bm.Handler {
	return func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
		isAdmin := IsAdmin(sess)

		pty, _, ok := sess.Pty()
		width := 80
		height := 24
		if ok {
			width = pty.Window.Width
			height = pty.Window.Height
		}

		app := tui.NewApp(s.repo, isAdmin, width, height)
		return app, []tea.ProgramOption{tea.WithAltScreen()}
	}
}
