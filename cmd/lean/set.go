package lean

import (
	"fmt"
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
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		raw := args[0]
		idx := strings.Index(raw, "=")
		if idx < 0 {
			fmt.Println(ui.Fail("Expected KEY=VALUE"))
			return
		}
		key, value := strings.TrimSpace(raw[:idx]), strings.TrimSpace(raw[idx+1:])
		if key == "" {
			fmt.Println(ui.Fail("Key cannot be empty."))
			return
		}

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run lean init first."))
			return
		}
		if err := engine.ScanDisk(); err != nil {
			fmt.Println(ui.Fail("Could not scan profiles: " + err.Error()))
			return
		}

		target := setProfile
		if target == "" {
			target = engine.State.Current
		}
		if target == "" {
			fmt.Println(ui.Fail("No active profile. Use --profile or apply a profile first."))
			return
		}

		path := env.ProfilePath(target)
		f, err := env.Parse(path)
		if err != nil {
			fmt.Println(ui.Fail("Could not read profile: " + err.Error()))
			return
		}

		_, existed := f.Get(key)
		f.Set(key, value)
		if err := f.Write(path); err != nil {
			fmt.Println(ui.Fail("Failed to write profile: " + err.Error()))
			return
		}

		if target == engine.State.Current {
			if resolved, err := env.Resolve(target); err == nil {
				_ = resolved.Write(".env")
			} else {
				fmt.Println(ui.Warn("Could not sync .env: " + err.Error()))
			}
		}

		verb := "added to"
		if existed {
			verb = "updated in"
		}
		fmt.Printf("%s %s %s %s\n", ui.Bolt(), ui.Active.Render(key), ui.Faint(verb), ui.Bold.Render(target))
	},
}

func init() {
	setCmd.Flags().StringVarP(&setProfile, "profile", "p", "", "Target profile")
}
