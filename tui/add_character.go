package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ByChanderZap/exile-tracker/repository"
)

type characterCreatedMsg struct {
	err error
}

type addCharacter struct {
	repo      *repository.Repository
	accountID string
	inputs    []textinput.Model
	focusIdx  int
	width     int
	height    int
	submitted bool
	success   bool
	err       error
}

func newAddCharacter(repo *repository.Repository, accountID string, width, height int) *addCharacter {
	nameInput := textinput.New()
	nameInput.Placeholder = "Character name (required)"
	nameInput.Prompt = "Character Name: "
	nameInput.CharLimit = 100
	nameInput.Focus()

	leagueInput := textinput.New()
	leagueInput.Placeholder = "Current league (required)"
	leagueInput.Prompt = "Current League: "
	leagueInput.CharLimit = 100

	return &addCharacter{
		repo:      repo,
		accountID: accountID,
		inputs:    []textinput.Model{nameInput, leagueInput},
		width:     width,
		height:    height,
	}
}

func (a *addCharacter) Init() tea.Cmd {
	return textinput.Blink
}

func (a *addCharacter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case characterCreatedMsg:
		a.submitted = false
		if msg.err != nil {
			a.err = msg.err
			return a, nil
		}
		a.success = true
		return a, nil

	case tea.KeyMsg:
		if a.success {
			switch msg.String() {
			case "enter", "esc", "q":
				return a, func() tea.Msg { return popViewMsg{} }
			}
			return a, nil
		}

		switch msg.String() {
		case "tab":
			a.focusIdx = (a.focusIdx + 1) % len(a.inputs)
			cmds := make([]tea.Cmd, len(a.inputs))
			for i := range a.inputs {
				if i == a.focusIdx {
					cmds[i] = a.inputs[i].Focus()
				} else {
					a.inputs[i].Blur()
				}
			}
			return a, tea.Batch(cmds...)

		case "enter":
			characterName := a.inputs[0].Value()
			currentLeague := a.inputs[1].Value()
			if characterName == "" || currentLeague == "" {
				a.err = nil
				return a, nil
			}
			a.submitted = true
			return a, func() tea.Msg {
				charID, err := a.repo.CreateCharacter(a.accountID, characterName, currentLeague)
				if err != nil {
					return characterCreatedMsg{err: err}
				}
				err = a.repo.AddCharacterToFetch(repository.AddCharactersToFetchParams{
					CharacterId: charID,
				})
				return characterCreatedMsg{err: err}
			}

		case "esc":
			return a, func() tea.Msg { return popViewMsg{} }

		case "q":
			if a.focusIdx != 0 || a.inputs[0].Value() == "" {
				return a, func() tea.Msg { return popViewMsg{} }
			}
		}
	}

	// Update focused input
	var cmd tea.Cmd
	a.inputs[a.focusIdx], cmd = a.inputs[a.focusIdx].Update(msg)
	return a, cmd
}

func (a *addCharacter) View() string {
	if a.success {
		return contentStyle.Render(
			successStyle.Render("Character created successfully!") + "\n\n" +
				helpStyle.Render("Press enter to go back"),
		)
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(gold).Render("Add New Character")

	var errMsg string
	if a.err != nil {
		errMsg = "\n" + errorStyle.Render("Error: "+a.err.Error())
	}

	status := ""
	if a.submitted {
		status = "\n" + helpStyle.Render("Creating character...")
	}

	help := helpStyle.Render("tab: next field | enter: submit | esc: cancel")

	return contentStyle.Render(
		title + "\n\n" +
			a.inputs[0].View() + "\n\n" +
			a.inputs[1].View() + "\n" +
			errMsg + status + "\n\n" +
			help,
	)
}
