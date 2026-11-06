package proxy

import (
	"fmt"
	"net/http"
)

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>QuotaDock — AI Quota & Failover Dashboard</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #0D0F12;
            --surface: #161A20;
            --surface-border: #262D38;
            --text-primary: #F3F5F7;
            --text-secondary: #8B949E;
            --accent: #388BFD;
            --success: #238636;
            --success-glow: rgba(35, 134, 54, 0.4);
            --warning: #D29922;
            --critical: #DA3633;
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            background-color: var(--bg);
            color: var(--text-primary);
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            padding: 32px 24px;
            display: flex;
            justify-content: center;
        }

        .container {
            max-width: 960px;
            width: 100%;
        }

        header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 28px;
            padding-bottom: 20px;
            border-bottom: 1px solid var(--surface-border);
        }

        .brand {
            display: flex;
            align-items: center;
            gap: 12px;
        }

        .brand-icon {
            font-size: 26px;
        }

        .brand-title {
            font-size: 22px;
            font-weight: 700;
            letter-spacing: -0.5px;
        }

        .pill {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            padding: 6px 14px;
            background: #11261B;
            border: 1px solid #1E4E30;
            border-radius: 9999px;
            font-size: 13px;
            font-weight: 500;
            color: #3FB950;
        }

        .pill-dot {
            width: 8px;
            height: 8px;
            background: #3FB950;
            border-radius: 50%;
            box-shadow: 0 0 10px var(--success-glow);
        }

        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 16px;
            margin-bottom: 28px;
        }

        .card {
            background: var(--surface);
            border: 1px solid var(--surface-border);
            border-radius: 12px;
            padding: 20px;
            transition: border-color 0.2s ease;
        }

        .card:hover {
            border-color: #3B4452;
        }

        .card-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 14px;
        }

        .provider-name {
            font-size: 15px;
            font-weight: 600;
        }

        .quota-percent {
            font-family: 'JetBrains Mono', monospace;
            font-weight: 600;
            font-size: 14px;
            color: #3FB950;
        }

        .progress-bar-bg {
            background: #21262D;
            height: 8px;
            border-radius: 9999px;
            overflow: hidden;
            margin-bottom: 12px;
        }

        .progress-bar-fill {
            height: 100%;
            background: linear-gradient(90deg, #2EA043, #3FB950);
            border-radius: 9999px;
            transition: width 0.4s ease;
        }

        .quota-meta {
            display: flex;
            justify-content: space-between;
            font-size: 12px;
            color: var(--text-secondary);
            font-family: 'JetBrains Mono', monospace;
        }

        .failover-section {
            background: var(--surface);
            border: 1px solid var(--surface-border);
            border-radius: 12px;
            padding: 24px;
        }

        .failover-title {
            font-size: 16px;
            font-weight: 600;
            margin-bottom: 16px;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .chain-list {
            display: flex;
            flex-direction: column;
            gap: 10px;
        }

        .chain-item {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 12px 16px;
            background: #111419;
            border: 1px solid var(--surface-border);
            border-radius: 8px;
            font-size: 13px;
        }

        .chain-rank {
            font-weight: 600;
            color: var(--accent);
            margin-right: 12px;
        }

        .chain-status {
            font-size: 12px;
            padding: 2px 8px;
            border-radius: 4px;
            background: #1E2838;
            color: #58A6FF;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="brand">
                <span class="brand-icon">⚡</span>
                <span class="brand-title">QuotaDock Gateway</span>
            </div>
            <div class="pill">
                <div class="pill-dot"></div>
                127.0.0.1:8080 Active
            </div>
        </header>

        <main>
            <div class="stats-grid" id="statsGrid">
                <!-- Dynamically populated via telemetry -->
            </div>

            <section class="failover-section">
                <div class="failover-title">
                    <span>🛡️ Active Failover Cascade</span>
                </div>
                <div class="chain-list" id="chainList">
                    <!-- Dynamically populated -->
                </div>
            </section>
        </main>
    </div>

    <script>
        async function fetchTelemetry() {
            try {
                const res = await fetch('/quotadock/telemetry');
                if (!res.ok) return;
                const data = await res.json();
                renderQuotas(data.quotas || {});
                renderChain(data.backends || []);
            } catch (err) {
                console.error("Telemetry fetch failed:", err);
            }
        }

        function renderQuotas(quotas) {
            const container = document.getElementById('statsGrid');
            if (!Object.keys(quotas).length) {
                // Default placeholders
                quotas = {
                    "anthropic": { provider: "Anthropic Claude 3.5", remaining_tokens: 88000, limit_tokens: 100000, status: "healthy" },
                    "openai": { provider: "OpenAI GPT-4o", remaining_tokens: 74000, limit_tokens: 100000, status: "healthy" },
                    "gemini": { provider: "Google Gemini 1.5 Pro", remaining_tokens: 950000, limit_tokens: 1000000, status: "healthy" },
                    "ollama": { provider: "Ollama Local (Offline Fallback)", remaining_tokens: 1, limit_tokens: 1, status: "healthy" }
                };
            }

            container.innerHTML = Object.entries(quotas).map(([k, q]) => {
                const rem = q.remaining_tokens || 0;
                const lim = q.limit_tokens || 1;
                const pct = Math.min(100, Math.round((rem / lim) * 100));
                const provName = q.provider || k;
                return `
                    <div class="card">
                        <div class="card-header">
                            <span class="provider-name">${provName}</span>
                            <span class="quota-percent">${pct}%</span>
                        </div>
                        <div class="progress-bar-bg">
                            <div class="progress-bar-fill" style="width: ${pct}%"></div>
                        </div>
                        <div class="quota-meta">
                            <span>${rem.toLocaleString()} tokens left</span>
                            <span>Limit: ${lim.toLocaleString()}</span>
                        </div>
                    </div>
                `;
            }).join('');
        }

        function renderChain(backends) {
            const list = document.getElementById('chainList');
            if (!backends.length) {
                backends = [
                    { provider: "Anthropic Claude", default_model: "claude-3-5-sonnet-20241022", enabled: true },
                    { provider: "OpenAI", default_model: "gpt-4o", enabled: true },
                    { provider: "Google Gemini", default_model: "gemini-1.5-pro", enabled: true },
                    { provider: "Local Ollama", default_model: "llama3.1:8b", enabled: true }
                ];
            }

            list.innerHTML = backends.map((b, idx) => `
                <div class="chain-item">
                    <div>
                        <span class="chain-rank">#${idx + 1}</span>
                        <strong>${b.provider}</strong>
                        <span style="color: var(--text-secondary); margin-left: 8px;">(${b.default_model})</span>
                    </div>
                    <span class="chain-status">${idx === 0 ? 'Primary' : 'Failover Ready'}</span>
                </div>
            `).join('');
        }

        fetchTelemetry();
        setInterval(fetchTelemetry, 2500);
    </script>
</body>
</html>`

func (s *GatewayServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, dashboardHTML)
}
