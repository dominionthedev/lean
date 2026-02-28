package lean

import (
	"fmt"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the active environment profile",
	Run: func(cmd *cobra.Command, args []string) {

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		if engine.State.Current == "" {
			fmt.Println(ui.Info("No active profile. Run `lean apply <profile>` to set one."))
			return
		}

		fmt.Printf("%s %s\n", ui.Bolt(), ui.Active.Render(engine.State.Current))
	},
}