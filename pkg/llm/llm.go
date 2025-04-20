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

// Client is the interface for our LLM backends.
type Client interface {
	// Prompt sends a text prompt and returns the model's reply.
	Prompt(ctx context.Context, prompt string) (string, error)
	// HealthCheck verifies that the backend is reachable/ready.
	HealthCheck(ctx context.Context) error
}

// New selects and instantiates the proper Client, then wraps it for metrics.
func New(cfg *config.AppConfig, lic *license.License) (Client, error) {
	var base Client
	var err error

	switch cfg.LLMProvider {
	case "openai":
		base = NewOpenAIClient(cfg.LLMConfig["openai_key"])
	case "offline":
		base, err = NewOfflineClient(cfg.LLMConfig)
		if err != nil {
			return nil, err
		}
	case "echo":
		base = &echoClient{}
	default:
		return nil, fmt.Errorf("unknown LLM provider: %q", cfg.LLMProvider)
	}

	// Wrap in metrics collector
	return &metricsClient{
		provider: cfg.LLMProvider,
		next:     base,
	}, nil
}

// --------------------
// metricsClient decorates any Client to capture Prometheus metrics.
// --------------------
type metricsClient struct {
	provider string
	next     Client
}

func (m *metricsClient) Prompt(ctx context.Context, prompt string) (string, error) {
	// increment request count
	metrics.LLMRequestsTotal.WithLabelValues(m.provider).Inc()
	// time the request
	timer := prometheus.NewTimer(metrics.LLMRequestDuration.WithLabelValues(m.provider))
	defer timer.ObserveDuration()

	resp, err := m.next.Prompt(ctx, prompt)
	if err != nil {
		metrics.LLMErrorsTotal.WithLabelValues(m.provider).Inc()
	}
	return resp, err
}

func (m *metricsClient) HealthCheck(ctx context.Context) error {
	return m.next.HealthCheck(ctx)
}

// --------------------
// echoClient simply echoes back the prompt.
// --------------------
type echoClient struct{}

func (e *echoClient) Prompt(ctx context.Context, prompt string) (string, error) {
	return prompt, nil
}

func (e *echoClient) HealthCheck(ctx context.Context) error {
	return nil
}
