package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/dkd-dobberkau/squirrel/internal/analyzer"
	"github.com/dkd-dobberkau/squirrel/internal/claude"
	"github.com/dkd-dobberkau/squirrel/internal/output"
)

var version = "dev"

var (
	depth   string
	days    int
	jsonOut bool
)

func claudeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

func runAnalysis() (analyzer.CategorizedProjects, error) {
	cDir := claudeDir()
	histPath := filepath.Join(cDir, "history.jsonl")

	entries, err := claude.ParseHistory(histPath)
	if err != nil {
		return analyzer.CategorizedProjects{}, fmt.Errorf("reading history: %w", err)
	}

	projects := claude.AggregateByProject(entries, days)

	claude.EnrichWithSessions(projects, filepath.Join(cDir, "projects"))

	if depth == "medium" || depth == "deep" {
		analyzer.EnrichWithGit(projects)
	}

	return analyzer.Categorize(projects), nil
}

func renderOutput(data analyzer.CategorizedProjects) error {
	if jsonOut {
		s, err := output.RenderJSON(data)
		if err != nil {
			return err
		}
		fmt.Println(s)
	} else {
		fmt.Print(output.RenderTerminal(data))
	}
	return nil
}

var rootCmd = &cobra.Command{
	Use:     "squirrel",
	Short:   "Find your forgotten Claude Code projects",
	Long:    "Squirrel helps you find projects you started in Claude Code but forgot about.",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusCmd.RunE(cmd, args)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project overview with open work, activity, and sleeping projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		resolveDepthShortcuts(cmd)
		data, err := runAnalysis()
		if err != nil {
			return err
		}
		return renderOutput(data)
	},
}

var stashCmd = &cobra.Command{
	Use:   "stash",
	Short: "Show only projects with uncommitted changes or feature branches",
	RunE: func(cmd *cobra.Command, args []string) error {
		if depth == "quick" {
			depth = "medium"
		}
		data, err := runAnalysis()
		if err != nil {
			return err
		}
		data.RecentActivity = nil
		data.Sleeping = nil
		return renderOutput(data)
	},
}

var timelineCmd = &cobra.Command{
	Use:   "timeline",
	Short: "Show chronological activity timeline",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := runAnalysis()
		if err != nil {
			return err
		}
		all := append(data.OpenWork, data.RecentActivity...)
		all = append(all, data.Sleeping...)
		data.OpenWork = nil
		data.RecentActivity = all
		data.Sleeping = nil
		return renderOutput(data)
	},
}

var projectCmd = &cobra.Command{
	Use:   "project [path]",
	Short: "Show detailed view for a specific project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("squirrel project %s - not yet implemented\n", args[0])
		return nil
	},
}

var installSkillCmd = &cobra.Command{
	Use:   "install-skill",
	Short: "Install the /squirrel Claude Code skill",
	RunE: func(cmd *cobra.Command, args []string) error {
		skillDir := filepath.Join(claudeDir(), "skills", "squirrel")
		skillPath := filepath.Join(skillDir, "SKILL.md")

		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return fmt.Errorf("creating skill directory: %w", err)
		}

		if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
			return fmt.Errorf("writing skill file: %w", err)
		}

		fmt.Printf("Skill installed to %s\n", skillPath)
		fmt.Println("You can now use /squirrel in any Claude Code session.")
		return nil
	},
}

const skillContent = `---
name: squirrel
description: Find forgotten Claude Code projects - shows open work, activity timeline, and recommendations
user_invocable: true
---

# Squirrel - Find Your Forgotten Projects

Run the squirrel CLI to analyze Claude Code history and present results.

## Steps

1. Run the squirrel command with JSON output and deep analysis:

   ` + "```bash" + `
   squirrel status --json --deep --days 14
   ` + "```" + `

2. Parse the JSON output and present the results in a structured way:

   **For each category (openWork, recentActivity, sleeping):**
   - Show the project name, last activity date, prompt count
   - For open work: highlight uncommitted files and feature branches
   - For sleeping projects: show days since last activity

3. After presenting the overview, provide:
   - **Top 3 Empfehlungen:** Which projects the user should focus on (highest score)
   - **Quick Summary:** "Du hast X offene Baustellen, Y aktive Projekte und Z schlafende Projekte"

4. Ask the user: "An welchem Projekt moechtest du weiterarbeiten?"

5. When the user picks a project:
   - Show the last session summary for that project
   - Show the last few prompts from history
   - Suggest: "Soll ich in das Projektverzeichnis wechseln?"

## Notes

- The --deep flag takes longer but provides richer context
- Use --days 30 for a broader view
`

var nutsCmd = &cobra.Command{
	Use:    "nuts",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := runAnalysis()
		if err != nil {
			return err
		}
		nutCount := len(data.OpenWork)
		if nutCount == 0 {
			fmt.Println("\n🐿️ No nuts buried — your filesystem is clean!")
			return nil
		}

		fmt.Println()
		for i := 1; i <= nutCount; i++ {
			nuts := strings.Repeat("🥜 ", i)
			fmt.Printf("\r\033[K%s🐿️💨", nuts)
			time.Sleep(80 * time.Millisecond)
			fmt.Println()
		}
		fmt.Printf("\n🐿️  %d nuts buried across your filesystem!\n\n", nutCount)
		return nil
	},
}

func resolveDepthShortcuts(cmd *cobra.Command) {
	if q, _ := cmd.Flags().GetBool("quick"); q {
		depth = "quick"
	}
	if m, _ := cmd.Flags().GetBool("medium"); m {
		depth = "medium"
	}
	if d, _ := cmd.Flags().GetBool("deep"); d {
		depth = "deep"
	}
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&depth, "depth", "medium", "Analysis depth: quick, medium, or deep")
	pf.BoolVar(&jsonOut, "json", false, "Output as JSON (for skill integration)")
	pf.IntVar(&days, "days", 14, "Number of days to look back")

	statusCmd.Flags().Bool("quick", false, "Shortcut for --depth=quick")
	statusCmd.Flags().Bool("medium", false, "Shortcut for --depth=medium")
	statusCmd.Flags().Bool("deep", false, "Shortcut for --depth=deep")

	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(stashCmd)
	rootCmd.AddCommand(timelineCmd)
	rootCmd.AddCommand(projectCmd)
	rootCmd.AddCommand(installSkillCmd)
	rootCmd.AddCommand(nutsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
