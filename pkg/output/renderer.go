// pkg/output/renderer.go
package output

import (
	"bytes"
	"fmt"

	"github.com/adi-ber/vjal-platform/pkg/metrics"
	"github.com/jung-kurt/gofpdf"
	"github.com/yuin/goldmark"

	// DejaVu Sans TTF data
	dejavusans "github.com/go-fonts/dejavu/dejavusans"
)

// dejavuTTF is the raw TTF bytes for DejaVu Sans, imported from the go-fonts package.
var dejavuTTF = dejavusans.TTF

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

// ToHTML converts Markdown to HTML.
func (r *Renderer) ToHTML(input string) (string, error) {
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(input), &buf); err != nil {
		return "", fmt.Errorf("HTML conversion failed: %w", err)
	}
	return buf.String(), nil
}

// ToPDF converts a Markdown string into a simple PDF with full UTF-8 support.
func (r *Renderer) ToPDF(input string) ([]byte, error) {
	metrics.OutputPDFTotal.Inc()

	// Create a new PDF, register our DejaVu Sans TrueType font, then use it.
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes("dejavu", "", dejavuTTF)
	pdf.SetFont("dejavu", "", 12)
	pdf.AddPage()

	// Write the markdown text into the PDF
	pdf.MultiCell(0, 6, input, "", "", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		metrics.OutputPDFErrors.Inc()
		return nil, fmt.Errorf("PDF generation failed: %w", err)
	}
	return buf.Bytes(), nil
}
