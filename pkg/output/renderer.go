package output

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/yuin/goldmark"
)

// Renderer handles converting Markdown to various formats.
type Renderer struct {
	md goldmark.Markdown
}

// NewRenderer creates a new Renderer instance.
func NewRenderer() *Renderer {
	return &Renderer{
		md: goldmark.New(),
	}
}

// ToHTML converts a Markdown string into HTML.
func (r *Renderer) ToHTML(input string) (string, error) {
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(input), &buf); err != nil {
		return "", fmt.Errorf("HTML conversion failed: %w", err)
	}
	return buf.String(), nil
}

// ToPDF converts a Markdown string into a PDF byte slice.
// Currently not implemented.
func (r *Renderer) ToPDF(input string) ([]byte, error) {
	return nil, errors.New("PDF generation not implemented")
}
