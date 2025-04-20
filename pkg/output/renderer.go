// pkg/output/renderer.go
package output

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/adi-ber/vjal-platform/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yuin/goldmark"
)

// Renderer handles converting Markdown to various formats.
type Renderer struct {
	md goldmark.Markdown
}

// NewRenderer creates a new Renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		md: goldmark.New(),
	}
}

// ToHTML converts Markdown to HTML, recording duration.
func (r *Renderer) ToHTML(input string) (string, error) {
	timer := prometheus.NewTimer(metrics.OutputHTMLDuration)
	defer timer.ObserveDuration()

	var buf bytes.Buffer
	if err := r.md.Convert([]byte(input), &buf); err != nil {
		return "", fmt.Errorf("HTML conversion failed: %w", err)
	}
	return buf.String(), nil
}

// ToPDF attempts to convert Markdown to PDF.
// Currently a stub that records an error.
func (r *Renderer) ToPDF(input string) ([]byte, error) {
	metrics.OutputPDFTotal.Inc()
	metrics.OutputPDFErrors.Inc()
	return nil, errors.New("PDF generation not implemented")
}