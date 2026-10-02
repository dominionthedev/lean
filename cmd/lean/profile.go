package lean

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
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
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run lean init first."))
			return
		}
		if err := engine.ScanDisk(); err != nil {
			fmt.Println(ui.Fail("Could not scan profiles: " + err.Error()))
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
		path := env.ProfilePath(name)
		if _, err := os.Stat(path); err != nil {
			confirmStaleProfile(engine, name)
			return
		}
		if err := os.Remove(path); err != nil {
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

var profileArchiveCmd = &cobra.Command{
	Use:   "archive NAME",
	Short: "Archive a profile into .lean/archive",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run lean init first."))
			return
		}
		if err := engine.ScanDisk(); err != nil {
			fmt.Println(ui.Fail("Could not scan profiles: " + err.Error()))
			return
		}
		if !engine.ProfileExists(name) {
			fmt.Println(ui.Fail(fmt.Sprintf("Profile '%s' not found.", name)))
			return
		}
		if engine.State.Current == name {
			fmt.Println(ui.Fail(fmt.Sprintf("Cannot archive active profile '%s'. Apply another profile first.", name)))
			return
		}

		path := env.ProfilePath(name)
		archivePath := filepath.Join(".lean", "archive", "profiles", name+".env")
		if _, err := os.Stat(path); err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Profile file '%s' does not exist.", path)))
			return
		}
		if err := os.MkdirAll(filepath.Dir(archivePath), 0700); err != nil {
			fmt.Println(ui.Fail("Failed to create archive: " + err.Error()))
			return
		}
		if _, err := os.Stat(archivePath); err == nil {
			fmt.Println(ui.Fail(fmt.Sprintf("An archived profile named '%s' already exists.", name)))
			return
		}
		if err := os.Rename(path, archivePath); err != nil {
			fmt.Println(ui.Fail("Failed to archive profile: " + err.Error()))
			return
		}
		if err := engine.DeleteProfile(name); err != nil {
			fmt.Println(ui.Fail("Failed to update profile state: " + err.Error()))
			return
		}
		fmt.Println(ui.Ok(fmt.Sprintf("Profile '%s' archived.", name)))
	},
}

var profileRestoreCmd = &cobra.Command{
	Use:   "restore NAME",
	Short: "Restore an archived profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run lean init first."))
			return
		}
		path := env.ProfilePath(name)
		archivePath := filepath.Join(".lean", "archive", "profiles", name+".env")
		if _, err := os.Stat(archivePath); err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("No archived profile '%s' found.", name)))
			return
		}
		if _, err := os.Stat(path); err == nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Profile '%s' already exists.", name)))
			return
		}
		if err := os.Rename(archivePath, path); err != nil {
			fmt.Println(ui.Fail("Failed to restore profile: " + err.Error()))
			return
		}
		if err := engine.AddProfile(name); err != nil {
			fmt.Println(ui.Fail("Failed to register restored profile: " + err.Error()))
			return
		}
		fmt.Println(ui.Ok(fmt.Sprintf("Profile '%s' restored.", name)))
	},
}

var profileArchiveListCmd = &cobra.Command{
	Use:   "list-archived",
	Short: "List archived profiles",
	Run: func(cmd *cobra.Command, args []string) {
		entries, err := os.ReadDir(filepath.Join(".lean", "archive", "profiles"))
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println(ui.Info("No archived profiles."))
				return
			}
			fmt.Println(ui.Fail("Failed to read archive: " + err.Error()))
			return
		}
		found := false
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".env") {
				continue
			}
			found = true
			fmt.Println("  " + strings.TrimSuffix(entry.Name(), ".env"))
		}
		if !found {
			fmt.Println(ui.Info("No archived profiles."))
		}
	},
}

var profileEditCmd = &cobra.Command{
	Use:   "edit NAME",
	Short: "Edit a profile interactively",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		path := env.ProfilePath(name)
		if _, err := os.Stat(path); err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Profile '%s' does not exist.", name)))
			return
		}
		f, err := env.Parse(path)
		if err != nil {
			fmt.Println(ui.Fail("Could not read profile: " + err.Error()))
			return
		}
		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run lean init first."))
			return
		}
		for {
			var action string
			form := huh.NewForm(huh.NewGroup(
				huh.NewSelect[string]().Title("Profile action").Options(
					huh.NewOption("Set a variable", "set"),
					huh.NewOption("Delete a variable", "delete"),
					huh.NewOption("Finish", "finish"),
				).Value(&action),
			))
			if err := form.Run(); err != nil || action == "finish" {
				return
			}
			var key, value string
			if action == "set" {
				form := huh.NewForm(huh.NewGroup(
					huh.NewInput().Title("Key").Value(&key),
					huh.NewInput().Title("Value").Value(&value),
				))
				if err := form.Run(); err != nil {
					return
				}
				if key == "" {
					continue
				}
				f.Set(key, value)
			} else {
				form := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Key").Value(&key)))
				if err := form.Run(); err != nil {
					return
				}
				f.Delete(key)
			}
			if err := f.Write(path); err != nil {
				fmt.Println(ui.Fail("Failed to write profile: " + err.Error()))
				return
			}
			if name == engine.State.Current {
				if resolved, err := env.Resolve(name); err == nil {
					_ = resolved.Write(".env")
				} else {
					fmt.Println(ui.Warn("Could not sync .env: " + err.Error()))
				}
			}
		}
	},
}

func init() {
	profileCmd.AddCommand(profileDeleteCmd)
	profileCmd.AddCommand(profileArchiveCmd)
	profileCmd.AddCommand(profileRestoreCmd)
	profileCmd.AddCommand(profileArchiveListCmd)
	profileCmd.AddCommand(profileEditCmd)
}
