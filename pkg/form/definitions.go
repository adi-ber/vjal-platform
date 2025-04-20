package form

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// PromptField describes one input in your form.
type PromptField struct {
	ID            string         `json:"id"`                      // lowercase field name
	Label         string         `json:"label"`                   // human‑readable label
	Type          string         `json:"type"`                    // html input type (text, number, etc.)
	Validations   *Validations   `json:"validations,omitempty"`   // optional validation rules
	Condition     *Condition     `json:"condition,omitempty"`     // optional display condition
	LLMValidation *LLMValidation `json:"llmValidation,omitempty"` // optional LLM‑based check
	Options       []string       `json:"options,omitempty"`       // for selects/radios
	Placeholder   string         `json:"placeholder,omitempty"`   // optional placeholder text
}

// Validations holds basic client/server rules.
type Validations struct {
	Required  bool    `json:"required,omitempty"`
	MinLength int     `json:"minLength,omitempty"`
	MaxLength int     `json:"maxLength,omitempty"`
	Min       float64 `json:"min,omitempty"`
	Max       float64 `json:"max,omitempty"`
	Pattern   string  `json:"pattern,omitempty"`
}

// Condition controls whether to show a field.
type Condition struct {
	FieldID string      `json:"fieldId"`
	Value   interface{} `json:"value"`
}

// LLMValidation flags a field for AI checks.
type LLMValidation struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
	Trigger string `json:"trigger"`
}

// FormDefinitions maps promptKey → slice of PromptField.
type FormDefinitions map[string][]PromptField

// LoadDefinitionsDir reads every .json file in dir, merges into one FormDefinitions.
// Invalid JSON or duplicate keys are logged and skipped.
func LoadDefinitionsDir(dir string) (FormDefinitions, error) {
	defs := make(FormDefinitions)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading definitions dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			log.Printf("⚠️  could not read %s: %v", path, err)
			continue
		}

		var tmp FormDefinitions
		if err := json.Unmarshal(b, &tmp); err != nil {
			log.Printf("⚠️  invalid JSON in %s: %v", path, err)
			continue
		}

		for key, fields := range tmp {
			if _, exists := defs[key]; exists {
				log.Printf("⚠️  duplicate form key %q in %s – skipping", key, path)
				continue
			}
			defs[key] = fields
		}
	}

	if len(defs) == 0 {
		return nil, fmt.Errorf("no valid form definitions found in %s", dir)
	}
	return defs, nil
}
