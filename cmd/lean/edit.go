package lean

import (
	"fmt"
	"os"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [profile]",
	Short: "Open a profile in your configured editor",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run lean init first."))
			return
		}
		profile := engine.State.Current
		if len(args) > 0 {
			profile = args[0]
		}
		if profile == "" {
			fmt.Println(ui.Fail("No profile specified and no active profile set."))
			return
		}

		path := env.ProfilePath(profile)
		if _, err := os.Stat(path); err != nil {
			confirmStaleProfile(engine, profile)
			return
		}
		if err := openEditor(path); err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Failed to open editor: %s", err)))
			return
		}

		if profile == engine.State.Current {
			if f, err := env.Parse(path); err == nil {
				_ = f.Write(".env")
			}
		}
	},
}
