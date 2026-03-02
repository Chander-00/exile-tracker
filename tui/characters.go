package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ByChanderZap/exile-tracker/repository"
)

type charactersDataMsg struct {
	characters []characterRow
	err        error
}

type characterRow struct {
	id            string
	characterName string
	league        string
	died          bool
	updatedAt     string
}

type charactersList struct {
	repo         *repository.Repository
	accountID    string
	accountName  string
	table        table.Model
	characterIDs []string
	width        int
	height       int
	loading      bool
	err          error
}

func newCharactersList(repo *repository.Repository, accountID, accountName string, width, height int) *charactersList {
	columns := []table.Column{
		{Title: "Character", Width: 25},
		{Title: "League", Width: 20},
		{Title: "Status", Width: 10},
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

	return &charactersList{
		repo:        repo,
		accountID:   accountID,
		accountName: accountName,
		table:       t,
		width:       width,
		height:      height,
		loading:     true,
	}
}

func (c *charactersList) Init() tea.Cmd {
	return c.loadCharacters()
}

func (c *charactersList) loadCharacters() tea.Cmd {
	return func() tea.Msg {
		chars, err := c.repo.GetCharactersByAccountId(c.accountID)
		if err != nil {
			return charactersDataMsg{err: err}
		}
		var rows []characterRow
		for _, ch := range chars {
			league := ""
			if ch.CurrentLeague != nil {
				league = *ch.CurrentLeague
			}
			rows = append(rows, characterRow{
				id:            ch.ID,
				characterName: ch.CharacterName,
				league:        league,
				died:          ch.Died,
				updatedAt:     ch.UpdatedAt.Format("2006-01-02 15:04"),
			})
		}
		return charactersDataMsg{characters: rows}
	}
}

func (c *charactersList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.table.SetHeight(c.height - 10)
		return c, nil

	case charactersDataMsg:
		c.loading = false
		if msg.err != nil {
			c.err = msg.err
			return c, nil
		}
		var rows []table.Row
		c.characterIDs = nil
		for _, ch := range msg.characters {
			status := "Alive"
			if ch.died {
				status = "Dead"
			}
			rows = append(rows, table.Row{ch.characterName, ch.league, status, ch.updatedAt})
			c.characterIDs = append(c.characterIDs, ch.id)
		}
		c.table.SetRows(rows)
		return c, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			idx := c.table.Cursor()
			if idx >= 0 && idx < len(c.characterIDs) {
				charID := c.characterIDs[idx]
				charName := ""
				if row := c.table.SelectedRow(); row != nil {
					charName = row[0]
				}
				return c, func() tea.Msg {
					return pushViewMsg{
						model: newSnapshots(c.repo, charID, charName, c.width, c.height),
						title: charName,
					}
				}
			}
		case "r":
			c.loading = true
			c.err = nil
			return c, c.loadCharacters()
		case "q":
			return c, func() tea.Msg { return popViewMsg{} }
		}
	}

	var cmd tea.Cmd
	c.table, cmd = c.table.Update(msg)
	return c, cmd
}

func (c *charactersList) View() string {
	if c.loading {
		return contentStyle.Render("Loading characters...")
	}
	if c.err != nil {
		return contentStyle.Render(errorStyle.Render("Error: " + c.err.Error()))
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(gold).Render(
		fmt.Sprintf("Characters for %s (%d)", c.accountName, len(c.characterIDs)),
	)

	help := helpStyle.Render("enter: view snapshots | r: refresh | esc: back | q: back")

	return contentStyle.Render(title + "\n\n" + c.table.View() + "\n\n" + help)
}
