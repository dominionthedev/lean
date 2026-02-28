package lean

import (
	"fmt"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var getProfile string

var getCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Get the value of a variable",
	Long: `Get the value of a variable from the active profile (or a named profile via --profile).

Examples:
  lean get DEBUG
  lean get API_KEY --profile prod`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		key := args[0]

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		target := getProfile
		if target == "" {
			target = engine.State.Current
		}
		if target == "" {
			fmt.Println(ui.Fail("No active profile. Use --profile or run `lean apply <profile>` first."))
			return
		}

		profilePath := ".env." + target
		f, err := env.Parse(profilePath)
		if err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Could not read profile '%s': %s", target, err)))
			return
		}

		value, found := f.Get(key)
		if !found {
			fmt.Println(ui.Warn(fmt.Sprintf("'%s' is not set in profile '%s'.", key, target)))
			return
		}

		// Plain output so it's pipeline-friendly: lean get KEY | xargs ...
		fmt.Println(value)
	},
}

func init() {
	getCmd.Flags().StringVarP(&getProfile, "profile", "p", "", "Target profile (defaults to active)")
}