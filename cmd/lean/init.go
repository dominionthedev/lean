package lean

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var initQuiet bool

const schemaTemplate = `# Lean environment schema
# Add rules under [keys.NAME].
#
# [keys.DATABASE_URL]
# required = true
# secret = true
# description = "Database connection string"
#
# [keys.NODE_ENV]
# values = ["development", "test", "production"]
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize lean interactively",
	Run: func(cmd *cobra.Command, args []string) {
		if initQuiet {
			fmt.Println(ui.Warn("I can't be silent."))
			fmt.Println(ui.Faint("Try `lean create` instead."))
			return
		}

		if _, err := os.Stat(".lean"); err == nil {
			fmt.Println(ui.Info("Already initialized."))
			return
		}

		fmt.Println(ui.Banner.Render("⚡ lean is waking up..."))
		fmt.Println()

		var profileName string
		var addVars bool
		form := huh.NewForm(huh.NewGroup(
			huh.NewInput().Title("Profile name").Description("What environment are we setting up?").Placeholder("dev").Value(&profileName),
			huh.NewConfirm().Title("Add variables now?").Description("You can always add them later.").Value(&addVars),
		))
		if err := form.Run(); err != nil {
			fmt.Println(ui.Fail("Interrupted."))
			return
		}
		if profileName == "" {
			profileName = "dev"
		}

		var varsInput string
		if addVars {
			varsForm := huh.NewForm(huh.NewGroup(
				huh.NewText().Title("Variables").Description("Enter KEY=VALUE pairs, one per line.").Placeholder("DATABASE_URL=\nAPI_KEY=\nDEBUG=true").Value(&varsInput),
			))
			if err := varsForm.Run(); err != nil {
				fmt.Println(ui.Fail("Interrupted."))
				return
			}
		}

		if err := core.Initialize(); err != nil {
			fmt.Println(ui.Fail("Failed to initialize: " + err.Error()))
			return
		}
		if err := os.MkdirAll(".lean", 0700); err != nil {
			fmt.Println(ui.Fail("Failed to create .lean: " + err.Error()))
			return
		}

		cfg := config.Default()
		cfg.Lean.DefaultProfile = profileName
		if err := config.Save(config.LocalPath(), cfg); err != nil {
			fmt.Println(ui.Fail("Failed to create local config: " + err.Error()))
			return
		}
		if err := os.WriteFile(cfg.Schema.Path, []byte(schemaTemplate), 0644); err != nil {
			fmt.Println(ui.Fail("Failed to create schema template: " + err.Error()))
			return
		}

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Failed to load initialized state: " + err.Error()))
			return
		}
		if err := engine.AddProfile(profileName); err != nil {
			fmt.Println(ui.Fail("Failed to register profile: " + err.Error()))
			return
		}

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
		if err := os.WriteFile(envPath, []byte(body), 0600); err != nil {
			fmt.Println(ui.Fail("Failed to write " + envPath + ": " + err.Error()))
			return
		}
		if err := os.WriteFile(".env", []byte(body), 0600); err != nil {
			fmt.Println(ui.Fail("Failed to write .env: " + err.Error()))
			return
		}
		if err := engine.SetCurrent(profileName); err != nil {
			fmt.Println(ui.Fail("Failed to set active profile: " + err.Error()))
			return
		}

		fmt.Println()
		fmt.Println(ui.Ok("lean is ready."))
		fmt.Printf("   Profile  : %s\n", ui.Active.Render(profileName))
		fmt.Printf("   Config   : %s\n", ui.Faint(config.LocalPath()))
		fmt.Printf("   Schema   : %s\n", ui.Faint(cfg.Schema.Path))

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
