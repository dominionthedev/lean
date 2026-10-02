package lean

import (
	"fmt"
	"os"
	"sort"

	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/schema"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var validateSchemaFlag string

var validateCmd = &cobra.Command{
	Use:   "validate [profile]",
	Short: "Validate a profile against the schema",
	Long: `Validate a profile against schema rules.

Advanced schema (.lean/schema.toml via lean.toml):
  required, values (enum), same_as, depends_on,
  required_when, deactivated_when, from, secret

Legacy fallback (key presence only):
  --schema, .env.schema, .env.example, .env.template

Examples:
  lean validate production
  lean validate staging
  lean validate --schema .env.example`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("")
		if err != nil {
			fmt.Println(ui.Fail("Config error: " + err.Error()))
			return
		}

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
			profile = cfg.Lean.DefaultProfile
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

		var values map[string]string
		if profile == engine.State.Current {
			if _, err := os.Stat(".env"); err == nil {
				if f, err := env.Parse(".env"); err == nil {
					values = f.ToMap()
				}
			}
		}
		if values == nil {
			resolved, err := env.Resolve(profile)
			if err != nil {
				fmt.Println(ui.Fail("Failed to resolve profile: " + err.Error()))
				return
			}
			values = resolved.ToMap()
		}

		if validateSchemaFlag == "" {
			if _, err := os.Stat(cfg.Schema.Path); err == nil {
				runAdvancedValidate(profile, cfg.Schema.Path, values)
				return
			}
		}

		runLegacyValidate(profile, values)
	},
}

func runAdvancedValidate(profile, schemaPath string, values map[string]string) {
	sch, err := schema.Load(schemaPath)
	if err != nil {
		fmt.Println(ui.Fail(fmt.Sprintf("Cannot load schema '%s': %s", schemaPath, err)))
		return
	}

	enriched := sch.ApplyDefaults(values)
	issues := sch.Validate(enriched)

	fmt.Printf("%s Validate  %s  against  %s\n\n",
		ui.Bolt(),
		ui.Bold.Render(profile),
		ui.Faint(schemaPath),
	)

	if len(issues) == 0 {
		req := sch.RequiredKeys()
		sort.Strings(req)
		for _, k := range req {
			fmt.Printf("  %s %s\n", ui.Success.Render("✓"), k)
		}
		fmt.Println()
		fmt.Println(ui.Ok("All schema rules passed."))
		return
	}

	errors, warns, infos := 0, 0, 0
	for _, iss := range issues {
		switch iss.Level {
		case "error":
			fmt.Printf("  %s %s  %s\n", ui.Err.Render("✗"), iss.Key, ui.Faint(iss.Message))
			errors++
		case "warn":
			fmt.Printf("  %s %s  %s\n", ui.Warning.Render("!"), iss.Key, ui.Faint(iss.Message))
			warns++
		default:
			fmt.Printf("  %s %s  %s\n", ui.Muted.Render("·"), iss.Key, ui.Faint(iss.Message))
			infos++
		}
	}

	fmt.Println()
	if errors > 0 {
		fmt.Println(ui.Fail(fmt.Sprintf("%d error(s), %d warning(s).", errors, warns)))
		os.Exit(1)
	}
	if warns > 0 {
		fmt.Println(ui.Warn(fmt.Sprintf("%d warning(s), %d info.", warns, infos)))
		return
	}
	fmt.Println(ui.Ok("Passed with info notes."))
}

func runLegacyValidate(profile string, values map[string]string) {
	schemaPath := validateSchemaFlag
	if schemaPath == "" {
		for _, candidate := range []string{".env.schema", ".env.example", ".env.template"} {
			if _, err := os.Stat(candidate); err == nil {
				schemaPath = candidate
				break
			}
		}
	}
	if schemaPath == "" {
		fmt.Println(ui.Fail("No schema found. Add .lean/schema.toml, or .env.schema / .env.example / .env.template"))
		return
	}

	schemaFile, err := env.Parse(schemaPath)
	if err != nil {
		fmt.Println(ui.Fail(fmt.Sprintf("Cannot read schema '%s': %s", schemaPath, err)))
		return
	}
	required := schemaFile.Keys()
	if len(required) == 0 {
		fmt.Println(ui.Warn(fmt.Sprintf("Schema '%s' has no keys.", schemaPath)))
		return
	}

	sort.Strings(required)

	var missing []string
	var ok []string
	for _, k := range required {
		if _, present := values[k]; present {
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
}

func init() {
	validateCmd.Flags().StringVar(&validateSchemaFlag, "schema", "", "Legacy schema file (key presence only)")
}
