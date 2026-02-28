package lean

import (
	"os"
	"fmt"
	"context"

	"github.com/spf13/cobra"
	"github.com/charmbracelet/fang"
)

var rootCmd = &cobra.Command{
	Use:   "lean",
	Short: "⚡ lean — Environment Orchestrator",
	Long:  "lean manages and protects your environment profiles safely.",
}

func Execute(version string) {
	// Set the version for the root command
	rootCmd.Version = version
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(currentCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(versionCmd)
}