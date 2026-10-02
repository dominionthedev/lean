package lean

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/ui"
)

func confirmStaleProfile(engine *core.Engine, profile string) bool {
	active := engine.State.Current == profile
	if active {
		fmt.Println(ui.Warn(fmt.Sprintf("The active profile '%s' no longer has a .env.%s file.", profile, profile)))
	} else {
		fmt.Println(ui.Warn(fmt.Sprintf("Profile '%s' no longer has a .env.%s file.", profile, profile)))
	}

	remove := false
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Remove this stale profile from Lean?").
				Affirmative("Remove").
				Negative("Keep it").
				Value(&remove),
		),
	)
	if err := form.Run(); err != nil {
		fmt.Println(ui.Fail("Interrupted."))
		return false
	}
	if !remove {
		fmt.Println(ui.Info("Profile kept in Lean state."))
		return false
	}
	if err := engine.DeleteProfile(profile); err != nil {
		fmt.Println(ui.Fail("Failed to remove stale profile: " + err.Error()))
		return false
	}
	fmt.Println(ui.Ok(fmt.Sprintf("Removed stale profile '%s'.", profile)))
	return true
}
