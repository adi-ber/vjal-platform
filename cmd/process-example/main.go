package main

import (
  "bytes"
  "context"
  "encoding/json"
  "fmt"
  "html/template"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "path/filepath"

  "github.com/adi-ber/vjal-platform/pkg/config"
  "github.com/adi-ber/vjal-platform/pkg/license"
  "github.com/adi-ber/vjal-platform/pkg/llm"
  _ "github.com/adi-ber/vjal-platform/pkg/metrics"
  "github.com/adi-ber/vjal-platform/pkg/output"
  "github.com/adi-ber/vjal-platform/pkg/storage"
  "github.com/prometheus/client_golang/prometheus/promhttp"
)

// promptTemplates holds your IP‑protected prompt strings.
var promptTemplates = map[string]string{
  "accountingClassifier": `You are a financial classifier.
Description: {{.description}}
Amount: {{.amount}}
Answer:`,
  "userSummary": `Form submission summary:
Name: {{.name}}
Age: {{.age}}
Answer:`,
}

// PromptField describes one variable the prompt expects.
type PromptField struct {
  ID    string `json:"id"`
  Label string `json:"label"`
  Type  string `json:"type"`
}

// promptFields defines, for each promptKey, the fields to render.
var promptFields = map[string][]PromptField{
  "accountingClassifier": {
    {ID: "description", Label: "Description", Type: "text"},
    {ID: "amount",       Label: "Amount",      Type: "number"},
  },
  "userSummary": {
    {ID: "name", Label: "Name", Type: "text"},
    {ID: "age",  Label: "Age",  Type: "number"},
  },
}

// processRequest defines the JSON payload for /process.
type processRequest struct {
  PromptKey string                 `json:"promptKey"`
  Data      map[string]interface{} `json:"data"`
  Format    string                 `json:"format"`
}

// formTmpl parses the HTML template for the generic schema‑driven form.
var formTmpl = template.Must(template.ParseFiles(
  "cmd/process-example/templates/form_demo.html",
))

// promptFormTmpl parses the HTML template for prompt‑driven forms.
var promptFormTmpl = template.Must(template.ParseFiles(
  "cmd/process-example/templates/prompt_form.html",
))

func main() {
  // 1) Load configuration
  cfg, err := config.Load("config.json")
  if err != nil {
    log.Fatalf("config load error: %v", err)
  }

  // 2) Ensure output directory exists
  if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
    log.Fatalf("failed to create output dir: %v", err)
  }

  // 3) Validate license
  validator := license.NewValidator(cfg)
  lic, err := validator.Validate(context.Background())
  if err != nil {
    log.Fatalf("license validation failed: %v", err)
  }

  // 4) Initialize storage (unused here but required)
  if _, err := storage.New(filepath.Join(cfg.OutputDir, "state.db")); err != nil {
    log.Fatalf("storage init error: %v", err)
  }

  // 5) Initialize LLM and renderer
  ai, err := llm.New(cfg, lic)
  if err != nil {
    log.Fatalf("LLM init error: %v", err)
  }
  renderer := output.NewRenderer()

  // 6) Serve the generic schema‑driven form
  http.HandleFunc("/form-demo", func(w http.ResponseWriter, r *http.Request) {
    schemaBytes, err := ioutil.ReadFile("forms/schema_v1.json")
    if err != nil {
      http.Error(w, "failed to read schema", http.StatusInternalServerError)
      return
    }
    data := struct {
      SchemaJSON template.JS
      PromptKeys []string
    }{
      SchemaJSON: template.JS(schemaBytes),
      PromptKeys: []string{"accountingClassifier", "userSummary"},
    }
    w.Header().Set("Content-Type", "text/html")
    if err := formTmpl.Execute(w, data); err != nil {
      log.Printf("[form-demo] template error: %v", err)
    }
  })

  // 6b) Serve a prompt‑driven form at /prompt-form
  http.HandleFunc("/prompt-form", func(w http.ResponseWriter, r *http.Request) {
    fieldsJSON, err := json.Marshal(promptFields)
    if err != nil {
      http.Error(w, "failed to encode fields", http.StatusInternalServerError)
      return
    }
    data := struct {
      PromptFieldsJSON template.JS
    }{
      PromptFieldsJSON: template.JS(fieldsJSON),
    }
    w.Header().Set("Content-Type", "text/html")
    if err := promptFormTmpl.Execute(w, data); err != nil {
      log.Printf("[prompt-form] template error: %v", err)
    }
  })

  // 7) Process endpoint with detailed logging
  http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
    var req processRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
      log.Printf("[process] JSON decode error: %v", err)
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
    }

    tmplStr, ok := promptTemplates[req.PromptKey]
    if !ok {
      err := fmt.Errorf("unknown promptKey %q", req.PromptKey)
      log.Printf("[process] %v", err)
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
    }

    var buf bytes.Buffer
    tmpl := template.Must(template.New("p").Parse(tmplStr))
    if err := tmpl.Execute(&buf, req.Data); err != nil {
      log.Printf("[process] template exec error: %v", err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    aiResp, err := ai.Prompt(context.Background(), buf.String())
    if err != nil {
      log.Printf("[process] LLM error: %v", err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    switch req.Format {
    case "md":
      w.Header().Set("Content-Type", "text/markdown")
      w.Write([]byte(aiResp))

    case "html":
      htmlOut, err := renderer.ToHTML(aiResp)
      if err != nil {
        log.Printf("[process] HTML render error: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
      w.Header().Set("Content-Type", "text/html")
      w.Write([]byte(htmlOut))

    default: // PDF
      pdfBytes, err := renderer.ToPDF(aiResp)
      if err != nil {
        log.Printf("[process] PDF render error: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
      w.Header().Set("Content-Type", "application/pdf")
      w.Header().Set("Content-Disposition", "attachment; filename=\"result.pdf\"")
      w.Write(pdfBytes)
    }
  })

  // 8) Metrics & health endpoints
  http.Handle("/metrics", promhttp.Handler())
  http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "OK")
  })

  // 9) Start the HTTP server
  addr := fmt.Sprintf(":%d", cfg.HTTPPort)
  log.Printf("listening on %s", addr)
  log.Fatal(http.ListenAndServe(addr, nil))
}
