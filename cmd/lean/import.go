package lean

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/dominionthedev/lean/internal/backup"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/dominionthedev/lean/internal/workspace"
	"github.com/spf13/cobra"
)

var importYes bool

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Apply the profile remembered for this directory",
	Long: `Workspace awareness: lean remembers which profile you last applied
in each project directory.

  lean apply api-dev          # remembers cwd → api-dev
  cd ~/Projects/api
  lean import                 # suggests / applies api-dev

Use --yes to skip the confirmation prompt.`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println(ui.Fail("Cannot determine current directory: " + err.Error()))
			return
		}

		profile, ok := workspace.Lookup(cwd)
		if !ok {
			fmt.Println(ui.Info("No profile remembered for this directory."))
			fmt.Println(ui.Faint("Apply a profile with `lean apply <name>` and lean will remember it."))
			return
		}

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		if engine.State.Current == profile {
			fmt.Println(ui.Ok(fmt.Sprintf("Already on '%s'.", profile)))
			return
		}

		// Confirm unless --yes
		if !importYes {
			var confirm bool
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Apply remembered profile '%s'?", profile)).
						Affirmative("Yes").
						Negative("No").
						Value(&confirm),
				),
			)
			if err := form.Run(); err != nil {
				fmt.Println(ui.Fail("Interrupted."))
				return
			}
			if !confirm {
				fmt.Println(ui.Faint("Cancelled."))
				return
			}
		}

		// Reuse apply logic
		src := env.ProfilePath(profile)
		if _, err := os.Stat(src); err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Remembered profile '%s' no longer exists on disk.", profile)))
			_ = workspace.Forget(cwd)
			return
		}

		resolved, err := env.Resolve(profile)
		if err != nil {
			fmt.Println(ui.Fail("Failed to resolve profile: " + err.Error()))
			return
		}

		prev := engine.State.Current
		if err := backup.Snapshot(prev); err != nil {
			fmt.Println(ui.Warn("Could not snapshot current .env: " + err.Error()))
		}

		if err := resolved.Write(".env"); err != nil {
			fmt.Println(ui.Fail("Failed to write .env: " + err.Error()))
			return
		}

		if !engine.ProfileExists(profile) {
			_ = engine.AddProfile(profile)
		}
		_ = engine.SetCurrent(profile)

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

func init() {
	importCmd.Flags().BoolVarP(&importYes, "yes", "y", false, "Skip confirmation prompt")
}
