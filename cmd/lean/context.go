package lean

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dominionthedev/lean/internal/backup"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/leanctx"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/dominionthedev/lean/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	ctxDesc    string
	ctxProfile string
)

var contextCmd = &cobra.Command{
	Use:     "context",
	Aliases: []string{"ctx"},
	Short:   "Manage multi-file environment contexts",
	Long: `Contexts bundle multiple files that should be applied together.

A context can map several sources onto targets:
  .env.production  → .env
  .env.secret.prod → .env.secret
  configs/prod.toml → config.toml

Examples:
  lean context create production --profile production
  lean context add production .env.secret.prod .env.secret
  lean context apply production
  lean context list
  lean context show production`,
}

var contextCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new context",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if leanctx.Exists(name) {
			fmt.Println(ui.Warn(fmt.Sprintf("Context '%s' already exists.", name)))
			return
		}

		c := &leanctx.Context{
			Name:        name,
			Description: ctxDesc,
			Profile:     ctxProfile,
			Files:       []leanctx.FileMapping{},
		}

		// If --profile is set, verify the profile file exists (warn only)
		if ctxProfile != "" {
			p := env.ProfilePath(ctxProfile)
			if _, err := os.Stat(p); err != nil {
				fmt.Println(ui.Warn(fmt.Sprintf("Profile file '%s' not found yet — add it before applying.", p)))
			}
		}

		if err := leanctx.Save(c); err != nil {
			fmt.Println(ui.Fail("Failed to create context: " + err.Error()))
			return
		}

		suffix := ""
		if ctxProfile != "" {
			suffix = fmt.Sprintf(" (profile: %s)", ctxProfile)
		}
		fmt.Println(ui.Ok(fmt.Sprintf("Context '%s' created%s.", name, suffix)))
		fmt.Println(ui.Faint("  Add files: lean context add " + name + " <source> <target>"))
		fmt.Println(ui.Faint("  Apply:     lean context apply " + name))
	},
}

var contextAddCmd = &cobra.Command{
	Use:   "add [name] [source] [target]",
	Short: "Add a file mapping to a context",
	Long: `Add a source → target mapping.

Examples:
  lean context add production .env.secret.prod .env.secret
  lean context add production configs/prod.toml config.toml`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		name, source, target := args[0], args[1], args[2]

		c, err := leanctx.Load(name)
		if err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Context '%s' not found. Create it first.", name)))
			return
		}

		if _, err := os.Stat(source); err != nil {
			fmt.Println(ui.Warn(fmt.Sprintf("Source '%s' not found on disk yet.", source)))
		}

		c.AddFile(source, target)
		if err := leanctx.Save(c); err != nil {
			fmt.Println(ui.Fail("Failed to update context: " + err.Error()))
			return
		}

		fmt.Printf("%s %s  →  %s  %s\n",
			ui.Bolt(),
			ui.Active.Render(source),
			ui.Bold.Render(target),
			ui.Faint("("+name+")"),
		)
	},
}

var contextRemoveCmd = &cobra.Command{
	Use:     "remove [name] [target]",
	Aliases: []string{"rm"},
	Short:   "Remove a file mapping from a context",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name, target := args[0], args[1]

		c, err := leanctx.Load(name)
		if err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Context '%s' not found.", name)))
			return
		}

		if !c.RemoveFile(target) {
			fmt.Println(ui.Warn(fmt.Sprintf("No mapping with target '%s' in context '%s'.", target, name)))
			return
		}

		if err := leanctx.Save(c); err != nil {
			fmt.Println(ui.Fail("Failed to update context: " + err.Error()))
			return
		}
		fmt.Println(ui.Ok(fmt.Sprintf("Removed target '%s' from context '%s'.", target, name)))
	},
}

var contextListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all contexts",
	Run: func(cmd *cobra.Command, args []string) {
		names, err := leanctx.List()
		if err != nil {
			fmt.Println(ui.Fail("Could not list contexts: " + err.Error()))
			return
		}
		if len(names) == 0 {
			fmt.Println(ui.Info("No contexts yet. Run `lean context create <name>`."))
			return
		}

		fmt.Println(ui.Bolt() + " " + ui.Bold.Render("Contexts"))
		fmt.Println()

		for _, name := range names {
			c, err := leanctx.Load(name)
			if err != nil {
				fmt.Printf("  %s %s\n", ui.Muted.Render("·"), name)
				continue
			}
			files := c.EffectiveFiles()
			extra := ui.Faint(fmt.Sprintf("(%d file%s)", len(files), plural(len(files))))
			desc := ""
			if c.Description != "" {
				desc = "  " + ui.Faint(c.Description)
			}
			fmt.Printf("  %s %s %s%s\n", ui.Muted.Render("·"), ui.Bold.Render(name), extra, desc)
		}
		fmt.Println()
	},
}

var contextShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show context details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		c, err := leanctx.Load(name)
		if err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Context '%s' not found.", name)))
			return
		}

		fmt.Printf("%s Context  %s\n\n", ui.Bolt(), ui.Bold.Render(c.Name))
		if c.Description != "" {
			fmt.Printf("  %s  %s\n", ui.Faint("description"), c.Description)
		}
		if c.Profile != "" {
			fmt.Printf("  %s      %s\n", ui.Faint("profile"), c.Profile)
		}
		fmt.Println()

		files := c.EffectiveFiles()
		if len(files) == 0 {
			fmt.Println(ui.Faint("  (no file mappings)"))
			fmt.Println(ui.Faint("  Add with: lean context add " + name + " <source> <target>"))
			return
		}

		fmt.Println("  " + ui.Faint("mappings"))
		for _, f := range files {
			srcStatus := ""
			if _, err := os.Stat(f.Source); err != nil {
				srcStatus = " " + ui.Err.Render("(missing)")
			}
			fmt.Printf("  %s  %s  →  %s%s\n",
				ui.Success.Render("→"),
				f.Source,
				ui.Bold.Render(f.Target),
				srcStatus,
			)
		}
		fmt.Println()
	},
}

var contextApplyCmd = &cobra.Command{
	Use:   "apply [name]",
	Short: "Apply a context — write all mapped files",
	Long: `Resolve and write every source → target mapping in the context.

For sources that look like env profiles (.env.*), inheritance is resolved
before writing. Other files are copied as-is.

A backup of the current .env is taken when .env is among the targets.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		c, err := leanctx.Load(name)
		if err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Context '%s' not found.", name)))
			return
		}

		files := c.EffectiveFiles()
		if len(files) == 0 {
			fmt.Println(ui.Fail("Context has no file mappings. Add some with `lean context add`."))
			return
		}

		// Validate sources exist
		var missing []string
		for _, f := range files {
			if _, err := os.Stat(f.Source); err != nil {
				missing = append(missing, f.Source)
			}
		}
		if len(missing) > 0 {
			fmt.Println(ui.Fail("Missing source files:"))
			for _, m := range missing {
				fmt.Printf("  %s %s\n", ui.Err.Render("✗"), m)
			}
			return
		}

		engine, _ := core.NewEngine()
		prev := ""
		if engine != nil {
			prev = engine.State.Current
		}

		// Backup .env if it's a target
		for _, f := range files {
			if f.Target == ".env" {
				if err := backup.Snapshot(prev); err != nil {
					fmt.Println(ui.Warn("Could not snapshot current .env: " + err.Error()))
				}
				break
			}
		}

		applied := 0
		for _, f := range files {
			if err := applyMapping(f); err != nil {
				fmt.Println(ui.Fail(fmt.Sprintf("%s → %s: %s", f.Source, f.Target, err)))
				return
			}
			fmt.Printf("  %s %s  →  %s\n", ui.Success.Render("✓"), f.Source, ui.Bold.Render(f.Target))
			applied++
		}

		// Update lean state if .env was written from a profile
		if engine != nil {
			profileName := c.Profile
			if profileName == "" {
				// Infer from first .env.* source targeting .env
				for _, f := range files {
					if f.Target == ".env" && strings.HasPrefix(f.Source, ".env.") {
						profileName = strings.TrimPrefix(f.Source, ".env.")
						break
					}
				}
			}
			if profileName != "" {
				if !engine.ProfileExists(profileName) {
					_ = engine.AddProfile(profileName)
				}
				_ = engine.SetCurrent(profileName)
				if cwd, err := os.Getwd(); err == nil {
					_ = workspace.Remember(cwd, profileName)
				}
			}
		}

		fmt.Println()
		fmt.Println(ui.Ok(fmt.Sprintf("Context '%s' applied (%d file%s).", name, applied, plural(applied))))
	},
}

var contextDeleteCmd = &cobra.Command{
	Use:     "delete [name]",
	Aliases: []string{"rm-context"},
	Short:   "Delete a context definition",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if !leanctx.Exists(name) {
			fmt.Println(ui.Fail(fmt.Sprintf("Context '%s' not found.", name)))
			return
		}
		if err := leanctx.Delete(name); err != nil {
			fmt.Println(ui.Fail("Delete failed: " + err.Error()))
			return
		}
		fmt.Println(ui.Ok(fmt.Sprintf("Context '%s' deleted.", name)))
	},
}

// applyMapping writes one source → target.
// Env-profile sources are resolved through inheritance; everything else is copied.
func applyMapping(f leanctx.FileMapping) error {
	// Ensure target directory exists
	if dir := filepath.Dir(f.Target); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	base := filepath.Base(f.Source)
	isEnvProfile := strings.HasPrefix(base, ".env.") &&
		base != ".env.template" && base != ".env.example" && base != ".env.schema"

	if isEnvProfile && f.Target == ".env" {
		profile := strings.TrimPrefix(base, ".env.")
		resolved, err := env.Resolve(profile)
		if err != nil {
			return err
		}
		return resolved.Write(f.Target)
	}

	// Plain copy (atomic)
	data, err := os.ReadFile(f.Source)
	if err != nil {
		return err
	}
	tmp := f.Target + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, f.Target)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func init() {
	contextCreateCmd.Flags().StringVar(&ctxDesc, "description", "", "Context description")
	contextCreateCmd.Flags().StringVar(&ctxProfile, "profile", "", "Primary env profile (maps to .env)")

	contextCmd.AddCommand(contextCreateCmd)
	contextCmd.AddCommand(contextAddCmd)
	contextCmd.AddCommand(contextRemoveCmd)
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextShowCmd)
	contextCmd.AddCommand(contextApplyCmd)
	contextCmd.AddCommand(contextDeleteCmd)
}
