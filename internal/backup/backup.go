package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const backupDir = ".lean/backups"

// Snapshot saves the current .env to .lean/backups/ before it gets overwritten.
func Snapshot(activeProfile string) error {
	data, err := os.ReadFile(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return nil // nothing to back up
		}
		return err
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	label := activeProfile
	if label == "" {
		label = "unknown"
	}

	ts := time.Now().Format("20060102-150405")
	name := fmt.Sprintf("%s-%s.env", label, ts)
	path := filepath.Join(backupDir, name)

	return os.WriteFile(path, data, 0644)
}

// List returns all backup filenames, newest first.
func List() ([]string, error) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".env") {
			names = append(names, e.Name())
		}
	}

	// Sort newest first (names are timestamp-prefixed so lexicographic reverse works)
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	return names, nil
}

// Restore atomically writes the chosen backup to .env.
func Restore(name string) error {
	path := filepath.Join(backupDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	tmp := ".env.tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, ".env")
}