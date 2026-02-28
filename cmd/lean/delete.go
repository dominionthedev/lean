package lean

import (
	"fmt"
	"os"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var deleteProfile string

var deleteCmd = &cobra.Command{
	Use:     "delete KEY",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a variable from a profile",
	Long: `Delete a variable from the active profile (or a named profile via --profile).
If the profile is currently active the change is also reflected in .env.

Examples:
  lean delete DEBUG
  lean delete API_KEY --profile staging`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		key := args[0]

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		target := deleteProfile
		if target == "" {
			target = engine.State.Current
		}
		if target == "" {
			fmt.Println(ui.Fail("No active profile. Use --profile or run `lean apply <profile>` first."))
			return
		}

		profilePath := ".env." + target

		if _, statErr := os.Stat(profilePath); statErr != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Profile '%s' not found on disk.", target)))
			return
		}

		f, err := env.Parse(profilePath)
		if err != nil {
			fmt.Println(ui.Fail("Could not read profile: " + err.Error()))
			return
		}

		if !f.Delete(key) {
			fmt.Println(ui.Warn(fmt.Sprintf("'%s' was not found in profile '%s'.", key, target)))
			return
		}

		if err := f.Write(profilePath); err != nil {
			fmt.Println(ui.Fail("Failed to write profile: " + err.Error()))
			return
		}

		// Sync .env if this is the active profile
		if target == engine.State.Current {
			active, err := env.Parse(profilePath)
			if err == nil {
				_ = active.Write(".env")
			}
		}

		fmt.Printf("%s %s removed from %s\n",
			ui.Bolt(),
			ui.Active.Render(key),
			ui.Bold.Render(target),
		)
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteProfile, "profile", "p", "", "Target profile (defaults to active)")
}