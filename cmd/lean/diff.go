package lean

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [left] [right]",
	Short: "Show differences between two profiles",
	Long: `Compare two environment profiles after resolving inheritance.

Examples:
  lean diff dev prod
  lean diff current prod
  lean diff staging production`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		leftName := args[0]
		rightName := args[1]

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println(ui.Fail("Not initialized. Run `lean init` first."))
			return
		}

		// Resolve "current" alias
		if leftName == "current" {
			if engine.State.Current == "" {
				fmt.Println(ui.Fail("No active profile. Apply one first or pass an explicit name."))
				return
			}
			leftName = engine.State.Current
		}
		if rightName == "current" {
			if engine.State.Current == "" {
				fmt.Println(ui.Fail("No active profile. Apply one first or pass an explicit name."))
				return
			}
			rightName = engine.State.Current
		}

		leftFile, err := resolveProfile(leftName)
		if err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}
		rightFile, err := resolveProfile(rightName)
		if err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}

		diffs := env.Diff(leftFile.ToMap(), rightFile.ToMap())
		if len(diffs) == 0 {
			fmt.Println(ui.Ok(fmt.Sprintf("%s and %s are identical.", leftName, rightName)))
			return
		}

		// Stable sort by key
		sort.Slice(diffs, func(i, j int) bool {
			return diffs[i].Key < diffs[j].Key
		})

		fmt.Printf("%s Diff  %s  ↔  %s\n\n",
			ui.Bolt(),
			ui.Bold.Render(leftName),
			ui.Bold.Render(rightName),
		)

		for _, d := range diffs {
			switch d.Status {
			case "changed":
				fmt.Printf("%s:\n", ui.Active.Render(d.Key))
				fmt.Printf("  %s: %s\n", ui.Faint(leftName), mask(d.Left))
				fmt.Printf("  %s: %s\n", ui.Faint(rightName), mask(d.Right))
			case "only-left":
				fmt.Printf("%s:\n", ui.Warning.Render(d.Key))
				fmt.Printf("  %s: %s\n", ui.Faint(leftName), mask(d.Left))
				fmt.Printf("  %s: %s\n", ui.Faint(rightName), ui.Muted.Render("(absent)"))
			case "only-right":
				fmt.Printf("%s:\n", ui.Warning.Render(d.Key))
				fmt.Printf("  %s: %s\n", ui.Faint(leftName), ui.Muted.Render("(absent)"))
				fmt.Printf("  %s: %s\n", ui.Faint(rightName), mask(d.Right))
			}
			fmt.Println()
		}

		fmt.Printf("%s %d difference(s)\n", ui.Bolt(), len(diffs))
	},
}

func resolveProfile(name string) (*env.File, error) {
	path := env.ProfilePath(name)
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("profile '%s' not found (%s)", name, path)
	}
	// For "current" / .env there is no inheritance to resolve beyond the file itself
	if name == "current" || path == ".env" {
		return env.Parse(path)
	}
	return env.Resolve(name)
}

// mask hides values that look secret-like for safer terminal output.
func mask(v string) string {
	lower := strings.ToLower(v)
	if len(v) > 24 && (strings.Contains(lower, "secret") ||
		strings.Contains(lower, "password") ||
		strings.Contains(lower, "token") ||
		strings.Contains(lower, "key")) {
		return v[:4] + "…" + v[len(v)-4:]
	}
	return v
}
