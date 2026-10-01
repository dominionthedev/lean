package lean

import (
	"fmt"
	"strings"
	"time"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var (
	metaDesc   string
	metaAuthor string
	metaTags   string
)

var metaCmd = &cobra.Command{
	Use:   "meta [profile]",
	Short: "View or set profile metadata",
	Long: `Show or update metadata for a profile (description, author, tags).

Examples:
  lean meta production
  lean meta production --description "Main production API" --author DominionDev --tags aws,production
  lean meta                          # active profile`,
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

		// Setting?
		setting := cmd.Flags().Changed("description") ||
			cmd.Flags().Changed("author") ||
			cmd.Flags().Changed("tags")

		if setting {
			existing, _ := engine.GetMeta(profile)
			if cmd.Flags().Changed("description") {
				existing.Description = metaDesc
			}
			if cmd.Flags().Changed("author") {
				existing.Author = metaAuthor
			}
			if cmd.Flags().Changed("tags") {
				if metaTags == "" {
					existing.Tags = nil
				} else {
					parts := strings.Split(metaTags, ",")
					var tags []string
					for _, t := range parts {
						t = strings.TrimSpace(t)
						if t != "" {
							tags = append(tags, t)
						}
					}
					existing.Tags = tags
				}
			}
			if existing.Created == "" {
				existing.Created = time.Now().Format("2006-01-02")
			}
			if err := engine.SetMeta(profile, existing); err != nil {
				fmt.Println(ui.Fail("Failed to save metadata: " + err.Error()))
				return
			}
			fmt.Println(ui.Ok(fmt.Sprintf("Metadata updated for '%s'.", profile)))
		}

		// Always show current meta
		m, ok := engine.GetMeta(profile)
		fmt.Printf("%s Profile  %s\n\n", ui.Bolt(), ui.Bold.Render(profile))
		if !ok || (m.Description == "" && m.Author == "" && len(m.Tags) == 0 && m.Created == "") {
			fmt.Println(ui.Faint("  (no metadata yet)"))
			fmt.Println()
			fmt.Println(ui.Faint("  Set with: lean meta " + profile + " --description \"...\" --tags a,b"))
			return
		}
		if m.Description != "" {
			fmt.Printf("  %s  %s\n", ui.Faint("description"), m.Description)
		}
		if m.Author != "" {
			fmt.Printf("  %s       %s\n", ui.Faint("author"), m.Author)
		}
		if len(m.Tags) > 0 {
			fmt.Printf("  %s         %s\n", ui.Faint("tags"), strings.Join(m.Tags, ", "))
		}
		if m.Created != "" {
			fmt.Printf("  %s      %s\n", ui.Faint("created"), m.Created)
		}
		fmt.Println()
	},
}

func init() {
	metaCmd.Flags().StringVar(&metaDesc, "description", "", "Profile description")
	metaCmd.Flags().StringVar(&metaAuthor, "author", "", "Author name")
	metaCmd.Flags().StringVar(&metaTags, "tags", "", "Comma-separated tags")
}
