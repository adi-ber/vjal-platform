// pkg/security/crypto_test.go
package security

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"testing"
)

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	plaintext := []byte("sensitive data")
	licenseKey := "TEST-LIC"
	deviceID := []byte("DEVICE1234")

	// Encrypt
	ciphertext, err := Encrypt(plaintext, licenseKey, deviceID)
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}

	// Decrypt
	decrypted, err := Decrypt(ciphertext, licenseKey, deviceID)
	if err != nil {
		t.Fatalf("Decrypt error: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("expected %s, got %s", plaintext, decrypted)
	}
}

func TestDecrypt_TamperDetect(t *testing.T) {
	plaintext := []byte("sensitive data")
	licenseKey := "TEST-LIC"
	deviceID := []byte("DEVICE1234")

	ciphertext, err := Encrypt(plaintext, licenseKey, deviceID)
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}

	// Tamper last byte
	ciphertext[len(ciphertext)-1] ^= 0xFF

	_, err = Decrypt(ciphertext, licenseKey, deviceID)
	if err == nil {
		t.Fatal("expected decryption error after tampering, got nil")
	}
}

func TestValidateHash_Success(t *testing.T) {
	// Create a temp file
	f, err := ioutil.TempFile("", "hash-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	content := []byte("hello world")
	ioutil.WriteFile(f.Name(), content, 0600)
	f.Close()

	// Compute expected hash
	h := sha256.Sum256(content)
	expected := hex.EncodeToString(h[:])

	// Validate
	err = ValidateHash(f.Name(), expected)
	if err != nil {
		t.Fatalf("ValidateHash failed: %v", err)
	}
}

func TestValidateHash_Failure(t *testing.T) {
	// Create a temp file
	f, err := ioutil.TempFile("", "hash-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	ioutil.WriteFile(f.Name(), []byte("data1"), 0600)
	f.Close()

	// Wrong expected hash
	expected := "abcdef123456"

	// Validate
	err = ValidateHash(f.Name(), expected)
	if err == nil {
		t.Fatal("expected ValidateHash to error on mismatch, got nil")
	}
}