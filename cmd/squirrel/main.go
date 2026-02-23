package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	depth   string
	days    int
	jsonOut bool
)

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
		fmt.Println("squirrel status - not yet implemented")
		return nil
	},
}

var stashCmd = &cobra.Command{
	Use:   "stash",
	Short: "Show only projects with uncommitted changes or feature branches",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("squirrel stash - not yet implemented")
		return nil
	},
}

var timelineCmd = &cobra.Command{
	Use:   "timeline",
	Short: "Show chronological activity timeline",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("squirrel timeline - not yet implemented")
		return nil
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
