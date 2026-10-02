package lean

import (
	"fmt"
	"os"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the active environment profile",
	Run: func(cmd *cobra.Command, args []string) {
		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run lean init first."))
			return
		}
		if err := engine.ScanDisk(); err != nil {
			fmt.Println(ui.Fail("Could not scan profiles: " + err.Error()))
			return
		}
		if engine.State.Current == "" {
			fmt.Println(ui.Info("No active profile. Run lean apply <profile> to set one."))
			return
		}

		profile := engine.State.Current
		if _, err := os.Stat(env.ProfilePath(profile)); err != nil {
			confirmStaleProfile(engine, profile)
			return
		}
		fmt.Printf("%s %s\n", ui.Bolt(), ui.Active.Render(profile))
	},
}
