// pkg/form/engine.go
package form

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/adi-ber/vjal-platform/pkg/storage"
	"github.com/adi-ber/vjal-platform/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// Form represents a multi-page form with persisted state.
type Form struct {
	Schema map[string]interface{} // raw JSON schema
	Values map[string]interface{} // in-memory merged values
	store  *storage.Store         // storage backend for state
	formID string                 // namespace for persisted state
}

// New loads the form schema, attaches storage under formID, and registers metrics.
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

// RenderPage returns placeholder HTML for the given pageID, with metrics.
func (f *Form) RenderPage(ctx context.Context, pageID string) (string, error) {
	metrics.FormRenderTotal.Inc()
	timer := prometheus.NewTimer(metrics.FormRenderDuration)
	defer timer.ObserveDuration()

	// Stub implementation: real implementation will generate HTML from schema
	return fmt.Sprintf("<div>Rendering page: %s</div>", pageID), nil
}

// Validate merges input into Values, records warnings, and updates metrics.
func (f *Form) Validate(ctx context.Context, pageID string, input map[string]interface{}) ([]string, error) {
	metrics.FormValidationTotal.Inc()
	// Stub: merge input to Values
	for k, v := range input {
		f.Values[k] = v
	}

	// No actual warnings in stub
	warnings := []string{}
	metrics.FormValidationWarnings.Add(float64(len(warnings)))

	return warnings, nil
}

// NextPage stub—always returns "end".
func (f *Form) NextPage(currentPageID string) (string, error) {
	return "end", nil
}

// SaveState persists the input map for the given pageID and updates metrics.
func (f *Form) SaveState(ctx context.Context, pageID string, input map[string]interface{}) error {
	if err := f.store.Save(f.formID, pageID, input); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	return nil
}

// LoadState retrieves the persisted input map for pageID and updates metrics.
func (f *Form) LoadState(ctx context.Context, pageID string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	if err := f.store.Load(f.formID, pageID, &data); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}
	// Merge into in-memory Values
	for k, v := range data {
		f.Values[k] = v
	}
	return data, nil
}