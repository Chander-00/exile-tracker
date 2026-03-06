package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ByChanderZap/exile-tracker/repository"
)

type dashboardDataMsg struct {
	stats  repository.DashboardStats
	recent []repository.RecentCharacter
	err    error
}

type dashboard struct {
	repo    *repository.Repository
	isAdmin bool
	width   int
	height  int
	stats   repository.DashboardStats
	recent  []repository.RecentCharacter
	loading bool
	err     error
	cursor  int
}

func newDashboard(repo *repository.Repository, isAdmin bool, width, height int) *dashboard {
	return &dashboard{
		repo:    repo,
		isAdmin: isAdmin,
		width:   width,
		height:  height,
		loading: true,
	}
}

func (d *dashboard) Init() tea.Cmd {
	return d.loadData()
}

func (d *dashboard) loadData() tea.Cmd {
	return func() tea.Msg {
		stats, err := d.repo.GetDashboardStats()
		if err != nil {
			return dashboardDataMsg{err: err}
		}
		recent, err := d.repo.GetRecentlyUpdatedCharacters(10)
		if err != nil {
			return dashboardDataMsg{err: err}
		}
		return dashboardDataMsg{stats: stats, recent: recent}
	}
}

func (d *dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		return d, nil

	case dashboardDataMsg:
		d.loading = false
		if msg.err != nil {
			d.err = msg.err
			return d, nil
		}
		d.stats = msg.stats
		d.recent = msg.recent
		return d, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return d, func() tea.Msg {
				return pushViewMsg{
					model: newAccountsList(d.repo, d.isAdmin, d.width, d.height),
					title: "Accounts",
				}
			}
		case "a":
			if d.isAdmin {
				return d, func() tea.Msg {
					return pushViewMsg{
						model: newAddAccount(d.repo, d.width, d.height),
						title: "Add Account",
					}
				}
			}
		case "t":
			return d, func() tea.Msg {
				return pushViewMsg{
					model: newTrashAccountsList(d.repo, d.isAdmin, d.width, d.height),
					title: "Trash",
				}
			}
		case "L":
			if d.isAdmin {
				return d, func() tea.Msg {
					return pushViewMsg{
						model: newLeagueReset(d.repo, d.width, d.height),
						title: "League Reset",
					}
				}
			}
		case "r":
			d.loading = true
			d.err = nil
			return d, d.loadData()
		case "j", "down":
			if d.cursor < len(d.recent)-1 {
				d.cursor++
			}
		case "k", "up":
			if d.cursor > 0 {
				d.cursor--
			}
		}
	}
	return d, nil
}

func (d *dashboard) View() string {
	if d.loading {
		return contentStyle.Render("Loading dashboard data...")
	}
	if d.err != nil {
		return contentStyle.Render(errorStyle.Render("Error: " + d.err.Error()))
	}

	var b strings.Builder

	// Stat boxes row
	boxes := []string{
		d.renderStatBox("Accounts", d.stats.AccountCount),
		d.renderStatBox("Characters", d.stats.CharacterCount),
		d.renderStatBox("Snapshots", d.stats.SnapshotCount),
		d.renderStatBox("Fetch Queue", d.stats.FetchQueueSize),
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, boxes...))
	b.WriteString("\n\n")

	// Recently updated characters
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(gold).Render("Recently Updated Characters"))
	b.WriteString("\n\n")

	if len(d.recent) == 0 {
		b.WriteString(helpStyle.Render("  No characters found"))
	} else {
		for i, rc := range d.recent {
			cursor := "  "
			if i == d.cursor {
				cursor = "> "
			}

			status := aliveStyle.Render("Alive")
			if rc.Died {
				status = deadStyle.Render("Dead")
			}

			league := ""
			if rc.CurrentLeague != nil {
				league = *rc.CurrentLeague
			}

			line := fmt.Sprintf("%s%-18.18s %-16.16s %-10.10s %s",
				cursor,
				rc.CharacterName,
				rc.AccountName,
				league,
				status,
			)
			b.WriteString(line + "\n")
		}
	}

	b.WriteString("\n")

	// Help
	help := "enter: accounts | t: trash"
	if d.isAdmin {
		help += " | a: add account | L: league reset"
	}
	help += " | r: refresh | q: quit"
	b.WriteString(helpStyle.Render(help))

	return contentStyle.Render(b.String())
}

func (d *dashboard) renderStatBox(label string, value int) string {
	content := statLabelStyle.Render(label) + "\n" + statValueStyle.Render(fmt.Sprintf("%d", value))
	return statBoxStyle.Render(content)
}
