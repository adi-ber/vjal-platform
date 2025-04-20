package llm

import (
	"context"
	"testing"

	"github.com/adi-ber/vjal-platform/pkg/config"
	"github.com/adi-ber/vjal-platform/pkg/license"
)

func TestEchoProvider_PromptAndHealth(t *testing.T) {
	// Prepare dummy config and license
	cfg := config.AppConfig{LLMProvider: "echo"}
	lic := &license.License{}

	// Instantiate using a pointer to config.AppConfig
	provider, err := New(&cfg, lic)
	if err != nil {
		t.Fatalf("expected no error getting echo provider, got %v", err)
	}

	// Test Prompt
	input := "hello"
	out, err := provider.Prompt(context.Background(), input)
	if err != nil {
		t.Fatalf("Prompt error: %v", err)
	}
	if out != input {
		t.Errorf("expected echo %q, got %q", input, out)
	}

	// Test HealthCheck
	if err := provider.HealthCheck(context.Background()); err != nil {
		t.Errorf("expected no healthcheck error, got %v", err)
	}
}

func TestUnknownProvider(t *testing.T) {
	cfg := config.AppConfig{LLMProvider: "no-such"}
	lic := &license.License{}

	// Use pointer to cfg
	_, err := New(&cfg, lic)
	if err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
}

func TestMetricsLabels(t *testing.T) {
	// provider labels shouldn't affect compilation; just ensure New accepts pointer
	cfg := config.AppConfig{LLMProvider: "echo"}
	lic := &license.License{}

	_, err := New(&cfg, lic)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
