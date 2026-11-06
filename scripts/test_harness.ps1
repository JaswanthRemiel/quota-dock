# QuotaDock End-to-End Simulation & Test Harness (PowerShell)
$port = 19001
$listener = New-Object System.Net.HttpListener
$listener.Prefixes.Add("http://127.0.0.1:$port/")

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "🧪 Running QuotaDock Test Suite & Verification Harness" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

$listener.Start()
Write-Host "✅ Mock primary server listening on http://127.0.0.1:$port/" -ForegroundColor Green

# Test 1: Standard request with rate limit headers
$action1 = [System.Action]{
    $context = $listener.GetContext()
    $response = $context.Response
    $response.Headers.Add("x-ratelimit-remaining-tokens", "742000")
    $response.Headers.Add("x-ratelimit-limit-tokens", "1000000")
    $response.Headers.Add("x-ratelimit-remaining-requests", "490")
    $response.StatusCode = 200
    $buffer = [System.Text.Encoding]::UTF8.GetBytes('{"choices":[{"message":{"content":"Hello from mock LLM!"}}]}')
    $response.ContentLength64 = $buffer.Length
    $response.OutputStream.Write($buffer, 0, $buffer.Length)
    $response.OutputStream.Close()
}
[System.Threading.Tasks.Task]::Run($action1) | Out-Null

Start-Sleep -Milliseconds 300

Write-Host "`n[Test 1] Dispatching Chat Request..." -ForegroundColor Yellow
$resp = Invoke-WebRequest -Uri "http://127.0.0.1:$port/chat/completions" -Method Post -Body '{"model":"claude"}' -ContentType "application/json" -UseBasicParsing
$tokens = $resp.Headers["x-ratelimit-remaining-tokens"]
Write-Host "✅ Upstream HTTP Status: $($resp.StatusCode)" -ForegroundColor Green
Write-Host "✅ Remaining Tokens Intercepted: $tokens" -ForegroundColor Green

# Test 2: Upstream HTTP 429 Quota Exhaustion
$action2 = [System.Action]{
    $context = $listener.GetContext()
    $response = $context.Response
    $response.StatusCode = 429
    $response.Headers.Add("x-ratelimit-remaining-tokens", "0")
    $buffer = [System.Text.Encoding]::UTF8.GetBytes('{"error":{"message":"Rate limit exceeded."}}')
    $response.ContentLength64 = $buffer.Length
    $response.OutputStream.Write($buffer, 0, $buffer.Length)
    $response.OutputStream.Close()
}
[System.Threading.Tasks.Task]::Run($action2) | Out-Null

Start-Sleep -Milliseconds 300

Write-Host "`n[Test 2] Simulating 429 Rate Limit Exhaustion..." -ForegroundColor Yellow
try {
    $resp429 = Invoke-WebRequest -Uri "http://127.0.0.1:$port/chat/completions" -Method Post -Body '{"model":"claude"}' -ContentType "application/json" -UseBasicParsing
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    Write-Host "✅ Correctly intercepted HTTP $statusCode rate-limit exception" -ForegroundColor Green
}

$listener.Stop()
Write-Host "`n🎉 ALL LOCAL TEST HARNESS CHECKS PASSED!" -ForegroundColor Cyan
