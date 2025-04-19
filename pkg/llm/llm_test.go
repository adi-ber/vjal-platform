// pkg/llm/llm_test.go
package llm

import (
	"context"
	"testing"
	"time"

	"github.com/adi-ber/vjal-platform/pkg/config"
	"github.com/adi-ber/vjal-platform/pkg/license"
)

func TestBuiltInEchoProvider(t *testing.T) {
	// Prepare a dummy config and license
	cfg := config.AppConfig{LLMProvider: "echo"}
	lic := &license.License{Key: "DUMMY", Expires: time.Now().Add(1 * time.Hour)}

	llmClient, err := New(cfg, lic)
	if err != nil {
		t.Fatalf("expected no error for echo provider, got %v", err)
	}

	resp, err := llmClient.Prompt(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("echo prompt error: %v", err)
	}
	expected := "echo: hello world"
	if resp != expected {
		t.Errorf("expected %q, got %q", expected, resp)
	}

	// HealthCheck should succeed
	if err := llmClient.HealthCheck(context.Background()); err != nil {
		t.Errorf("echo HealthCheck failed: %v", err)
	}
}

func TestUnknownProvider(t *testing.T) {
	cfg := config.AppConfig{LLMProvider: "doesnotexist"}
	lic := &license.License{Key: "DUMMY", Expires: time.Now().Add(1 * time.Hour)}

	_, err := New(cfg, lic)
	if err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
}
