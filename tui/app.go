package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ByChanderZap/exile-tracker/repository"
)

// Messages for stack-based navigation.
type pushViewMsg struct {
	model tea.Model
	title string
}

type popViewMsg struct{}

// stackEntry holds a view and its display title.
type stackEntry struct {
	model tea.Model
	title string
}

// App is the root Bubbletea model managing a stack of sub-views.
type App struct {
	repo    *repository.Repository
	stack   []stackEntry
	isAdmin bool
	width   int
	height  int
}

func NewApp(repo *repository.Repository, isAdmin bool, width, height int) *App {
	dash := newDashboard(repo, isAdmin, width, height-3) // reserve space for header
	return &App{
		repo:    repo,
		stack:   []stackEntry{{model: dash, title: "Dashboard"}},
		isAdmin: isAdmin,
		width:   width,
		height:  height,
	}
}

func (a *App) Init() tea.Cmd {
	if len(a.stack) > 0 {
		return a.stack[0].model.Init()
	}
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Forward resize to active view
		if len(a.stack) > 0 {
			top := a.stack[len(a.stack)-1]
			updated, cmd := top.model.Update(msg)
			a.stack[len(a.stack)-1].model = updated
			return a, cmd
		}
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "esc", "backspace":
			if len(a.stack) > 1 {
				a.stack = a.stack[:len(a.stack)-1]
				return a, nil
			}
			return a, tea.Quit
		}

	case pushViewMsg:
		cmd := msg.model.Init()
		a.stack = append(a.stack, stackEntry{model: msg.model, title: msg.title})
		return a, cmd

	case popViewMsg:
		if len(a.stack) > 1 {
			a.stack = a.stack[:len(a.stack)-1]
			return a, nil
		}
		return a, tea.Quit
	}

	// Forward all other messages to the top-of-stack view
	if len(a.stack) > 0 {
		idx := len(a.stack) - 1
		updated, cmd := a.stack[idx].model.Update(msg)
		a.stack[idx].model = updated
		return a, cmd
	}

	return a, nil
}

func (a *App) View() string {
	if len(a.stack) == 0 {
		return ""
	}

	header := a.renderHeader()
	body := a.stack[len(a.stack)-1].model.View()

	return header + "\n" + body
}

func (a *App) renderHeader() string {
	title := titleStyle.Render(" Exile Tracker ")

	var badge string
	if a.isAdmin {
		badge = " " + adminBadge.Render("ADMIN")
	}

	breadcrumb := a.renderBreadcrumb()

	left := title + badge
	right := breadcrumbStyle.Render(breadcrumb)

	gap := a.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return headerStyle.Width(a.width).Render(left + strings.Repeat(" ", gap) + right)
}

func (a *App) renderBreadcrumb() string {
	if len(a.stack) <= 1 {
		return ""
	}
	parts := make([]string, len(a.stack))
	for i, entry := range a.stack {
		parts[i] = entry.title
	}
	return strings.Join(parts, " > ")
}
