package sync

import (
	"context"
	"fmt"

	"github.com/your-org/vault-sync/internal/vault"
)

// LoaderFunc is a function type that satisfies a simple loader interface.
type LoaderFunc func(ctx context.Context, path string) (map[string]string, error)

// SecretsLoader is the interface for loading secrets from a path.
type SecretsLoader interface {
	Load(ctx context.Context, path string) (map[string]string, error)
}

// PipelineStage represents a named transformation step applied to secrets.
type PipelineStage struct {
	Name    string
	Apply   func(secrets map[string]string) (map[string]string, error)
}

// Pipeline chains multiple transformation stages over loaded secrets.
type Pipeline struct {
	loader SecretsLoader
	stages []PipelineStage
}

// NewPipeline creates a Pipeline with the given loader and stages.
func NewPipeline(loader SecretsLoader, stages ...PipelineStage) (*Pipeline, error) {
	if loader == nil {
		return nil, fmt.Errorf("pipeline: loader must not be nil")
	}
	return &Pipeline{loader: loader, stages: stages}, nil
}

// Run loads secrets from path and applies all stages in order.
func (p *Pipeline) Run(ctx context.Context, path string) (map[string]string, error) {
	secrets, err := p.loader.Load(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("pipeline: load %q: %w", path, err)
	}

	for _, stage := range p.stages {
		secrets, err = stage.Apply(secrets)
		if err != nil {
			return nil, fmt.Errorf("pipeline: stage %q: %w", stage.Name, err)
		}
	}
	return secrets, nil
}

// FilterStage returns a PipelineStage that filters secrets by prefixes.
func FilterStage(prefixes []string) PipelineStage {
	return PipelineStage{
		Name: "filter",
		Apply: func(secrets map[string]string) (map[string]string, error) {
			return vault.FilterSecrets(secrets, prefixes), nil
		},
	}
}

// RenameStage returns a PipelineStage that applies rename rules.
func RenameStage(rules []vault.RenameRule) PipelineStage {
	return PipelineStage{
		Name: "rename",
		Apply: func(secrets map[string]string) (map[string]string, error) {
			return vault.ApplyRenames(secrets, rules), nil
		},
	}
}

// TransformStage returns a PipelineStage that applies key transform rules.
func TransformStage(rules []vault.TransformRule) PipelineStage {
	return PipelineStage{
		Name: "transform",
		Apply: func(secrets map[string]string) (map[string]string, error) {
			return vault.TransformSecrets(secrets, rules)
		},
	}
}
