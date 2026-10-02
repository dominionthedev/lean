package lean

import (
	"fmt"

	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/ui"
)

func init() {
	secretCmd.PersistentPreRun = func(cmd interface{}) {
		// Kept as a separate hook in the command setup so secret operations can
		// bootstrap security configuration without making config.Load interactive.
	}
}

func ensureSecretConfiguration() error {
	if config.Exists("") {
		return nil
	}
	cfg := config.Default()
	fmt.Println(ui.Info("Lean has no local configuration yet. Let's configure secret storage first."))
	if err := configureSecurity(cfg); err != nil {
		return err
	}
	if err := config.Save(config.LocalPath(), cfg); err != nil {
		return err
	}
	fmt.Println(ui.Ok("Created " + config.LocalPath()))
	return nil
}
