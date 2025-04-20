// cmd/example/main.go
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
	"github.com/adi-ber/vjal-platform/pkg/form"
	"github.com/adi-ber/vjal-platform/pkg/license"
	"github.com/adi-ber/vjal-platform/pkg/llm"
	_ "github.com/adi-ber/vjal-platform/pkg/metrics"
	"github.com/adi-ber/vjal-platform/pkg/output"
	"github.com/adi-ber/vjal-platform/pkg/storage"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// processRequest is the JSON payload for our /process endpoint.
type processRequest struct {
	PromptKey string                 `json:"promptKey"`
	Data      map[string]interface{} `json:"data"`
	Format    string                 `json:"format"` // "html" or "pdf"
}

func main() {
	// 1) Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	// 1a) Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		log.Fatalf("failed to create output dir %q: %v", cfg.OutputDir, err)
	}

	// 2) Load all form definitions from definitions/
	formDefs, err := form.LoadDefinitionsDir("definitions")
	if err != nil {
		log.Fatalf("cannot load form definitions: %v", err)
	}

	// 3) Load LLM prompts
	promptBytes, err := ioutil.ReadFile("llm_prompts.enc")
	if err != nil {
		log.Fatalf("failed to read llm_prompts.enc: %v", err)
	}
	var promptTemplates map[string]string
	if err := json.Unmarshal(promptBytes, &promptTemplates); err != nil {
		log.Fatalf("invalid JSON in llm_prompts.enc: %v", err)
	}

	// 4) Validate license
	validator := license.NewValidator(cfg)
	lic, err := validator.Validate(context.Background())
	if err != nil {
		log.Fatalf("license validation failed: %v", err)
	}

	// 5) Initialize storage (for form state, unused here but required)
	if _, err := storage.New(filepath.Join(cfg.OutputDir, "state.db")); err != nil {
		log.Fatalf("storage init error: %v", err)
	}

	// 6) Initialize LLM client & renderer
	ai, err := llm.New(cfg, lic)
	if err != nil {
		log.Fatalf("LLM init error: %v", err)
	}
	renderer := output.NewRenderer()

	// 7) Parse our prompt‑form template
	promptFormTmpl := template.Must(template.ParseFiles(
		"cmd/example/templates/prompt_form.html",
	))

	// --- Serve the prompt‑driven form ---
	http.HandleFunc("/prompt-form", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("promptKey")
		fields, ok := formDefs[key]
		if !ok {
			http.NotFound(w, r)
			return
		}
		fieldsJSON, _ := json.Marshal(fields)
		data := struct {
			PromptFieldsJSON template.JS
		}{
			PromptFieldsJSON: template.JS(fieldsJSON),
		}
		w.Header().Set("Content-Type", "text/html")
		if err := promptFormTmpl.Execute(w, data); err != nil {
			log.Printf("[prompt-form] template exec error: %v", err)
		}
	})

	// --- Process form → prompt → LLM → HTML or PDF ---
	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		var req processRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// 1) Lookup prompt template
		tpl, ok := promptTemplates[req.PromptKey]
		if !ok {
			http.Error(w, "unknown promptKey", http.StatusBadRequest)
			return
		}

		// 2) Render the prompt by merging in user data
		var buf bytes.Buffer
		t := template.Must(template.New("p").Parse(tpl))
		if err := t.Execute(&buf, req.Data); err != nil {
			http.Error(w, fmt.Sprintf("prompt render error: %v", err), http.StatusInternalServerError)
			return
		}

		// 3) Call the LLM
		aiResp, err := ai.Prompt(context.Background(), buf.String())
		if err != nil {
			http.Error(w, fmt.Sprintf("LLM error: %v", err), http.StatusInternalServerError)
			return
		}

		// 4) Return in requested format
		switch req.Format {
		case "html":
			out, err := renderer.ToHTML(aiResp)
			if err != nil {
				http.Error(w, fmt.Sprintf("HTML render error: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(out))

		default: // pdf
			pdf, err := renderer.ToPDF(aiResp)
			if err != nil {
				http.Error(w, fmt.Sprintf("PDF error: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", "attachment; filename=\"result.pdf\"")
			w.Write(pdf)
		}
	})

	// --- Metrics & health ---
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// --- Start server ---
	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	log.Printf("starting example server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
