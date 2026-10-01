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
// Used automatically by `lean apply`.
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

// NamedSnapshot saves the current .env under an explicit human-readable name.
// If a snapshot with that name already exists it is overwritten.
// The file is stored as <name>.env inside .lean/backups/.
func NamedSnapshot(name string) error {
	name = sanitizeName(name)
	if name == "" {
		return fmt.Errorf("snapshot name cannot be empty")
	}

	data, err := os.ReadFile(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no .env to snapshot")
		}
		return err
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	path := filepath.Join(backupDir, name+".env")
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

	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	return names, nil
}

// ListNamed returns only named snapshots (those without a -YYYYMMDD-HHMMSS suffix).
func ListNamed() ([]string, error) {
	all, err := List()
	if err != nil {
		return nil, err
	}
	var named []string
	for _, n := range all {
		base := strings.TrimSuffix(n, ".env")
		if !isTimestamped(base) {
			named = append(named, base)
		}
	}
	sort.Strings(named)
	return named, nil
}

// Restore atomically writes the chosen backup to .env.
// Accepts either a full filename (prod-20260301-120000.env) or a bare
// named-snapshot label (before-migration).
func Restore(name string) error {
	path := resolvePath(name)
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

// Exists reports whether a backup (named or auto) can be resolved.
func Exists(name string) bool {
	_, err := os.Stat(resolvePath(name))
	return err == nil
}

// Delete removes a backup by name or full filename.
func Delete(name string) error {
	path := resolvePath(name)
	return os.Remove(path)
}

func resolvePath(name string) string {
	if strings.HasSuffix(name, ".env") {
		return filepath.Join(backupDir, name)
	}
	candidate := filepath.Join(backupDir, name+".env")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return filepath.Join(backupDir, name)
}

func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.TrimSuffix(name, ".env")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	return name
}

// isTimestamped detects the auto-snapshot pattern: something-YYYYMMDD-HHMMSS
func isTimestamped(base string) bool {
	parts := strings.Split(base, "-")
	if len(parts) < 2 {
		return false
	}
	date, timePart := parts[len(parts)-2], parts[len(parts)-1]
	if len(date) != 8 || len(timePart) != 6 {
		return false
	}
	for _, c := range date + timePart {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
