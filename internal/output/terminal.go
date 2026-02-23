package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dkd-dobberkau/squirrel/internal/analyzer"
	"github.com/dkd-dobberkau/squirrel/internal/claude"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF8C00")).
		MarginBottom(1)

	sectionStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#87CEEB")).
		MarginTop(1)

	warnStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700"))

	okStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#98FB98"))

	sleepStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#808080"))

	dimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))
)

// RenderTerminal prints the categorized projects as styled terminal output.
func RenderTerminal(data analyzer.CategorizedProjects) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Squirrel - Deine vergessenen Nuesse"))
	b.WriteString("\n\n")

	if len(data.OpenWork) > 0 {
		b.WriteString(sectionStyle.Render(fmt.Sprintf("Offene Baustellen (%d)", len(data.OpenWork))))
		b.WriteString("\n")
		for _, p := range data.OpenWork {
			b.WriteString(warnStyle.Render("  ! "))
			b.WriteString(formatProject(p))
			b.WriteString("\n")
		}
	}

	if len(data.RecentActivity) > 0 {
		b.WriteString(sectionStyle.Render(fmt.Sprintf("\nLetzte Aktivitaet (%d)", len(data.RecentActivity))))
		b.WriteString("\n")
		for _, p := range data.RecentActivity {
			b.WriteString(okStyle.Render("  + "))
			b.WriteString(formatProject(p))
			b.WriteString("\n")
		}
	}

	if len(data.Sleeping) > 0 {
		b.WriteString(sectionStyle.Render(fmt.Sprintf("\nSchlafende Projekte (%d)", len(data.Sleeping))))
		b.WriteString("\n")
		for _, p := range data.Sleeping {
			b.WriteString(sleepStyle.Render("  ~ "))
			b.WriteString(formatProject(p))
			b.WriteString("\n")
		}
	}

	if len(data.OpenWork) == 0 && len(data.RecentActivity) == 0 && len(data.Sleeping) == 0 {
		b.WriteString(dimStyle.Render("  Keine Projekte im gewaehlten Zeitraum gefunden."))
		b.WriteString("\n")
	}

	return b.String()
}

func formatProject(p claude.ProjectInfo) string {
	date := p.LastActivity.Format("02.01.")
	name := fmt.Sprintf("%-22s", truncate(p.ShortName, 22))
	prompts := fmt.Sprintf("%4d prompts", p.PromptCount)

	details := []string{name, date, prompts}

	if p.UncommittedFiles > 0 {
		details = append(details, warnStyle.Render(fmt.Sprintf("%d uncommitted", p.UncommittedFiles)))
	}

	branch := p.GitBranch
	if branch == "" {
		branch = p.LatestBranch
	}
	if branch != "" && branch != "main" && branch != "master" {
		details = append(details, dimStyle.Render("branch: "+branch))
	}

	if p.DaysSinceActive > 0 {
		details = append(details, dimStyle.Render(fmt.Sprintf("%d Tage inaktiv", p.DaysSinceActive)))
	}

	return strings.Join(details, " | ")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "~"
}
