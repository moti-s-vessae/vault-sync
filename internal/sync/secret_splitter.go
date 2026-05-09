package sync

import (
	"fmt"
	"regexp"
	"strings"
)

// SplitRule defines how a single secret value should be split into multiple keys.
type SplitRule struct {
	// Pattern is a glob or regex pattern to match secret keys.
	Pattern string
	// Separator is the delimiter used to split the value.
	Separator string
	// KeyTemplate is a format string for generated keys, e.g. "{{.Key}}_{{.Index}}".
	KeyTemplate string
}

// SecretSplitter splits secret values into multiple keys based on a separator.
type SecretSplitter struct {
	rules []compiledSplitRule
}

type compiledSplitRule struct {
	pattern     *regexp.Regexp
	separator   string
	keyTemplate string
}

// NewSecretSplitter creates a SecretSplitter from the provided rules.
func NewSecretSplitter(rules []SplitRule) (*SecretSplitter, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("at least one split rule is required")
	}
	compiled := make([]compiledSplitRule, 0, len(rules))
	for _, r := range rules {
		if r.Pattern == "" {
			return nil, fmt.Errorf("split rule pattern must not be empty")
		}
		if r.Separator == "" {
			return nil, fmt.Errorf("split rule separator must not be empty")
		}
		if r.KeyTemplate == "" {
			r.KeyTemplate = "{{.Key}}_{{.Index}}"
		}
		glob := "^" + strings.ReplaceAll(regexp.QuoteMeta(r.Pattern), "\\*", ".*") + "$"
		re, err := regexp.Compile(glob)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", r.Pattern, err)
		}
		compiled = append(compiled, compiledSplitRule{
			pattern:     re,
			separator:   r.Separator,
			keyTemplate: r.KeyTemplate,
		})
	}
	return &SecretSplitter{rules: compiled}, nil
}

// Apply splits matching secret values, returning an expanded map.
func (s *SecretSplitter) Apply(secrets map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		matched := false
		for _, r := range s.rules {
			if r.pattern.MatchString(k) {
				parts := strings.Split(v, r.separator)
				for i, part := range parts {
					newKey := expandKeyTemplate(r.keyTemplate, k, i+1)
					out[newKey] = part
				}
				matched = true
				break
			}
		}
		if !matched {
			out[k] = v
		}
	}
	return out, nil
}

func expandKeyTemplate(tmpl, key string, index int) string {
	s := strings.ReplaceAll(tmpl, "{{.Key}}", key)
	s = strings.ReplaceAll(s, "{{.Index}}", fmt.Sprintf("%d", index))
	return s
}

// SplitStage returns a pipeline stage that applies the given splitter.
func SplitStage(splitter *SecretSplitter) func(map[string]string) (map[string]string, error) {
	return func(secrets map[string]string) (map[string]string, error) {
		return splitter.Apply(secrets)
	}
}
