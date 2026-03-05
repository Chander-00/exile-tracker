package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ByChanderZap/exile-tracker/repository"
)

type characterUpdatedMsg struct {
	err error
}

type editCharacter struct {
	repo        *repository.Repository
	characterID string
	died        bool
	inputs      []textinput.Model
	focusIdx    int
	width       int
	height      int
	submitted   bool
	success     bool
	err         error
}

func newEditCharacter(repo *repository.Repository, characterID, characterName, league string, died bool, width, height int) *editCharacter {
	nameInput := textinput.New()
	nameInput.Placeholder = "Character name (required)"
	nameInput.Prompt = "Character Name: "
	nameInput.CharLimit = 100
	nameInput.SetValue(characterName)
	nameInput.Focus()

	leagueInput := textinput.New()
	leagueInput.Placeholder = "Current league (required)"
	leagueInput.Prompt = "Current League: "
	leagueInput.CharLimit = 100
	leagueInput.SetValue(league)

	return &editCharacter{
		repo:        repo,
		characterID: characterID,
		died:        died,
		inputs:      []textinput.Model{nameInput, leagueInput},
		width:       width,
		height:      height,
	}
}

func (a *editCharacter) Init() tea.Cmd {
	return textinput.Blink
}

func (a *editCharacter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case characterUpdatedMsg:
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
				err := a.repo.UpdateCharacter(repository.UpdateCharacterParams{
					ID:            a.characterID,
					CharacterName: characterName,
					Died:          a.died,
					CurrentLeague: currentLeague,
					UpdatedAt:     time.Now().UTC().Format(time.RFC3339),
				})
				return characterUpdatedMsg{err: err}
			}

		case "esc":
			return a, func() tea.Msg { return popViewMsg{} }

		case "q":
			if a.focusIdx != 0 || a.inputs[0].Value() == "" {
				return a, func() tea.Msg { return popViewMsg{} }
			}
		}
	}

	var cmd tea.Cmd
	a.inputs[a.focusIdx], cmd = a.inputs[a.focusIdx].Update(msg)
	return a, cmd
}

func (a *editCharacter) View() string {
	if a.success {
		return contentStyle.Render(
			successStyle.Render("Character updated successfully!") + "\n\n" +
				helpStyle.Render("Press enter to go back"),
		)
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(gold).Render("Edit Character")

	var errMsg string
	if a.err != nil {
		errMsg = "\n" + errorStyle.Render("Error: "+a.err.Error())
	}

	status := ""
	if a.submitted {
		status = "\n" + helpStyle.Render("Updating character...")
	}

	help := helpStyle.Render("tab: next field | enter: submit | esc: cancel")

	return contentStyle.Render(
		title + "\n\n" +
			renderInput(a.inputs[0], a.focusIdx == 0) + "\n\n" +
			renderInput(a.inputs[1], a.focusIdx == 1) + "\n" +
			errMsg + status + "\n\n" +
			help,
	)
}
