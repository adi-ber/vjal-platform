package output

import (
	"bytes"
	"testing"
)

func TestToHTML_Basic(t *testing.T) {
	r := NewRenderer()
	html, err := r.ToHTML("**bold**")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains([]byte(html), []byte("<strong>bold</strong>")) {
		t.Errorf("expected HTML to contain <strong>bold</strong>, got %q", html)
	}
}

func TestToPDF_ValidPDF(t *testing.T) {
	r := NewRenderer()
	pdf, err := r.ToPDF("Hello PDF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Must start with PDF header
	if !bytes.HasPrefix(pdf, []byte("%PDF")) {
		t.Errorf("PDF should start with %%PDF, got %q...", pdf[:4])
	}
	// And must be non-trivially sized
	if len(pdf) < 200 {
		t.Errorf("PDF size too small: %d bytes", len(pdf))
	}
}
