package proxy

import (
	_ "embed"
	"fmt"
	"net/http"
)

//go:embed dashboard.html
var dashboardHTML string

func (s *GatewayServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, dashboardHTML)
}
