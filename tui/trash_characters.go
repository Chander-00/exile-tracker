package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ByChanderZap/exile-tracker/repository"
)

type trashCharactersDataMsg struct {
	characters []characterRow
	err        error
}

type characterRestoredMsg struct {
	err error
}

type trashCharactersList struct {
	repo         *repository.Repository
	isAdmin      bool
	accountID    string
	accountName  string
	table        table.Model
	characterIDs []string
	allRows      []characterRow
	width        int
	height       int
	loading      bool
	err          error
}

func newTrashCharactersList(repo *repository.Repository, isAdmin bool, accountID, accountName string, width, height int) *trashCharactersList {
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

	enableJKNav(&t)

	return &trashCharactersList{
		repo:        repo,
		isAdmin:     isAdmin,
		accountID:   accountID,
		accountName: accountName,
		table:       t,
		width:       width,
		height:      height,
		loading:     true,
	}
}

func (c *trashCharactersList) Init() tea.Cmd {
	return c.loadCharacters()
}

func (c *trashCharactersList) loadCharacters() tea.Cmd {
	return func() tea.Msg {
		chars, err := c.repo.GetDeletedCharactersByAccountId(c.accountID)
		if err != nil {
			return trashCharactersDataMsg{err: err}
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
		return trashCharactersDataMsg{characters: rows}
	}
}

func (c *trashCharactersList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.table.SetHeight(c.height - 10)
		return c, nil

	case trashCharactersDataMsg:
		c.loading = false
		if msg.err != nil {
			c.err = msg.err
			return c, nil
		}
		c.allRows = msg.characters
		c.characterIDs = nil
		var rows []table.Row
		for _, ch := range c.allRows {
			status := "Alive"
			if ch.died {
				status = "Dead"
			}
			rows = append(rows, table.Row{ch.characterName, ch.league, status, ch.updatedAt})
			c.characterIDs = append(c.characterIDs, ch.id)
		}
		c.table.SetRows(rows)
		return c, nil

	case characterRestoredMsg:
		if msg.err != nil {
			c.err = msg.err
			return c, nil
		}
		c.loading = true
		return c, c.loadCharacters()

	case tea.KeyMsg:
		switch msg.String() {
		case "u":
			if !c.isAdmin {
				return c, nil
			}
			idx := c.table.Cursor()
			if idx >= 0 && idx < len(c.characterIDs) {
				charID := c.characterIDs[idx]
				return c, func() tea.Msg {
					err := c.repo.RestoreCharacter(charID)
					return characterRestoredMsg{err: err}
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

func (c *trashCharactersList) View() string {
	if c.loading {
		return contentStyle.Render("Loading deleted characters...")
	}
	if c.err != nil {
		return contentStyle.Render(errorStyle.Render("Error: " + c.err.Error()))
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(red).Render(
		fmt.Sprintf("Deleted Characters for %s (%d)", c.accountName, len(c.characterIDs)),
	)

	help := ""
	if c.isAdmin {
		help = "u: restore | "
	}
	help += "r: refresh | esc: back | q: back"
	helpLine := helpStyle.Render(help)

	return contentStyle.Render(title + "\n\n" + c.table.View() + "\n\n" + helpLine)
}
