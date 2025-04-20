// pkg/license/validator.go
package license

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/adi-ber/vjal-platform/pkg/config"
	"github.com/adi-ber/vjal-platform/pkg/metrics"
)

// License represents a customer's license file structure.
type License struct {
	Key      string    `json:"license_key"`      // matches JSON field in your license.json
	Expires  time.Time `json:"expires"`          // RFC3339 timestamp
	Features []string  `json:"features"`         // list of enabled features
	DeviceID string    `json:"deviceID,omitempty"` // optional device fingerprint
}

// Validator handles license validation and feature checks.
type Validator struct {
	cfg *config.AppConfig
}

// NewValidator creates a license Validator using the application config.
func NewValidator(cfg *config.AppConfig) *Validator {
	return &Validator{cfg: cfg}
}

// Validate reads, parses, and checks the license file. It records
// total attempts and errors in Prometheus, and returns the License.
func (v *Validator) Validate(ctx context.Context) (*License, error) {
	metrics.LicenseValidationTotal.Inc()

	// Read the file
	data, err := ioutil.ReadFile(v.cfg.LicensePath)
	if err != nil {
		metrics.LicenseValidationErrors.Inc()
		return nil, fmt.Errorf("failed to read license file %s: %w", v.cfg.LicensePath, err)
	}

	// Parse JSON
	var lic License
	if err := json.Unmarshal(data, &lic); err != nil {
		metrics.LicenseValidationErrors.Inc()
		return nil, fmt.Errorf("invalid JSON in license file: %w", err)
	}

	// Check expiry
	if time.Now().After(lic.Expires) {
		metrics.LicenseValidationErrors.Inc()
		return &lic, fmt.Errorf("license expired on %s", lic.Expires.Format(time.RFC3339))
	}

	return &lic, nil
}

// CheckFeature returns true if the given feature is present in the license.
// It re-reads the license file for simplicity; a real implementation might cache.
func (v *Validator) CheckFeature(feature string) bool {
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

// HealthCheck allows you to verify license validity as a readiness check.
func (v *Validator) HealthCheck(ctx context.Context) error {
	_, err := v.Validate(ctx)
	return err
}