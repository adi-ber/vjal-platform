// pkg/license/validator.go
package license

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/adi-ber/vjal-platform/pkg/config"
)

// License represents a customer's license file structure.
type License struct {
	Key      string    `json:"key"`
	Expires  time.Time `json:"expires"`  // RFC3339 format in JSON
	Features []string  `json:"features"`
	DeviceID string    `json:"deviceID,omitempty"` // optional device fingerprint
}

// Validator handles license validation and feature checks.
type Validator struct {
	cfg config.AppConfig
}

// NewValidator creates a license Validator using the application config.
func NewValidator(cfg config.AppConfig) *Validator {
	return &Validator{cfg: cfg}
}

// Validate reads and validates the license file, returning a License or an error.
func (v *Validator) Validate(ctx context.Context) (*License, error) {
	// Read license file
	data, err := ioutil.ReadFile(v.cfg.LicensePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read license file %s: %w", v.cfg.LicensePath, err)
	}

	// Unmarshal JSON
	var lic License
	if err := json.Unmarshal(data, &lic); err != nil {
		return nil, fmt.Errorf("invalid JSON in license file: %w", err)
	}

	// Check expiry
	if time.Now().After(lic.Expires) {
		return &lic, fmt.Errorf("license expired on %s", lic.Expires.Format(time.RFC3339))
	}

	return &lic, nil
}

// CheckFeature returns true if the license includes the given feature.
func (v *Validator) CheckFeature(feature string) bool {
	if lic := v; lic != nil {
		// This method expects Validate to have been called previously.
	}
	// features are checked on the last validated license
	// For simplicity, re-read license file (could cache in real impl)
	data, err := ioutil.ReadFile(v.cfg.LicensePath)
	if err != nil {
		return false
	}
	var lic License
	if err := json.Unmarshal(data, &lic); err != nil {
		return false
	}
	for _, f := range lic.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// HealthCheck can be used to verify local license cache readiness in the future.
func (v *Validator) HealthCheck(ctx context.Context) error {
	// Currently, Validate does all checks; stub for readiness
	_, err := v.Validate(ctx)
	return err
}