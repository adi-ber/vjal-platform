// pkg/storage/storage_test.go
package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestStore_SaveLoad(t *testing.T) {
	// Setup temp DB file
	tmpDir, err := ioutil.TempDir("", "storagetest")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "state.db")
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	namespace := "form"
	key := "page1"
	// Note: JSON unmarshals numbers as float64
	original := map[string]interface{}{
		"name":  "Alice",
		"score": 100.0,  // float64 literal
	}

	// Save state
	if err := store.Save(namespace, key, original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load back
	loaded := make(map[string]interface{})
	if err := store.Load(namespace, key, &loaded); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !reflect.DeepEqual(loaded, original) {
		t.Errorf("Loaded %v, want %v", loaded, original)
	}

	// Loading non-existent key should leave the map empty
	empty := make(map[string]interface{})
	if err := store.Load("nope", "none", &empty); err != nil {
		t.Fatalf("Load(nonexistent) error: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected empty map for nonexistent key, got %v", empty)
	}
}