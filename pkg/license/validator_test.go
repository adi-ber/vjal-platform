// pkg/license/validator_test.go
package license

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/adi-ber/vjal-platform/pkg/config"
)

func writeTempLicense(t *testing.T, content string) string {
	t.Helper()
	f, err := ioutil.TempFile("", "license-*.json")
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

func TestValidate_ValidLicense(t *testing.T) {
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	json := fmt.Sprintf(`{
	"key": "ABC-123",
	"expires": "%s",
	"features": ["offline", "pdf"]
}`, future)
	path := writeTempLicense(t, json)
	defer os.Remove(path)

	cfg := config.AppConfig{LicensePath: path}
	validator := NewValidator(cfg)
	lic, err := validator.Validate(context.Background())
	if err != nil {
		t.Fatalf("expected valid license, got error: %v", err)
	}
	if lic.Key != "ABC-123" {
		t.Errorf("expected key ABC-123, got %s", lic.Key)
	}
}

func TestValidate_ExpiredLicense(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	json := fmt.Sprintf(`{
	"key": "ABC-123",
	"expires": "%s",
	"features": ["offline"]
}`, past)
	path := writeTempLicense(t, json)
	defer os.Remove(path)

	cfg := config.AppConfig{LicensePath: path}
	validator := NewValidator(cfg)
	_, err := validator.Validate(context.Background())
	if err == nil {
		t.Fatal("expected error for expired license, got nil")
	}
}

func TestCheckFeature(t *testing.T) {
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	json := fmt.Sprintf(`{
	"key": "ABC-123",
	"expires": "%s",
	"features": ["offline", "pdf"]
}`, future)
	path := writeTempLicense(t, json)
	defer os.Remove(path)

	cfg := config.AppConfig{LicensePath: path}
	validator := NewValidator(cfg)

	if !validator.CheckFeature("pdf") {
		t.Errorf("expected CheckFeature(pdf) to be true")
	}
	if validator.CheckFeature("nonexistent") {
		t.Errorf("expected CheckFeature(nonexistent) to be false")
	}
}