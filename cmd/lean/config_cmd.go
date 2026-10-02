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

		fmt.Println(ui.Bolt() + " config")
		fmt.Println()
		fmt.Printf("  %s %d\n", ui.Faint("version"), cfg.Lean.Version)
		fmt.Printf("  %s %s\n", ui.Faint("default_profile"), orDash(cfg.Lean.DefaultProfile))
		fmt.Printf("  %s %s\n", ui.Faint("editor"), orDash(cfg.Lean.Editor))
		fmt.Printf("  %s %s\n", ui.Faint("schema.path"), cfg.Schema.Path)
		fmt.Printf("  %s %s\n", ui.Faint("output.format"), cfg.Output.Format)
		fmt.Printf("  %s %s\n", ui.Faint("secrets.backend"), cfg.Secrets.Backend)
		fmt.Printf("  %s %s\n", ui.Faint("secrets.recipient"), orDash(cfg.Secrets.Recipient))
		fmt.Printf("  %s %s\n", ui.Faint("secrets.identity"), orDash(cfg.Secrets.Identity))
		if cfg.Secrets.Backend == "local" {
			fmt.Printf("  %s %s\n", ui.Faint("secrets.master_key_file"), orDash(cfg.Secrets.MasterKeyFile))
		}
		fmt.Println()

		if path, err := config.GlobalPath(); err == nil {
			fmt.Printf("  %s %s\n", ui.Faint("global"), path)
		}
		fmt.Printf("  %s %s\n", ui.Faint("local"), config.LocalPath())
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a local or global configuration interactively",
	Run: func(cmd *cobra.Command, args []string) {
		if cfgGlobal {
			initGlobalConfig()
			return
		}
		initLocalConfig()
	},
}

func initLocalConfig() {
	var defaultProfile string
	var schemaPath = ".lean/schema.toml"
	var outputFormat = "text"

	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Default profile").Description("Used when no profile is specified.").Value(&defaultProfile),
		huh.NewInput().Title("Schema path").Value(&schemaPath),
		huh.NewSelect[string]().Title("Output format").Options(
			huh.NewOption("Text", "text"),
		huh.NewOption("JSON", "json"),
	).Value(&outputFormat),
	))
	if err := form.Run(); err != nil {
		fmt.Println(ui.Fail("Interrupted."))
		return
	}

	path := config.LocalPath()
	if config.Exists(path) {
		fmt.Println(ui.Warn(path + " already exists. Use `lean config edit` to change it."))
		return
	}

	cfg := config.Default()
	cfg.Lean.DefaultProfile = defaultProfile
	cfg.Schema.Path = schemaPath
	cfg.Output.Format = outputFormat
	if err := config.Save(path, cfg); err != nil {
		fmt.Println(ui.Fail("Failed to write config: " + err.Error()))
		return
	}
	fmt.Println(ui.Ok("Wrote " + path))
}

func initGlobalConfig() {
	var editor string
	cfg := config.Default()
	if path, err := config.GlobalPath(); err == nil && config.Exists(path) {
		fmt.Println(ui.Warn(path + " already exists. Use `lean config edit --global` to change it."))
		return
	}

	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Editor command (optional)").Description("Leave empty to use $EDITOR or a detected editor.").Value(&editor),
	))
	if err := form.Run(); err != nil {
		fmt.Println(ui.Fail("Interrupted."))
		return
	}
	cfg.Lean.Editor = editor

	if err := configureSecurity(cfg); err != nil {
		fmt.Println(ui.Fail("Security setup failed: " + err.Error()))
		return
	}

	path, err := config.GlobalPath()
	if err != nil {
		fmt.Println(ui.Fail("Cannot determine global config path: " + err.Error()))
		return
	}
	if err := config.Save(path, cfg); err != nil {
		fmt.Println(ui.Fail("Failed to write config: " + err.Error()))
		return
	}
	fmt.Println(ui.Ok("Wrote " + path))
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

		if cfgGlobal {
			setGlobalConfig(cmd, cfg)
			return
		}
		setLocalConfig(cmd, cfg)
	},
}

func setLocalConfig(cmd *cobra.Command, cfg *config.Config) {
	changed := false
	if cmd.Flags().Changed("default-profile") {
		cfg.Lean.DefaultProfile = cfgDefaultProfile
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
	if !changed {
		fmt.Println(ui.Info("Nothing to change."))
		return
	}

	if err := config.Save(config.LocalPath(), cfg); err != nil {
		fmt.Println(ui.Fail("Failed to write config: " + err.Error()))
		return
	}
	fmt.Println(ui.Ok("Local config updated."))
}

func setGlobalConfig(cmd *cobra.Command, cfg *config.Config) {
	changed := false
	if cmd.Flags().Changed("editor") {
		cfg.Lean.Editor = cfgEditor
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

	path, err := config.GlobalPath()
	if err != nil {
		fmt.Println(ui.Fail("Cannot determine global config path: " + err.Error()))
		return
	}
	if err := config.Save(path, cfg); err != nil {
		fmt.Println(ui.Fail("Failed to write config: " + err.Error()))
		return
	}
	fmt.Println(ui.Ok("Global config updated."))
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func init() {
	configInitCmd.Flags().BoolVarP(&cfgGlobal, "global", "g", false, "Create global config")
	configSetCmd.Flags().BoolVarP(&cfgGlobal, "global", "g", false, "Update global config")
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
