package lean

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var (
	createName        string
	createFrom        string
	createStrip       bool
	createInteractive bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new environment profile",
	Run: func(cmd *cobra.Command, args []string) {

		// Accept name as positional arg too
		if createName == "" && len(args) > 0 {
			createName = args[0]
		}

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		// Interactive mode — ask for name if not provided
		if createInteractive || createName == "" {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Profile name").
						Placeholder("staging").
						Value(&createName),
				),
			)
			if err := form.Run(); err != nil {
				fmt.Println(ui.Fail("Interrupted."))
				return
			}
		}

		if createName == "" {
			fmt.Println(ui.Fail("Profile name is required. Use --name or pass it as an argument."))
			return
		}

		if engine.ProfileExists(createName) {
			fmt.Println(ui.Warn(fmt.Sprintf("Profile '%s' already exists.", createName)))
			return
		}

		envPath := ".env." + createName

		// Check if file already exists on disk
		if _, err := os.Stat(envPath); err == nil {
			fmt.Println(ui.Warn(fmt.Sprintf("%s already exists on disk. Registering it as a profile.", envPath)))
			engine.AddProfile(createName)
			fmt.Println(ui.Ok(fmt.Sprintf("Profile '%s' registered.", createName)))
			return
		}

		// Create from template
		if createFrom != "" {
			source, err := env.Parse(createFrom)
			if err != nil {
				fmt.Println(ui.Fail(fmt.Sprintf("Cannot read template '%s': %s", createFrom, err)))
				return
			}

			if createStrip {
				source = source.Strip()
			}

			if err := source.Write(envPath); err != nil {
				fmt.Println(ui.Fail("Failed to write profile: " + err.Error()))
				return
			}
		} else {
			// Empty profile
			if err := os.WriteFile(envPath, []byte(""), 0644); err != nil {
				fmt.Println(ui.Fail("Failed to create profile file: " + err.Error()))
				return
			}
		}

		if err := engine.AddProfile(createName); err != nil {
			fmt.Println(ui.Fail("Failed to register profile: " + err.Error()))
			return
		}

		suffix := ""
		if createFrom != "" {
			suffix = fmt.Sprintf(" from %s", createFrom)
			if createStrip {
				suffix += ui.Faint(" (values stripped)")
			}
		}

		fmt.Println(ui.Ok(fmt.Sprintf("Profile '%s' created%s.", createName, strings.TrimSpace(suffix))))
	},
}

func init() {
	createCmd.Flags().StringVarP(&createName, "name", "n", "", "Profile name")
	createCmd.Flags().StringVar(&createFrom, "from", "", "Create from a template file")
	createCmd.Flags().BoolVarP(&createStrip, "strip", "s", false, "Strip values from template (keys only)")
	createCmd.Flags().BoolVarP(&createInteractive, "interactive", "i", false, "Interactive mode")
}