package lean

import (
	"fmt"
	"os"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize lean",
	Run: func(cmd *cobra.Command, args []string) {

		_, err := os.Stat(".lean")
		if err == nil {
			fmt.Println("⚡️ Already initialized.")
			return
		}

		err = core.Initialize()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Println("⚡️ lean has awakened.")
	},
}
