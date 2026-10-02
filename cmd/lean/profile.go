package lean

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage environment profiles",
}

var profileDeleteCmd = &cobra.Command{
	Use:     "delete NAME",
	Aliases: []string{"rm"},
	Short:   "Delete an environment profile",
	Long: "Delete a profile and its associated local state.\n\nThe profile file (.env.NAME), profile metadata, and profile-scoped\nencrypted secrets are removed. The active profile and profiles that\nare still used as inheritance parents cannot be deleted.",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run lean init first."))
			return
		}

		if !engine.ProfileExists(name) {
			fmt.Println(ui.Fail(fmt.Sprintf("Profile '%s' not found.", name)))
			return
		}

		if engine.State.Current == name {
			fmt.Println(ui.Fail(fmt.Sprintf("Cannot delete active profile '%s'. Apply another profile first.", name)))
			return
		}

		profilePath := env.ProfilePath(name)
		if _, err := os.Stat(profilePath); err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Profile file '%s' not found.", profilePath)))
			return
		}

		entries, err := os.ReadDir(".")
		if err != nil {
			fmt.Println(ui.Fail("Could not inspect profiles: " + err.Error()))
			return
		}

		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name(), ".env.") || entry.Name() == profilePath {
				continue
			}
			child := strings.TrimPrefix(entry.Name(), ".env.")
			if child == "tmp" || child == "template" || child == "example" || child == "schema" {
				continue
			}

			f, err := env.Parse(entry.Name())
			if err != nil {
				continue
			}
			if f.Extends() == name {
				fmt.Println(ui.Fail(fmt.Sprintf("Cannot delete '%s': profile '%s' extends it.", name, child)))
				return
			}
		}

		secretDir := filepath.Join(".lean", "secrets", name)
		if err := os.RemoveAll(secretDir); err != nil {
			fmt.Println(ui.Fail("Failed to remove profile secrets: " + err.Error()))
			return
		}

		if err := os.Remove(profilePath); err != nil {
			fmt.Println(ui.Fail("Failed to delete profile: " + err.Error()))
			return
		}

		if err := engine.DeleteProfile(name); err != nil {
			fmt.Println(ui.Fail("Failed to update profile state: " + err.Error()))
			return
		}

		fmt.Println(ui.Ok(fmt.Sprintf("Profile '%s' deleted.", name)))
	},
}

func init() {
	profileCmd.AddCommand(profileDeleteCmd)
}
