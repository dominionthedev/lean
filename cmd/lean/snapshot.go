package lean

import (
	"fmt"
	"os"
	"strings"

	"github.com/dominionthedev/lean/internal/backup"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot [name]",
	Short: "Save a named snapshot of the current .env",
	Long: `Create a named snapshot of the current .env so you can restore it later.

Examples:
  lean snapshot before-migration
  lean snapshot before-testing
  lean snapshot                  # lists snapshots (alias for lean snapshots)

Named snapshots live in .lean/backups/<name>.env and can be restored with:
  lean restore before-migration`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			listSnapshots()
			return
		}

		name := args[0]
		if err := backup.NamedSnapshot(name); err != nil {
			fmt.Println(ui.Fail("Snapshot failed: " + err.Error()))
			return
		}
		fmt.Println(ui.Ok(fmt.Sprintf("Snapshot '%s' saved.", name)))
	},
}

var snapshotsCmd = &cobra.Command{
	Use:     "snapshots",
	Short:   "List named and automatic snapshots",
	Aliases: []string{"snaps"},
	Run: func(cmd *cobra.Command, args []string) {
		listSnapshots()
	},
}

func listSnapshots() {
	all, err := backup.List()
	if err != nil {
		fmt.Println(ui.Fail("Could not list snapshots: " + err.Error()))
		return
	}
	if len(all) == 0 {
		fmt.Println(ui.Info("No snapshots yet. Run `lean snapshot <name>` or `lean apply`."))
		return
	}

	named, _ := backup.ListNamed()
	namedSet := make(map[string]bool)
	for _, n := range named {
		namedSet[n] = true
	}

	fmt.Println(ui.Bolt() + " " + ui.Bold.Render("Snapshots"))
	fmt.Println()

	if len(named) > 0 {
		fmt.Println("  " + ui.Faint("named"))
		for _, n := range named {
			fmt.Printf("  %s %s\n", ui.Success.Render("◆"), n)
		}
		fmt.Println()
	}

	var auto []string
	for _, f := range all {
		base := strings.TrimSuffix(f, ".env")
		if !namedSet[base] {
			auto = append(auto, f)
		}
	}
	if len(auto) > 0 {
		fmt.Println("  " + ui.Faint("automatic (from lean apply)"))
		for _, f := range auto {
			fmt.Printf("  %s %s\n", ui.Muted.Render("·"), f)
		}
		fmt.Println()
	}
}

var snapshotDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a named snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if !backup.Exists(name) {
			fmt.Println(ui.Fail(fmt.Sprintf("Snapshot '%s' not found.", name)))
			os.Exit(1)
		}
		if err := backup.Delete(name); err != nil {
			fmt.Println(ui.Fail("Delete failed: " + err.Error()))
			return
		}
		fmt.Println(ui.Ok(fmt.Sprintf("Snapshot '%s' deleted.", name)))
	},
}

func init() {
	snapshotCmd.AddCommand(snapshotDeleteCmd)
}
