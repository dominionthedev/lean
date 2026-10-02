package config

import (
	"os"
	"path/filepath"

	"github.com/dominionthedev/lean/internal/globaldir"
	"github.com/pelletier/go-toml/v2"
)

const DefaultPath = ".lean/lean.toml"

const legacyGlobalPath = "lean.toml"

type Config struct {
	Lean    LeanSection    `toml:"lean"`
	Schema  SchemaSection  `toml:"schema"`
	Secrets SecretsSection `toml:"secrets"`
	Output  OutputSection  `toml:"output"`
}

type LeanSection struct {
	Version        int    `toml:"version"`
	DefaultProfile string `toml:"default_profile"`
	Editor         string `toml:"editor"`
}

type SchemaSection struct {
	Path string `toml:"path"`
}

type SecretsSection struct {
	Backend       string `toml:"backend"`
	Recipient     string `toml:"recipient"`
	Identity      string `toml:"identity"`
	MasterKeyFile string `toml:"master_key_file"`
}

type OutputSection struct {
	Format string `toml:"format"`
}

func Default() *Config {
	return &Config{
		Lean:   LeanSection{Version: 1},
		Schema: SchemaSection{Path: ".lean/schema.toml"},
		Secrets: SecretsSection{
			Backend: "local",
		},
		Output: OutputSection{Format: "text"},
	}
}

func GlobalPath() (string, error) {
	return globaldir.ConfigPath()
}

func LocalPath() string {
	return DefaultPath
}

func Load(path string) (*Config, error) {
	if path != "" {
		return loadFile(path)
	}

	cfg := Default()
	global, err := GlobalPath()
	if err == nil {
		if err := mergeGlobalFile(cfg, global); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if os.IsNotExist(err) {
			if legacy, legacyErr := globaldir.Path(legacyGlobalPath); legacyErr == nil {
				_ = mergeGlobalFile(cfg, legacy)
			}
		}
	}
	if err := mergeLocalFile(cfg, LocalPath()); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	normalize(cfg)
	return cfg, nil
}

func loadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := Default()
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	normalize(cfg)
	return cfg, nil
}

func mergeGlobalFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var overlay struct {
		Lean    globalLean    `toml:"lean"`
		Secrets SecretsSection `toml:"secrets"`
	}
	if err := toml.Unmarshal(data, &overlay); err != nil {
		return err
	}
	if overlay.Lean.Version != 0 {
		cfg.Lean.Version = overlay.Lean.Version
	}
	if overlay.Lean.Editor != "" {
		cfg.Lean.Editor = overlay.Lean.Editor
	}
	if overlay.Secrets.Backend != "" {
		cfg.Secrets.Backend = overlay.Secrets.Backend
	}
	if overlay.Secrets.Recipient != "" {
		cfg.Secrets.Recipient = overlay.Secrets.Recipient
	}
	if overlay.Secrets.Identity != "" {
		cfg.Secrets.Identity = overlay.Secrets.Identity
	}
	if overlay.Secrets.MasterKeyFile != "" {
		cfg.Secrets.MasterKeyFile = overlay.Secrets.MasterKeyFile
	}
	return nil
}

func mergeLocalFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var overlay struct {
		Lean   localLean     `toml:"lean"`
		Schema SchemaSection `toml:"schema"`
		Output OutputSection `toml:"output"`
	}
	if err := toml.Unmarshal(data, &overlay); err != nil {
		return err
	}
	if overlay.Lean.Version != 0 {
		cfg.Lean.Version = overlay.Lean.Version
	}
	if overlay.Lean.DefaultProfile != "" {
		cfg.Lean.DefaultProfile = overlay.Lean.DefaultProfile
	}
	if overlay.Schema.Path != "" {
		cfg.Schema.Path = overlay.Schema.Path
	}
	if overlay.Output.Format != "" {
		cfg.Output.Format = overlay.Output.Format
	}
	return nil
}

func normalize(cfg *Config) {
	if cfg.Lean.Version == 0 {
		cfg.Lean.Version = 1
	}
	if cfg.Schema.Path == "" {
		cfg.Schema.Path = ".lean/schema.toml"
	}
	if cfg.Secrets.Backend == "" {
		cfg.Secrets.Backend = "local"
	}
	if cfg.Output.Format == "" {
		cfg.Output.Format = "text"
	}
}

func Save(path string, cfg *Config) error {
	if path == "" {
		path = LocalPath()
	}

	var value any = localConfig{
		Lean: localLean{
			Version:        cfg.Lean.Version,
			DefaultProfile: cfg.Lean.DefaultProfile,
		},
		Schema: cfg.Schema,
		Output: cfg.Output,
	}

	if global, err := GlobalPath(); err == nil && filepath.Clean(path) == filepath.Clean(global) {
		value = globalConfig{
			Lean: globalLean{
				Version: cfg.Lean.Version,
				Editor:  cfg.Lean.Editor,
			},
			Secrets: cfg.Secrets,
		}
	}

	data, err := toml.Marshal(value)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	mode := os.FileMode(0644)
	if global, err := GlobalPath(); err == nil && filepath.Clean(path) == filepath.Clean(global) {
		mode = 0600
	}
	return os.WriteFile(path, data, mode)
}

type localConfig struct {
	Lean   localLean     `toml:"lean"`
	Schema SchemaSection `toml:"schema"`
	Output OutputSection `toml:"output"`
}

type localLean struct {
	Version        int    `toml:"version"`
	DefaultProfile string `toml:"default_profile"`
}

type globalConfig struct {
	Lean    globalLean    `toml:"lean"`
	Secrets SecretsSection `toml:"secrets"`
}

type globalLean struct {
	Version int    `toml:"version"`
	Editor  string `toml:"editor"`
}

func Exists(path string) bool {
	if path == "" {
		path = LocalPath()
	}
	_, err := os.Stat(path)
	return err == nil
}
