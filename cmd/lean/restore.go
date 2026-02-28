package lean

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/dominionthedev/lean/internal/backup"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore .env from a backup",
	Run: func(cmd *cobra.Command, args []string) {

		backups, err := backup.List()
		if err != nil {
			fmt.Println(ui.Fail("Could not read backups: " + err.Error()))
			return
		}

		if len(backups) == 0 {
			fmt.Println(ui.Info("No backups found. Lean creates one every time you run `lean apply`."))
			return
		}

		// If a name was passed directly, use it
		if len(args) > 0 {
			if err := backup.Restore(args[0]); err != nil {
				fmt.Println(ui.Fail("Restore failed: " + err.Error()))
				return
			}
			fmt.Println(ui.Ok(fmt.Sprintf("Restored from '%s'.", args[0])))
			return
		}

		// Interactive selection
		var chosen string
		options := make([]huh.Option[string], len(backups))
		for i, b := range backups {
			options[i] = huh.NewOption(b, b)
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Which backup would you like to restore?").
					Options(options...).
					Value(&chosen),
			),
		)

		if err := form.Run(); err != nil {
			fmt.Println(ui.Fail("Interrupted."))
			return
		}

		if err := backup.Restore(chosen); err != nil {
			fmt.Println(ui.Fail("Restore failed: " + err.Error()))
			return
		}

		fmt.Println(ui.Ok(fmt.Sprintf("Restored from '%s'. Your .env is back.", chosen)))
	},
}
