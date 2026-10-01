package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is the global lean config stored under the user's config directory.
type Config struct {
	Workspaces map[string]string `json:"workspaces"` // abs path → profile name
}

func configDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "lean"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "lean"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return &Config{Workspaces: map[string]string{}}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Workspaces: map[string]string{}}, nil
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if c.Workspaces == nil {
		c.Workspaces = map[string]string{}
	}
	return &c, nil
}

func Save(c *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Remember records that this directory prefers the given profile.
func Remember(dir, profile string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	c, err := Load()
	if err != nil {
		return err
	}
	c.Workspaces[abs] = profile
	return Save(c)
}

// Lookup returns the remembered profile for a directory, if any.
func Lookup(dir string) (string, bool) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", false
	}
	c, err := Load()
	if err != nil {
		return "", false
	}
	p, ok := c.Workspaces[abs]
	return p, ok
}

// Forget removes the workspace mapping for a directory.
func Forget(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	c, err := Load()
	if err != nil {
		return err
	}
	delete(c.Workspaces, abs)
	return Save(c)
}
