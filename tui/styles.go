package tui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	// PoE-inspired color palette
	gold     = lipgloss.Color("#AF8700")
	steel    = lipgloss.Color("#5F87AF")
	dimGray  = lipgloss.Color("#585858")
	darkBg   = lipgloss.Color("#1C1C1C")
	white    = lipgloss.Color("#E4E4E4")
	red      = lipgloss.Color("#D75F5F")
	green    = lipgloss.Color("#87AF5F")
	darkGold = lipgloss.Color("#875F00")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(gold).
			Background(darkBg).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(dimGray).
			Background(darkBg)

	helpStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	adminBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(gold).
			Padding(0, 1)

	statBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(steel).
			Padding(0, 1).
			Width(20)

	statLabelStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	statValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(gold)

	deadStyle = lipgloss.NewStyle().
			Foreground(red)

	aliveStyle = lipgloss.NewStyle().
			Foreground(green)

	contentStyle = lipgloss.NewStyle().
			Padding(1, 2)

	breadcrumbStyle = lipgloss.NewStyle().
			Foreground(darkGold)

	headerStyle = lipgloss.NewStyle().
			Background(darkBg).
			Width(80)

	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	focusedInputStyle = lipgloss.NewStyle().
				Foreground(gold).
				Bold(true)

	blurredInputStyle = lipgloss.NewStyle().
				Foreground(dimGray)
)

// renderInput renders a text input with a visual focus indicator.
func renderInput(ti textinput.Model, focused bool) string {
	indicator := "  "
	if focused {
		indicator = focusedInputStyle.Render("▸ ")
	}
	return indicator + ti.View()
}

// enableJKNav adds j/k keys to a table's up/down keybindings.
func enableJKNav(t *table.Model) {
	km := t.KeyMap
	km.LineDown.SetKeys("down", "j")
	km.LineUp.SetKeys("up", "k")
	t.KeyMap = km
}
