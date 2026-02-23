package lean

import (
	"fmt"
	"os"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/spf13/cobra"
)

var profileName string

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new environment profile",
	Run: func(cmd *cobra.Command, args []string) {

		if profileName == "" && len(args) > 0 {
			profileName = args[0]
		}

		if profileName == "" {
			fmt.Println("Profile name required.")
			return
		}

		engine, err := core.NewEngine()
		if err != nil {
			fmt.Println("Not initialized. Run `lean init`.")
			return
		}

		err = engine.AddProfile(profileName)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		envFile := ".env." + profileName
		os.WriteFile(envFile, []byte(""), 0644)

		fmt.Printf("⚡️ Profile '%s' created.\n", profileName)
	},
}

func init() {
	createCmd.Flags().StringVarP(&profileName, "name", "n", "", "Profile name")
}
