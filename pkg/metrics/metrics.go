package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// License
	LicenseValidationTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "vjal", Subsystem: "license", Name: "validation_total",
		Help: "Total license validation attempts",
	})
	LicenseValidationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "vjal", Subsystem: "license", Name: "validation_errors_total",
		Help: "Total number of license validation failures",
	})

	// Config
	ConfigLoadDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "vjal", Subsystem: "config", Name: "load_duration_seconds",
		Help:    "Duration of config loading",
		Buckets: prometheus.DefBuckets,
	})
	ConfigLoadErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "vjal", Subsystem: "config", Name: "load_errors_total",
		Help: "Number of config load errors",
	})

	// Form
	FormRenderTotal    = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "vjal", Subsystem: "form", Name: "render_total",
		Help: "Total number of form render calls",
	})
	FormRenderDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "vjal", Subsystem: "form", Name: "render_duration_seconds",
		Help:    "Duration of form render calls",
		Buckets: prometheus.DefBuckets,
	})
	FormValidationTotal    = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "vjal", Subsystem: "form", Name: "validation_total",
		Help: "Total number of form validation calls",
	})
	FormValidationWarnings = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "vjal", Subsystem: "form", Name: "validation_warnings_total",
		Help: "Total number of validation warnings issued",
	})

	// Storage
	StateSaveTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "vjal", Subsystem: "storage", Name: "state_save_total",
		Help: "Total number of state save calls",
	})
	StateLoadTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "vjal", Subsystem: "storage", Name: "state_load_total",
		Help: "Total number of state load calls",
	})

	// LLM (with provider label)
	LLMRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "vjal", Subsystem: "llm", Name: "requests_total",
		Help: "Total number of LLM requests",
	}, []string{"provider"})
	LLMRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "vjal", Subsystem: "llm", Name: "request_duration_seconds",
		Help:    "Duration of LLM requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"provider"})
	LLMErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "vjal", Subsystem: "llm", Name: "errors_total",
		Help: "Number of LLM errors",
	}, []string{"provider"})

	// Output
	OutputHTMLDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "vjal", Subsystem: "output", Name: "html_duration_seconds",
		Help:    "Duration of Markdown→HTML conversion",
		Buckets: prometheus.DefBuckets,
	})
	OutputPDFTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "vjal", Subsystem: "output", Name: "pdf_total",
		Help: "Total PDF generation attempts",
	})
	OutputPDFErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "vjal", Subsystem: "output", Name: "pdf_errors_total",
		Help: "Total PDF generation failures",
	})
)
