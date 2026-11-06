package tracker

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ProviderQuota holds telemetry metrics for an AI provider.
type ProviderQuota struct {
	Provider          string    `json:"provider"`
	LimitRequests     int64     `json:"limit_requests"`
	RemainingRequests int64     `json:"remaining_requests"`
	LimitTokens       int64     `json:"limit_tokens"`
	RemainingTokens   int64     `json:"remaining_tokens"`
	ResetTokens       time.Time `json:"reset_tokens"`
	ResetRequests     time.Time `json:"reset_requests"`
	TokensPerSecond   float64   `json:"tokens_per_second"`
	EstimatedSeconds  float64   `json:"estimated_seconds_remaining"`
	Status            string    `json:"status"` // "healthy", "warning", "critical", "exhausted"
	LastUpdated       time.Time `json:"last_updated"`
}

// QuotaTracker maintains thread-safe live quotas across providers.
type QuotaTracker struct {
	mu     sync.RWMutex
	quotas map[string]*ProviderQuota
}

// NewQuotaTracker initializes a tracker with default supported providers.
func NewQuotaTracker() *QuotaTracker {
	t := &QuotaTracker{
		quotas: make(map[string]*ProviderQuota),
	}
	// Seed default providers
	defaults := []string{"anthropic", "openai", "gemini", "ollama"}
	for _, p := range defaults {
		t.quotas[p] = &ProviderQuota{
			Provider:          p,
			LimitRequests:     1000,
			RemainingRequests: 1000,
			LimitTokens:       1000000,
			RemainingTokens:   1000000,
			Status:            "healthy",
			LastUpdated:       time.Now(),
		}
	}
	return t
}

// SniffResponse parses rate limit headers from an HTTP response and updates the provider telemetry.
func (t *QuotaTracker) SniffResponse(provider string, header http.Header, statusCode int) *ProviderQuota {
	t.mu.Lock()
	defer t.mu.Unlock()

	pq, exists := t.quotas[provider]
	if !exists {
		pq = &ProviderQuota{
			Provider:    provider,
			Status:      "healthy",
			LastUpdated: time.Now(),
		}
		t.quotas[provider] = pq
	}

	previousRemaining := pq.RemainingTokens
	previousTime := pq.LastUpdated
	now := time.Now()

	// 1. OpenAI style headers
	if v := header.Get("x-ratelimit-remaining-tokens"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			pq.RemainingTokens = val
		}
	}
	if v := header.Get("x-ratelimit-limit-tokens"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			pq.LimitTokens = val
		}
	}
	if v := header.Get("x-ratelimit-remaining-requests"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			pq.RemainingRequests = val
		}
	}
	if v := header.Get("x-ratelimit-limit-requests"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			pq.LimitRequests = val
		}
	}

	// 2. Anthropic style headers
	if v := header.Get("anthropic-ratelimit-tokens-remaining"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			pq.RemainingTokens = val
		}
	}
	if v := header.Get("anthropic-ratelimit-tokens-limit"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			pq.LimitTokens = val
		}
	}

	// 3. Status determination & 429 handler
	if statusCode == http.StatusTooManyRequests {
		pq.Status = "exhausted"
		pq.RemainingTokens = 0
		pq.RemainingRequests = 0
	} else if pq.LimitTokens > 0 {
		pct := float64(pq.RemainingTokens) / float64(pq.LimitTokens)
		if pct <= 0.05 {
			pq.Status = "critical"
		} else if pct <= 0.35 {
			pq.Status = "warning"
		} else {
			pq.Status = "healthy"
		}
	}

	// 4. Velocity computation (Tokens per second)
	elapsed := now.Sub(previousTime).Seconds()
	if elapsed > 0.1 && previousRemaining > pq.RemainingTokens {
		consumed := previousRemaining - pq.RemainingTokens
		velocity := float64(consumed) / elapsed
		// Exponential moving average (alpha = 0.3)
		if pq.TokensPerSecond == 0 {
			pq.TokensPerSecond = velocity
		} else {
			pq.TokensPerSecond = 0.3*velocity + 0.7*pq.TokensPerSecond
		}

		if pq.TokensPerSecond > 0 && pq.RemainingTokens > 0 {
			pq.EstimatedSeconds = float64(pq.RemainingTokens) / pq.TokensPerSecond
		}
	}

	pq.LastUpdated = now
	return pq
}

// GetAll returns a snapshot of all provider quotas.
func (t *QuotaTracker) GetAll() []ProviderQuota {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]ProviderQuota, 0, len(t.quotas))
	for _, q := range t.quotas {
		result = append(result, *q)
	}
	return result
}

// GetProvider returns a copy of a single provider quota.
func (t *QuotaTracker) GetProvider(provider string) (ProviderQuota, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	q, exists := t.quotas[strings.ToLower(provider)]
	if !exists {
		return ProviderQuota{}, false
	}
	return *q, true
}
