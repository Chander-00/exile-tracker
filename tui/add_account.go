package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ByChanderZap/exile-tracker/repository"
)

type accountCreatedMsg struct {
	err error
}

type addAccount struct {
	repo      *repository.Repository
	inputs    []textinput.Model
	focusIdx  int
	width     int
	height    int
	submitted bool
	success   bool
	err       error
}

func newAddAccount(repo *repository.Repository, width, height int) *addAccount {
	nameInput := textinput.New()
	nameInput.Placeholder = "Account name (required)"
	nameInput.Prompt = "Account Name: "
	nameInput.CharLimit = 100
	nameInput.Focus()

	playerInput := textinput.New()
	playerInput.Placeholder = "Player name (optional)"
	playerInput.Prompt = "Player: "
	playerInput.CharLimit = 100

	return &addAccount{
		repo:   repo,
		inputs: []textinput.Model{nameInput, playerInput},
		width:  width,
		height: height,
	}
}

func (a *addAccount) Init() tea.Cmd {
	return textinput.Blink
}

func (a *addAccount) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case accountCreatedMsg:
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
				err := a.repo.CreateAccount(accountName, player)
				return accountCreatedMsg{err: err}
			}

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

func (a *addAccount) View() string {
	if a.success {
		return contentStyle.Render(
			successStyle.Render("Account created successfully!") + "\n\n" +
				helpStyle.Render("Press enter to go back"),
		)
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(gold).Render("Add New Account")

	var errMsg string
	if a.err != nil {
		errMsg = "\n" + errorStyle.Render("Error: "+a.err.Error())
	}

	status := ""
	if a.submitted {
		status = "\n" + helpStyle.Render("Creating account...")
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
