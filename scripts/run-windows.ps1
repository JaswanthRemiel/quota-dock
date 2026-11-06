# QuotaDock Windows Launch Script
# Starts the background proxy daemon and launches the system tray companion

Write-Host "⚡ Starting QuotaDock for Windows..." -ForegroundColor Cyan

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir = Split-Path -Parent $scriptDir
$daemonExe = Join-Path $rootDir "daemon\daemon.exe"

if (Test-Path $daemonExe) {
    Write-Host "Starting background daemon ($daemonExe)..." -ForegroundColor Green
    Start-Process -FilePath $daemonExe -ArgumentList "-port 8080" -WindowStyle Hidden
} else {
    Write-Host "daemon.exe not found. Starting daemon using go run..." -ForegroundColor Yellow
    Start-Process -FilePath "go" -ArgumentList "run . -port 8080" -WorkingDirectory (Join-Path $rootDir "daemon") -WindowStyle Hidden
}

Start-Sleep -Seconds 1

Write-Host "Opening QuotaDock Dashboard..." -ForegroundColor Green
Start-Process "http://127.0.0.1:8080/dashboard"

Write-Host "✅ QuotaDock is running on http://127.0.0.1:8080" -ForegroundColor Green
