package lean

import (
	"fmt"
	"os"

	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var (
	cfgDefaultProfile string
	cfgSchemaPath     string
	cfgSecretsBackend   string
	cfgSecretsRecipient string
	cfgSecretsIdentity  string
	cfgMasterKeyFile    string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or initialize lean.toml configuration",
	Long: `Manage project-level lean.toml configuration.

Examples:
  lean config                  # show current config
  lean config init             # write a default lean.toml
  lean config set --default-profile development
  lean config set --secrets-backend gpg --secrets-recipient you@example.com`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("")
		if err != nil {
			fmt.Println(ui.Fail("Failed to load config: " + err.Error()))
			return
		}

		fmt.Printf("%s lean.toml\n\n", ui.Bolt())
		if !config.Exists("") {
			fmt.Println(ui.Faint("  (using defaults — run `lean config init` to write lean.toml)"))
			fmt.Println()
		}

		fmt.Printf("  %s            %d\n", ui.Faint("version"), cfg.Lean.Version)
		fmt.Printf("  %s   %s\n", ui.Faint("default_profile"), orDash(cfg.Lean.DefaultProfile))
		fmt.Printf("  %s       %s\n", ui.Faint("schema.path"), cfg.Schema.Path)
		fmt.Printf("  %s   %s\n", ui.Faint("secrets.backend"), cfg.Secrets.Backend)
		if cfg.Secrets.Recipient != "" {
			fmt.Printf("  %s %s\n", ui.Faint("secrets.recipient"), cfg.Secrets.Recipient)
		}
		if cfg.Secrets.Identity != "" {
			fmt.Printf("  %s  %s\n", ui.Faint("secrets.identity"), cfg.Secrets.Identity)
		}
		fmt.Printf("  %s %s\n", ui.Faint("secrets.master_key_file"), orDash(cfg.Secrets.MasterKeyFile))
		fmt.Println()
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Write a default lean.toml",
	Run: func(cmd *cobra.Command, args []string) {
		if config.Exists("") {
			fmt.Println(ui.Warn("lean.toml already exists."))
			return
		}
		cfg := config.Default()
		if err := config.Save("", cfg); err != nil {
			fmt.Println(ui.Fail("Failed to write lean.toml: " + err.Error()))
			return
		}
		fmt.Println(ui.Ok("Wrote lean.toml"))
		fmt.Println(ui.Faint("  Edit it or use `lean config set` to configure."))
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Update lean.toml settings",
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
		if cmd.Flags().Changed("schema-path") {
			cfg.Schema.Path = cfgSchemaPath
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
			fmt.Println(ui.Info("Nothing to change. Pass flags like --default-profile or --secrets-backend."))
			return
		}

		if err := config.Save("", cfg); err != nil {
			fmt.Println(ui.Fail("Failed to write lean.toml: " + err.Error()))
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
	configSetCmd.Flags().StringVar(&cfgDefaultProfile, "default-profile", "", "Default profile name")
	configSetCmd.Flags().StringVar(&cfgSchemaPath, "schema-path", "", "Path to schema.toml")
	configSetCmd.Flags().StringVar(&cfgSecretsBackend, "secrets-backend", "", "Secrets backend: local | gpg | age | ssh")
	configSetCmd.Flags().StringVar(&cfgSecretsRecipient, "secrets-recipient", "", "GPG/age recipient or SSH .pub path")
	configSetCmd.Flags().StringVar(&cfgSecretsIdentity, "secrets-identity", "", "Age identity / SSH private key path")
	configSetCmd.Flags().StringVar(&cfgMasterKeyFile, "master-key-file", "", "Path to master key file (default ~/.lean/key)")

	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSetCmd)

	// silence unused import if os is only for future use
	_ = os.Stderr
}
