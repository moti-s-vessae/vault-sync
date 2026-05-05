package vault

import (
	"fmt"
	"regexp"
)

// TransformRule describes a regex-based key transformation.
type TransformRule struct {
	Pattern     string `yaml:"pattern"`
	Replacement string `yaml:"replacement"`
}

// TransformSecrets applies each rule to all secret keys using regex replacement.
func TransformSecrets(secrets map[string]string, rules []TransformRule) (map[string]string, error) {
	if len(rules) == 0 {
		return secrets, nil
	}
	compiled, err := compileRules(rules)
	if err != nil {
		return nil, err
	}
	return applyTransforms(secrets, compiled), nil
}

type compiledRule struct {
	re          *regexp.Regexp
	replacement string
}

func compileRules(rules []TransformRule) ([]compiledRule, error) {
	out := make([]compiledRule, 0, len(rules))
	for i, r := range rules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("transform rule[%d]: invalid pattern %q: %w", i, r.Pattern, err)
		}
		out = append(out, compiledRule{re: re, replacement: r.Replacement})
	}
	return out, nil
}

func applyTransforms(secrets map[string]string, rules []compiledRule) map[string]string {
	result := make(map[string]string, len(secrets))
	for k, v := range secrets {
		newKey := k
		for _, r := range rules {
			newKey = r.re.ReplaceAllString(newKey, r.replacement)
		}
		result[newKey] = v
	}
	return result
}
