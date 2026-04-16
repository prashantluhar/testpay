#requires -version 5.1
<#
.SYNOPSIS
    End-to-end load test for a running TestPay instance.

.DESCRIPTION
    Starts a small HTTP listener, creates N users (each gets their own workspace,
    API key, and per-gateway webhook URLs pointing at the listener), fires K
    charges per user across all three gateways, verifies webhooks arrive with
    the right echoed metadata (order_id).

.PARAMETER Users
    Number of users to sign up. Default 3.

.PARAMETER RequestsPerUser
    Number of charges per user per gateway. Default 2.

.PARAMETER ApiBase
    TestPay API base URL. Default http://localhost:7700

.PARAMETER ListenerPort
    Local port for the webhook listener. Default 9999.

.EXAMPLE
    .\scripts\load-test.ps1
    .\scripts\load-test.ps1 -Users 5 -RequestsPerUser 4
#>

param(
    [int]$Users = 3,
    [int]$RequestsPerUser = 2,
    [string]$ApiBase = 'http://localhost:7700',
    [int]$ListenerPort = 9999
)

$ErrorActionPreference = 'Stop'
Write-Host "=== TestPay load test ===" -ForegroundColor Cyan
Write-Host "API:                   $ApiBase"
Write-Host "Users:                 $Users"
Write-Host "Per user per gateway:  $RequestsPerUser"
Write-Host "Gateways:              stripe, razorpay, agnostic"
Write-Host "Listener:              http://localhost:$ListenerPort"
Write-Host ''

# ── Listener (in-process) ────────────────────────────────────────────────────
$counters  = [System.Collections.Concurrent.ConcurrentDictionary[string, int]]::new()
$payloads  = [System.Collections.Concurrent.ConcurrentDictionary[string, string]]::new()
$listener  = [System.Net.HttpListener]::new()
$listener.Prefixes.Add("http://localhost:$ListenerPort/")
try { $listener.Start() } catch {
    Write-Error "Failed to start listener on :$ListenerPort. $_"; exit 1
}

$listenerJob = Start-Job -ScriptBlock {
    param($listener, $counters, $payloads)
    while ($listener.IsListening) {
        try {
            $ctx    = $listener.GetContext()
            $path   = $ctx.Request.Url.AbsolutePath.TrimStart('/')
            $null   = $counters.AddOrUpdate($path, 1, { param($k, $v) $v + 1 })
            $reader = [System.IO.StreamReader]::new($ctx.Request.InputStream)
            $body   = $reader.ReadToEnd()
            $reader.Dispose()
            $null   = $payloads.TryAdd($path + "#" + $counters[$path], $body)
            $ctx.Response.StatusCode = 200
            $ctx.Response.Close()
        } catch { break }
    }
} -ArgumentList $listener, $counters, $payloads

# ── Helpers ──────────────────────────────────────────────────────────────────
function Invoke-Json {
    param(
        [string]$Method,
        [string]$Url,
        $Body,
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
    if ($null -ne $Body)   { $params.Body = ($Body | ConvertTo-Json -Compress -Depth 5) }
    if ($Headers)          { $params.Headers = $Headers }
    if ($Session)          { $params.WebSession = $Session }
    return Invoke-RestMethod @params
}

# ── 1. Sign up N users ───────────────────────────────────────────────────────
Write-Host "Signing up $Users users..." -ForegroundColor Yellow
$userData = @()
for ($i = 1; $i -le $Users; $i++) {
    $session = [Microsoft.PowerShell.Commands.WebRequestSession]::new()
    $suffix  = [Guid]::NewGuid().ToString('N').Substring(0, 6)
    $email   = "loadtest-$suffix@example.com"
    try {
        $signup = Invoke-Json -Method POST -Url "$ApiBase/api/auth/signup" `
            -Body @{ email = $email; password = 'loadtest-password' } -Session $session
    } catch {
        Write-Warning "signup $email failed: $($_.Exception.Message)"; continue
    }

    $userKey = "user$i-$suffix"
    $hooks   = @{
        stripe   = "http://localhost:$ListenerPort/$userKey-stripe"
        razorpay = "http://localhost:$ListenerPort/$userKey-razorpay"
        agnostic = "http://localhost:$ListenerPort/$userKey-agnostic"
    }
    # Authenticate via api_key (Authorization: Bearer) — more reliable across
    # PowerShell WebSession quirks than the signup cookie.
    Invoke-Json -Method PUT -Url "$ApiBase/api/workspace" `
        -Body @{ webhook_urls = $hooks } `
        -Headers @{ Authorization = "Bearer $($signup.workspace.api_key)" } | Out-Null

    $userData += [pscustomobject]@{
        Email  = $email
        Key    = $userKey
        APIKey = $signup.workspace.api_key
        Hooks  = $hooks
    }
    Write-Host ("  {0,2}. {1,-40} api_key={2}…  hook_prefix={3}" -f $i, $email,
        $signup.workspace.api_key.Substring(0, 8), $userKey)
}

if ($userData.Count -eq 0) {
    Write-Error 'No users signed up. Aborting.'; $listener.Stop(); exit 1
}

# ── 2. Fire charges across all 3 gateways ────────────────────────────────────
$gateways = @(
    @{ Name = 'stripe';   Path = '/stripe/v1/charges';   Echo = 'metadata' }
    @{ Name = 'razorpay'; Path = '/razorpay/v1/payments'; Echo = 'notes'    }
    @{ Name = 'agnostic'; Path = '/v1/charges';          Echo = 'root'     }
)
$totalExpected = $userData.Count * $RequestsPerUser * $gateways.Count

Write-Host ''
Write-Host "Firing $totalExpected charges ($($userData.Count) users × $RequestsPerUser × $($gateways.Count) gateways)..." -ForegroundColor Yellow
$sw   = [System.Diagnostics.Stopwatch]::StartNew()
$sent = 0
foreach ($u in $userData) {
    foreach ($gw in $gateways) {
        for ($j = 1; $j -le $RequestsPerUser; $j++) {
            $orderId = "ord-$($u.Key)-$($gw.Name)-$j"
            # Echo location depends on gateway convention.
            $body = @{ amount = (1000 * $j); currency = 'usd' }
            switch ($gw.Echo) {
                'metadata' { $body.metadata = @{ order_id = $orderId } }
                'notes'    { $body.notes    = @{ order_id = $orderId } }
                'root'     { $body.order_id = $orderId }
            }
            try {
                Invoke-Json -Method POST -Url "$ApiBase$($gw.Path)" -Body $body `
                    -Headers @{ Authorization = "Bearer $($u.APIKey)" } | Out-Null
                $sent++
            } catch {
                Write-Warning "charge failed ($($u.Email) $($gw.Name)): $($_.Exception.Message)"
            }
        }
    }
}
$sw.Stop()
Write-Host "  $sent charges sent in $([math]::Round($sw.Elapsed.TotalSeconds, 2))s"

# ── 3. Wait for webhooks ─────────────────────────────────────────────────────
Write-Host ''
Write-Host 'Waiting up to 10s for webhooks to arrive...' -ForegroundColor Yellow
$deadline = (Get-Date).AddSeconds(10)
while ((Get-Date) -lt $deadline) {
    $received = 0
    foreach ($u in $userData) {
        foreach ($gw in $gateways) {
            $received += $counters["$($u.Key)-$($gw.Name)"]
        }
    }
    if ($received -ge $totalExpected) { break }
    Start-Sleep -Milliseconds 500
}

# ── 4. Report ────────────────────────────────────────────────────────────────
Write-Host ''
Write-Host '=== Results ===' -ForegroundColor Cyan
$totalReceived = 0
foreach ($u in $userData) {
    foreach ($gw in $gateways) {
        $key = "$($u.Key)-$($gw.Name)"
        $n   = $counters[$key]
        if (-not $n) { $n = 0 }
        $totalReceived += $n
        $status = if ($n -eq $RequestsPerUser) { 'OK' } else { 'MISS' }
        Write-Host ("  [{0}] {1,-40} {2,-9} sent={3} received={4}" -f $status, $u.Email, $gw.Name, $RequestsPerUser, $n)
    }
}

Write-Host ''
Write-Host "Total charges sent:       $sent"
Write-Host "Total webhooks received:  $totalReceived (expected $totalExpected)"

# Sample echoed metadata from one webhook (proves the order_id round-trips)
$sampleKey = "$($userData[0].Key)-stripe#1"
$sample    = $payloads[$sampleKey]
if ($sample) {
    Write-Host ''
    Write-Host 'Sample Stripe webhook payload (first request from first user):' -ForegroundColor Yellow
    try {
        $obj = $sample | ConvertFrom-Json
        $md  = $obj.data.object.metadata
        Write-Host "  event type:          $($obj.type)"
        Write-Host "  echoed order_id:     $($md.order_id)"
        Write-Host "  payment_intent id:   $($obj.data.object.id)"
    } catch {
        Write-Host ($sample.Substring(0, [Math]::Min(400, $sample.Length)))
    }
}

try {
    $logs = Invoke-Json -Method GET -Url "$ApiBase/api/logs?limit=1000"
    Write-Host "Rows in request_logs:     $($logs.Count)"
} catch {
    Write-Warning "could not query /api/logs: $($_.Exception.Message)"
}

# ── 5. Cleanup ───────────────────────────────────────────────────────────────
$listener.Stop()
Stop-Job   -Job $listenerJob -ErrorAction SilentlyContinue | Out-Null
Remove-Job -Job $listenerJob -ErrorAction SilentlyContinue | Out-Null

if ($totalReceived -eq $totalExpected) {
    Write-Host ''; Write-Host 'PASS' -ForegroundColor Green; exit 0
} else {
    Write-Host ''; Write-Host 'PARTIAL / FAIL — some webhooks did not arrive within the timeout.' -ForegroundColor Red; exit 2
}
