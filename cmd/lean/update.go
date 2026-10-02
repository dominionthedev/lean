package lean

import (
	"fmt"
	"os"

	"github.com/dominionthedev/lean/internal/backup"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Sync the active profile to .env",
	Long: `Resolve the active profile, including inheritance, and write the result to .env.
A snapshot of the current .env is taken before it is replaced.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		if err := engine.ScanDisk(); err != nil {
			fmt.Println(ui.Fail("Could not scan profiles: " + err.Error()))
			return
		}
		profile := engine.State.Current
		if profile == "" {
			fmt.Println(ui.Fail("No active profile. Run `lean apply <profile>` first."))
			return
		}

		src := env.ProfilePath(profile)
		if _, err := os.Stat(src); err != nil {
			confirmStaleProfile(engine, profile)
			return
		}

		resolved, err := env.Resolve(profile)
		if err != nil {
			fmt.Println(ui.Fail("Failed to resolve profile: " + err.Error()))
			return
		}

		if err := backup.Snapshot(profile); err != nil {
			fmt.Println(ui.Warn("Could not snapshot current .env: " + err.Error()))
		}

		if err := resolved.Write(".env"); err != nil {
			fmt.Println(ui.Fail("Failed to update .env: " + err.Error()))
			return
		}

		fmt.Println(ui.Ok(fmt.Sprintf("Updated .env from '%s'.", profile)))
	},
}
