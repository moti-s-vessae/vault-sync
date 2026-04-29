package vault

import (
	"regexp"
	"strings"
)

// TransformRule defines a transformation to apply to secret keys.
type TransformRule struct {
	// Pattern is a regex pattern to match against the key.
	Pattern string `yaml:"pattern"`
	// Replace is the replacement string (supports capture groups like $1).
	Replace string `yaml:"replace"`
}

// TransformSecrets applies a list of transformation rules to secret keys.
// Rules are applied in order; all matching rules are applied (unlike rename).
// Keys that don't match any rule are left unchanged.
func TransformSecrets(secrets map[string]string, rules []TransformRule) (map[string]string, error) {
	if len(rules) == 0 {
		return secrets, nil
	}

	compiled := make([]*compiledRule, 0, len(rules))
	for _, r := range rules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, &compiledRule{re: re, replace: r.Replace})
	}

	result := make(map[string]string, len(secrets))
	for k, v := range secrets {
		transformed := applyTransforms(k, compiled)
		result[transformed] = v
	}
	return result, nil
}

type compiledRule struct {
	re      *regexp.Regexp
	replace string
}

func applyTransforms(key string, rules []*compiledRule) string {
	for _, r := range rules {
		if r.re.MatchString(key) {
			key = r.re.ReplaceAllString(key, r.replace)
		}
	}
	return strings.ToUpper(key)
}
