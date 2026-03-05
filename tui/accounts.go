package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ByChanderZap/exile-tracker/repository"
)

type accountsDataMsg struct {
	accounts []accountRow
	err      error
}

type accountDeletedMsg struct {
	err error
}

type accountRow struct {
	id          string
	accountName string
	player      string
	createdAt   string
	updatedAt   string
}

type accountsList struct {
	repo          *repository.Repository
	isAdmin       bool
	table         table.Model
	accountIDs    []string
	allRows       []accountRow
	search        textinput.Model
	searching     bool
	confirmDelete bool
	width         int
	height        int
	loading       bool
	err           error
}

func newAccountsList(repo *repository.Repository, isAdmin bool, width, height int) *accountsList {
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

	ti := textinput.New()
	ti.Placeholder = "Search accounts..."
	ti.CharLimit = 64

	enableJKNav(&t)

	return &accountsList{
		repo:    repo,
		isAdmin: isAdmin,
		table:   t,
		search:  ti,
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
		a.allRows = msg.accounts
		a.applyFilter()
		return a, nil

	case accountDeletedMsg:
		if msg.err != nil {
			a.err = msg.err
			return a, nil
		}
		a.loading = true
		return a, a.loadAccounts()

	case tea.KeyMsg:
		if a.confirmDelete {
			switch msg.String() {
			case "y", "Y":
				a.confirmDelete = false
				idx := a.table.Cursor()
				if idx >= 0 && idx < len(a.accountIDs) {
					accountID := a.accountIDs[idx]
					return a, func() tea.Msg {
						err := a.repo.SoftDeleteAccount(accountID)
						return accountDeletedMsg{err: err}
					}
				}
				return a, nil
			default:
				a.confirmDelete = false
				return a, nil
			}
		}

		if a.searching {
			switch msg.String() {
			case "esc":
				a.searching = false
				a.search.Blur()
				a.search.SetValue("")
				a.applyFilter()
				return a, nil
			case "enter":
				a.searching = false
				a.search.Blur()
				return a, nil
			default:
				var cmd tea.Cmd
				a.search, cmd = a.search.Update(msg)
				a.applyFilter()
				return a, cmd
			}
		}

		switch msg.String() {
		case "/":
			a.searching = true
			a.search.Focus()
			return a, nil
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
						model: newCharactersList(a.repo, a.isAdmin, accountID, accountName, a.width, a.height),
						title: accountName,
					}
				}
			}
		case "e":
			if !a.isAdmin {
				return a, nil
			}
			idx := a.table.Cursor()
			if idx >= 0 && idx < len(a.accountIDs) {
				row := a.allRows[idx]
				return a, func() tea.Msg {
					return pushViewMsg{
						model: newEditAccount(a.repo, row.id, row.accountName, row.player, a.width, a.height),
						title: "Edit " + row.accountName,
					}
				}
			}
		case "x":
			if !a.isAdmin {
				return a, nil
			}
			idx := a.table.Cursor()
			if idx >= 0 && idx < len(a.accountIDs) {
				a.confirmDelete = true
				return a, nil
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

func (a *accountsList) applyFilter() {
	query := strings.ToLower(a.search.Value())
	var rows []table.Row
	a.accountIDs = nil
	for _, acc := range a.allRows {
		if query != "" &&
			!strings.Contains(strings.ToLower(acc.accountName), query) &&
			!strings.Contains(strings.ToLower(acc.player), query) {
			continue
		}
		rows = append(rows, table.Row{acc.accountName, acc.player, acc.createdAt, acc.updatedAt})
		a.accountIDs = append(a.accountIDs, acc.id)
	}
	a.table.SetRows(rows)
	a.table.GotoTop()
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

	var searchLine string
	if a.searching {
		searchLine = "\n" + a.search.View() + "\n"
	} else if a.search.Value() != "" {
		searchLine = "\n" + helpStyle.Render(fmt.Sprintf("filter: %s", a.search.Value())) + "\n"
	}

	var confirmLine string
	if a.confirmDelete {
		idx := a.table.Cursor()
		name := ""
		if idx >= 0 && idx < len(a.allRows) {
			name = a.allRows[idx].accountName
		}
		confirmLine = "\n" + errorStyle.Render(fmt.Sprintf("Delete account %q and all its characters? (y/n)", name)) + "\n"
	}

	help := "enter: view | /: search"
	if a.isAdmin {
		help += " | e: edit | x: delete"
	}
	help += " | r: refresh | esc: back | q: back"
	helpLine := helpStyle.Render(help)

	return contentStyle.Render(title + searchLine + confirmLine + "\n" + a.table.View() + "\n\n" + helpLine)
}
