package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ByChanderZap/exile-tracker/repository"
)

type trashAccountsDataMsg struct {
	accounts []accountRow
	err      error
}

type accountRestoredMsg struct {
	err error
}

type trashAccountsList struct {
	repo       *repository.Repository
	isAdmin    bool
	table      table.Model
	accountIDs []string
	allRows    []accountRow
	width      int
	height     int
	loading    bool
	err        error
}

func newTrashAccountsList(repo *repository.Repository, isAdmin bool, width, height int) *trashAccountsList {
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

	enableJKNav(&t)

	return &trashAccountsList{
		repo:    repo,
		isAdmin: isAdmin,
		table:   t,
		width:   width,
		height:  height,
		loading: true,
	}
}

func (a *trashAccountsList) Init() tea.Cmd {
	return a.loadAccounts()
}

func (a *trashAccountsList) loadAccounts() tea.Cmd {
	return func() tea.Msg {
		accounts, err := a.repo.GetDeletedAccounts()
		if err != nil {
			return trashAccountsDataMsg{err: err}
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
		return trashAccountsDataMsg{accounts: rows}
	}
}

func (a *trashAccountsList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.table.SetHeight(a.height - 10)
		return a, nil

	case trashAccountsDataMsg:
		a.loading = false
		if msg.err != nil {
			a.err = msg.err
			return a, nil
		}
		a.allRows = msg.accounts
		a.accountIDs = nil
		var rows []table.Row
		for _, acc := range a.allRows {
			rows = append(rows, table.Row{acc.accountName, acc.player, acc.createdAt, acc.updatedAt})
			a.accountIDs = append(a.accountIDs, acc.id)
		}
		a.table.SetRows(rows)
		return a, nil

	case accountRestoredMsg:
		if msg.err != nil {
			a.err = msg.err
			return a, nil
		}
		a.loading = true
		return a, a.loadAccounts()

	case tea.KeyMsg:
		switch msg.String() {
		case "u":
			if !a.isAdmin {
				return a, nil
			}
			idx := a.table.Cursor()
			if idx >= 0 && idx < len(a.accountIDs) {
				accountID := a.accountIDs[idx]
				return a, func() tea.Msg {
					err := a.repo.RestoreAccount(accountID)
					return accountRestoredMsg{err: err}
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

func (a *trashAccountsList) View() string {
	if a.loading {
		return contentStyle.Render("Loading deleted accounts...")
	}
	if a.err != nil {
		return contentStyle.Render(errorStyle.Render("Error: " + a.err.Error()))
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(red).Render(
		fmt.Sprintf("Deleted Accounts (%d)", len(a.accountIDs)),
	)

	help := ""
	if a.isAdmin {
		help = "u: restore | "
	}
	help += "r: refresh | esc: back | q: back"
	helpLine := helpStyle.Render(help)

	return contentStyle.Render(title + "\n\n" + a.table.View() + "\n\n" + helpLine)
}
