package lean

import (
	"fmt"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environment profiles",
	Run: func(cmd *cobra.Command, args []string) {

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		// Pick up any .env.* files that exist on disk but aren't registered
		engine.ScanDisk()

		if len(engine.State.Profiles) == 0 {
			fmt.Println(ui.Info("No profiles yet. Run `lean create` to make one."))
			return
		}

		fmt.Println(ui.Bolt() + " " + ui.Bold.Render("Profiles"))
		fmt.Println()

		for _, profile := range engine.State.Profiles {
			if profile == engine.State.Current {
				fmt.Printf("  %s %s\n",
					ui.Success.Render("▶"),
					ui.Active.Render(profile)+" "+ui.Faint("(active)"),
				)
			} else {
				fmt.Printf("  %s %s\n",
					ui.Muted.Render("·"),
					profile,
				)
			}
		}

		fmt.Println()
	},
}