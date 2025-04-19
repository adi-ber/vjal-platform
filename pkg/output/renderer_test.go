package output

import (
	"strings"
	"testing"
)

func TestToHTML_Basic(t *testing.T) {
	r := NewRenderer()
	html, err := r.ToHTML("**bold**")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "<strong>bold</strong>") {
		t.Errorf("expected HTML to contain <strong>bold</strong>, got %q", html)
	}
}

func TestToPDF_Stub(t *testing.T) {
	r := NewRenderer()
	_, err := r.ToPDF("anything")
	if err == nil {
		t.Fatal("expected error from ToPDF stub, got nil")
	}
	if err.Error() != "PDF generation not implemented" {
		t.Errorf("unexpected error message: %v", err)
	}
}