package sync

import (
	"encoding/base64"
	"fmt"
	"regexp"
)

// EncodeRule defines a pattern and encoding format for matching secret keys.
type EncodeRule struct {
	Pattern string
	Format  string // "base64" or "base64url"
}

// SecretEncoder encodes secret values for matching keys.
type SecretEncoder struct {
	rules []compiledEncodeRule
}

type compiledEncodeRule struct {
	pattern *regexp.Regexp
	format  string
}

// NewSecretEncoder creates a SecretEncoder from the given rules.
// Returns an error if any pattern is invalid or format is unsupported.
func NewSecretEncoder(rules []EncodeRule) (*SecretEncoder, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("encoder: at least one rule is required")
	}
	compiled := make([]compiledEncodeRule, 0, len(rules))
	for _, r := range rules {
		if r.Format != "base64" && r.Format != "base64url" {
			return nil, fmt.Errorf("encoder: unsupported format %q (must be base64 or base64url)", r.Format)
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("encoder: invalid pattern %q: %w", r.Pattern, err)
		}
		compiled = append(compiled, compiledEncodeRule{pattern: re, format: r.Format})
	}
	return &SecretEncoder{rules: compiled}, nil
}

// Encode returns a new map with matching secret values encoded.
// Non-matching keys are passed through unchanged.
func (e *SecretEncoder) Encode(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[k] = v
	}
	for k, v := range out {
		for _, rule := range e.rules {
			if rule.pattern.MatchString(k) {
				if rule.format == "base64url" {
					out[k] = base64.URLEncoding.EncodeToString([]byte(v))
				} else {
					out[k] = base64.StdEncoding.EncodeToString([]byte(v))
				}
				break
			}
		}
	}
	return out
}

// EncodeStage returns a pipeline stage that encodes matching secret values.
func EncodeStage(encoder *SecretEncoder) func(map[string]string) (map[string]string, error) {
	return func(secrets map[string]string) (map[string]string, error) {
		return encoder.Encode(secrets), nil
	}
}
