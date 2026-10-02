package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dominionthedev/lean/internal/globaldir"
)

// Config is the global workspace map (cwd → profile).
type Config struct {
	Workspaces map[string]string `json:"workspaces"`
}

func configPath() (string, error) {
	// Prefer ~/.lean/workspaces.json; fall back to legacy XDG path for migration.
	p, err := globaldir.WorkspacesPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}
	// legacy
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		legacy := filepath.Join(xdg, "lean", "config.json")
		if _, err := os.Stat(legacy); err == nil {
			return legacy, nil
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		legacy := filepath.Join(home, ".config", "lean", "config.json")
		if _, err := os.Stat(legacy); err == nil {
			return legacy, nil
		}
	}
	return p, nil
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
	path, err := globaldir.WorkspacesPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

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
