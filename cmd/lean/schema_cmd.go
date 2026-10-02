package lean

import (
	"fmt"
	"os"

	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage the environment schema",
}

var schemaEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit the configured schema",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("")
		if err != nil {
			fmt.Println(ui.Fail("Failed to load config: " + err.Error()))
			return
		}
		path := cfg.Schema.Path
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte("# Lean environment schema\n"), 0644); err != nil {
				fmt.Println(ui.Fail("Failed to create schema: " + err.Error()))
				return
			}
		}
		if err := openEditor(path); err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}
		fmt.Println(ui.Ok("Schema saved."))
	},
}

var schemaPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show the configured schema path",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("")
		if err != nil {
			fmt.Println(ui.Fail("Failed to load config: " + err.Error()))
			return
		}
		fmt.Println(cfg.Schema.Path)
	},
}

func init() {
	schemaCmd.AddCommand(schemaEditCmd)
	schemaCmd.AddCommand(schemaPathCmd)
	rootCmd.AddCommand(schemaCmd)
}
