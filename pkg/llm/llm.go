// pkg/llm/llm.go
package llm

import (
	"context"
	"fmt"

	"github.com/adi-ber/vjal-platform/pkg/config"
	"github.com/adi-ber/vjal-platform/pkg/license"
	"github.com/adi-ber/vjal-platform/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type LLM interface {
	Prompt(ctx context.Context, input string) (string, error)
	HealthCheck(ctx context.Context) error
}

type factory func(cfg *config.AppConfig, lic *license.License) (LLM, error)

var providers = make(map[string]factory)

// Register registers an LLM provider under the given name.
func Register(name string, fn factory) {
	providers[name] = fn
}

// New constructs the chosen LLM based on cfg.LLMProvider.
func New(cfg *config.AppConfig, lic *license.License) (LLM, error) {
	if fn, ok := providers[cfg.LLMProvider]; ok {
		return fn(cfg, lic)
	}
	return nil, fmt.Errorf("unknown LLM provider: %s", cfg.LLMProvider)
}

// echoProvider is a simple stub that echoes the input.
type echoProvider struct{}

// Prompt echoes the input and records metrics.
func (e *echoProvider) Prompt(ctx context.Context, input string) (string, error) {
	metrics.LLMRequestsTotal.WithLabelValues("echo").Inc()
	timer := prometheus.NewTimer(metrics.LLMRequestDuration.WithLabelValues("echo"))
	defer timer.ObserveDuration()

	return input, nil
}

// HealthCheck always succeeds for the echo stub.
func (e *echoProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func init() {
	// Register the echo stub as the "echo" and "openai" provider (alias)
	Register("echo", func(_ *config.AppConfig, _ *license.License) (LLM, error) {
		return &echoProvider{}, nil
	})
	Register("openai", func(_ *config.AppConfig, _ *license.License) (LLM, error) {
		return &echoProvider{}, nil
	})
}