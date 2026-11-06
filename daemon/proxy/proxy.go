package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/JaswanthRemiel/quota-dock/daemon/failover"
	"github.com/JaswanthRemiel/quota-dock/daemon/tracker"
)

// GatewayServer handles HTTP proxying and telemetry dispatch.
type GatewayServer struct {
	Tracker *tracker.QuotaTracker
	Router  *failover.Router
	Client  *http.Client
}

// NewGatewayServer creates a new GatewayServer instance.
func NewGatewayServer(t *tracker.QuotaTracker, r *failover.Router) *GatewayServer {
	return &GatewayServer{
		Tracker: t,
		Router:  r,
		Client:  &http.Client{},
	}
}

// RegisterRoutes attaches endpoints to the multiplexer.
func (s *GatewayServer) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/v1/chat/completions", s.handleChatCompletions)
	mux.HandleFunc("/quotadock/telemetry", s.handleTelemetry)
	mux.HandleFunc("/quotadock/health", s.handleHealth)
	mux.HandleFunc("/dashboard", s.handleDashboard)
	mux.HandleFunc("/", s.handleDashboard)
}

func (s *GatewayServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "QuotaDock Daemon"})
}

func (s *GatewayServer) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	quotas := s.Tracker.GetAll()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"quotas":   quotas,
		"backends": s.Router.GetChain(),
	})
}

func (s *GatewayServer) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse requested model to infer provider
	var payload struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	_ = json.Unmarshal(bodyBytes, &payload)

	inferredProvider := "openai"
	if strings.Contains(payload.Model, "claude") {
		inferredProvider = "anthropic"
	} else if strings.Contains(payload.Model, "gemini") {
		inferredProvider = "gemini"
	} else if strings.Contains(payload.Model, "llama") || strings.Contains(payload.Model, "mistral") {
		inferredProvider = "ollama"
	}

	backend, err := s.Router.SelectBackend(inferredProvider)
	if err != nil {
		http.Error(w, "No healthy AI backends available", http.StatusServiceUnavailable)
		return
	}

	// Attempt upstream call with failover retry on 429
	resp, providerUsed, err := s.forwardWithRetry(backend, r, bodyBytes)
	if err != nil {
		log.Printf("[Proxy Error] All failover attempts exhausted: %v", err)
		http.Error(w, "Upstream AI gateway failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy headers to client
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("X-QuotaDock-Provider-Used", providerUsed)
	w.WriteHeader(resp.StatusCode)

	// Stream or copy response body
	_, _ = io.Copy(w, resp.Body)
}

func (s *GatewayServer) forwardWithRetry(initialBackend failover.TargetBackend, originalReq *http.Request, body []byte) (*http.Response, string, error) {
	chain := s.Router.GetChain()
	currentBackend := initialBackend

	for attempt := 0; attempt < len(chain); attempt++ {
		targetURL, err := url.Parse(currentBackend.BaseURL + "/chat/completions")
		if err != nil {
			return nil, "", err
		}

		proxyReq, err := http.NewRequestWithContext(originalReq.Context(), http.MethodPost, targetURL.String(), bytes.NewReader(body))
		if err != nil {
			return nil, "", err
		}

		// Forward headers
		for k, vv := range originalReq.Header {
			for _, v := range vv {
				proxyReq.Header.Add(k, v)
			}
		}

		if currentBackend.APIKey != "" {
			proxyReq.Header.Set("Authorization", "Bearer "+currentBackend.APIKey)
		}

		resp, err := s.Client.Do(proxyReq)
		if err == nil {
			// Sniff headers and track rate limits
			s.Tracker.SniffResponse(currentBackend.Provider, resp.Header, resp.StatusCode)

			// If rate limited, trigger failover to next backend
			if resp.StatusCode == http.StatusTooManyRequests {
				resp.Body.Close()
				log.Printf("[Failover] Provider %s rate-limited (429). Attempting fallback...", currentBackend.Provider)
				next, selErr := s.Router.SelectBackend("")
				if selErr == nil && next.Provider != currentBackend.Provider {
					currentBackend = next
					continue
				}
			} else {
				return resp, currentBackend.Provider, nil
			}
		}

		// Select next backend
		next, selErr := s.Router.SelectBackend("")
		if selErr != nil {
			break
		}
		currentBackend = next
	}

	return nil, "", http.ErrHandlerTimeout
}
