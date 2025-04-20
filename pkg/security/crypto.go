// pkg/security/crypto.go
package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"

	"golang.org/x/crypto/scrypt"
)

const (
	saltLen = 16
	keyLen  = 32 // 256-bit AES key
)

// deriveKey uses scrypt KDF to derive an AES key from the licenseKey, deviceID, and salt.
func deriveKey(licenseKey string, deviceID, salt []byte) ([]byte, error) {
	// N=1<<15, r=8, p=1 as reasonable work factor
	return scrypt.Key([]byte(licenseKey), append(deviceID, salt...), 1<<15, 8, 1, keyLen)
}

// Encrypt encrypts plaintext using AES-GCM. Returns salt||nonce||ciphertext (no auth tag separately).
func Encrypt(plaintext []byte, licenseKey string, deviceID []byte) ([]byte, error) {
	// Generate random salt
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key
	key, err := deriveKey(licenseKey, deviceID, salt)
	if err != nil {
		return nil, fmt.Errorf("key derivation failed: %w", err)
	}

	// Create AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Seal authenticates and encrypts
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Return concatenated: salt||nonce||ciphertext
	out := make([]byte, 0, saltLen+len(nonce)+len(ciphertext))
	out = append(out, salt...)
	out = append(out, nonce...)
	out = append(out, ciphertext...)
	return out, nil
}

// Decrypt decrypts data produced by Encrypt, verifying authenticity.
func Decrypt(data []byte, licenseKey string, deviceID []byte) ([]byte, error) {
	if len(data) < saltLen {
		return nil, fmt.Errorf("ciphertext too short")
	}
	// Extract salt
	salt := data[:saltLen]
	// Derive key
	key, err := deriveKey(licenseKey, deviceID, salt)
	if err != nil {
		return nil, fmt.Errorf("key derivation failed: %w", err)
	}

	// AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < saltLen+nonceSize {
		return nil, fmt.Errorf("ciphertext too short for nonce")
	}
	// Extract nonce and ciphertext
	nonce := data[saltLen : saltLen+nonceSize]
	ciphertext := data[saltLen+nonceSize:]

	// Decrypt and authenticate
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}
	return plaintext, nil
}

// ValidateHash checks that the SHA-256 hash of the file at path matches the expected hex string.
func ValidateHash(path, expectedHex string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}
	actual := sha256.Sum256(data)
	actualHex := hex.EncodeToString(actual[:])
	if !bytes.Equal([]byte(actualHex), []byte(expectedHex)) {
		return fmt.Errorf("hash mismatch: expected %s, got %s", expectedHex, actualHex)
	}
	return nil
}