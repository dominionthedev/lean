package lean

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dominionthedev/lean/internal/backup"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var (
	createName string
	createFrom string
	createStrip bool
	createInteractive bool
	createExtends string
)

var createCmd = &cobra.Command{
	Use: "create",
	Short: "Create a new environment profile",
	Run: func(cmd *cobra.Command, args []string) {
		if createName == "" && len(args) > 0 { createName = args[0] }
		engine, err := core.NewEngine()
		if err != nil { fmt.Println(ui.Fail("Not initialized. Run lean init first.")); return }
		_ = engine.ScanDisk()
		if createInteractive || createName == "" {
			form := huh.NewForm(huh.NewGroup(
				huh.NewInput().Title("Profile name").Placeholder("staging").Value(&createName),
			))
			if err := form.Run(); err != nil { fmt.Println(ui.Fail("Interrupted.")); return }
		}
		if createName == "" { fmt.Println(ui.Fail("Profile name is required.")); return }
		if engine.ProfileExists(createName) { fmt.Println(ui.Warn(fmt.Sprintf("Profile '%s' already exists.", createName))); return }

		envPath := env.ProfilePath(createName)
		var content *env.File
		if createFrom != "" {
			source, err := env.Parse(createFrom)
			if err != nil { fmt.Println(ui.Fail(fmt.Sprintf("Cannot read template '%s': %s", createFrom, err))); return }
			if createStrip { source = source.Strip() }
			content = source
			_ = engine.AddTemplate(createFrom)
		} else {
			content = &env.File{Path: envPath}
		}
		if createExtends != "" {
			parentPath := env.ProfilePath(createExtends)
			if _, err := os.Stat(parentPath); err != nil {
				fmt.Println(ui.Warn(fmt.Sprintf("Parent profile '%s' not found yet.", createExtends)))
			}
			content.Entries = append([]env.Entry{{Comment: "# lean:extends " + createExtends}, {Blank: true}}, content.Entries...)
		}
		if err := content.Write(envPath); err != nil { fmt.Println(ui.Fail("Failed to write profile: " + err.Error())); return }
		if err := engine.AddProfile(createName); err != nil { fmt.Println(ui.Fail("Failed to register profile: " + err.Error())); return }
		_ = backup.SnapshotProfile(createName, envPath)
		var parts []string
		if createFrom != "" {
			part := "from " + createFrom
			if createStrip { part += " (values stripped)" }
			parts = append(parts, part)
		}
		if createExtends != "" { parts = append(parts, "extends " + createExtends) }
		suffix := ""
		if len(parts) > 0 { suffix = " " + strings.Join(parts, ", ") }
		fmt.Println(ui.Ok(fmt.Sprintf("Profile '%s' created%s.", createName, suffix)))
	},
}

func init() {
	createCmd.Flags().StringVarP(&createName, "name", "n", "", "Profile name")
	createCmd.Flags().StringVar(&createFrom, "from", "", "Create from a template file")
	createCmd.Flags().BoolVarP(&createStrip, "strip", "s", false, "Strip values from template (keys only)")
	createCmd.Flags().BoolVarP(&createInteractive, "interactive", "i", false, "Interactive mode")
	createCmd.Flags().StringVar(&createExtends, "extends", "", "Inherit from a parent profile")
}
