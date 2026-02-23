package lean

import (
	"fmt"
	"io"
	"os"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [profile]",
	Short: "Apply environment profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		profile := args[0]

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println("Not initialized.")
			return
		}

		src := ".env." + profile
		dst := ".env"

		if _, err := os.Stat(src); err != nil {
			fmt.Println("Profile does not exist.")
			return
		}

		err = copyFile(src, dst)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		engine.SetCurrent(profile)

		fmt.Printf("⚡️ Switched to '%s'\n", profile)
	},
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
