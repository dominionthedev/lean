package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

const DefaultPath = "lean.toml"

// Config is the project-level lean.toml configuration.
type Config struct {
	Lean    LeanSection    `toml:"lean"`
	Schema  SchemaSection  `toml:"schema"`
	Secrets SecretsSection `toml:"secrets"`
}

type LeanSection struct {
	Version        int    `toml:"version"`
	DefaultProfile string `toml:"default_profile"`
}

type SchemaSection struct {
	Path string `toml:"path"`
}

type SecretsSection struct {
	Backend       string `toml:"backend"`         // local | gpg | age | ssh
	Recipient     string `toml:"recipient"`       // gpg/age/ssh recipient
	Identity      string `toml:"identity"`        // age identity / ssh private key
	MasterKeyEnv  string `toml:"master_key_env"`  // env var name (default LEAN_MASTER_KEY)
	MasterKeyFile string `toml:"master_key_file"` // optional path; default ~/.lean/key
}

func Default() *Config {
	return &Config{
		Lean:    LeanSection{Version: 1},
		Schema:  SchemaSection{Path: ".lean/schema.toml"},
		Secrets: SecretsSection{Backend: "local"},
	}
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}
	cfg := Default()
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Schema.Path == "" {
		cfg.Schema.Path = ".lean/schema.toml"
	}
	if cfg.Secrets.Backend == "" {
		cfg.Secrets.Backend = "local"
	}
	if cfg.Lean.Version == 0 {
		cfg.Lean.Version = 1
	}
	return cfg, nil
}

func Save(path string, cfg *Config) error {
	if path == "" {
		path = DefaultPath
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func Exists(path string) bool {
	if path == "" {
		path = DefaultPath
	}
	_, err := os.Stat(path)
	return err == nil
}
