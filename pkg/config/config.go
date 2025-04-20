// pkg/config/config.go
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/adi-ber/vjal-platform/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
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

// Load reads the JSON config file at the given path, overrides via env vars,
// validates required fields, and records metrics on load duration and errors.
func Load(path string) (*AppConfig, error) {
	// Start timer for load duration
	timer := prometheus.NewTimer(metrics.ConfigLoadDuration)
	defer timer.ObserveDuration()

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		metrics.ConfigLoadErrors.Inc()
		return nil, fmt.Errorf("invalid config path %q: %w", path, err)
	}

	// Read file
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		metrics.ConfigLoadErrors.Inc()
		return nil, fmt.Errorf("failed to read config file %s: %w", absPath, err)
	}

	// Parse JSON
	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		metrics.ConfigLoadErrors.Inc()
		return nil, fmt.Errorf("invalid JSON in config file: %w", err)
	}

	// Override with environment variables
	overrideEnv(&cfg)

	// Validate required fields
	if cfg.LicensePath == "" {
		metrics.ConfigLoadErrors.Inc()
		return nil, errors.New("licensePath is required in config")
	}
	if cfg.FormSchema == "" {
		metrics.ConfigLoadErrors.Inc()
		return nil, errors.New("formSchema is required in config")
	}
	if cfg.LLMProvider == "" {
		metrics.ConfigLoadErrors.Inc()
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