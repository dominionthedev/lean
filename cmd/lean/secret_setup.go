package lean

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/ui"
)

func init() {
	secretCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd == secretKeygenCmd || configConfigured() {
			return nil
		}
		cfg := config.Default()
		fmt.Println(ui.Info("Lean has no configuration yet. Let's configure secret storage first."))
		if err := configureSecurity(cfg); err != nil {
			return err
		}
		if err := config.Save(config.LocalPath(), cfg); err != nil {
			return err
		}
		fmt.Println(ui.Ok("Created " + config.LocalPath()))
		return nil
	}
}

func configConfigured() bool {
	if config.Exists("") {
		return true
	}
	path, err := config.GlobalPath()
	if err != nil {
		return false
	}
	return config.Exists(path)
}
