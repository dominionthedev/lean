package lean

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var (
	cfgDefaultProfile   string
	cfgSchemaPath       string
	cfgEditor           string
	cfgOutputFormat     string
	cfgSecretsBackend   string
	cfgSecretsRecipient string
	cfgSecretsIdentity  string
	cfgMasterKeyFile    string
	cfgGlobal           bool
	cfgInteractive      bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and manage Lean configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("")
		if err != nil {
			fmt.Println(ui.Fail("Failed to load config: " + err.Error()))
			return
		}
		fmt.Printf("%s lean.toml\n\n", ui.Bolt())
		if !config.Exists("") {
			fmt.Println(ui.Faint("  (using defaults — run `lean config init` to create local config)"))
			fmt.Println()
		}
		fmt.Printf("  %s %d\n", ui.Faint("version"), cfg.Lean.Version)
		fmt.Printf("  %s %s\n", ui.Faint("default_profile"), orDash(cfg.Lean.DefaultProfile))
		fmt.Printf("  %s %s\n", ui.Faint("editor"), orDash(cfg.Lean.Editor))
		fmt.Printf("  %s %s\n", ui.Faint("profiles.active_file"), cfg.Profiles.ActiveFile)
		fmt.Printf("  %s %s\n", ui.Faint("schema.path"), cfg.Schema.Path)
		fmt.Printf("  %s %s\n", ui.Faint("output.format"), cfg.Output.Format)
		fmt.Printf("  %s %s\n", ui.Faint("secrets.backend"), cfg.Secrets.Backend)
		fmt.Printf("  %s %s\n", ui.Faint("secrets.recipient"), orDash(cfg.Secrets.Recipient))
		fmt.Printf("  %s %s\n", ui.Faint("secrets.identity"), orDash(cfg.Secrets.Identity))
		if cfg.Secrets.Backend == "local" {
			fmt.Printf("  %s %s\n", ui.Faint("secrets.master_key_file"), orDash(cfg.Secrets.MasterKeyFile))
		}
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a local or global configuration interactively",
	Run: func(cmd *cobra.Command, args []string) {
		cfgDefaultProfile = ""
		cfgSchemaPath = ".lean/schema.toml"
		cfgOutputFormat = "text"
		cfgEditor = ""

		var scope string
		if cfgGlobal {
			scope = "global"
		} else {
			scope = "local"
		}
		form := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Title("Configuration scope").Description("Local settings override global settings.").Options(
				huh.NewOption("Local project (.lean/lean.toml)", "local"),
				huh.NewOption("Global user (~/.lean/lean.toml)", "global"),
			).Value(&scope),
			huh.NewInput().Title("Default profile").Description("Used when no profile is specified.").Value(&cfgDefaultProfile),
			huh.NewInput().Title("Schema path").Value(&cfgSchemaPath),
			huh.NewInput().Title("Editor command (optional)").Description("Leave empty to use $EDITOR or a detected editor.").Value(&cfgEditor),
			huh.NewSelect[string]().Title("Output format").Options(
				huh.NewOption("Text", "text"),
				huh.NewOption("JSON", "json"),
			).Value(&cfgOutputFormat),
		))
		if err := form.Run(); err != nil {
			fmt.Println(ui.Fail("Interrupted."))
			return
		}

		path := config.LocalPath()
		if scope == "global" {
			var err error
			path, err = config.GlobalPath()
			if err != nil {
				fmt.Println(ui.Fail("Cannot determine global config path: " + err.Error()))
				return
			}
		}
		if config.Exists(path) {
			fmt.Println(ui.Warn(path + " already exists. Use `lean config edit` to change it."))
			return
		}

		cfg := config.Default()
		cfg.Lean.DefaultProfile = cfgDefaultProfile
		cfg.Lean.Editor = cfgEditor
		cfg.Output.Format = cfgOutputFormat
		cfg.Schema.Path = cfgSchemaPath
		if err := configureSecurity(cfg); err != nil {
			fmt.Println(ui.Fail("Security setup failed: " + err.Error()))
			return
		}
		if err := config.Save(path, cfg); err != nil {
			fmt.Println(ui.Fail("Failed to write config: " + err.Error()))
			return
		}
		fmt.Println(ui.Ok("Wrote " + path))
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Update individual configuration settings",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("")
		if err != nil {
			fmt.Println(ui.Fail("Failed to load config: " + err.Error()))
			return
		}
		changed := false
		if cmd.Flags().Changed("default-profile") {
			cfg.Lean.DefaultProfile = cfgDefaultProfile
			changed = true
		}
		if cmd.Flags().Changed("editor") {
			cfg.Lean.Editor = cfgEditor
			changed = true
		}
		if cmd.Flags().Changed("schema-path") {
			cfg.Schema.Path = cfgSchemaPath
			changed = true
		}
		if cmd.Flags().Changed("output-format") {
			cfg.Output.Format = cfgOutputFormat
			changed = true
		}
		if cmd.Flags().Changed("secrets-backend") {
			cfg.Secrets.Backend = cfgSecretsBackend
			changed = true
		}
		if cmd.Flags().Changed("secrets-recipient") {
			cfg.Secrets.Recipient = cfgSecretsRecipient
			changed = true
		}
		if cmd.Flags().Changed("secrets-identity") {
			cfg.Secrets.Identity = cfgSecretsIdentity
			changed = true
		}
		if cmd.Flags().Changed("master-key-file") {
			cfg.Secrets.MasterKeyFile = cfgMasterKeyFile
			changed = true
		}
		if !changed {
			fmt.Println(ui.Info("Nothing to change."))
			return
		}
		if err := config.Save("", cfg); err != nil {
			fmt.Println(ui.Fail("Failed to write config: " + err.Error()))
			return
		}
		fmt.Println(ui.Ok("lean.toml updated."))
	},
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func init() {
	configInitCmd.Flags().BoolVarP(&cfgGlobal, "global", "g", false, "Create global config")
	configInitCmd.Flags().BoolVarP(&cfgInteractive, "interactive", "i", true, "Create the config interactively")
	configSetCmd.Flags().StringVar(&cfgDefaultProfile, "default-profile", "", "Default profile name")
	configSetCmd.Flags().StringVar(&cfgEditor, "editor", "", "Editor command")
	configSetCmd.Flags().StringVar(&cfgSchemaPath, "schema-path", "", "Path to schema.toml")
	configSetCmd.Flags().StringVar(&cfgOutputFormat, "output-format", "", "Output format: text | json")
	configSetCmd.Flags().StringVar(&cfgSecretsBackend, "secrets-backend", "", "Secrets backend: local | gpg | age | ssh")
	configSetCmd.Flags().StringVar(&cfgSecretsRecipient, "secrets-recipient", "", "GPG/age recipient or SSH public-key path")
	configSetCmd.Flags().StringVar(&cfgSecretsIdentity, "secrets-identity", "", "Age identity or SSH private-key path")
	configSetCmd.Flags().StringVar(&cfgMasterKeyFile, "master-key-file", "", "Local backend master key path")
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSetCmd)
}
