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
	allRows    []accountRow
	search     textinput.Model
	searching  bool
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

	ti := textinput.New()
	ti.Placeholder = "Search accounts..."
	ti.CharLimit = 64

	return &accountsList{
		repo:    repo,
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

	case tea.KeyMsg:
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

	help := helpStyle.Render("enter: view characters | /: search | r: refresh | esc: back | q: back")

	return contentStyle.Render(title + searchLine + "\n" + a.table.View() + "\n\n" + help)
}
