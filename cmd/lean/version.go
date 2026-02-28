package lean

import (
	"fmt"

	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

// Version is injected at build time via ldflags.
// e.g. -ldflags="-X github.com/dominionthedev/lean/cmd/lean.Version=v1.0.0"
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print lean version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s lean %s\n", ui.Bolt(), ui.Active.Render(version))
	},
}