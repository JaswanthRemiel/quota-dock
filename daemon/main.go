package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/JaswanthRemiel/quota-dock/daemon/failover"
	"github.com/JaswanthRemiel/quota-dock/daemon/proxy"
	"github.com/JaswanthRemiel/quota-dock/daemon/tracker"
)

func main() {
	port := flag.Int("port", 8080, "Port on which the QuotaDock daemon listens")
	flag.Parse()

	log.Println("⚡ Starting QuotaDock Local Gateway Daemon...")

	quotaTracker := tracker.NewQuotaTracker()
	failoverRouter := failover.NewRouter(quotaTracker)
	gateway := proxy.NewGatewayServer(quotaTracker, failoverRouter)

	mux := http.NewServeMux()
	gateway.RegisterRoutes(mux)

	serverAddr := fmt.Sprintf("127.0.0.1:%d", *port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}

	go func() {
		log.Printf("🚀 QuotaDock Daemon listening on http://%s", serverAddr)
		log.Printf("👉 Point your OpenAI clients to http://%s/v1", serverAddr)
		log.Printf("📊 Live Telemetry API active on http://%s/quotadock/telemetry", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Fatal server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down QuotaDock Daemon...")
}
