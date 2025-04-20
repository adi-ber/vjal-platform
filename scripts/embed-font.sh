#!/usr/bin/env bash
set -euo pipefail

# 1) Copy the font into your package
mkdir -p pkg/output/fonts
cp assets/fonts/DejaVuSans.ttf pkg/output/fonts/DejaVuSans.ttf

# 2) Patch pkg/output/renderer.go
patch -p1 << 'EOF'
*** Begin Patch
*** Update File: pkg/output/renderer.go
@@
-import (
+import (
+   "embed"
    "bytes"
    "fmt"

    "github.com/adi-ber/vjal-platform/pkg/metrics"
@@
 // Renderer handles converting Markdown to various formats.
 type Renderer struct {
@@
 // NewRenderer creates a new Renderer.
@@
 // ToPDF converts a Markdown string into a simple PDF.
 // It writes the raw markdown as text into the PDF.
 func (r *Renderer) ToPDF(input string) ([]byte, error) {
     metrics.OutputPDFTotal.Inc()

-    pdf := gofpdf.New("P", "mm", "A4", "")
-    pdf.AddPage()
-    pdf.SetFont("Arial", "", 12)
+    // embed DejaVuSans for full UTF-8
+    //go:embed fonts/DejaVuSans.ttf
+    var dejavuTTF []byte
+    pdf := gofpdf.New("P", "mm", "A4", "")
+    pdf.AddUTF8FontFromBytes("dejavu", "", dejavuTTF)
+    pdf.SetFont("dejavu", "", 12)
+    pdf.AddPage()

     // Write the markdown text into the PDF
     pdf.MultiCell(0, 6, input, "", "", false)

     var buf bytes.Buffer
*** End Patch
EOF

# 3) Rebuild your demo binary
go build -o demo-embed cmd/process-example/main.go

echo "✅ Embedded DejaVuSans and updated PDF renderer."
echo "Now run: export VJAL_HTTP_PORT=9090 && ./demo-embed"