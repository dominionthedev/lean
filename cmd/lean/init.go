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

var initQuiet bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize lean interactively",
	Run: func(cmd *cobra.Command, args []string) {

		// Easter egg
		if initQuiet {
			fmt.Println(ui.Warn("I can't be silent."))
			fmt.Println(ui.Faint("Try `lean create` instead."))
			return
		}

		// Already initialized
		if _, err := os.Stat(".lean"); err == nil {
			fmt.Println(ui.Info("Already initialized."))
			return
		}

		fmt.Println(ui.Banner.Render("⚡ lean is waking up..."))
		fmt.Println()

		var profileName string
		var addVars bool

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Profile name").
					Description("What environment are we setting up?").
					Placeholder("dev").
					Value(&profileName),
				huh.NewConfirm().
					Title("Add variables now?").
					Description("You can always add them later.").
					Value(&addVars),
			),
		)

		if err := form.Run(); err != nil {
			fmt.Println(ui.Fail("Interrupted."))
			return
		}

		if profileName == "" {
			profileName = "dev"
		}

		var varsInput string
		if addVars {
			varsForm := huh.NewForm(
				huh.NewGroup(
					huh.NewText().
						Title("Variables").
						Description("Enter KEY=VALUE pairs, one per line.").
						Placeholder("DATABASE_URL=\nAPI_KEY=\nDEBUG=true").
						Value(&varsInput),
				),
			)
			if err := varsForm.Run(); err != nil {
				fmt.Println(ui.Fail("Interrupted."))
				return
			}
		}

		// Initialize lean state
		if err := core.Initialize(); err != nil {
			fmt.Println(ui.Fail("Failed to initialize: " + err.Error()))
			return
		}

		engine, _ := core.NewEngine()
		engine.AddProfile(profileName)

		// Build env file content
		var content strings.Builder
		if varsInput != "" {
			for _, line := range strings.Split(strings.TrimSpace(varsInput), "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					content.WriteString(line + "\n")
				}
			}
		}

		envPath := ".env." + profileName
		body := content.String()

		if err := os.WriteFile(envPath, []byte(body), 0644); err != nil {
			fmt.Println(ui.Fail("Failed to write " + envPath + ": " + err.Error()))
			return
		}

		// Atomic write to .env
		tmp := ".env.tmp"
		if err := os.WriteFile(tmp, []byte(body), 0644); err != nil {
			fmt.Println(ui.Fail("Failed to write .env: " + err.Error()))
			return
		}
		if err := os.Rename(tmp, ".env"); err != nil {
			fmt.Println(ui.Fail("Failed to finalize .env: " + err.Error()))
			return
		}

		engine.SetCurrent(profileName)

		// Summary
		fmt.Println()
		fmt.Println(ui.Ok("lean is ready."))
		fmt.Printf("   Profile  : %s\n", ui.Active.Render(profileName))

		if envFile, err := env.Parse(envPath); err == nil {
			keys := envFile.Keys()
			if len(keys) > 0 {
				fmt.Printf("   Variables: %s\n", ui.Faint(fmt.Sprintf("%d added", len(keys))))
			} else {
				fmt.Printf("   Variables: %s\n", ui.Faint("none yet — add them anytime"))
			}
		}
	},
}

func init() {
	initCmd.Flags().BoolVarP(&initQuiet, "quiet", "q", false, "Quiet mode (not supported — lean init is always interactive)")
}