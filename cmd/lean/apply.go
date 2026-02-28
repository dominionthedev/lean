package lean

import (
	"fmt"
	"os"

	"github.com/dominionthedev/lean/internal/backup"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [profile]",
	Short: "Apply an environment profile → .env",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		profile := args[0]

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		src := ".env." + profile

		if _, err := os.Stat(src); err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Profile '%s' not found. Is there a .env.%s file?", profile, profile)))
			return
		}

		// Backup current .env before overwriting
		if err := backup.Snapshot(engine.State.Current); err != nil {
			fmt.Println(ui.Warn("Could not snapshot current .env: " + err.Error()))
			// non-fatal — continue
		}

		// Read source
		data, err := os.ReadFile(src)
		if err != nil {
			fmt.Println(ui.Fail("Failed to read profile: " + err.Error()))
			return
		}

		// Atomic write to .env
		tmp := ".env.tmp"
		if err := os.WriteFile(tmp, data, 0644); err != nil {
			fmt.Println(ui.Fail("Failed to write .env: " + err.Error()))
			return
		}
		if err := os.Rename(tmp, ".env"); err != nil {
			fmt.Println(ui.Fail("Failed to finalize .env: " + err.Error()))
			return
		}

		// Register if not already known
		if !engine.ProfileExists(profile) {
			engine.AddProfile(profile)
		}

		if err := engine.SetCurrent(profile); err != nil {
			fmt.Println(ui.Warn("State not saved: " + err.Error()))
		}

		prev := engine.State.Current
		if prev != "" && prev != profile {
			fmt.Printf("%s Switched %s → %s\n",
				ui.Bolt(),
				ui.Faint(prev),
				ui.Active.Render(profile),
			)
		} else {
			fmt.Println(ui.Ok(fmt.Sprintf("Now on '%s'.", profile)))
		}
	},
}