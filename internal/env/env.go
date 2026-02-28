package env

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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