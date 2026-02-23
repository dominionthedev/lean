package lean

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lean",
	Short: "⚡️ lean — Environment Orchestrator",
	Long:  "lean manages and protects your environment profiles safely.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(currentCmd)
}
