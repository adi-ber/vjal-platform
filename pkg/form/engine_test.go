// pkg/form/engine_test.go
package form

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/adi-ber/vjal-platform/pkg/storage"
)

// writeTempSchema creates a temp JSON schema file for tests.
func writeTempSchema(t *testing.T, content string) string {
	t.Helper()
	f, err := ioutil.TempFile("", "schema-*.json")
	if err != nil {
		t.Fatalf("failed to create temp schema file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		t.Fatalf("failed to write schema file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestSaveLoadState(t *testing.T) {
	// Prepare a temp DB
	tmpDir, err := ioutil.TempDir("", "formstate")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "state.db")
	store, err := storage.New(dbPath)
	if err != nil {
		t.Fatalf("storage.New error: %v", err)
	}

	// Create a minimal form schema file
	schema := `{"pages":[{"id":"start","fields":[]}]}`
	schemaPath := writeTempSchema(t, schema)
	defer os.Remove(schemaPath)

	// Instantiate the Form with storage
	formID := "testForm"
	f, err := New(schemaPath, store, formID)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	pageID := "start"
	input := map[string]interface{}{"field": "value", "num": 5.0}

	// Save
	if err := f.SaveState(context.Background(), pageID, input); err != nil {
		t.Fatalf("SaveState error: %v", err)
	}

	// Load and compare
	loaded, err := f.LoadState(context.Background(), pageID)
	if err != nil {
		t.Fatalf("LoadState error: %v", err)
	}
	if !reflect.DeepEqual(loaded, input) {
		t.Errorf("Loaded %v, want %v", loaded, input)
	}
}