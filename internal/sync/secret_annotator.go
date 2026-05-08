package sync

import (
	"fmt"
	"regexp"
)

// AnnotationRule maps a key pattern to a metadata tag to inject as a derived key.
type AnnotationRule struct {
	Pattern string
	TagKey  string
	TagValue string
}

// SecretAnnotator injects derived annotation keys alongside matched secrets.
type SecretAnnotator struct {
	rules []compiledAnnotation
}

type compiledAnnotation struct {
	re       *regexp.Regexp
	tagKey   string
	tagValue string
}

// NewSecretAnnotator creates a SecretAnnotator from the provided rules.
// Returns an error if any pattern is invalid or if a rule has empty fields.
func NewSecretAnnotator(rules []AnnotationRule) (*SecretAnnotator, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("annotator: at least one rule is required")
	}
	compiled := make([]compiledAnnotation, 0, len(rules))
	for _, r := range rules {
		if r.Pattern == "" {
			return nil, fmt.Errorf("annotator: pattern must not be empty")
		}
		if r.TagKey == "" {
			return nil, fmt.Errorf("annotator: tag_key must not be empty")
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("annotator: invalid pattern %q: %w", r.Pattern, err)
		}
		compiled = append(compiled, compiledAnnotation{re: re, tagKey: r.TagKey, tagValue: r.TagValue})
	}
	return &SecretAnnotator{rules: compiled}, nil
}

// Annotate returns a new map that includes the original secrets plus annotation
// entries for every key that matches a rule. Annotation keys are formatted as
// "<originalKey>__<tagKey>".
func (a *SecretAnnotator) Annotate(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[k] = v
	}
	for k := range secrets {
		for _, rule := range a.rules {
			if rule.re.MatchString(k) {
				annotationKey := k + "__" + rule.tagKey
				out[annotationKey] = rule.tagValue
			}
		}
	}
	return out
}

// AnnotateStage returns a pipeline Stage that annotates secrets in-place.
func AnnotateStage(a *SecretAnnotator) Stage {
	return func(secrets map[string]string) (map[string]string, error) {
		return a.Annotate(secrets), nil
	}
}
