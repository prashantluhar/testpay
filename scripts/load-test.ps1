#requires -version 5.1
<#
.SYNOPSIS
    End-to-end load test for a running TestPay instance.

.DESCRIPTION
    Starts a small HTTP listener, creates N users (each gets their own workspace
    + API key + webhook_url pointing at the listener), fires K charges per user
    against the mock gateway, and waits for webhooks to arrive. Reports:
      - # requests sent
      - # webhooks received per user
      - # request_logs rows in the DB
      - # webhook_logs rows in the DB

.PARAMETER Users
    Number of users to sign up. Default 3.

.PARAMETER RequestsPerUser
    Number of charges per user. Default 5.

.PARAMETER ApiBase
    TestPay API base URL. Default http://localhost:7700

.PARAMETER ListenerPort
    Local port for the webhook listener. Default 9999.

.EXAMPLE
    .\scripts\load-test.ps1
    .\scripts\load-test.ps1 -Users 5 -RequestsPerUser 10
#>

param(
    [int]$Users = 3,
    [int]$RequestsPerUser = 5,
    [string]$ApiBase = 'http://localhost:7700',
    [int]$ListenerPort = 9999
)

$ErrorActionPreference = 'Stop'
Write-Host "=== TestPay load test ===" -ForegroundColor Cyan
Write-Host "API:        $ApiBase"
Write-Host "Users:      $Users"
Write-Host "Per user:   $RequestsPerUser"
Write-Host "Listener:   http://localhost:$ListenerPort"
Write-Host ''

# ── Shared state for the listener ────────────────────────────────────────────
$counters = [System.Collections.Concurrent.ConcurrentDictionary[string, int]]::new()
$listener = [System.Net.HttpListener]::new()
$listener.Prefixes.Add("http://localhost:$ListenerPort/")
try { $listener.Start() }
catch {
    Write-Error "Failed to start listener on :$ListenerPort. Is the port busy? $_"
    exit 1
}

$listenerJob = Start-Job -ScriptBlock {
    param($listener, $counters)
    while ($listener.IsListening) {
        try {
            $ctx = $listener.GetContext()
            $path = $ctx.Request.Url.AbsolutePath.TrimStart('/')
            $null = $counters.AddOrUpdate($path, 1, { param($k, $v) $v + 1 })
            $reader = [System.IO.StreamReader]::new($ctx.Request.InputStream)
            $null = $reader.ReadToEnd()
            $reader.Dispose()
            $ctx.Response.StatusCode = 200
            $ctx.Response.Close()
        } catch {
            # listener stopped
            break
        }
    }
} -ArgumentList $listener, $counters

# ── Helpers ──────────────────────────────────────────────────────────────────
function Invoke-Json {
    param(
        [string]$Method,
        [string]$Url,
        [hashtable]$Body,
        [hashtable]$Headers,
        [Microsoft.PowerShell.Commands.WebRequestSession]$Session
    )
    $params = @{
        Method          = $Method
        Uri             = $Url
        ContentType     = 'application/json'
        UseBasicParsing = $true
        ErrorAction     = 'Stop'
    }
    if ($Body) { $params.Body = ($Body | ConvertTo-Json -Compress) }
    if ($Headers) { $params.Headers = $Headers }
    if ($Session) { $params.WebSession = $Session }
    return Invoke-RestMethod @params
}

# ── Sign up N users ──────────────────────────────────────────────────────────
Write-Host "Signing up $Users users..." -ForegroundColor Yellow
$userData = @()
for ($i = 1; $i -le $Users; $i++) {
    $session = [Microsoft.PowerShell.Commands.WebRequestSession]::new()
    $suffix = [Guid]::NewGuid().ToString('N').Substring(0, 6)
    $email = "loadtest-$suffix@example.com"
    $password = 'loadtest-password'
    try {
        $signup = Invoke-Json -Method POST -Url "$ApiBase/api/auth/signup" `
            -Body @{ email = $email; password = $password } -Session $session
    } catch {
        Write-Warning "signup $email failed: $($_.Exception.Message)"
        continue
    }
    $userKey = "user$i-$suffix"
    $webhookUrl = "http://localhost:$ListenerPort/$userKey"

    # Set the workspace webhook_url via PUT /api/workspace
    Invoke-Json -Method PUT -Url "$ApiBase/api/workspace" `
        -Body @{ webhook_url = $webhookUrl } -Session $session | Out-Null

    $userData += [pscustomobject]@{
        Email      = $email
        Key        = $userKey
        APIKey     = $signup.workspace.api_key
        WebhookURL = $webhookUrl
    }
    Write-Host ("  {0,2}. {1,-40} api_key={2} webhook={3}" -f $i, $email, $signup.workspace.api_key.Substring(0, 8), $userKey)
}

if ($userData.Count -eq 0) {
    Write-Error 'No users signed up. Aborting.'
    $listener.Stop(); exit 1
}

# ── Fire charges per user ────────────────────────────────────────────────────
Write-Host ''
Write-Host "Firing $($userData.Count * $RequestsPerUser) charges..." -ForegroundColor Yellow
$sw = [System.Diagnostics.Stopwatch]::StartNew()
$sent = 0
foreach ($u in $userData) {
    for ($j = 1; $j -le $RequestsPerUser; $j++) {
        try {
            Invoke-Json -Method POST -Url "$ApiBase/stripe/v1/charges" `
                -Body @{ amount = (1000 * $j); currency = 'usd' } `
                -Headers @{ Authorization = "Bearer $($u.APIKey)" } | Out-Null
            $sent++
        } catch {
            Write-Warning "charge failed for $($u.Email): $($_.Exception.Message)"
        }
    }
}
$sw.Stop()
Write-Host "  $sent charges sent in $([math]::Round($sw.Elapsed.TotalSeconds, 2))s"

# ── Wait for webhooks ────────────────────────────────────────────────────────
Write-Host ''
Write-Host 'Waiting up to 10s for webhooks to arrive...' -ForegroundColor Yellow
$expected = $userData.Count * $RequestsPerUser
$deadline = (Get-Date).AddSeconds(10)
while ((Get-Date) -lt $deadline) {
    $received = 0
    foreach ($u in $userData) { $received += $counters[$u.Key] }
    if ($received -ge $expected) { break }
    Start-Sleep -Milliseconds 500
}

# ── Report ───────────────────────────────────────────────────────────────────
Write-Host ''
Write-Host '=== Results ===' -ForegroundColor Cyan
$totalReceived = 0
foreach ($u in $userData) {
    $n = $counters[$u.Key]
    if (-not $n) { $n = 0 }
    $totalReceived += $n
    $status = if ($n -eq $RequestsPerUser) { 'OK' } else { 'MISS' }
    Write-Host ("  [{0}] {1,-40} sent={2} webhooks_received={3}" -f $status, $u.Email, $RequestsPerUser, $n)
}

Write-Host ''
Write-Host "Total charges sent:       $sent"
Write-Host "Total webhooks received:  $totalReceived (expected $expected)"

# DB verification via the control API
try {
    $logs = Invoke-Json -Method GET -Url "$ApiBase/api/logs?limit=1000"
    Write-Host "Rows in request_logs:     $($logs.Count)"
} catch {
    Write-Warning "could not query /api/logs: $($_.Exception.Message)"
}

# ── Cleanup ──────────────────────────────────────────────────────────────────
$listener.Stop()
Stop-Job -Job $listenerJob -ErrorAction SilentlyContinue | Out-Null
Remove-Job -Job $listenerJob -ErrorAction SilentlyContinue | Out-Null

if ($totalReceived -eq $expected) {
    Write-Host ''
    Write-Host 'PASS' -ForegroundColor Green
    exit 0
} else {
    Write-Host ''
    Write-Host 'PARTIAL / FAIL — some webhooks did not arrive within the timeout.' -ForegroundColor Red
    exit 2
}
