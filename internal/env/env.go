package env

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type Entry struct {
	Key     string
	Value   string
	Comment string
	Blank   bool
}

type File struct {
	Entries []Entry
	Path    string
}

// Parse reads a .env file into structured entries, preserving comments and blank lines.
func Parse(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			entries = append(entries, Entry{Blank: true})
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			entries = append(entries, Entry{Comment: trimmed})
			continue
		}

		idx := strings.Index(trimmed, "=")
		if idx < 0 {
			entries = append(entries, Entry{Key: trimmed})
			continue
		}

		key := strings.TrimSpace(trimmed[:idx])
		value := strings.TrimSpace(trimmed[idx+1:])
		entries = append(entries, Entry{Key: key, Value: value})
	}

	return &File{Entries: entries, Path: path}, scanner.Err()
}

// Get returns the value for a key and whether it was found.
func (f *File) Get(key string) (string, bool) {
	for _, e := range f.Entries {
		if e.Key == key {
			return e.Value, true
		}
	}
	return "", false
}

// Set updates an existing key or appends a new one.
func (f *File) Set(key, value string) {
	for i, e := range f.Entries {
		if e.Key == key {
			f.Entries[i].Value = value
			return
		}
	}
	f.Entries = append(f.Entries, Entry{Key: key, Value: value})
}

// Delete removes a key. Returns true if the key existed.
func (f *File) Delete(key string) bool {
	for i, e := range f.Entries {
		if e.Key == key {
			f.Entries = append(f.Entries[:i], f.Entries[i+1:]...)
			return true
		}
	}
	return false
}

// Strip returns a copy with all values cleared (keys only).
func (f *File) Strip() *File {
	stripped := &File{Path: f.Path}
	for _, e := range f.Entries {
		if e.Key != "" {
			stripped.Entries = append(stripped.Entries, Entry{Key: e.Key})
		} else {
			stripped.Entries = append(stripped.Entries, e)
		}
	}
	return stripped
}

// Write atomically writes the file to path.
func (f *File) Write(path string) error {
	var sb strings.Builder
	for _, e := range f.Entries {
		switch {
		case e.Blank:
			sb.WriteString("\n")
		case e.Comment != "":
			sb.WriteString(e.Comment + "\n")
		default:
			sb.WriteString(fmt.Sprintf("%s=%s\n", e.Key, e.Value))
		}
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(sb.String()), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Keys returns all variable names (no comments, no blanks).
func (f *File) Keys() []string {
	var keys []string
	for _, e := range f.Entries {
		if e.Key != "" {
			keys = append(keys, e.Key)
		}
	}
	return keys
}

// ToMap converts the file entries into a simple key-value map.
func (f *File) ToMap() map[string]string {
	m := make(map[string]string)
	for _, e := range f.Entries {
		if e.Key != "" {
			m[e.Key] = e.Value
		}
	}
	return m
}

// ToJSON returns the JSON representation of the environment variables.
func (f *File) ToJSON() (string, error) {
	data, err := json.MarshalIndent(f.ToMap(), "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToYAML returns the YAML representation of the environment variables.
func (f *File) ToYAML() (string, error) {
	data, err := yaml.Marshal(f.ToMap())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToTOML returns the TOML representation of the environment variables.
func (f *File) ToTOML() (string, error) {
	data, err := toml.Marshal(f.ToMap())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToString returns the env-style string representation of the file.
func (f *File) ToString() string {
	var sb strings.Builder
	for _, e := range f.Entries {
		switch {
		case e.Blank:
			sb.WriteString("\n")
		case e.Comment != "":
			sb.WriteString(e.Comment + "\n")
		default:
			sb.WriteString(fmt.Sprintf("%s=%s\n", e.Key, e.Value))
		}
	}
	return sb.String()
}

// Extends returns the parent profile name if a lean inheritance directive is present.
// Supported forms (first match wins):
//
//	# lean:extends base
//	# @extends base
//	LEAN_EXTENDS=base
func (f *File) Extends() string {
	for _, e := range f.Entries {
		if e.Comment != "" {
			c := strings.TrimSpace(e.Comment)
			c = strings.TrimPrefix(c, "#")
			c = strings.TrimSpace(c)
			lower := strings.ToLower(c)
			if strings.HasPrefix(lower, "lean:extends") {
				parts := strings.Fields(c)
				if len(parts) >= 2 {
					return parts[len(parts)-1]
				}
			}
			if strings.HasPrefix(lower, "@extends") {
				parts := strings.Fields(c)
				if len(parts) >= 2 {
					return parts[len(parts)-1]
				}
			}
		}
		if e.Key == "LEAN_EXTENDS" && e.Value != "" {
			return e.Value
		}
	}
	return ""
}

// Merge overlays child entries onto a copy of the parent.
// Child keys override parent keys. Parent-only keys are kept;
// LEAN_EXTENDS is stripped from the result.
func (f *File) Merge(parent *File) *File {
	if parent == nil {
		return f
	}

	childKeys := make(map[string]bool)
	for _, e := range f.Entries {
		if e.Key != "" {
			childKeys[e.Key] = true
		}
	}

	merged := &File{Path: f.Path}

	// Parent keys not overridden by child
	for _, e := range parent.Entries {
		if e.Key != "" && !childKeys[e.Key] {
			merged.Entries = append(merged.Entries, e)
		}
	}

	// All child entries — skip LEAN_EXTENDS and inheritance comments
	for _, e := range f.Entries {
		if e.Key == "LEAN_EXTENDS" {
			continue
		}
		if e.Comment != "" {
			c := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(e.Comment), "#"))
			lower := strings.ToLower(c)
			if strings.HasPrefix(lower, "lean:extends") || strings.HasPrefix(lower, "@extends") {
				continue
			}
		}
		merged.Entries = append(merged.Entries, e)
	}

	return merged
}

// Resolve loads a profile and walks its inheritance chain, returning the fully
// merged File. Cycle detection is included. maxDepth guards against runaway chains.
func Resolve(profile string) (*File, error) {
	const maxDepth = 16
	visited := make(map[string]bool)
	return resolve(profile, visited, maxDepth)
}

func resolve(profile string, visited map[string]bool, depth int) (*File, error) {
	if depth <= 0 {
		return nil, fmt.Errorf("inheritance chain too deep (possible cycle involving '%s')", profile)
	}
	if visited[profile] {
		return nil, fmt.Errorf("circular inheritance detected at profile '%s'", profile)
	}
	visited[profile] = true

	path := ProfilePath(profile)
	f, err := Parse(path)
	if err != nil {
		return nil, fmt.Errorf("profile '%s': %w", profile, err)
	}

	parent := f.Extends()
	if parent == "" {
		return f, nil
	}

	base, err := resolve(parent, visited, depth-1)
	if err != nil {
		return nil, err
	}

	return f.Merge(base), nil
}

// ProfilePath returns the conventional on-disk path for a named profile.
func ProfilePath(name string) string {
	if name == "current" || name == ".env" {
		return ".env"
	}
	return ".env." + name
}

// DiffEntry holds one differing key between two profiles.
type DiffEntry struct {
	Key    string
	Left   string
	Right  string
	Status string // "changed", "only-left", "only-right"
}

// Diff compares two key-value maps and returns entries that differ.
func Diff(left, right map[string]string) []DiffEntry {
	seen := make(map[string]bool)
	var out []DiffEntry

	for k, lv := range left {
		seen[k] = true
		rv, ok := right[k]
		if !ok {
			out = append(out, DiffEntry{Key: k, Left: lv, Status: "only-left"})
		} else if lv != rv {
			out = append(out, DiffEntry{Key: k, Left: lv, Right: rv, Status: "changed"})
		}
	}
	for k, rv := range right {
		if !seen[k] {
			out = append(out, DiffEntry{Key: k, Right: rv, Status: "only-right"})
		}
	}
	return out
}