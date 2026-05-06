package lean

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [profile]",
	Short: "Open a profile in your editor",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

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

		path := ".env." + profile
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Println(ui.Fail(fmt.Sprintf("Profile file %s does not exist.", path)))
			return
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			if runtime.GOOS == "windows" {
				editor = "notepad"
			} else {
				// try some common editors
				editors := []string{"nano", "vim", "vi"}
				for _, e := range editors {
					if _, err := exec.LookPath(e); err == nil {
						editor = e
						break
					}
				}
			}
		}

		if editor == "" {
			fmt.Println(ui.Fail("Could not find an editor. Please set your $EDITOR environment variable."))
			return
		}

		c := exec.Command(editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		if err := c.Run(); err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Failed to open editor: %s", err)))
		}
	},
}
