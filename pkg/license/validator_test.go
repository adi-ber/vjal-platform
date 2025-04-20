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

func writeTempLicenseFile(t *testing.T, content string) string {
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

func TestValidate_Valid(t *testing.T) {
	expires := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	jsonData := fmt.Sprintf(
		`{"license_key":"TEST","expires":"%s","features":["f1"]}`, expires,
	)
	path := writeTempLicenseFile(t, jsonData)
	defer os.Remove(path)

	cfg := config.AppConfig{LicensePath: path}
	v := NewValidator(&cfg)
	lic, err := v.Validate(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if lic.Key != "TEST" {
		t.Errorf("expected key TEST, got %s", lic.Key)
	}
}

func TestValidate_Expired(t *testing.T) {
	expires := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	jsonData := fmt.Sprintf(
		`{"license_key":"OLD","expires":"%s","features":[]}`, expires,
	)
	path := writeTempLicenseFile(t, jsonData)
	defer os.Remove(path)

	cfg := config.AppConfig{LicensePath: path}
	v := NewValidator(&cfg)
	_, err := v.Validate(context.Background())
	if err == nil {
		t.Fatal("expected error for expired license, got nil")
	}
}

func TestCheckFeature(t *testing.T) {
	jsonData := `{"license_key":"F","expires":"2099-12-31T23:59:59Z","features":["a","b"]}`
	path := writeTempLicenseFile(t, jsonData)
	defer os.Remove(path)

	cfg := config.AppConfig{LicensePath: path}
	v := NewValidator(&cfg)
	if !v.CheckFeature("a") {
		t.Error("expected feature 'a' to be present")
	}
	if v.CheckFeature("z") {
		t.Error("did not expect feature 'z'")
	}
}

func TestHealthCheck(t *testing.T) {
	jsonData := `{"license_key":"H","expires":"2099-12-31T23:59:59Z","features":[]}`
	path := writeTempLicenseFile(t, jsonData)
	defer os.Remove(path)

	cfg := config.AppConfig{LicensePath: path}
	v := NewValidator(&cfg)
	if err := v.HealthCheck(context.Background()); err != nil {
		t.Errorf("expected HealthCheck to succeed, got %v", err)
	}
}
