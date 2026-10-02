package core

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ProfileMeta holds optional metadata for a profile.
type ProfileMeta struct {
	Description string   `json:"description,omitempty"`
	Author      string   `json:"author,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Created     string   `json:"created,omitempty"`
}

type State struct {
	Initialized bool                   `json:"initialized"`
	Current     string                 `json:"current"`
	Profiles    []string               `json:"profiles"`
	Templates   []string               `json:"templates,omitempty"`
	Meta        map[string]ProfileMeta `json:"meta,omitempty"`
	Version     string                 `json:"version"`
}

const stateDir = ".lean"
const stateFile = "state.json"

func statePath() string {
	return filepath.Join(stateDir, stateFile)
}

func LoadState() (*State, error) {
	data, err := os.ReadFile(statePath())
	if err != nil {
		return nil, err
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if s.Meta == nil {
		s.Meta = make(map[string]ProfileMeta)
	}
	return &s, nil
}

func SaveState(s *State) error {
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statePath(), data, 0644)
}
