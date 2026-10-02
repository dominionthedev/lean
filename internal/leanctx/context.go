package leanctx

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const contextsDir = ".lean/contexts"

// FileMapping maps a source file onto a target path when the context is applied.
type FileMapping struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// Context is a named bundle of file mappings.
// Applying it copies each source → target (with env inheritance resolution for .env.* sources).
type Context struct {
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	// Profile is the primary env profile to resolve into .env (optional convenience).
	// If set and no explicit ".env" target exists in Files, it is treated as
	// {source: ".env.<profile>", target: ".env"}.
	Profile string        `json:"profile,omitempty"`
	Files   []FileMapping `json:"files"`
}

func pathFor(name string) string {
	return filepath.Join(contextsDir, name+".json")
}

// Load reads a context by name.
func Load(name string) (*Context, error) {
	data, err := os.ReadFile(pathFor(name))
	if err != nil {
		return nil, err
	}
	var c Context
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if c.Name == "" {
		c.Name = name
	}
	return &c, nil
}

// Save writes a context to disk.
func Save(c *Context) error {
	if c.Name == "" {
		return fmt.Errorf("context name is required")
	}
	if err := os.MkdirAll(contextsDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pathFor(c.Name), data, 0644)
}

// Delete removes a context definition.
func Delete(name string) error {
	return os.Remove(pathFor(name))
}

// Exists reports whether a context definition is on disk.
func Exists(name string) bool {
	_, err := os.Stat(pathFor(name))
	return err == nil
}

// List returns all context names, sorted.
func List() ([]string, error) {
	entries, err := os.ReadDir(contextsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasSuffix(n, ".json") {
			names = append(names, strings.TrimSuffix(n, ".json"))
		}
	}
	sort.Strings(names)
	return names, nil
}

// EffectiveFiles returns the full list of source→target mappings,
// injecting the Profile → .env mapping when needed.
func (c *Context) EffectiveFiles() []FileMapping {
	files := make([]FileMapping, len(c.Files))
	copy(files, c.Files)

	if c.Profile == "" {
		return files
	}

	hasEnvTarget := false
	for _, f := range files {
		if f.Target == ".env" {
			hasEnvTarget = true
			break
		}
	}
	if !hasEnvTarget {
		files = append([]FileMapping{{
			Source: ".env." + c.Profile,
			Target: ".env",
		}}, files...)
	}
	return files
}

// AddFile appends or updates a source→target mapping.
func (c *Context) AddFile(source, target string) {
	for i, f := range c.Files {
		if f.Target == target {
			c.Files[i].Source = source
			return
		}
	}
	c.Files = append(c.Files, FileMapping{Source: source, Target: target})
}

// RemoveFile drops the mapping whose target matches.
func (c *Context) RemoveFile(target string) bool {
	for i, f := range c.Files {
		if f.Target == target {
			c.Files = append(c.Files[:i], c.Files[i+1:]...)
			return true
		}
	}
	return false
}
