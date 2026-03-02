package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ByChanderZap/exile-tracker/models"
	"github.com/ByChanderZap/exile-tracker/repository"
)

type snapshotsDataMsg struct {
	snapshots []models.POBSnapshot
	err       error
}

type snapshots struct {
	repo        *repository.Repository
	characterID string
	charName    string
	table       table.Model
	viewport    viewport.Model
	data        []models.POBSnapshot
	showExport  bool
	width       int
	height      int
	loading     bool
	err         error
}

func newSnapshots(repo *repository.Repository, characterID, charName string, width, height int) *snapshots {
	columns := []table.Column{
		{Title: "#", Width: 5},
		{Title: "Created", Width: 25},
		{Title: "Export (preview)", Width: 40},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(height-10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(steel).
		BorderBottom(true).
		Bold(true).
		Foreground(gold)
	s.Selected = s.Selected.
		Foreground(white).
		Background(steel).
		Bold(false)
	t.SetStyles(s)

	vp := viewport.New(width-4, height-6)

	return &snapshots{
		repo:        repo,
		characterID: characterID,
		charName:    charName,
		table:       t,
		viewport:    vp,
		width:       width,
		height:      height,
		loading:     true,
	}
}

func (sn *snapshots) Init() tea.Cmd {
	return sn.loadSnapshots()
}

func (sn *snapshots) loadSnapshots() tea.Cmd {
	return func() tea.Msg {
		snaps, err := sn.repo.GetSnapshotsByCharacter(sn.characterID)
		if err != nil {
			return snapshotsDataMsg{err: err}
		}
		return snapshotsDataMsg{snapshots: snaps}
	}
}

func (sn *snapshots) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		sn.width = msg.Width
		sn.height = msg.Height
		sn.table.SetHeight(sn.height - 10)
		sn.viewport.Width = sn.width - 4
		sn.viewport.Height = sn.height - 6
		return sn, nil

	case snapshotsDataMsg:
		sn.loading = false
		if msg.err != nil {
			sn.err = msg.err
			return sn, nil
		}
		sn.data = msg.snapshots
		var rows []table.Row
		for i, snap := range sn.data {
			preview := snap.ExportString
			if len(preview) > 37 {
				preview = preview[:37] + "..."
			}
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", i+1),
				snap.CreatedAt.Format("2006-01-02 15:04:05"),
				preview,
			})
		}
		sn.table.SetRows(rows)
		return sn, nil

	case tea.KeyMsg:
		if sn.showExport {
			switch msg.String() {
			case "esc", "backspace", "q":
				sn.showExport = false
				return sn, nil
			}
			var cmd tea.Cmd
			sn.viewport, cmd = sn.viewport.Update(msg)
			return sn, cmd
		}

		switch msg.String() {
		case "enter":
			idx := sn.table.Cursor()
			if idx >= 0 && idx < len(sn.data) {
				sn.viewport.SetContent(sn.data[idx].ExportString)
				sn.viewport.GotoTop()
				sn.showExport = true
				return sn, nil
			}
		case "r":
			sn.loading = true
			sn.err = nil
			return sn, sn.loadSnapshots()
		case "q":
			return sn, func() tea.Msg { return popViewMsg{} }
		}
	}

	var cmd tea.Cmd
	sn.table, cmd = sn.table.Update(msg)
	return sn, cmd
}

func (sn *snapshots) View() string {
	if sn.loading {
		return contentStyle.Render("Loading snapshots...")
	}
	if sn.err != nil {
		return contentStyle.Render(errorStyle.Render("Error: " + sn.err.Error()))
	}

	if sn.showExport {
		title := lipgloss.NewStyle().Bold(true).Foreground(gold).Render("PoB Export String")
		help := helpStyle.Render("scroll: up/down | esc: back to list")
		return contentStyle.Render(title + "\n\n" + sn.viewport.View() + "\n\n" + help)
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(gold).Render(
		fmt.Sprintf("Snapshots for %s (%d)", sn.charName, len(sn.data)),
	)

	help := helpStyle.Render("enter: view export | r: refresh | esc: back | q: back")

	return contentStyle.Render(title + "\n\n" + sn.table.View() + "\n\n" + help)
}
