package lean

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var (
	templateName  string
	templateStrip bool
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage environment templates",
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all known templates",
	Run: func(cmd *cobra.Command, args []string) {
		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		engine.ScanTemplates()

		if len(engine.State.Templates) == 0 {
			fmt.Println(ui.Info("No templates found. Create .env.template or .env.example, or use `lean template add`."))
			return
		}

		fmt.Println(ui.Bolt() + " " + ui.Bold.Render("Templates"))
		fmt.Println()

		for _, t := range engine.State.Templates {
			fmt.Printf("  %s %s\n", ui.Muted.Render("·"), t)
		}
		fmt.Println()
	},
}

var templateAddCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Register a file as a template",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Println(ui.Fail(fmt.Sprintf("File %s does not exist.", path)))
			return
		}

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		if err := engine.AddTemplate(path); err != nil {
			fmt.Println(ui.Fail("Failed to add template: " + err.Error()))
			return
		}

		fmt.Println(ui.Ok(fmt.Sprintf("Template '%s' registered.", path)))
	},
}

var templateCreateCmd = &cobra.Command{
	Use:   "create-from [template]",
	Short: "Create a new profile from a template",
	Run: func(cmd *cobra.Command, args []string) {
		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		engine.ScanTemplates()

		selectedTemplate := ""
		if len(args) > 0 {
			selectedTemplate = args[0]
		}

		if selectedTemplate == "" {
			if len(engine.State.Templates) == 0 {
				fmt.Println(ui.Fail("No templates available."))
				return
			}

			options := make([]huh.Option[string], len(engine.State.Templates))
			for i, t := range engine.State.Templates {
				options[i] = huh.NewOption(t, t)
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Select a template").
						Options(options...).
						Value(&selectedTemplate),
				),
			)

			if err := form.Run(); err != nil {
				fmt.Println(ui.Fail("Interrupted."))
				return
			}
		}

		if templateName == "" {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("New profile name").
						Placeholder("prod").
						Value(&templateName),
				),
			)

			if err := form.Run(); err != nil {
				fmt.Println(ui.Fail("Interrupted."))
				return
			}
		}

		if templateName == "" {
			fmt.Println(ui.Fail("Profile name is required."))
			return
		}

		if engine.ProfileExists(templateName) {
			fmt.Println(ui.Warn(fmt.Sprintf("Profile '%s' already exists.", templateName)))
			return
		}

		envPath := ".env." + templateName
		source, err := env.Parse(selectedTemplate)
		if err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Cannot read template '%s': %s", selectedTemplate, err)))
			return
		}

		if templateStrip {
			source = source.Strip()
		}

		if err := source.Write(envPath); err != nil {
			fmt.Println(ui.Fail("Failed to write profile: " + err.Error()))
			return
		}

		if err := engine.AddProfile(templateName); err != nil {
			fmt.Println(ui.Fail("Failed to register profile: " + err.Error()))
			return
		}

		suffix := ""
		if templateStrip {
			suffix = ui.Faint(" (values stripped)")
		}

		fmt.Println(ui.Ok(fmt.Sprintf("Profile '%s' created from %s%s.", templateName, selectedTemplate, suffix)))
	},
}

func init() {
	templateCreateCmd.Flags().StringVarP(&templateName, "name", "n", "", "New profile name")
	templateCreateCmd.Flags().BoolVarP(&templateStrip, "strip", "s", false, "Strip values (keys only)")

	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateAddCmd)
	templateCmd.AddCommand(templateCreateCmd)
}
