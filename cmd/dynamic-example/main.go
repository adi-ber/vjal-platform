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
  Done      bool       `json:"done,omitempty"`
  Report    string     `json:"report,omitempty"`
}

var tmpl = template.Must(template.ParseFiles("cmd/dynamic-example/templates/dynamic_form.html"))

func main() {
  // 1) Load config
  cfg, err := config.Load("config.json")
  if err != nil {
    log.Fatalf("config load: %v", err)
  }

  // 2) Validate license
  lic, err := license.NewValidator(cfg).Validate(context.Background())
  if err != nil {
    log.Fatalf("license: %v", err)
  }

  // 3) Initialize LLM
  ai, err := llm.New(cfg, lic)
  if err != nil {
    log.Fatalf("llm init: %v", err)
  }

  // 4) Load form definitions
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

  // 5) Metrics & health
  http.Handle("/metrics", promhttp.Handler())
  http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
  })

  // 6) Serve initial page
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

  // 7) Handle iteration & final report
  http.HandleFunc("/dynamic-submit", func(w http.ResponseWriter, r *http.Request) {
    var req dynamicRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
      http.Error(w, "invalid JSON", http.StatusBadRequest)
      return
    }

    // Build history text
    history := ""
    count := 0
    for q, ans := range req.Answers {
      count++
      history += fmt.Sprintf("%d. %s: %s\n", count, q, ans)
    }

    next := req.Round + 1
    if next <= maxRounds {
      // ask follow‑up
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
        Done:      false,
      }
      json.NewEncoder(w).Encode(resp)
      return
    }

    // final report (JSON only)
    prompt := fmt.Sprintf(
      "You have completed %d rounds. Compile a final, comprehensive report based solely on these answers:\n%s\n\nReturn only the report text.",
      maxRounds, history,
    )
    report, err := ai.Prompt(r.Context(), prompt)
    if err != nil {
      http.Error(w, "LLM error: "+err.Error(), http.StatusInternalServerError)
      return
    }
    resp := dynamicResponse{
      Done:   true,
      Report: report,
    }
    json.NewEncoder(w).Encode(resp)
  })

  // 8) Start server
  addr := fmt.Sprintf(":%d", cfg.HTTPPort)
  log.Printf("listening on %s", addr)
  log.Fatal(http.ListenAndServe(addr, nil))
}
