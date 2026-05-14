package sync

import (
	"fmt"
	"regexp"
)

// TagRule defines a pattern and the tag value to assign to matching keys.
type TagRule struct {
	Pattern string
	Tag     string
}

// SecretTagger attaches a metadata tag prefix to matching secret keys.
type SecretTagger struct {
	rules []compiledTagRule
}

type compiledTagRule struct {
	re  *regexp.Regexp
	tag string
}

// NewSecretTagger creates a SecretTagger from the provided rules.
// Returns an error if any rule has an empty pattern, empty tag, or invalid regex.
func NewSecretTagger(rules []TagRule) (*SecretTagger, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("secret tagger: at least one rule is required")
	}
	compiled := make([]compiledTagRule, 0, len(rules))
	for i, r := range rules {
		if r.Pattern == "" {
			return nil, fmt.Errorf("secret tagger: rule[%d] has empty pattern", i)
		}
		if r.Tag == "" {
			return nil, fmt.Errorf("secret tagger: rule[%d] has empty tag", i)
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("secret tagger: rule[%d] invalid pattern %q: %w", i, r.Pattern, err)
		}
		compiled = append(compiled, compiledTagRule{re: re, tag: r.Tag})
	}
	return &SecretTagger{rules: compiled}, nil
}

// Apply returns a new map where matching keys are prefixed with "<tag>:".
// Non-matching keys are copied unchanged. The original map is not mutated.
func (t *SecretTagger) Apply(secrets map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		newKey := k
		for _, r := range t.rules {
			if r.re.MatchString(k) {
				newKey = r.tag + ":" + k
				break
			}
		}
		if _, exists := out[newKey]; exists {
			return nil, fmt.Errorf("secret tagger: key collision after tagging: %q", newKey)
		}
		out[newKey] = v
	}
	return out, nil
}

// TagStage returns a pipeline stage that applies the tagger to secrets.
func TagStage(tagger *SecretTagger) Stage {
	return func(secrets map[string]string) (map[string]string, error) {
		return tagger.Apply(secrets)
	}
}
