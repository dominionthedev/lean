package lean

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var configEditGlobal bool

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit lean.toml interactively",
	Run: func(cmd *cobra.Command, args []string) {
		path := config.LocalPath()
		if configEditGlobal {
			var err error
			path, err = config.GlobalPath()
			if err != nil {
				fmt.Println(ui.Fail("Cannot determine global config path: " + err.Error()))
				return
			}
		}
		if !config.Exists(path) {
			if err := config.Save(path, config.Default()); err != nil {
				fmt.Println(ui.Fail("Failed to create config: " + err.Error()))
				return
			}
		}
		if err := openEditor(path); err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}
		fmt.Println(ui.Ok("Configuration saved."))
	},
}

func openEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			for _, candidate := range []string{"nano", "vim", "vi"} {
				if _, err := exec.LookPath(candidate); err == nil {
					editor = candidate
					break
				}
			}
		}
	}
	if editor == "" {
		return fmt.Errorf("could not find an editor; set EDITOR")
	}
	c := exec.Command(editor, path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func init() {
	configEditCmd.Flags().BoolVarP(&configEditGlobal, "global", "g", false, "Edit global config")
	configCmd.AddCommand(configEditCmd)
}
