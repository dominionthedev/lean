package lean

import (
	"fmt"
	"os"
	"strings"

	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/secrets"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var secretProfile string

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Encrypt and store secrets in .lean/secrets/",
	Long: `Manage encrypted secrets stored under .lean/secrets/<profile>/.

Backends (lean.toml [secrets]):
  local  — AES-256-GCM
  gpg    — gpg -e -r <recipient>
  age    — age -r <recipient>
  ssh    — age with an SSH public key (same as age)

Master key for local backend:
  ~/.lean/key  (create with: lean secret keygen)
  or secrets.master_key_file in lean.toml

Examples:
  lean secret keygen
  lean secret put JWT_SECRET=supersecret
  lean secret get JWT_SECRET
  lean secret inject`,
}

var secretPutCmd = &cobra.Command{
	Use:   "put KEY=VALUE",
	Short: "Encrypt and store a secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		raw := args[0]
		idx := strings.Index(raw, "=")
		if idx < 0 {
			fmt.Println(ui.Fail("Expected KEY=VALUE"))
			return
		}
		key, value := strings.TrimSpace(raw[:idx]), raw[idx+1:]
		if key == "" {
			fmt.Println(ui.Fail("Key cannot be empty."))
			return
		}

		profile, store, err := resolveSecretStore()
		if err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}

		if err := store.Put(profile, key, value); err != nil {
			fmt.Println(ui.Fail("Failed to store secret: " + err.Error()))
			return
		}

		fmt.Printf("%s Stored %s in %s  %s\n",
			ui.Bolt(),
			ui.Active.Render(key),
			ui.Bold.Render(profile),
			ui.Faint("(hash "+secrets.Hash(value)+")"),
		)
	},
}

var secretGetCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Decrypt and print a secret (plain output)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		profile, store, err := resolveSecretStore()
		if err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}

		val, err := store.Get(profile, key)
		if err != nil {
			fmt.Println(ui.Fail(fmt.Sprintf("Cannot read secret '%s': %s", key, err)))
			os.Exit(1)
		}
		fmt.Println(val)
	},
}

var secretListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored secret keys for a profile",
	Run: func(cmd *cobra.Command, args []string) {
		profile, store, err := resolveSecretStore()
		if err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}

		keys, err := store.List(profile)
		if err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}
		if len(keys) == 0 {
			fmt.Println(ui.Info(fmt.Sprintf("No secrets stored for '%s'.", profile)))
			return
		}

		fmt.Printf("%s Secrets  %s\n\n", ui.Bolt(), ui.Bold.Render(profile))
		for _, k := range keys {
			fmt.Printf("  %s %s\n", ui.Success.Render("◆"), k)
		}
		fmt.Println()
	},
}

var secretDeleteCmd = &cobra.Command{
	Use:     "delete KEY",
	Aliases: []string{"rm"},
	Short:   "Delete a stored secret",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		profile, store, err := resolveSecretStore()
		if err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}
		if !store.Exists(profile, key) {
			fmt.Println(ui.Fail(fmt.Sprintf("Secret '%s' not found in '%s'.", key, profile)))
			return
		}
		if err := store.Delete(profile, key); err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}
		fmt.Println(ui.Ok(fmt.Sprintf("Deleted secret '%s' from '%s'.", key, profile)))
	},
}

var secretInjectCmd = &cobra.Command{
	Use:   "inject",
	Short: "Inject stored secrets into the active .env",
	Long: `Decrypt all secrets for the profile and write them into .env
(or the profile file). Existing keys are overwritten.`,
	Run: func(cmd *cobra.Command, args []string) {
		profile, store, err := resolveSecretStore()
		if err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}

		keys, err := store.List(profile)
		if err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}
		if len(keys) == 0 {
			fmt.Println(ui.Info("No secrets to inject."))
			return
		}

		// Write into .env if it exists, else into profile file
		target := ".env"
		if _, err := os.Stat(target); err != nil {
			target = ".env." + profile
		}

		f, err := openEnvFile(target)
		if err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}

		for _, k := range keys {
			val, err := store.Get(profile, k)
			if err != nil {
				fmt.Println(ui.Warn(fmt.Sprintf("Skip %s: %s", k, err)))
				continue
			}
			f.Set(k, val)
			fmt.Printf("  %s %s\n", ui.Success.Render("✓"), k)
		}

		if err := f.Write(target); err != nil {
			fmt.Println(ui.Fail("Failed to write: " + err.Error()))
			return
		}
		fmt.Println(ui.Ok(fmt.Sprintf("Injected %d secret(s) into %s.", len(keys), target)))
	},
}


var secretKeygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate ~/.lean/key (mode 0600) for the local secrets backend",
	Long: `Creates a random 256-bit master key at ~/.lean/key with permissions 0600.

This is the recommended way to set up the local backend — no shell export needed.
The directory ~/.lean is created with mode 0700.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, err := secrets.Keygen()
		if err != nil {
			fmt.Println(ui.Fail(err.Error()))
			return
		}
		fmt.Println(ui.Ok(fmt.Sprintf("Master key written to %s (mode 0600)", path)))
		fmt.Println(ui.Faint("  local backend will use this automatically."))
	},
}

func resolveSecretStore() (string, *secrets.Store, error) {
	cfg, err := config.Load("")
	if err != nil {
		return "", nil, fmt.Errorf("config: %w", err)
	}

	profile := secretProfile
	if profile == "" {
		engine, err := core.NewEngine()
		if err != nil {
			return "", nil, fmt.Errorf("not initialized — run lean init first")
		}
		profile = engine.State.Current
		if profile == "" {
			profile = cfg.Lean.DefaultProfile
		}
		if profile == "" {
			return "", nil, fmt.Errorf("no profile specified; use --profile or apply a profile first")
		}
	}

	return profile, secrets.NewStore(cfg), nil
}

func init() {
	for _, c := range []*cobra.Command{secretPutCmd, secretGetCmd, secretListCmd, secretDeleteCmd, secretInjectCmd} {
		c.Flags().StringVarP(&secretProfile, "profile", "p", "", "Target profile (default: active)")
	}

	secretCmd.AddCommand(secretPutCmd)
	secretCmd.AddCommand(secretGetCmd)
	secretCmd.AddCommand(secretListCmd)
	secretCmd.AddCommand(secretDeleteCmd)
	secretCmd.AddCommand(secretInjectCmd)
	secretCmd.AddCommand(secretKeygenCmd)
}

func openEnvFile(path string) (*env.File, error) {
	if _, err := os.Stat(path); err == nil {
		return env.Parse(path)
	}
	return &env.File{Path: path}, nil
}
