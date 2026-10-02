package lean

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var manOutDir string
var manGenerate bool

var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Generate or view man pages",
	Long: `Generate man pages for lean and its subcommands, or open the system man page.

Examples:
  lean man                  # open 'man lean' if installed
  lean man --generate       # write man pages to ./man
  lean man --generate -o /usr/local/share/man/man1`,
	Run: func(cmd *cobra.Command, args []string) {
		if manGenerate || manOutDir != "" {
			generateMan()
			return
		}
		c := exec.Command("man", "lean")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		if err := c.Run(); err != nil {
			fmt.Println(ui.Info("No system man page found. Generate one with:"))
			fmt.Println(ui.Faint("  lean man --generate"))
			fmt.Println(ui.Faint("  sudo cp man/*.1 /usr/local/share/man/man1/"))
		}
	},
}

func generateMan() {
	dir := manOutDir
	if dir == "" {
		dir = "man"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Println(ui.Fail("Cannot create man dir: " + err.Error()))
		return
	}
	header := &doc.GenManHeader{
		Title:   "LEAN",
		Section: "1",
		Source:  "lean",
		Manual:  "lean Manual",
	}
	if err := doc.GenManTree(rootCmd, header, dir); err != nil {
		fmt.Println(ui.Fail("Failed to generate man pages: " + err.Error()))
		return
	}
	count := 0
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".1" {
			count++
		}
	}
	fmt.Println(ui.Ok(fmt.Sprintf("Wrote %d man page(s) to %s/", count, dir)))
	fmt.Println(ui.Faint("  Install: sudo cp " + dir + "/*.1 /usr/local/share/man/man1/ && sudo mandb"))
}

func init() {
	manCmd.Flags().BoolVar(&manGenerate, "generate", false, "Generate man pages instead of opening them")
	manCmd.Flags().StringVarP(&manOutDir, "output", "o", "", "Output directory for generated man pages (implies --generate)")
}
