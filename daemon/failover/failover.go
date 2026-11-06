package failover

import (
	"fmt"
	"strings"
	"sync"

	"github.com/JaswanthRemiel/quota-dock/daemon/tracker"
)

// TargetBackend represents an upstream endpoint configuration.
type TargetBackend struct {
	Provider    string `json:"provider"`
	BaseURL     string `json:"base_url"`
	DefaultModel string `json:"default_model"`
	APIKey      string `json:"api_key"`
	Enabled     bool   `json:"enabled"`
}

// Router determines which backend should handle an incoming request.
type Router struct {
	mu            sync.RWMutex
	chain         []TargetBackend
	tracker       *tracker.QuotaTracker
	modelMappings map[string]string
}

// NewRouter initializes the failover router with default backends.
func NewRouter(t *tracker.QuotaTracker) *Router {
	r := &Router{
		tracker: t,
		chain: []TargetBackend{
			{
				Provider:     "anthropic",
				BaseURL:      "https://api.anthropic.com/v1",
				DefaultModel: "claude-3-5-sonnet-20241022",
				Enabled:      true,
			},
			{
				Provider:     "openai",
				BaseURL:      "https://api.openai.com/v1",
				DefaultModel: "gpt-4o",
				Enabled:      true,
			},
			{
				Provider:     "gemini",
				BaseURL:      "https://generativelanguage.googleapis.com/v1beta/openai",
				DefaultModel: "gemini-1.5-pro",
				Enabled:      true,
			},
			{
				Provider:     "ollama",
				BaseURL:      "http://localhost:11434/v1",
				DefaultModel: "llama3.1:latest",
				Enabled:      true,
			},
		},
		modelMappings: map[string]string{
			"claude-3-5-sonnet": "gpt-4o",
			"gpt-4o":            "gemini-1.5-pro",
			"gemini-1.5-pro":    "llama3.1:latest",
		},
	}
	return r
}

// SelectBackend returns the best available backend based on quota health.
func (r *Router) SelectBackend(requestedProvider string) (TargetBackend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// If a specific provider was requested and is healthy, try it first
	var preferred TargetBackend
	var found bool
	for _, b := range r.chain {
		if strings.EqualFold(b.Provider, requestedProvider) && b.Enabled {
			preferred = b
			found = true
			break
		}
	}

	if found {
		if q, exists := r.tracker.GetProvider(preferred.Provider); exists {
			if q.Status != "exhausted" {
				return preferred, nil
			}
		} else {
			return preferred, nil
		}
	}

	// Fallback through chain for the first healthy backend
	for _, b := range r.chain {
		if !b.Enabled {
			continue
		}
		if q, exists := r.tracker.GetProvider(b.Provider); exists {
			if q.Status != "exhausted" {
				return b, nil
			}
		} else {
			return b, nil
		}
	}

	// If all are exhausted, return the local Ollama backend or default
	if len(r.chain) > 0 {
		return r.chain[len(r.chain)-1], nil
	}

	return TargetBackend{}, fmt.Errorf("no backends configured")
}

// GetChain returns a copy of current routing priority.
func (r *Router) GetChain() []TargetBackend {
	r.mu.RLock()
	defer r.mu.RUnlock()
	copied := make([]TargetBackend, len(r.chain))
	copy(copied, r.chain)
	return copied
}

// SetChain updates the failover priority chain.
func (r *Router) SetChain(chain []TargetBackend) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.chain = chain
}
