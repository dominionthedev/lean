package lean

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/globaldir"
	"github.com/dominionthedev/lean/internal/secrets"
	"github.com/dominionthedev/lean/internal/ui"
)

func init() {
	secretCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd == secretKeygenCmd {
			return nil
		}

		cfg, err := config.Load("")
		if err != nil {
			return err
		}

		if !configConfigured() {
			fmt.Println(ui.Info("Lean has no global configuration yet. Let's configure secret storage first."))
			if err := configureSecurity(cfg); err != nil {
				return err
			}
			path, err := config.GlobalPath()
			if err != nil {
				return err
			}
			if err := config.Save(path, cfg); err != nil {
				return err
			}
			fmt.Println(ui.Ok("Created " + path))
		}

		if cfg.Secrets.Backend == "local" && cfg.Secrets.MasterKeyFile == "" && !secrets.KeyExists() {
			path, err := secrets.Keygen()
			if err != nil {
				return fmt.Errorf("create local secrets key: %w", err)
			}
			fmt.Println(ui.Ok("Created local Lean master key at " + path))
		}
		return nil
	}
}

func configConfigured() bool {
	if path, err := config.GlobalPath(); err == nil && config.Exists(path) {
		return true
	}
	legacy, err := globaldir.Path("lean.toml")
	return err == nil && config.Exists(legacy)
}
