package lean

import (
	"fmt"
	"os"
	"sort"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var validateSchema string

var validateCmd = &cobra.Command{
	Use:   "validate [profile]",
	Short: "Check that required keys exist in a profile",
	Long: `Validate a profile against a schema of required keys.

The schema is loaded from (first match wins):
  1. --schema <file>
  2. .env.schema
  3. .env.example
  4. .env.template

Only key presence is checked — values may be empty.

Examples:
  lean validate production
  lean validate staging --schema .env.schema
  lean validate current`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		profile := ""
		if len(args) > 0 {
			profile = args[0]
		} else {
			profile = engine.State.Current
		}
		if profile == "" {
			fmt.Println(ui.Fail("No profile specified and no active profile set."))
			return
		}
		if profile == "current" {
			profile = engine.State.Current
			if profile == "" {
				fmt.Println(ui.Fail("No active profile."))
				return
			}
		}

		// Load schema
		schemaPath := validateSchema
		if schemaPath == "" {
			for _, candidate := range []string{".env.schema", ".env.example", ".env.template"} {
				if _, err := os.Stat(candidate); err == nil {
					schemaPath = candidate
					break
				}
			}
		}
		if schemaPath == "" {
			fmt.Println(ui.Fail("No schema found. Provide --schema, or add .env.schema / .env.example / .env.template"))
			return
		}

		schema, err := env.Parse(schemaPath)
		if err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Cannot read schema '%s': %s", schemaPath, err)))
			return
		}
		required := schema.Keys()
		if len(required) == 0 {
			fmt.Println(ui.Warn(fmt.Sprintf("Schema '%s' has no keys.", schemaPath)))
			return
		}

		// Resolve the profile (with inheritance)
		var resolved *env.File
		if profile == engine.State.Current {
			// Prefer live .env when validating the active profile
			if _, err := os.Stat(".env"); err == nil {
				resolved, err = env.Parse(".env")
				if err != nil {
					fmt.Println(ui.Fail("Cannot read .env: " + err.Error()))
					return
				}
			}
		}
		if resolved == nil {
			resolved, err = env.Resolve(profile)
			if err != nil {
				fmt.Println(ui.Fail("Failed to resolve profile: " + err.Error()))
				return
			}
		}

		present := make(map[string]bool)
		for _, k := range resolved.Keys() {
			present[k] = true
		}

		sort.Strings(required)

		var missing []string
		var ok []string
		for _, k := range required {
			if present[k] {
				ok = append(ok, k)
			} else {
				missing = append(missing, k)
			}
		}

		fmt.Printf("%s Validate  %s  against  %s\n\n",
			ui.Bolt(),
			ui.Bold.Render(profile),
			ui.Faint(schemaPath),
		)

		for _, k := range ok {
			fmt.Printf("  %s %s\n", ui.Success.Render("✓"), k)
		}
		for _, k := range missing {
			fmt.Printf("  %s %s %s\n", ui.Err.Render("✗"), k, ui.Faint("missing"))
		}

		fmt.Println()
		if len(missing) == 0 {
			fmt.Println(ui.Ok(fmt.Sprintf("All %d required keys present.", len(required))))
		} else {
			fmt.Println(ui.Fail(fmt.Sprintf("%d of %d keys missing.", len(missing), len(required))))
			os.Exit(1)
		}
	},
}

func init() {
	validateCmd.Flags().StringVar(&validateSchema, "schema", "", "Schema file (default: .env.schema, .env.example, or .env.template)")
}
