package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ByChanderZap/exile-tracker/repository"
)

type accountsDataMsg struct {
	accounts []accountRow
	err      error
}

type accountRow struct {
	id          string
	accountName string
	player      string
	createdAt   string
	updatedAt   string
}

type accountsList struct {
	repo       *repository.Repository
	table      table.Model
	accountIDs []string
	width      int
	height     int
	loading    bool
	err        error
}

func newAccountsList(repo *repository.Repository, width, height int) *accountsList {
	columns := []table.Column{
		{Title: "Account Name", Width: 25},
		{Title: "Player", Width: 20},
		{Title: "Created", Width: 20},
		{Title: "Updated", Width: 20},
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

	return &accountsList{
		repo:    repo,
		table:   t,
		width:   width,
		height:  height,
		loading: true,
	}
}

func (a *accountsList) Init() tea.Cmd {
	return a.loadAccounts()
}

func (a *accountsList) loadAccounts() tea.Cmd {
	return func() tea.Msg {
		accounts, err := a.repo.GetAllAccounts()
		if err != nil {
			return accountsDataMsg{err: err}
		}
		var rows []accountRow
		for _, acc := range accounts {
			player := ""
			if acc.Player != nil {
				player = *acc.Player
			}
			rows = append(rows, accountRow{
				id:          acc.ID,
				accountName: acc.AccountName,
				player:      player,
				createdAt:   acc.CreatedAt.Format("2006-01-02 15:04"),
				updatedAt:   acc.UpdatedAt.Format("2006-01-02 15:04"),
			})
		}
		return accountsDataMsg{accounts: rows}
	}
}

func (a *accountsList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.table.SetHeight(a.height - 10)
		return a, nil

	case accountsDataMsg:
		a.loading = false
		if msg.err != nil {
			a.err = msg.err
			return a, nil
		}
		var rows []table.Row
		a.accountIDs = nil
		for _, acc := range msg.accounts {
			rows = append(rows, table.Row{acc.accountName, acc.player, acc.createdAt, acc.updatedAt})
			a.accountIDs = append(a.accountIDs, acc.id)
		}
		a.table.SetRows(rows)
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			idx := a.table.Cursor()
			if idx >= 0 && idx < len(a.accountIDs) {
				accountID := a.accountIDs[idx]
				accountName := ""
				if row := a.table.SelectedRow(); row != nil {
					accountName = row[0]
				}
				return a, func() tea.Msg {
					return pushViewMsg{
						model: newCharactersList(a.repo, accountID, accountName, a.width, a.height),
						title: accountName,
					}
				}
			}
		case "r":
			a.loading = true
			a.err = nil
			return a, a.loadAccounts()
		case "q":
			return a, func() tea.Msg { return popViewMsg{} }
		}
	}

	var cmd tea.Cmd
	a.table, cmd = a.table.Update(msg)
	return a, cmd
}

func (a *accountsList) View() string {
	if a.loading {
		return contentStyle.Render("Loading accounts...")
	}
	if a.err != nil {
		return contentStyle.Render(errorStyle.Render("Error: " + a.err.Error()))
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(gold).Render(
		fmt.Sprintf("Accounts (%d)", len(a.accountIDs)),
	)

	help := helpStyle.Render("enter: view characters | r: refresh | esc: back | q: back")

	return contentStyle.Render(title + "\n\n" + a.table.View() + "\n\n" + help)
}
