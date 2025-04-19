// pkg/llm/llm.go
package llm

import (
	"context"
	"fmt"

	"github.com/adi-ber/vjal-platform/pkg/config"
	"github.com/adi-ber/vjal-platform/pkg/license"
)

// LLM defines the interface for interacting with language models.
type LLM interface {
	Prompt(ctx context.Context, input string) (string, error)
	HealthCheck(ctx context.Context) error
}

// factory is a function that creates an LLM based on config and license.
type factory func(cfg config.AppConfig, lic *license.License) (LLM, error)

var providers = make(map[string]factory)

// Register makes a new LLM provider available by name.
func Register(name string, f factory) {
	providers[name] = f
}

// New returns an LLM instance based on the configured provider.
func New(cfg config.AppConfig, lic *license.License) (LLM, error) {
	factoryFn, ok := providers[cfg.LLMProvider]
	if !ok {
		return nil, fmt.Errorf("unknown LLM provider: %s", cfg.LLMProvider)
	}
	return factoryFn(cfg, lic)
}

// Built-in echo provider for testing and defaults
type echoLLM struct{}

func (e *echoLLM) Prompt(ctx context.Context, input string) (string, error) {
	return fmt.Sprintf("echo: %s", input), nil
}

func (e *echoLLM) HealthCheck(ctx context.Context) error {
	return nil
}

func init() {
	// Register built-in "echo" provider
	Register("echo", func(cfg config.AppConfig, lic *license.License) (LLM, error) {
		return &echoLLM{}, nil
	})
}