package vault

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ValidationRule defines a rule for validating secret keys or values.
type ValidationRule struct {
	// Key is a regex pattern matched against the secret key.
	Key string `yaml:"key"`
	// Pattern is a regex pattern the value must match (optional).
	Pattern string `yaml:"pattern,omitempty"`
	// Required means the key must be present in the secrets map.
	Required bool `yaml:"required,omitempty"`
}

// ValidationError holds all violations found during validation.
type ValidationError struct {
	Violations []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("secret validation failed: %s", strings.Join(e.Violations, "; "))
}

// ValidateSecrets checks secrets against the provided rules.
// It returns a *ValidationError if any violations are found, nil otherwise.
func ValidateSecrets(secrets map[string]string, rules []ValidationRule) error {
	if len(rules) == 0 {
		return nil
	}

	var violations []string

	for _, rule := range rules {
		keyRe, err := regexp.Compile(rule.Key)
		if err != nil {
			violations = append(violations, fmt.Sprintf("invalid key pattern %q: %v", rule.Key, err))
			continue
		}

		matched := false
		for k, v := range secrets {
			if !keyRe.MatchString(k) {
				continue
			}
			matched = true
			if rule.Pattern == "" {
				continue
			}
			valRe, err := regexp.Compile(rule.Pattern)
			if err != nil {
				violations = append(violations, fmt.Sprintf("invalid value pattern %q: %v", rule.Pattern, err))
				continue
			}
			if !valRe.MatchString(v) {
				violations = append(violations, fmt.Sprintf("key %q value does not match pattern %q", k, rule.Pattern))
			}
		}

		if rule.Required && !matched {
			violations = append(violations, fmt.Sprintf("required key matching %q not found", rule.Key))
		}
	}

	if len(violations) > 0 {
		return &ValidationError{Violations: violations}
	}
	return nil
}

// IsValidationError reports whether err is a *ValidationError.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
