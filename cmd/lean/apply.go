package lean

import (
	"fmt"
	"os"

	"github.com/dominionthedev/lean/internal/backup"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/dominionthedev/lean/internal/workspace"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [profile]",
	Short: "Apply an environment profile → .env",
	Long: `Resolve a profile (including inheritance) and write it to .env.
A backup of the current .env is taken before overwriting.

Inheritance is declared in the profile file:
  # lean:extends base
  # @extends base
  LEAN_EXTENDS=base`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		profile := args[0]

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		_ = engine.ScanDisk()

		src := env.ProfilePath(profile)
		if _, err := os.Stat(src); err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Profile '%s' not found. Is there a %s file?", profile, src)))
			return
		}

		// Resolve inheritance chain
		resolved, err := env.Resolve(profile)
		if err != nil {
			fmt.Println(ui.Fail("Failed to resolve profile: " + err.Error()))
			return
		}

		prev := engine.State.Current

		// Backup current .env before overwriting
		if err := backup.Snapshot(prev); err != nil {
			fmt.Println(ui.Warn("Could not snapshot current .env: " + err.Error()))
			// non-fatal — continue
		}

		// Atomic write of the resolved (merged) content
		if err := resolved.Write(".env"); err != nil {
			fmt.Println(ui.Fail("Failed to write .env: " + err.Error()))
			return
		}

		// Register if not already known
		if !engine.ProfileExists(profile) {
			engine.AddProfile(profile)
		}

		if err := engine.SetCurrent(profile); err != nil {
			fmt.Println(ui.Warn("State not saved: " + err.Error()))
		}

		parent := ""
		if raw, parseErr := env.Parse(src); parseErr == nil {
			parent = raw.Extends()
		}

		// Remember this workspace → profile mapping
		if cwd, err := os.Getwd(); err == nil {
			_ = workspace.Remember(cwd, profile)
		}

		if prev != "" && prev != profile {
			msg := fmt.Sprintf("%s Switched %s → %s",
				ui.Bolt(),
				ui.Faint(prev),
				ui.Active.Render(profile),
			)
			if parent != "" {
				msg += ui.Faint(fmt.Sprintf(" (extends %s)", parent))
			}
			fmt.Println(msg)
		} else {
			suffix := ""
			if parent != "" {
				suffix = ui.Faint(fmt.Sprintf(" (extends %s)", parent))
			}
			fmt.Println(ui.Ok(fmt.Sprintf("Now on '%s'.%s", profile, suffix)))
		}
	},
}
