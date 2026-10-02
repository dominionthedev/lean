package schema

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// KeyRule describes constraints and behaviors for a single env key.
type KeyRule struct {
	Required         bool              `toml:"required"`
	Secret           bool              `toml:"secret"`
	Description      string            `toml:"description"`
	Default          string            `toml:"default"`
	Values           []string          `toml:"values"`             // allowed enum values
	SameAs           string            `toml:"same_as"`            // default-link to another key
	DependsOn        []string          `toml:"depends_on"`         // keys that must be present/truthy
	RequiredWhen     map[string]string `toml:"required_when"`      // key→value conditions
	DeactivatedWhen  map[string]string `toml:"deactivated_when"`   // key→value: this key is ignored
	From             string            `toml:"from"`               // command: | file: | env:
}

// TemplateRule describes template-level defaults.
type TemplateRule struct {
	ResolvesTo string `toml:"resolves_to"` // default output file, e.g. .env.production
	Extends    string `toml:"extends"`
	Description string `toml:"description"`
}

// Schema is the full advanced schema document.
type Schema struct {
	Keys      map[string]KeyRule      `toml:"keys"`
	Templates map[string]TemplateRule `toml:"templates"`
}

// Load reads a schema TOML file.
func Load(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Schema
	if err := toml.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if s.Keys == nil {
		s.Keys = map[string]KeyRule{}
	}
	if s.Templates == nil {
		s.Templates = map[string]TemplateRule{}
	}
	return &s, nil
}

// Save writes the schema to path.
func Save(path string, s *Schema) error {
	data, err := toml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Issue is one validation finding.
type Issue struct {
	Key     string
	Level   string // "error" | "warn" | "info"
	Message string
}

// Validate checks values against the schema rules.
// values is the resolved key→value map of the profile under test.
func (s *Schema) Validate(values map[string]string) []Issue {
	var issues []Issue

	for key, rule := range s.Keys {
		val, present := values[key]
		active := !rule.isDeactivated(values)

		if !active {
			continue
		}

		// Required (unconditional or conditional)
		if rule.Required || rule.isRequiredWhen(values) {
			if !present || strings.TrimSpace(val) == "" {
				msg := "required but missing or empty"
				if len(rule.RequiredWhen) > 0 {
					msg = "required by condition but missing or empty"
				}
				issues = append(issues, Issue{Key: key, Level: "error", Message: msg})
				continue
			}
		}

		if !present {
			// depends_on: warn if dependency is active but this key is absent
			if len(rule.DependsOn) > 0 && rule.depsSatisfied(values) {
				issues = append(issues, Issue{
					Key: key, Level: "warn",
					Message: fmt.Sprintf("depends on %s but is not set", strings.Join(rule.DependsOn, ", ")),
				})
			}
			continue
		}

		// Enum values
		if len(rule.Values) > 0 && !contains(rule.Values, val) {
			issues = append(issues, Issue{
				Key: key, Level: "error",
				Message: fmt.Sprintf("value %q not in allowed set %v", val, rule.Values),
			})
		}

		// same_as check (warn if diverged — allowed, but notify)
		if rule.SameAs != "" {
			if other, ok := values[rule.SameAs]; ok && other != val {
				issues = append(issues, Issue{
					Key: key, Level: "info",
					Message: fmt.Sprintf("differs from same_as %s (%q ≠ %q)", rule.SameAs, val, other),
				})
			}
		}
	}

	return issues
}

// ApplyDefaults returns a copy of values with defaults and same_as links filled in
// for keys that are absent. Does not overwrite existing values.
func (s *Schema) ApplyDefaults(values map[string]string) map[string]string {
	out := make(map[string]string, len(values))
	for k, v := range values {
		out[k] = v
	}

	// First pass: static defaults
	for key, rule := range s.Keys {
		if _, ok := out[key]; ok {
			continue
		}
		if rule.Default != "" && !rule.isDeactivated(out) {
			out[key] = rule.Default
		}
	}

	// Second pass: same_as (only if still absent)
	for key, rule := range s.Keys {
		if _, ok := out[key]; ok {
			continue
		}
		if rule.SameAs != "" {
			if v, ok := out[rule.SameAs]; ok {
				out[key] = v
			}
		}
	}

	return out
}

// ResolveFrom evaluates from: sources for keys that are absent or empty.
// Returns a map of key → resolved value for keys that were filled.
func (s *Schema) ResolveFrom(values map[string]string) (map[string]string, []Issue) {
	filled := map[string]string{}
	var issues []Issue

	for key, rule := range s.Keys {
		if rule.From == "" {
			continue
		}
		if rule.isDeactivated(values) {
			continue
		}
		existing, present := values[key]
		if present && strings.TrimSpace(existing) != "" {
			continue // already set — don't overwrite
		}

		val, err := resolveSource(rule.From)
		if err != nil {
			issues = append(issues, Issue{
				Key: key, Level: "error",
				Message: fmt.Sprintf("from %q failed: %s", rule.From, err),
			})
			continue
		}
		filled[key] = val
	}
	return filled, issues
}

// RequiredKeys returns keys that are unconditionally required.
func (s *Schema) RequiredKeys() []string {
	var keys []string
	for k, rule := range s.Keys {
		if rule.Required {
			keys = append(keys, k)
		}
	}
	return keys
}

// SecretKeys returns keys marked as secret.
func (s *Schema) SecretKeys() []string {
	var keys []string
	for k, rule := range s.Keys {
		if rule.Secret {
			keys = append(keys, k)
		}
	}
	return keys
}

func (r KeyRule) isDeactivated(values map[string]string) bool {
	if len(r.DeactivatedWhen) == 0 {
		return false
	}
	for k, expect := range r.DeactivatedWhen {
		if values[k] == expect {
			return true
		}
	}
	return false
}

func (r KeyRule) isRequiredWhen(values map[string]string) bool {
	if len(r.RequiredWhen) == 0 {
		return false
	}
	for k, expect := range r.RequiredWhen {
		if values[k] != expect {
			return false
		}
	}
	return true
}

func (r KeyRule) depsSatisfied(values map[string]string) bool {
	for _, dep := range r.DependsOn {
		v, ok := values[dep]
		if !ok || strings.TrimSpace(v) == "" || v == "false" || v == "0" {
			return false
		}
	}
	return true
}

func contains(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

// resolveSource handles from: directives.
//
//	command:<shell>   — run command, use stdout
//	file:<path>       — read file contents (trimmed)
//	env:<VAR>         — read from process environment
func resolveSource(spec string) (string, error) {
	spec = strings.TrimSpace(spec)
	idx := strings.Index(spec, ":")
	if idx < 0 {
		return "", fmt.Errorf("invalid from spec (want command:|file:|env:)")
	}
	kind, rest := spec[:idx], strings.TrimSpace(spec[idx+1:])
	if rest == "" {
		return "", fmt.Errorf("empty %s source", kind)
	}

	switch kind {
	case "command":
		cmd := exec.Command("sh", "-c", rest)
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	case "file":
		data, err := os.ReadFile(rest)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	case "env":
		v, ok := os.LookupEnv(rest)
		if !ok {
			return "", fmt.Errorf("environment variable %s not set", rest)
		}
		return v, nil
	default:
		return "", fmt.Errorf("unknown from kind %q (want command|file|env)", kind)
	}
}
