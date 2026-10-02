package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Engine struct {
	State *State
}

func NewEngine() (*Engine, error) {
	state, err := LoadState()
	if err != nil {
		return nil, err
	}
	return &Engine{State: state}, nil
}

func Initialize() error {
	s := &State{
		Initialized: true,
		Version:     "1.0.0",
		Profiles:    []string{},
		Templates:   []string{},
	}
	return SaveState(s)
}

func (e *Engine) AddProfile(name string) error {
	for _, p := range e.State.Profiles {
		if p == name {
			return errors.New("profile already exists")
		}
	}
	e.State.Profiles = append(e.State.Profiles, name)
	return SaveState(e.State)
}

func (e *Engine) DeleteProfile(name string) error {
	for i, p := range e.State.Profiles {
		if p == name {
			e.State.Profiles = append(e.State.Profiles[:i], e.State.Profiles[i+1:]...)
			if e.State.Meta != nil {
				delete(e.State.Meta, name)
			}
			if e.State.Current == name {
				e.State.Current = ""
			}
			return SaveState(e.State)
		}
	}
	return errors.New("profile not found")
}

func (e *Engine) AddTemplate(path string) error {
	for _, t := range e.State.Templates {
		if t == path {
			return nil
		}
	}
	e.State.Templates = append(e.State.Templates, path)
	return SaveState(e.State)
}

func (e *Engine) ProfileExists(name string) bool {
	for _, p := range e.State.Profiles {
		if p == name {
			return true
		}
	}
	return false
}

func (e *Engine) SetCurrent(name string) error {
	e.State.Current = name
	return SaveState(e.State)
}

// ScanDisk discovers root .env.<profile> files and keeps the state registry in sync
// with profiles that exist. It intentionally does not delete missing profiles from
// state; callers must ask the user before removing stale profile registrations.
func (e *Engine) ScanDisk() error {
	entries, err := os.ReadDir(".")
	if err != nil {
		return err
	}

	// Older development versions briefly stored profiles under .lean/profiles.
	// Move those files back to the normal root .env.<name> layout when possible.
	legacyDir := filepath.Join(".lean", "profiles")
	if legacyEntries, err := os.ReadDir(legacyDir); err == nil {
		for _, entry := range legacyEntries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".env") {
				continue
			}
			name := strings.TrimSuffix(entry.Name(), ".env")
			src := filepath.Join(legacyDir, entry.Name())
			dst := ProfileFilePath(name)
			if _, err := os.Stat(dst); os.IsNotExist(err) {
				if err := os.Rename(src, dst); err != nil {
					data, readErr := os.ReadFile(src)
					if readErr == nil {
						if writeErr := os.WriteFile(dst, data, 0600); writeErr == nil {
							_ = os.Remove(src)
						}
					}
				}
			}
		}
		_ = os.Remove(legacyDir)
	}

	known := make(map[string]bool, len(e.State.Profiles))
	for _, p := range e.State.Profiles {
		known[p] = true
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, ".env.") {
			continue
		}
		profile := strings.TrimPrefix(name, ".env.")
		if profile == "tmp" || profile == "template" || profile == "example" || profile == "" {
			continue
		}
		if !known[profile] {
			e.State.Profiles = append(e.State.Profiles, profile)
			known[profile] = true
		}
	}

	return SaveState(e.State)
}

func (e *Engine) MissingProfiles() []string {
	var missing []string
	for _, profile := range e.State.Profiles {
		if _, err := os.Stat(ProfileFilePath(profile)); os.IsNotExist(err) {
			missing = append(missing, profile)
		}
	}
	return missing
}

func ProfileFilePath(name string) string {
	return ".env." + name
}

func (e *Engine) ScanTemplates() error {
	entries, err := os.ReadDir(".")
	if err != nil {
		return err
	}
	known := make(map[string]bool)
	for _, t := range e.State.Templates {
		known[t] = true
	}
	for _, entry := range entries {
		name := entry.Name()
		if (name == ".env.template" || name == ".env.example") && !known[name] {
			e.State.Templates = append(e.State.Templates, name)
			known[name] = true
		}
	}
	return SaveState(e.State)
}

func (e *Engine) SetMeta(name string, meta ProfileMeta) error {
	if e.State.Meta == nil {
		e.State.Meta = make(map[string]ProfileMeta)
	}
	e.State.Meta[name] = meta
	return SaveState(e.State)
}

func (e *Engine) GetMeta(name string) (ProfileMeta, bool) {
	if e.State.Meta == nil {
		return ProfileMeta{}, false
	}
	m, ok := e.State.Meta[name]
	return m, ok
}
