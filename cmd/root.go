package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "backup_ollama",
	Short: "A command-line application for backing up and restoring models",
	Long:  `This application allows you to backup and restore models with specified names.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// init initializes the root command.
// Note: Subcommands are added in their respective files.
func init() {
	// Commands are registered in their own files
}
