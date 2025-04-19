// pkg/form/engine.go
package form

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/adi-ber/vjal-platform/pkg/storage"
)

// Form represents a multi-page form with persisted state.
type Form struct {
	Schema map[string]interface{}       // raw JSON schema
	Values map[string]interface{}       // in-memory merged values
	store  *storage.Store               // storage backend for state
	formID string                       // namespace for persisted state
}

// New loads the form schema and attaches the given storage.Store under formID.
func New(schemaPath string, store *storage.Store, formID string) (*Form, error) {
	data, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read form schema %s: %w", schemaPath, err)
	}
	var schema map[string]interface{}
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("invalid JSON in form schema: %w", err)
	}
	return &Form{
		Schema: schema,
		Values: make(map[string]interface{}),
		store:  store,
		formID: formID,
	}, nil
}

// RenderPage returns placeholder HTML for the given pageID.
func (f *Form) RenderPage(ctx context.Context, pageID string) (string, error) {
	return fmt.Sprintf("<div>Rendering page: %s</div>", pageID), nil
}

// Validate merges input into Values and returns any warnings.
func (f *Form) Validate(ctx context.Context, pageID string, input map[string]interface{}) ([]string, error) {
	for k, v := range input {
		f.Values[k] = v
	}
	return []string{}, nil    // return an empty slice instead of nil
}

// NextPage stub—always returns "end".
func (f *Form) NextPage(currentPageID string) (string, error) {
	return "end", nil
}

// SaveState persists the input map for the given pageID.
func (f *Form) SaveState(ctx context.Context, pageID string, input map[string]interface{}) error {
	f.Values[pageID] = input
	if err := f.store.Save(f.formID, pageID, input); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	return nil
}

// LoadState retrieves the persisted input map for pageID.
func (f *Form) LoadState(ctx context.Context, pageID string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	if err := f.store.Load(f.formID, pageID, &data); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}
	for k, v := range data {
		f.Values[k] = v
	}
	return data, nil
}