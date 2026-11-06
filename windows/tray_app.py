"""
QuotaDock Windows Lightweight System Tray Companion
Provides quick access to QuotaDock status, telemetry, and web dashboard from the Windows notification area.
"""

import sys
import webbrowser
import urllib.request
import json
import time

DAEMON_URL = "http://127.0.0.1:8080"
DASHBOARD_URL = "http://127.0.0.1:8080/dashboard"
TELEMETRY_URL = "http://127.0.0.1:8080/quotadock/telemetry"

def check_daemon_status():
    try:
        req = urllib.request.Request(f"{DAEMON_URL}/quotadock/health", headers={"User-Agent": "QuotaDock-Win"})
        with urllib.request.urlopen(req, timeout=2) as resp:
            if resp.status == 200:
                return True
    except Exception:
        pass
    return False

def open_dashboard():
    print(f"Opening QuotaDock Dashboard at {DASHBOARD_URL}")
    webbrowser.open(DASHBOARD_URL)

def print_telemetry():
    try:
        req = urllib.request.Request(TELEMETRY_URL, headers={"User-Agent": "QuotaDock-Win"})
        with urllib.request.urlopen(req, timeout=2) as resp:
            data = json.loads(resp.read().decode())
            print("\n" + "=" * 50)
            print("⚡ QuotaDock Telemetry Snapshot")
            print("=" * 50)
            quotas = data.get("quotas", {})
            for provider, metrics in quotas.items():
                rem_tok = metrics.get("remaining_tokens", 0)
                lim_tok = metrics.get("limit_tokens", 1)
                status = metrics.get("status", "healthy")
                pct = (rem_tok / lim_tok * 100) if lim_tok > 0 else 100
                print(f"[{provider.upper():<10}] Status: {status:<10} | Tokens: {rem_tok:,} / {lim_tok:,} ({pct:.1f}%)")
            print("=" * 50 + "\n")
    except Exception as e:
        print(f"Could not reach QuotaDock daemon: {e}")

if __name__ == "__main__":
    print("⚡ QuotaDock Windows Tray Companion")
    is_active = check_daemon_status()
    print(f"Daemon on {DAEMON_URL}: {'ONLINE (Green)' if is_active else 'OFFLINE (Start daemon first)'}")
    print("Commands:")
    print("  [1] Open Web Dashboard in Browser")
    print("  [2] Print Live Telemetry")
    print("  [3] Exit")
    
    if len(sys.argv) > 1 and sys.argv[1] == "--dashboard":
        open_dashboard()
        sys.exit(0)
    elif len(sys.argv) > 1 and sys.argv[1] == "--telemetry":
        print_telemetry()
        sys.exit(0)
