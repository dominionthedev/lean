package lean

import (
	"fmt"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environment profiles",
	Run: func(cmd *cobra.Command, args []string) {

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println("⚡️ lean is not initialized.")
			fmt.Println("Run: lean init")
			return
		}

		if len(engine.State.Profiles) == 0 {
			fmt.Println("⚡️ No profiles found.")
			return
		}

		fmt.Println("⚡️ Profiles:")

		for _, profile := range engine.State.Profiles {

			if profile == engine.State.Current {
				fmt.Printf("• %s (current)\n", profile)
			} else {
				fmt.Printf("• %s\n", profile)
			}

		}
	},
}
