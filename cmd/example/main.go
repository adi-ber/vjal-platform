// cmd/example/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/adi-ber/vjal-platform/pkg/config"
	"github.com/adi-ber/vjal-platform/pkg/form"
	"github.com/adi-ber/vjal-platform/pkg/license"
	"github.com/adi-ber/vjal-platform/pkg/storage"
	"github.com/adi-ber/vjal-platform/pkg/llm"
	"github.com/adi-ber/vjal-platform/pkg/output"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 1. Load config
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Validate license
	validator := license.NewValidator(*cfg)
	lic, err := validator.Validate(context.Background())
	if err != nil {
		log.Fatalf("license validation failed: %v", err)
	}

	// 3. Init storage for form state
	stateDB := filepath.Join(cfg.OutputDir, "form_state.db")
	store, err := storage.New(stateDB)
	if err != nil {
		log.Fatalf("failed to init storage: %v", err)
	}

	// 4. Load form schema & attach storage
	frm, err := form.New(cfg.FormSchema, store, "default")
	if err != nil {
		log.Fatalf("failed to load form schema: %v", err)
	}

	// 5. Init LLM
	ai, err := llm.New(*cfg, lic)
	if err != nil {
		log.Fatalf("failed to init LLM: %v", err)
	}

	// 6. Init renderer
	renderer := output.NewRenderer()

	// 7. HTTP Handlers

	// GET /form?page=<pageID>
	http.HandleFunc("/form", func(w http.ResponseWriter, r *http.Request) {
		pageID := r.URL.Query().Get("page")
		if pageID == "" {
			pageID = "start"
		}

		// Load saved state
		state, err := frm.LoadState(context.Background(), pageID)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to load state: %v", err), 500)
			return
		}

		// Render page (stub)
		html, err := frm.RenderPage(context.Background(), pageID)
		if err != nil {
			http.Error(w, fmt.Sprintf("render error: %v", err), 500)
			return
		}

		// Show saved values and HTML
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<h3>Saved Data: %v</h3>%s", state, html)
	})

	// POST /submit?page=<pageID>  body: JSON map[string]interface{}
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		pageID := r.URL.Query().Get("page")
		if pageID == "" {
			pageID = "start"
		}

		var input map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON body", 400)
			return
		}

		// Validate page input
		warnings, err := frm.Validate(context.Background(), pageID, input)
		if err != nil {
			http.Error(w, fmt.Sprintf("validation error: %v", err), 500)
			return
		}

		// Persist page state
		if err := frm.SaveState(context.Background(), pageID, input); err != nil {
			http.Error(w, fmt.Sprintf("save state error: %v", err), 500)
			return
		}

		// Compute next page
		nextPage, err := frm.NextPage(pageID)
		if err != nil {
			http.Error(w, fmt.Sprintf("next page error: %v", err), 500)
			return
		}

		// Respond with warnings and next page
		resp := map[string]interface{}{
			"warnings": warnings,
			"nextPage": nextPage,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /run?prompt=<text>
	http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		prompt := r.URL.Query().Get("prompt")
		if prompt == "" {
			prompt = "Hello, world!"
		}
		result, err := ai.Prompt(context.Background(), prompt)
		if err != nil {
			http.Error(w, fmt.Sprintf("LLM error: %v", err), 500)
			return
		}
		html, err := renderer.ToHTML(result)
		if err != nil {
			http.Error(w, fmt.Sprintf("render error: %v", err), 500)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	// Metrics & health
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// Start server
	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	log.Printf("starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}