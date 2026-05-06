package sync

import (
	"bytes"
	"fmt"
	"text/template"
)

// SecretTemplater renders Go templates using secrets as template data.
// It allows injecting secret values into arbitrary string templates,
// useful for generating config snippets or connection strings.
type SecretTemplater struct {
	tmpl *template.Template
}

// NewSecretTemplater parses the given template string and returns a
// SecretTemplater. Returns an error if the template is invalid.
func NewSecretTemplater(tmplStr string) (*SecretTemplater, error) {
	if tmplStr == "" {
		return nil, fmt.Errorf("template string must not be empty")
	}
	t, err := template.New("secret").Option("missingkey=error").Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}
	return &SecretTemplater{tmpl: t}, nil
}

// Render executes the template with the provided secrets map as data.
// Returns the rendered string or an error if a referenced key is missing.
func (st *SecretTemplater) Render(secrets map[string]string) (string, error) {
	if secrets == nil {
		secrets = map[string]string{}
	}
	var buf bytes.Buffer
	if err := st.tmpl.Execute(&buf, secrets); err != nil {
		return "", fmt.Errorf("template render failed: %w", err)
	}
	return buf.String(), nil
}

// TemplateStage returns a pipeline Stage that renders a template string
// using the current secrets and stores the result under outputKey.
// The secrets map is passed through unchanged.
func TemplateStage(tmplStr, outputKey string) func(map[string]string) (map[string]string, error) {
	return func(secrets map[string]string) (map[string]string, error) {
		st, err := NewSecretTemplater(tmplStr)
		if err != nil {
			return nil, err
		}
		rendered, err := st.Render(secrets)
		if err != nil {
			return nil, err
		}
		out := make(map[string]string, len(secrets)+1)
		for k, v := range secrets {
			out[k] = v
		}
		out[outputKey] = rendered
		return out, nil
	}
}
