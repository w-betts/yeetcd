package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yeetcd",
	Short: "yeetcd - Continuous deployment with container-based pipeline execution",
	Long: `yeetcd is a continuous deployment solution with container-based pipeline execution.

It supports three execution modes:
  1. CLI with Docker executor for local development
  2. CLI with Mock executor for testing
  3. Kubernetes Operator for production deployment

Use 'yeetcd run' to execute a pipeline from source code,
or 'yeetcd list' to see available pipelines.`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("command execution failed", "error", err)
		os.Exit(1)
	}
}

func init() {
	// Initialize structured logging with slog
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Add commands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listCmd)
}

func main() {
	Execute()
}
