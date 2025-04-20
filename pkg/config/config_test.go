package config

import (
	"io/ioutil"
	"os"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := ioutil.TempFile("", "config-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	json := `{
		"env": "development",
		"httpPort": 9090,
		"licensePath": "/tmp/license.json",
		"llmProvider": "openai",
		"llmConfig": {"apiKey": "testkey"},
		"formSchema": "schema.json",
		"outputDir": "out",
		"metricsEndpoint": "http://localhost:9091"
	}`
	path := writeTempFile(t, json)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.HTTPPort != 9090 {
		t.Errorf("expected HTTPPort 9090, got %d", cfg.HTTPPort)
	}
	if cfg.LLMProvider != "openai" {
		t.Errorf("expected LLMProvider openai, got %s", cfg.LLMProvider)
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	json := `{"env": "prod"}`
	path := writeTempFile(t, json)
	defer os.Remove(path)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing required fields, got nil")
	}
}
