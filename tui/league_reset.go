package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ByChanderZap/exile-tracker/repository"
)

type leagueResetDoneMsg struct{ err error }

type leagueReset struct {
	repo      *repository.Repository
	width     int
	height    int
	confirmed bool
	done      bool
	err       error
}

func newLeagueReset(repo *repository.Repository, width, height int) *leagueReset {
	return &leagueReset{
		repo:   repo,
		width:  width,
		height: height,
	}
}

func (l *leagueReset) Init() tea.Cmd {
	return nil
}

func (l *leagueReset) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.width = msg.Width
		l.height = msg.Height
		return l, nil

	case leagueResetDoneMsg:
		l.done = true
		l.err = msg.err
		return l, nil

	case tea.KeyMsg:
		if l.done {
			return l, func() tea.Msg { return popViewMsg{} }
		}

		switch msg.String() {
		case "y", "Y":
			l.confirmed = true
			return l, func() tea.Msg {
				err := l.repo.LeagueReset()
				return leagueResetDoneMsg{err: err}
			}
		case "n", "N", "esc":
			return l, func() tea.Msg { return popViewMsg{} }
		}
	}
	return l, nil
}

func (l *leagueReset) View() string {
	var b strings.Builder

	if l.done {
		if l.err != nil {
			b.WriteString(errorStyle.Render("League reset failed: " + l.err.Error()))
		} else {
			b.WriteString(successStyle.Render("League reset complete! All characters and snapshots have been deleted."))
		}
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press any key to go back"))
		return contentStyle.Render(b.String())
	}

	if l.confirmed {
		b.WriteString("Resetting league data...")
		return contentStyle.Render(b.String())
	}

	b.WriteString(errorStyle.Render("WARNING: League Reset"))
	b.WriteString("\n\n")
	b.WriteString("This will permanently delete ALL:\n")
	b.WriteString("  - Characters\n")
	b.WriteString("  - PoB Snapshots\n")
	b.WriteString("  - Fetch queue entries\n")
	b.WriteString("\n")
	b.WriteString("Accounts will be kept.\n\n")
	b.WriteString("Are you sure? (y/n)")

	return contentStyle.Render(b.String())
}
