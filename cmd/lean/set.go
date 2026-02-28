package lean

import (
	"fmt"
	"os"
	"strings"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var setProfile string

var setCmd = &cobra.Command{
	Use:   "set KEY=VALUE",
	Short: "Set a variable in a profile",
	Long: `Set a variable in the active profile (or a named profile via --profile).
If the key already exists it is updated in place. If the profile is currently
active the change is also written to .env immediately.

Examples:
  lean set DEBUG=true
  lean set API_KEY=secret --profile prod`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		raw := args[0]
		idx := strings.Index(raw, "=")
		if idx < 0 {
			fmt.Println(ui.Fail("Expected KEY=VALUE — e.g. `lean set DEBUG=true`"))
			return
		}
		key := strings.TrimSpace(raw[:idx])
		value := strings.TrimSpace(raw[idx+1:])

		if key == "" {
			fmt.Println(ui.Fail("Key cannot be empty."))
			return
		}

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		// Resolve which profile to target
		target := setProfile
		if target == "" {
			target = engine.State.Current
		}
		if target == "" {
			fmt.Println(ui.Fail("No active profile. Use --profile or run `lean apply <profile>` first."))
			return
		}

		profilePath := ".env." + target

		// Parse (or create fresh if file doesn't exist yet)
		var f *env.File
		if _, statErr := os.Stat(profilePath); statErr == nil {
			f, err = env.Parse(profilePath)
			if err != nil {
				fmt.Println(ui.Fail("Could not read profile: " + err.Error()))
				return
			}
		} else {
			f = &env.File{Path: profilePath}
		}

		_, existed := f.Get(key)
		f.Set(key, value)

		if err := f.Write(profilePath); err != nil {
			fmt.Println(ui.Fail("Failed to write profile: " + err.Error()))
			return
		}

		// If this is the active profile, sync .env too
		if target == engine.State.Current {
			active, err := env.Parse(profilePath)
			if err == nil {
				_ = active.Write(".env")
			}
		}

		verb := "added to"
		if existed {
			verb = "updated in"
		}
		fmt.Printf("%s %s %s %s\n",
			ui.Bolt(),
			ui.Active.Render(key),
			ui.Faint(verb),
			ui.Bold.Render(target),
		)
	},
}

func init() {
	setCmd.Flags().StringVarP(&setProfile, "profile", "p", "", "Target profile (defaults to active)")
}