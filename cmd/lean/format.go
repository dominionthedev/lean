package lean

import (
	"fmt"
	"strings"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var formatType string

var formatCmd = &cobra.Command{
	Use:   "format [profile]",
	Short: "Convert a profile to different formats (json, yaml, toml, env)",
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
		f, err := env.Parse(path)
		if err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Failed to read profile: %s", err)))
			return
		}

		var output string
		var formatErr error

		switch strings.ToLower(formatType) {
		case "json":
			output, formatErr = f.ToJSON()
		case "yaml", "yml":
			output, formatErr = f.ToYAML()
		case "toml":
			output, formatErr = f.ToTOML()
		case "env":
			output = f.ToString()
		default:
			fmt.Println(ui.Fail(fmt.Sprintf("Unsupported format: %s", formatType)))
			return
		}

		if formatErr != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Failed to format: %s", formatErr)))
			return
		}

		fmt.Print(output)
	},
}

func init() {
	formatCmd.Flags().StringVarP(&formatType, "type", "t", "json", "Output format (json, yaml, toml, env)")
}
