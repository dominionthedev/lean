package lean

import (
	"fmt"
	"os"

	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/schema"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var fillProfile string
var fillDryRun bool

var fillCmd = &cobra.Command{
	Use:   "fill [profile]",
	Short: "Fill missing keys from schema defaults, same_as, and from: sources",
	Long: `Apply schema rules to fill in missing values:

  • default       — static default value
  • same_as       — copy from another key
  • from:command  — run a shell command
  • from:file     — read a file
  • from:env      — read a process environment variable

Existing values are never overwritten.

Examples:
  lean fill
  lean fill production
  lean fill --dry-run`,
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

		_ = engine.ScanDisk()

		profile := fillProfile
		if len(args) > 0 {
			profile = args[0]
		}
		if profile == "" {
			profile = engine.State.Current
		}
		if profile == "" {
			profile = cfg.Lean.DefaultProfile
		}
		if profile == "" {
			fmt.Println(ui.Fail("No profile specified."))
			return
		}

		sch, err := schema.Load(cfg.Schema.Path)
		if err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Cannot load schema %s: %s", cfg.Schema.Path, err)))
			fmt.Println(ui.Faint("  Create .lean/schema.toml or set schema.path in lean.toml"))
			return
		}

		resolved, err := env.Resolve(profile)
		if err != nil {
			fmt.Println(ui.Fail("Failed to resolve profile: " + err.Error()))
			return
		}

		values := resolved.ToMap()

		// Defaults + same_as
		withDefaults := sch.ApplyDefaults(values)

		// from: sources
		filled, fromIssues := sch.ResolveFrom(withDefaults)
		for k, v := range filled {
			withDefaults[k] = v
		}

		// Diff: what actually changed
		var changes []string
		for k, v := range withDefaults {
			old, ok := values[k]
			if !ok || old != v {
				changes = append(changes, k)
				resolved.Set(k, v)
			}
		}

		fmt.Printf("%s Fill  %s\n\n", ui.Bolt(), ui.Bold.Render(profile))

		for _, issue := range fromIssues {
			fmt.Printf("  %s %s: %s\n", ui.Err.Render("✗"), issue.Key, issue.Message)
		}

		if len(changes) == 0 {
			fmt.Println(ui.Info("Nothing to fill — all schema keys already set."))
			return
		}

		for _, k := range changes {
			fmt.Printf("  %s %s=%s\n", ui.Success.Render("✓"), k, maskSecret(k, withDefaults[k], sch))
		}

		if fillDryRun {
			fmt.Println()
			fmt.Println(ui.Faint("Dry run — no files written."))
			return
		}

		path := env.ProfilePath(profile)
		if err := resolved.Write(path); err != nil {
			fmt.Println(ui.Fail("Failed to write profile: " + err.Error()))
			return
		}

		// Sync .env if this is the active profile
		if profile == engine.State.Current {
			_ = resolved.Write(".env")
		}

		fmt.Println()
		fmt.Println(ui.Ok(fmt.Sprintf("Filled %d key(s) in %s.", len(changes), path)))
	},
}

func maskSecret(key, val string, sch *schema.Schema) string {
	if rule, ok := sch.Keys[key]; ok && rule.Secret {
		if len(val) <= 4 {
			return "****"
		}
		return val[:2] + "…" + val[len(val)-2:]
	}
	return val
}

func init() {
	fillCmd.Flags().StringVarP(&fillProfile, "profile", "p", "", "Target profile")
	fillCmd.Flags().BoolVar(&fillDryRun, "dry-run", false, "Show what would be filled without writing")
	_ = os.Stderr
}
