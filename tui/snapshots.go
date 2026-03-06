package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ByChanderZap/exile-tracker/models"
	"github.com/ByChanderZap/exile-tracker/pobparser"
	"github.com/ByChanderZap/exile-tracker/repository"
)

type snapshotsDataMsg struct {
	snapshots []models.POBSnapshot
	err       error
}

type snapshots struct {
	repo        *repository.Repository
	characterID string
	charName    string
	table       table.Model
	viewport    viewport.Model
	data        []models.POBSnapshot
	showDetail  bool
	width       int
	height      int
	loading     bool
	err         error
}

func newSnapshots(repo *repository.Repository, characterID, charName string, width, height int) *snapshots {
	columns := []table.Column{
		{Title: "#", Width: 5},
		{Title: "Created", Width: 25},
		{Title: "Export (preview)", Width: 40},
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

	vp := viewport.New(width-4, height-6)

	enableJKNav(&t)

	return &snapshots{
		repo:        repo,
		characterID: characterID,
		charName:    charName,
		table:       t,
		viewport:    vp,
		width:       width,
		height:      height,
		loading:     true,
	}
}

func (sn *snapshots) Init() tea.Cmd {
	return sn.loadSnapshots()
}

func (sn *snapshots) loadSnapshots() tea.Cmd {
	return func() tea.Msg {
		snaps, err := sn.repo.GetSnapshotsByCharacter(sn.characterID)
		if err != nil {
			return snapshotsDataMsg{err: err}
		}
		return snapshotsDataMsg{snapshots: snaps}
	}
}

func (sn *snapshots) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		sn.width = msg.Width
		sn.height = msg.Height
		sn.table.SetHeight(sn.height - 10)
		sn.viewport.Width = sn.width - 4
		sn.viewport.Height = sn.height - 6
		return sn, nil

	case snapshotsDataMsg:
		sn.loading = false
		if msg.err != nil {
			sn.err = msg.err
			return sn, nil
		}
		sn.data = msg.snapshots
		var rows []table.Row
		for i, snap := range sn.data {
			preview := snap.ExportString
			if len(preview) > 37 {
				preview = preview[:37] + "..."
			}
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", i+1),
				snap.CreatedAt.Format("2006-01-02 15:04:05"),
				preview,
			})
		}
		sn.table.SetRows(rows)
		return sn, nil

	case tea.KeyMsg:
		if sn.showDetail {
			switch msg.String() {
			case "esc", "backspace", "q":
				sn.showDetail = false
				return sn, nil
			}
			var cmd tea.Cmd
			sn.viewport, cmd = sn.viewport.Update(msg)
			return sn, cmd
		}

		switch msg.String() {
		case "enter":
			idx := sn.table.Cursor()
			if idx >= 0 && idx < len(sn.data) {
				snap := sn.data[idx]
				content := sn.renderSnapshotDetail(snap)
				sn.viewport.SetContent(content)
				sn.viewport.GotoTop()
				sn.showDetail = true
				return sn, nil
			}
		case "r":
			sn.loading = true
			sn.err = nil
			return sn, sn.loadSnapshots()
		case "q":
			return sn, func() tea.Msg { return popViewMsg{} }
		}
	}

	var cmd tea.Cmd
	sn.table, cmd = sn.table.Update(msg)
	return sn, cmd
}

func (sn *snapshots) View() string {
	if sn.loading {
		return contentStyle.Render("Loading snapshots...")
	}
	if sn.err != nil {
		return contentStyle.Render(errorStyle.Render("Error: " + sn.err.Error()))
	}

	if sn.showDetail {
		title := lipgloss.NewStyle().Bold(true).Foreground(gold).Render("Build Details")
		help := helpStyle.Render("scroll: up/down/j/k | esc: back to list")
		return contentStyle.Render(title + "\n\n" + sn.viewport.View() + "\n\n" + help)
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(gold).Render(
		fmt.Sprintf("Snapshots for %s (%d)", sn.charName, len(sn.data)),
	)

	help := helpStyle.Render("enter: view build | r: refresh | esc: back | q: back")

	return contentStyle.Render(title + "\n\n" + sn.table.View() + "\n\n" + help)
}

func (sn *snapshots) renderSnapshotDetail(snap models.POBSnapshot) string {
	var b strings.Builder

	// Always show link
	if snap.ExportString != "" {
		b.WriteString(dimStyle.Render("PoB Link: ") + snap.ExportString + "\n")
		b.WriteString(dimStyle.Render("Created:  ") + snap.CreatedAt.Format("2006-01-02 15:04:05") + "\n")
	}

	// If no pob_code, show raw export string
	if snap.PobCode == "" {
		b.WriteString("\n" + dimStyle.Render("No PoB data available for this snapshot."))
		return b.String()
	}

	pob, err := pobparser.Decode(snap.PobCode)
	if err != nil {
		b.WriteString("\n" + errorStyle.Render("Failed to decode PoB data: "+err.Error()))
		return b.String()
	}

	summary := pob.Summarize()

	// Header
	b.WriteString("\n")
	header := fmt.Sprintf("%s %s  Level %d", summary.Ascendancy, summary.Class, summary.Level)
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(gold).Render(header))
	b.WriteString("\n\n")

	// Stats in two columns
	leftCol := sn.renderDefenceStats(summary)
	rightCol := sn.renderOffenceStats(summary)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "    ", rightCol))
	b.WriteString("\n\n")

	// Equipment
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(gold).Render("Equipment"))
	b.WriteString("\n")
	for _, item := range summary.Items {
		rarity := dimStyle.Render(fmt.Sprintf("%-12s", item.Slot))
		name := sn.rarityStyle(item.Rarity).Render(item.Name)
		base := ""
		if item.BaseName != "" && item.BaseName != item.Name {
			base = " " + dimStyle.Render(item.BaseName)
		}
		b.WriteString(rarity + " " + name + base + "\n")
	}

	// Skills
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(gold).Render("Skills"))
	b.WriteString("\n")
	for _, sg := range summary.SkillGroups {
		slot := dimStyle.Render(fmt.Sprintf("%-14s", sg.Slot))
		var gems []string
		for i, gem := range sg.Gems {
			if i == 0 {
				gems = append(gems, lipgloss.NewStyle().Foreground(gold).Render(gem))
			} else {
				gems = append(gems, dimStyle.Render("+ ")+gem)
			}
		}
		b.WriteString(slot + " " + strings.Join(gems, " ") + "\n")
	}

	// Tree info
	if summary.NodeCount > 0 {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(gold).Render("Passive Tree"))
		b.WriteString(fmt.Sprintf("  %d nodes\n", summary.NodeCount))
		b.WriteString(fmt.Sprintf("  Str: %.0f  Dex: %.0f  Int: %.0f\n",
			summary.Stats["Str"], summary.Stats["Dex"], summary.Stats["Int"]))
	}

	return b.String()
}

func (sn *snapshots) renderDefenceStats(s pobparser.BuildSummary) string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(gold).Render("Defence") + "\n")
	b.WriteString(sn.statLine("Life", s.Stats["Life"], red))
	b.WriteString(sn.statLine("Mana", s.Stats["Mana"], steel))
	b.WriteString(sn.statLine("Energy Shield", s.Stats["EnergyShield"], lipgloss.Color("#00CED1")))
	b.WriteString(sn.statLine("Evasion", s.Stats["Evasion"], green))
	b.WriteString(sn.statLine("Armour", s.Stats["Armour"], gold))
	b.WriteString(fmt.Sprintf("  %-18s %s/%s/%s/%s\n",
		"Resistances",
		sn.resistStr(s.Stats["FireResist"]),
		sn.resistStr(s.Stats["ColdResist"]),
		sn.resistStr(s.Stats["LightningResist"]),
		sn.resistStr(s.Stats["ChaosResist"]),
	))
	b.WriteString(sn.statLine("Block", s.Stats["EffectiveBlockChance"], white))
	b.WriteString(sn.statLine("Suppression", s.Stats["EffectiveSpellSuppressionChance"], lipgloss.Color("#9370DB")))
	b.WriteString(sn.statLine("Effective HP", s.Stats["TotalEHP"], white))
	return b.String()
}

func (sn *snapshots) renderOffenceStats(s pobparser.BuildSummary) string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(gold).Render("Offence") + "\n")
	b.WriteString(sn.statLine("Combined DPS", s.Stats["CombinedDPS"], gold))
	b.WriteString(sn.statLine("Total DPS", s.Stats["TotalDPS"], white))
	b.WriteString(sn.statLine("Hit Chance", s.Stats["HitChance"], white))
	b.WriteString(fmt.Sprintf("  %-18s %.1f%%\n", "Crit Chance", s.Stats["CritChance"]))
	b.WriteString(fmt.Sprintf("  %-18s %.0f%%\n", "Crit Multi", s.Stats["CritMultiplier"]*100))
	b.WriteString(fmt.Sprintf("  %-18s %.2f/s\n", "Speed", s.Stats["Speed"]))
	b.WriteString(fmt.Sprintf("  %-18s %.0f%%\n", "Move Speed", (s.Stats["EffectiveMovementSpeedMod"]-1)*100))
	b.WriteString(fmt.Sprintf("  %-18s %.0f/%.0f/%.0f\n", "Charges",
		s.Stats["PowerChargesMax"],
		s.Stats["FrenzyChargesMax"],
		s.Stats["EnduranceChargesMax"],
	))
	return b.String()
}

func (sn *snapshots) statLine(label string, val float64, color lipgloss.Color) string {
	formatted := formatNumber(val)
	return fmt.Sprintf("  %-18s %s\n", label, lipgloss.NewStyle().Foreground(color).Render(formatted))
}

func (sn *snapshots) resistStr(val float64) string {
	color := red
	if val >= 75 {
		color = green
	}
	return lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%.0f", val))
}

func (sn *snapshots) rarityStyle(rarity string) lipgloss.Style {
	switch strings.ToUpper(rarity) {
	case "UNIQUE":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#AF6025"))
	case "RARE":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7"))
	case "MAGIC":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#88F"))
	default:
		return lipgloss.NewStyle().Foreground(white)
	}
}

func formatNumber(val float64) string {
	if val >= 1_000_000 {
		return fmt.Sprintf("%.1fM", val/1_000_000)
	}
	if val >= 1_000 {
		return fmt.Sprintf("%.1fk", val/1_000)
	}
	return fmt.Sprintf("%.0f", val)
}

var dimStyle = lipgloss.NewStyle().Foreground(dimGray)
