package core

import (
	"errors"
	"os"
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

// ScanDisk discovers any .env.* files on disk that aren't registered yet.
func (e *Engine) ScanDisk() error {
	entries, err := os.ReadDir(".")
	if err != nil {
		return err
	}

	known := make(map[string]bool)
	for _, p := range e.State.Profiles {
		known[p] = true
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, ".env.") {
			continue
		}
		profile := strings.TrimPrefix(name, ".env.")
		// skip temp files and template-like names
		if profile == "tmp" || profile == "template" || profile == "example" {
			continue
		}
		if !known[profile] {
			e.State.Profiles = append(e.State.Profiles, profile)
			known[profile] = true
		}
	}

	return SaveState(e.State)
}