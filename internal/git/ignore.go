package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const excludeRules = `.lean
.env.*
!.env.template
!.env.example
`

// ProtectProjectFiles adds Lean's local files and environment profiles to the
// repository's local Git exclude file. It does nothing when the project is not
// inside a Git worktree.
func ProtectProjectFiles() (bool, error) {
	gitDir, err := gitDir()
	if err != nil {
		if errors.Is(err, errNotRepository) {
			return false, nil
		}
		return false, err
	}

	excludePath := filepath.Join(gitDir, "info", "exclude")
	if err := os.MkdirAll(filepath.Dir(excludePath), 0700); err != nil {
		return false, err
	}

	data, err := os.ReadFile(excludePath)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	current := string(data)
	missing := missingRules(current)
	if len(missing) == 0 {
		return true, nil
	}

	var b strings.Builder
	b.WriteString(current)
	if b.Len() > 0 && !strings.HasSuffix(current, "\n") {
		b.WriteByte('\n')
	}
	b.WriteString("# lean local files and environment profiles\n")
	for _, rule := range missing {
		b.WriteString(rule)
		b.WriteByte('\n')
	}

	if err := os.WriteFile(excludePath, []byte(b.String()), 0600); err != nil {
		return false, err
	}
	return true, nil
}

var errNotRepository = errors.New("not a git repository")

func gitDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Stderr = nil
	out, err := cmd.Output()
	if err != nil {
		return "", errNotRepository
	}

	dir := strings.TrimSpace(string(out))
	if dir == "" {
		return "", errNotRepository
	}
	if !filepath.IsAbs(dir) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(cwd, dir)
	}
	return filepath.Clean(dir), nil
}

func missingRules(content string) []string {
	lines := make(map[string]bool)
	for _, line := range strings.Split(content, "\n") {
		lines[strings.TrimSpace(line)] = true
	}

	var missing []string
	for _, rule := range []string{".lean", ".env.*", "!.env.template", "!.env.example"} {
		if !lines[rule] {
			missing = append(missing, rule)
		}
	}
	return missing
}
