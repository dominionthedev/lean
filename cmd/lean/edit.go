package lean

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/dominionthedev/lean/internal/backup"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use: "edit [profile]", Short: "Open a profile in your editor", Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run lean init first."))
			return
		}
		_ = engine.ScanDisk()
		profile := ""
		if len(args) > 0 {
			profile = args[0]
		} else {
			profile = engine.State.Current
		}
		if profile == "" {
			fmt.Println(ui.Fail("No profile specified and no active profile set."))
			return
		}
		path := env.ProfilePath(profile)
		if _, err := os.Stat(path); err != nil {
			fmt.Println(ui.Warn(fmt.Sprintf("Profile '%s' no longer exists. Removing it from Lean.", profile)))
			_ = engine.DeleteProfile(profile)
			return
		}
		editor := os.Getenv("EDITOR")
		if editor == "" {
			if runtime.GOOS == "windows" {
				editor = "notepad"
			} else {
				for _, e := range []string{"nano", "vim", "vi"} {
					if _, err := exec.LookPath(e); err == nil {
						editor = e
						break
					}
				}
			}
		}
		if editor == "" {
			fmt.Println(ui.Fail("Could not find an editor. Set EDITOR."))
			return
		}
		if err := backup.SnapshotProfile(profile, path); err != nil {
			fmt.Println(ui.Warn("Could not back up profile: " + err.Error()))
		}
		c := exec.Command(editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
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
