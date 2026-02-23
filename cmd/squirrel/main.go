package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/dkd-dobberkau/squirrel/internal/analyzer"
	"github.com/dkd-dobberkau/squirrel/internal/claude"
	"github.com/dkd-dobberkau/squirrel/internal/output"
)

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
	Use:   "squirrel",
	Short: "Find your forgotten Claude Code projects",
	Long:  "Squirrel helps you find projects you started in Claude Code but forgot about.",
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
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
