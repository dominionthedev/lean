package globaldir

import (
	"os"
	"path/filepath"
)

// Dir returns ~/.lean (mode 0700).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".lean")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

func Path(elem ...string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{dir}, elem...)...), nil
}

func KeyPath() (string, error)       { return Path("key") }
func ConfigPath() (string, error)    { return Path("config.toml") }
func WorkspacesPath() (string, error) { return Path("workspaces.json") }

func TemplatesDir() (string, error) {
	dir, err := Path("templates")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}
