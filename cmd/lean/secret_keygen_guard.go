package lean

import (
	"fmt"

	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	original := secretKeygenCmd.Run
	secretKeygenCmd.Run = func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("")
		if err != nil {
			fmt.Println(ui.Fail("Failed to load config: " + err.Error()))
			return
		}
		if cfg.Secrets.Backend != "local" {
			fmt.Println(ui.Fail("The Lean master key is only used by the local secrets backend."))
			return
		}
		original(cmd, args)
	}
}
