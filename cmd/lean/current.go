package lean

import (
	"fmt"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current environment profile",
	Run: func(cmd *cobra.Command, args []string) {

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println("⚡️ lean is not initialized.")
			return
		}

		if engine.State.Current == "" {
			fmt.Println("⚡️ No active profile.")
			return
		}

		fmt.Println(engine.State.Current)
	},
}
