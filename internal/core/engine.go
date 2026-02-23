package core

import "errors"

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

func (e *Engine) SetCurrent(name string) error {
	e.State.Current = name
	return SaveState(e.State)
}