// pkg/config/config.go
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// AppConfig holds application configuration loaded from JSON and environment variables.
type AppConfig struct {
	Env             string            `json:"env"`             // "development" or "production"
	HTTPPort        int               `json:"httpPort"`        // port for HTTP server
	LicensePath     string            `json:"licensePath"`     // path to license.json
	LLMProvider     string            `json:"llmProvider"`     // "openai", "vjal", or "offline"
	LLMConfig       map[string]string `json:"llmConfig"`       // provider-specific settings
	FormSchema      string            `json:"formSchema"`      // path to JSON form schema
	OutputDir       string            `json:"outputDir"`       // path to write outputs
	MetricsEndpoint string            `json:"metricsEndpoint"` // pushgateway URL or empty
}

// Load reads the JSON config file at path, merges in environment variables, and validates required fields.
func Load(path string) (*AppConfig, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid config path: %w", err)
	}

	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", absPath, err)
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid JSON in config file: %w", err)
	}

	// Override with environment variables if set
	overrideEnv(&cfg)

	// Validate required fields
	if cfg.LicensePath == "" {
		return nil, errors.New("licensePath is required in config")
	}
	if cfg.FormSchema == "" {
		return nil, errors.New("formSchema is required in config")
	}
	if cfg.LLMProvider == "" {
		return nil, errors.New("llmProvider is required in config")
	}
	if cfg.HTTPPort == 0 {
		cfg.HTTPPort = 8080 // default port
	}

	return &cfg, nil
}

// overrideEnv checks for environment variables prefixed with VJAL_ and overrides fields.
func overrideEnv(cfg *AppConfig) {
	if v := os.Getenv("VJAL_ENV"); v != "" {
		cfg.Env = v
	}
	if v := os.Getenv("VJAL_HTTP_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.HTTPPort)
	}
	if v := os.Getenv("VJAL_LICENSE_PATH"); v != "" {
		cfg.LicensePath = v
	}
	if v := os.Getenv("VJAL_LLM_PROVIDER"); v != "" {
		cfg.LLMProvider = v
	}
	if v := os.Getenv("VJAL_FORM_SCHEMA"); v != "" {
		cfg.FormSchema = v
	}
	// Add overrides for other fields as needed
}