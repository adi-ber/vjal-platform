package main

import (
  "context"
  "encoding/json"
  "fmt"
  "html/template"
  "log"
  "net/http"

  "github.com/adi-ber/vjal-platform/pkg/config"
  "github.com/adi-ber/vjal-platform/pkg/form"
  "github.com/adi-ber/vjal-platform/pkg/license"
  "github.com/adi-ber/vjal-platform/pkg/llm"
  "github.com/adi-ber/vjal-platform/pkg/output"
  "github.com/prometheus/client_golang/prometheus/promhttp"
)

const maxRounds = 4

type pageData struct {
  Round         int
  MaxRounds     int
  QuestionsJSON template.JS
}

type dynamicRequest struct {
  Round   int               `json:"round"`
  Answers map[string]string `json:"answers"`
}

type Question struct {
  ID    string `json:"id"`
  Label string `json:"label"`
}

type dynamicResponse struct {
  Questions []Question `json:"questions,omitempty"`
  NextRound int        `json:"nextRound,omitempty"`
}

var tmpl = template.Must(template.ParseFiles("cmd/dynamic-example/templates/dynamic_form.html"))

func main() {
  cfg, err := config.Load("config.json")
  if err != nil {
    log.Fatalf("config load: %v", err)
  }

  lic, err := license.NewValidator(cfg).Validate(context.Background())
  if err != nil {
    log.Fatalf("license: %v", err)
  }

  ai, err := llm.New(cfg, lic)
  if err != nil {
    log.Fatalf("llm init: %v", err)
  }

  renderer := output.NewRenderer()

  defs, err := form.LoadDefinitionsDir("definitions")
  if err != nil {
    log.Fatalf("load definitions: %v", err)
  }
  initialFields, ok := defs["complexProcess"]
  if !ok {
    log.Fatalf("no definition for key complexProcess")
  }
  questionsJSON, err := json.Marshal(initialFields)
  if err != nil {
    log.Fatalf("marshal initial questions: %v", err)
  }

  http.Handle("/metrics", promhttp.Handler())
  http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
  })

  http.HandleFunc("/dynamic", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    data := pageData{
      Round:         1,
      MaxRounds:     maxRounds,
      QuestionsJSON: template.JS(questionsJSON),
    }
    if err := tmpl.Execute(w, data); err != nil {
      log.Printf("template exec: %v", err)
    }
  })

  http.HandleFunc("/dynamic-submit", func(w http.ResponseWriter, r *http.Request) {
    var req dynamicRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
      http.Error(w, "invalid JSON", http.StatusBadRequest)
      return
    }

    history := ""
    count := 0
    for q, ans := range req.Answers {
      count++
      history += fmt.Sprintf("%d. %s: %s\n", count, q, ans)
    }

    next := req.Round + 1
    if next <= maxRounds {
      prompt := fmt.Sprintf(
        "Round %d of %d.\n\nThe user has provided the following answers so far:\n%s\n\nPlease ask exactly one concise follow‑up question to clarify or gather missing details.",
        next, maxRounds, history,
      )
      question, err := ai.Prompt(r.Context(), prompt)
      if err != nil {
        http.Error(w, "LLM error: "+err.Error(), http.StatusInternalServerError)
        return
      }
      resp := dynamicResponse{
        Questions: []Question{{ID: fmt.Sprintf("q%d", next), Label: question}},
        NextRound: next,
      }
      json.NewEncoder(w).Encode(resp)
      return
    }

    prompt := fmt.Sprintf(
      "You have completed %d rounds. Compile a final, comprehensive report based solely on these answers:\n%s\n\nReturn only the report text.",
      maxRounds, history,
    )
    report, err := ai.Prompt(r.Context(), prompt)
    if err != nil {
      http.Error(w, "LLM error: "+err.Error(), http.StatusInternalServerError)
      return
    }
    pdfBytes, err := renderer.ToPDF(report)
    if err != nil {
      http.Error(w, "PDF error: "+err.Error(), http.StatusInternalServerError)
      return
    }
    w.Header().Set("Content-Type", "application/pdf")
    w.Header().Set("Content-Disposition", "attachment; filename=\"report.pdf\"")
    w.Write(pdfBytes)
  })

  addr := fmt.Sprintf(":%d", cfg.HTTPPort)
  log.Printf("listening on %s", addr)
  log.Fatal(http.ListenAndServe(addr, nil))
}
