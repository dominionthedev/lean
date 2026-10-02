package lean

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/secrets"
	"github.com/dominionthedev/lean/internal/ui"
)

func configureSecurity(cfg *config.Config) error {
	var backend string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("Secrets backend").Description("Choose how Lean encrypts stored secrets.").Options(
			huh.NewOption("Local — AES-256-GCM with a Lean master key", "local"),
			huh.NewOption("GPG — use your GPG keyring", "gpg"),
			huh.NewOption("age — use an age identity", "age"),
			huh.NewOption("SSH — use an SSH key through age", "ssh"),
		).Value(&backend),
	))
	if err := form.Run(); err != nil {
		return err
	}

	cfg.Secrets = config.SecretsSection{Backend: backend}
	switch backend {
	case "local":
		if secrets.KeyExists() {
			useExisting := true
			form := huh.NewForm(huh.NewGroup(
				huh.NewConfirm().Title("Use the existing ~/.lean/key?").Value(&useExisting),
			))
			if err := form.Run(); err != nil {
				return err
			}
			if useExisting {
				return nil
			}
		}
		path, err := secrets.Keygen()
		if err != nil {
			return err
		}
		fmt.Println(ui.Ok("Created local Lean master key at " + path))

	case "gpg":
		return configureGPG(&cfg.Secrets)
	case "age":
		return configureAge(&cfg.Secrets)
	case "ssh":
		return configureSSH(&cfg.Secrets)
	}
	return nil
}

func configureGPG(dst *config.SecretsSection) error {
	var action string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("GPG setup").Options(
			huh.NewOption("Use an existing GPG identity", "existing"),
			huh.NewOption("Generate a Lean GPG key", "generate"),
		).Value(&action),
	))
	if err := form.Run(); err != nil {
		return err
	}

	if action == "existing" {
		return huh.NewInput().Title("GPG recipient or key ID").Value(&dst.Recipient).Run()
	}

	var identity string
	if err := huh.NewInput().Title("Identity email/name").Description("Lean uses GPG's default key parameters.").Value(&identity).Run(); err != nil {
		return err
	}
	if identity == "" {
		return fmt.Errorf("GPG identity cannot be empty")
	}
	cmd := exec.Command("gpg", "--quick-generate-key", identity, "default", "default", "0")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gpg key generation failed: %w", err)
	}
	dst.Recipient = identity
	return nil
}

func configureAge(dst *config.SecretsSection) error {
	var action string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("age setup").Options(
			huh.NewOption("Use an existing age identity", "existing"),
			huh.NewOption("Generate a new age identity", "generate"),
		).Value(&action),
	))
	if err := form.Run(); err != nil {
		return err
	}
	if action == "existing" {
		return huh.NewInput().Title("age recipient").Value(&dst.Recipient).Run()
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".config", "age", "lean-keys.txt")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	cmd := exec.Command("age-keygen", "-o", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("age-keygen failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	identity, err := findAgeRecipient(path)
	if err != nil {
		return err
	}
	dst.Identity = path
	dst.Recipient = identity
	fmt.Println(ui.Ok("Created age identity at " + path))
	return nil
}

func configureSSH(dst *config.SecretsSection) error {
	var action string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("SSH setup").Options(
			huh.NewOption("Use an existing SSH key", "existing"),
			huh.NewOption("Generate a dedicated Lean SSH key", "generate"),
		).Value(&action),
	))
	if err := form.Run(); err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	private := filepath.Join(home, ".ssh", "lean_ed25519")
	if action == "existing" {
		if err := huh.NewInput().Title("SSH private key path").Value(&private).Run(); err != nil {
			return err
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(private), 0700); err != nil {
			return err
		}
		cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", private, "-N", "", "-C", "lean")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("ssh-keygen failed: %w", err)
		}
	}
	dst.Identity = private
	dst.Recipient = private + ".pub"
	return nil
}

func findAgeRecipient(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# public key:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# public key:")), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("age-keygen did not produce a public recipient")
}
