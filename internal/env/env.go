package env

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

const profileDir = ".lean/profiles"

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
		entries = append(entries, Entry{Key: strings.TrimSpace(trimmed[:idx]), Value: strings.TrimSpace(trimmed[idx+1:])})
	}
	return &File{Entries: entries, Path: path}, scanner.Err()
}

func (f *File) Get(key string) (string, bool) {
	for _, e := range f.Entries {
		if e.Key == key {
			return e.Value, true
		}
	}
	return "", false
}

func (f *File) Set(key, value string) {
	for i, e := range f.Entries {
		if e.Key == key {
			f.Entries[i].Value = value
			return
		}
	}
	f.Entries = append(f.Entries, Entry{Key: key, Value: value})
}

func (f *File) Delete(key string) bool {
	for i, e := range f.Entries {
		if e.Key == key {
			f.Entries = append(f.Entries[:i], f.Entries[i+1:]...)
			return true
		}
	}
	return false
}

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
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(sb.String()), 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (f *File) Keys() []string {
	var keys []string
	for _, e := range f.Entries {
		if e.Key != "" {
			keys = append(keys, e.Key)
		}
	}
	return keys
}

func (f *File) ToMap() map[string]string {
	m := make(map[string]string)
	for _, e := range f.Entries {
		if e.Key != "" {
			m[e.Key] = e.Value
		}
	}
	return m
}

func (f *File) ToJSON() (string, error) {
	d, e := json.MarshalIndent(f.ToMap(), "", "  ")
	return string(d), e
}

func (f *File) ToYAML() (string, error) { d, e := yaml.Marshal(f.ToMap()); return string(d), e }

func (f *File) ToTOML() (string, error) { d, e := toml.Marshal(f.ToMap()); return string(d), e }

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

func (f *File) Extends() string {
	for _, e := range f.Entries {
		if e.Comment != "" {
			c := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(e.Comment), "#"))
			lower := strings.ToLower(c)
			if strings.HasPrefix(lower, "lean:extends") || strings.HasPrefix(lower, "@extends") {
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
	for _, e := range parent.Entries {
		if e.Key != "" && !childKeys[e.Key] {
			merged.Entries = append(merged.Entries, e)
		}
	}
	for _, e := range f.Entries {
		if e.Key == "LEAN_EXTENDS" {
			continue
		}
		if e.Comment != "" {
			c := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(e.Comment), "#")))
			if strings.HasPrefix(c, "lean:extends") || strings.HasPrefix(c, "@extends") {
				continue
			}
		}
		merged.Entries = append(merged.Entries, e)
	}
	return merged
}

func Resolve(profile string) (*File, error) { return resolve(profile, make(map[string]bool), 16) }

func resolve(profile string, visited map[string]bool, depth int) (*File, error) {
	if depth <= 0 {
		return nil, fmt.Errorf("inheritance chain too deep (possible cycle involving '%s')", profile)
	}
	if visited[profile] {
		return nil, fmt.Errorf("circular inheritance detected at profile '%s'", profile)
	}
	visited[profile] = true
	f, err := Parse(ProfilePath(profile))
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

func ProfilePath(name string) string {
	if name == "current" || name == ".env" {
		return ".env"
	}
	return filepath.Join(profileDir, name+".env")
}

func LegacyProfilePath(name string) string {
	if name == "current" || name == ".env" {
		return ".env"
	}
	return ".env." + name
}

type DiffEntry struct {
	Key    string
	Left   string
	Right  string
	Status string
}

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
