package main

import (
	"net/http"
	"testing"

	"github.com/JaswanthRemiel/quota-dock/daemon/failover"
	"github.com/JaswanthRemiel/quota-dock/daemon/tracker"
)

func TestHeaderSniffingOpenAI(t *testing.T) {
	tr := tracker.NewQuotaTracker()

	header := make(http.Header)
	header.Set("x-ratelimit-remaining-tokens", "450000")
	header.Set("x-ratelimit-limit-tokens", "1000000")
	header.Set("x-ratelimit-remaining-requests", "480")
	header.Set("x-ratelimit-limit-requests", "500")

	quota := tr.SniffResponse("openai", header, http.StatusOK)
	if quota.RemainingTokens != 450000 {
		t.Fatalf("Expected 450000 remaining tokens, got %d", quota.RemainingTokens)
	}
	if quota.Status != "healthy" {
		t.Fatalf("Expected status healthy, got %s", quota.Status)
	}
}

func TestHeaderSniffingAnthropic(t *testing.T) {
	tr := tracker.NewQuotaTracker()

	header := make(http.Header)
	header.Set("anthropic-ratelimit-tokens-remaining", "25000")
	header.Set("anthropic-ratelimit-tokens-limit", "1000000")

	quota := tr.SniffResponse("anthropic", header, http.StatusOK)
	if quota.RemainingTokens != 25000 {
		t.Fatalf("Expected 25000 remaining tokens, got %d", quota.RemainingTokens)
	}
	// 25000 / 1000000 = 2.5%, so critical
	if quota.Status != "critical" {
		t.Fatalf("Expected critical status, got %s", quota.Status)
	}
}

func TestFailoverOn429(t *testing.T) {
	tr := tracker.NewQuotaTracker()
	r := failover.NewRouter(tr)

	// Simulate rate limit 429 on Anthropic
	tr.SniffResponse("anthropic", http.Header{}, http.StatusTooManyRequests)

	q, _ := tr.GetProvider("anthropic")
	if q.Status != "exhausted" {
		t.Fatalf("Expected anthropic to be exhausted, got %s", q.Status)
	}

	// Requesting anthropic should now dynamically failover to openai
	selected, err := r.SelectBackend("anthropic")
	if err != nil {
		t.Fatalf("Failover router returned error: %v", err)
	}
	if selected.Provider != "openai" {
		t.Fatalf("Expected failover to openai, got %s", selected.Provider)
	}
}
