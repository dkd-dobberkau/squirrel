package output

import (
	"fmt"
	"strings"
	"time"

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

	if len(data.Acknowledged) > 0 {
		b.WriteString(sectionStyle.Render(fmt.Sprintf("\nAcknowledged (%d)", len(data.Acknowledged))))
		b.WriteString("\n")
		for _, p := range data.Acknowledged {
			b.WriteString(dimStyle.Render("  ✓ "))
			b.WriteString(dimStyle.Render(formatProjectAck(p)))
			b.WriteString("\n")
		}
	}

	if len(data.OpenWork) == 0 && len(data.RecentActivity) == 0 && len(data.Sleeping) == 0 && len(data.Acknowledged) == 0 {
		b.WriteString(dimStyle.Render("  Keine Projekte im gewaehlten Zeitraum gefunden."))
		b.WriteString("\n")
	}

	return b.String()
}

// RenderProjectDetail renders a detailed view for a single project.
func RenderProjectDetail(p claude.ProjectInfo, prompts []claude.HistoryEntry) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Squirrel - Projekt-Detail"))
	b.WriteString("\n\n")

	// Project metadata
	b.WriteString(sectionStyle.Render("Projekt"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Pfad:       %s\n", p.Path))
	b.WriteString(fmt.Sprintf("  Name:       %s\n", p.ShortName))

	branch := p.GitBranch
	if branch == "" {
		branch = p.LatestBranch
	}
	if branch != "" {
		b.WriteString(fmt.Sprintf("  Branch:     %s\n", branch))
	}

	if p.GitDirty {
		b.WriteString(fmt.Sprintf("  Git-Status: %s\n", warnStyle.Render(fmt.Sprintf("%d uncommitted", p.UncommittedFiles))))
	} else if p.UncommittedFiles == 0 && p.GitBranch != "" {
		b.WriteString(fmt.Sprintf("  Git-Status: %s\n", okStyle.Render("clean")))
	}

	b.WriteString(fmt.Sprintf("  Score:      %.1f\n", p.Score))
	b.WriteString(fmt.Sprintf("  Prompts:    %d\n", p.PromptCount))

	// Activity period
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("Aktivitaet"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Erste:      %s\n", p.FirstActivity.Format("02.01.2006 15:04")))
	b.WriteString(fmt.Sprintf("  Letzte:     %s\n", p.LastActivity.Format("02.01.2006 15:04")))
	if p.DaysSinceActive > 0 {
		b.WriteString(fmt.Sprintf("  Inaktiv:    %d Tage\n", p.DaysSinceActive))
	}

	// Sessions
	if len(p.Sessions) > 0 {
		b.WriteString("\n")
		b.WriteString(sectionStyle.Render(fmt.Sprintf("Sessions (%d)", len(p.Sessions))))
		b.WriteString("\n")
		for _, s := range p.Sessions {
			summary := s.Summary
			if summary == "" {
				summary = s.FirstPrompt
			}
			summary = truncate(summary, 60)
			date := s.Modified
			if len(date) > 10 {
				date = date[:10]
			}
			b.WriteString(fmt.Sprintf("  %s  %s  %d msgs\n",
				dimStyle.Render(date),
				summary,
				s.MsgCount,
			))
		}
	}

	// Recent prompts from history
	if len(prompts) > 0 {
		b.WriteString("\n")
		b.WriteString(sectionStyle.Render(fmt.Sprintf("Letzte Prompts (%d)", len(prompts))))
		b.WriteString("\n")
		for _, pr := range prompts {
			ts := time.UnixMilli(pr.Timestamp).Format("02.01. 15:04")
			display := truncate(pr.Display, 70)
			b.WriteString(fmt.Sprintf("  %s  %s\n", dimStyle.Render(ts), display))
		}
	}

	// TODOs from deep mode
	if len(p.Todos) > 0 {
		b.WriteString("\n")
		b.WriteString(sectionStyle.Render(fmt.Sprintf("TODOs (%d)", len(p.Todos))))
		b.WriteString("\n")
		for _, todo := range p.Todos {
			b.WriteString(fmt.Sprintf("  %s %s\n",
				warnStyle.Render("["+todo.Source+"]"),
				todo.Text,
			))
		}
	}

	return b.String()
}

func formatProjectAck(p claude.ProjectInfo) string {
	date := p.LastActivity.Format("02.01.")
	name := fmt.Sprintf("%-22s", truncate(p.ShortName, 22))
	prompts := fmt.Sprintf("%4d prompts", p.PromptCount)
	return strings.Join([]string{name, date, prompts}, " | ")
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
