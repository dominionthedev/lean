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
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run lean init first."))
			return
		}
		if err := engine.ScanDisk(); err != nil {
			fmt.Println(ui.Fail("Could not scan profiles: " + err.Error()))
			return
		}

		target := deleteProfile
		if target == "" {
			target = engine.State.Current
		}
		if target == "" {
			fmt.Println(ui.Fail("No active profile. Use --profile or apply a profile first."))
			return
		}

		path := env.ProfilePath(target)
		if _, err := os.Stat(path); err != nil {
			if !confirmStaleProfile(engine, target) {
				return
			}
			return
		}

		f, err := env.Parse(path)
		if err != nil {
			fmt.Println(ui.Fail("Could not read profile: " + err.Error()))
			return
		}
		if !f.Delete(key) {
			fmt.Println(ui.Warn(fmt.Sprintf("'%s' was not found in profile '%s'.", key, target)))
			return
		}
		if err := f.Write(path); err != nil {
			fmt.Println(ui.Fail("Failed to write profile: " + err.Error()))
			return
		}
		if target == engine.State.Current {
			if active, err := env.Parse(path); err == nil {
				_ = active.Write(".env")
			}
		}
		fmt.Printf("%s %s removed from %s\n", ui.Bolt(), ui.Active.Render(key), ui.Bold.Render(target))
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteProfile, "profile", "p", "", "Target profile")
}
