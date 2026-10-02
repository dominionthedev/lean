package config

import (
	"os"
	"path/filepath"

	"github.com/dominionthedev/lean/internal/globaldir"
	"github.com/pelletier/go-toml/v2"
)

const DefaultPath = ".lean/lean.toml"

type Config struct {
	Lean    LeanSection    \`toml:"lean"\`
	Schema  SchemaSection  \`toml:"schema"\`
	Secrets SecretsSection \`toml:"secrets"\`
}

type LeanSection struct {
	Version int \`toml:"version"\`
	DefaultProfile string \`toml:"default_profile"\`
}

type SchemaSection struct { Path string \`toml:"path"\` }

type SecretsSection struct {
	Backend string \`toml:"backend"\`
	Recipient string \`toml:"recipient"\`
	Identity string \`toml:"identity"\`
	MasterKeyFile string \`toml:"master_key_file"\`
}

func Default() *Config {
	return &Config{
		Lean: LeanSection{Version: 1},
		Schema: SchemaSection{Path: ".lean/schema.toml"},
		Secrets: SecretsSection{Backend: "local"},
	}
}

func GlobalPath() (string, error) { return globaldir.Path("lean.toml") }
func LocalPath() string { return DefaultPath }

func Load(path string) (*Config, error) {
	if path != "" { return loadFile(path) }
	if cfg, err := loadFile(LocalPath()); err == nil { return cfg,nil } else if !os.IsNotExist(err) { return nil,err }
	global, err := GlobalPath()
	if err == nil {
		if cfg, err := loadFile(global); err == nil { return cfg,nil } else if !os.IsNotExist(err) { return nil,err }
	}
	return Default(),nil
}

func loadFile(path string) (*Config,error) {
	data,err:=os.ReadFile(path)
	if err!=nil{return nil,err}
	cfg:=Default()
	if err:=toml.Unmarshal(data,cfg);err!=nil{return nil,err}
	normalize(cfg)
	return cfg,nil
}

func normalize(cfg *Config) {
	if cfg.Schema.Path==""{cfg.Schema.Path=".lean/schema.toml"}
	if cfg.Secrets.Backend==""{cfg.Secrets.Backend="local"}
	if cfg.Lean.Version==0{cfg.Lean.Version=1}
}

func Save(path string,cfg *Config) error {
	if path==""{path=LocalPath()}
	data,err:=toml.Marshal(cfg);if err!=nil{return err}
	if err:=os.MkdirAll(filepath.Dir(path),0700);err!=nil{return err}
	mode:=os.FileMode(0644)
	if global,err:=GlobalPath();err==nil&&filepath.Clean(path)==filepath.Clean(global){mode=0600}
	return os.WriteFile(path,data,mode)
}

func Exists(path string) bool {
	if path==""{path=LocalPath()}
	_,err:=os.Stat(path)
	return err==nil
}
