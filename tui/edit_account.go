package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ByChanderZap/exile-tracker/repository"
)

type accountUpdatedMsg struct {
	err error
}

type editAccount struct {
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

func newEditAccount(repo *repository.Repository, accountID, accountName, player string, width, height int) *editAccount {
	nameInput := textinput.New()
	nameInput.Placeholder = "Account name (required)"
	nameInput.Prompt = "Account Name: "
	nameInput.CharLimit = 100
	nameInput.SetValue(accountName)
	nameInput.Focus()

	playerInput := textinput.New()
	playerInput.Placeholder = "Player name (optional)"
	playerInput.Prompt = "Player: "
	playerInput.CharLimit = 100
	playerInput.SetValue(player)

	return &editAccount{
		repo:      repo,
		accountID: accountID,
		inputs:    []textinput.Model{nameInput, playerInput},
		width:     width,
		height:    height,
	}
}

func (a *editAccount) Init() tea.Cmd {
	return textinput.Blink
}

func (a *editAccount) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case accountUpdatedMsg:
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
			accountName := a.inputs[0].Value()
			if accountName == "" {
				a.err = nil
				return a, nil
			}
			player := a.inputs[1].Value()
			a.submitted = true
			return a, func() tea.Msg {
				err := a.repo.UpdateAccount(repository.UpdateAccountParams{
					ID:          a.accountID,
					AccountName: accountName,
					Player:      player,
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				})
				return accountUpdatedMsg{err: err}
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

func (a *editAccount) View() string {
	if a.success {
		return contentStyle.Render(
			successStyle.Render("Account updated successfully!") + "\n\n" +
				helpStyle.Render("Press enter to go back"),
		)
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(gold).Render("Edit Account")

	var errMsg string
	if a.err != nil {
		errMsg = "\n" + errorStyle.Render("Error: "+a.err.Error())
	}

	status := ""
	if a.submitted {
		status = "\n" + helpStyle.Render("Updating account...")
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
