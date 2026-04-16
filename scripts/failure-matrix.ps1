#requires -version 5.1
<#
.SYNOPSIS
    Drive every mock gateway through a matrix of failure scenarios for one user.

.DESCRIPTION
    1. Logs into the specified account (or signs it up if not found).
    2. Sets a default webhook URL on the workspace.
    3. Pre-creates one scenario per failure mode being tested.
    4. For every (gateway x scenario) combination: pins the scenario via a
       session, fires a charge, records the HTTP status. Webhook delivery is
       async and visible on the configured URL.

.PARAMETER Email
    Existing or new account email. Default: loadtest-b1f55f@example.com

.PARAMETER Password
    Default "loadtest-password" (matches load-test.ps1).

.PARAMETER ApiBase
    TestPay API base URL.

.PARAMETER WebhookURL
    Single webhook URL applied as the workspace default. Empty = skip configure.

.PARAMETER Gateways
    Comma-separated gateway names to test. Default: all known ones.

.PARAMETER Modes
    Comma-separated outcome modes to cover. Default: a curated mix.

.EXAMPLE
    .\scripts\failure-matrix.ps1
    .\scripts\failure-matrix.ps1 -Email alice@example.com -Gateways stripe,adyen
    .\scripts\failure-matrix.ps1 -Modes success,bank_decline_hard,network_error
#>

param(
    [string]$Email    = 'loadtest-b1f55f@example.com',
    [string]$Password = 'loadtest-password',
    [string]$ApiBase  = 'http://localhost:7700',
    [string]$WebhookURL = 'https://webhook.site/d2c50235-b68a-4af3-a182-206cc01052cb',
    [string[]]$Gateways = @(),
    [string[]]$Modes = @()
)

$ErrorActionPreference = 'Stop'

# Default gateway set: everything except agnostic (covered via /v1/*)
if ($Gateways.Count -eq 0) {
    $Gateways = @('stripe', 'razorpay', 'adyen', 'omise', 'mastercard', 'komoju',
                  'instamojo', 'tillpay', 'tappay', 'payletter', 'paynamics',
                  'epay', 'espay')
}
# Default mode mix - one success, bank, PG, webhook anomaly, and network
if ($Modes.Count -eq 0) {
    $Modes = @('success', 'bank_decline_hard', 'bank_decline_soft',
              'bank_timeout', 'pg_rate_limited', 'pg_server_error',
              'network_error', 'webhook_delayed', 'webhook_missing')
}

# URL path per gateway - most are /{gateway}/v1/charges; a few differ.
$GatewayPath = @{
    'stripe'     = '/stripe/v1/charges'
    'razorpay'   = '/razorpay/v1/payments'
    'adyen'      = '/adyen/v1/payments'
    'omise'      = '/omise/v1/charges'
    'mastercard' = '/mastercard/v1/payments'
    'komoju'     = '/komoju/v1/payments'
    'instamojo'  = '/instamojo/v1/payments'
    'tillpay'    = '/tillpay/v1/charges'
    'tappay'     = '/tappay/v1/payments'
    'payletter'  = '/payletter/v1/payments'
    'paynamics'  = '/paynamics/v1/payments'
    'epay'       = '/epay/v1/payments'
    'espay'      = '/espay/v1/payments'
    'agnostic'   = '/v1/charges'
}

Write-Host "=== TestPay failure matrix ===" -ForegroundColor Cyan
Write-Host "Account:    $Email"
Write-Host "API:        $ApiBase"
Write-Host "Gateways:   $($Gateways -join ', ')"
Write-Host "Modes:      $($Modes -join ', ')"
Write-Host "Webhook:    $WebhookURL"
Write-Host ''

# Helpers
$session = [Microsoft.PowerShell.Commands.WebRequestSession]::new()

# Compose a plausible request body per gateway (+ surface order_id for echo).
function Build-RequestBody {
    param([string]$Gateway, [string]$Mode)
    $ordId = "ord-$Gateway-$Mode-$([guid]::NewGuid().ToString('N').Substring(0,6))"
    switch ($Gateway) {
        'stripe' {
            return @{ amount = 5000; currency = 'usd'; metadata = @{ order_id = $ordId } }
        }
        'razorpay' {
            return @{ amount = 5000; currency = 'INR'; notes = @{ order_id = $ordId } }
        }
        'adyen' {
            return @{ amount = @{ value = 5000; currency = 'USD' }; additionalData = @{ order_id = $ordId } }
        }
        'mastercard' {
            return @{ order = @{ amount = 5000; currency = 'USD' }; metadata = @{ order_id = $ordId } }
        }
        default {
            return @{ amount = 5000; currency = 'USD'; order_id = $ordId }
        }
    }
}

function Invoke-Json {
    param(
        [string]$Method,
        [string]$Url,
        $Body,
        [hashtable]$Headers
    )
    $p = @{
        Method          = $Method
        Uri             = $Url
        ContentType     = 'application/json'
        UseBasicParsing = $true
        WebSession      = $session
        ErrorAction     = 'Stop'
    }
    if ($null -ne $Body) { $p.Body = ($Body | ConvertTo-Json -Compress -Depth 5) }
    if ($Headers)        { $p.Headers = $Headers }
    return Invoke-RestMethod @p
}

# 1. Log in (or sign up if login fails)
Write-Host "Logging in as $Email..." -ForegroundColor Yellow
$me = $null
try {
    $me = Invoke-Json -Method POST -Url "$ApiBase/api/auth/login" `
        -Body @{ email = $Email; password = $Password }
    Write-Host "  logged in  api_key=$($me.workspace.api_key.Substring(0,8))..."
} catch {
    Write-Host "  login failed ($($_.Exception.Message)) - trying signup" -ForegroundColor Yellow
    $me = Invoke-Json -Method POST -Url "$ApiBase/api/auth/signup" `
        -Body @{ email = $Email; password = $Password }
    Write-Host "  signed up   api_key=$($me.workspace.api_key.Substring(0,8))..."
}
$apiKey = $me.workspace.api_key

# 2. Set workspace default webhook URL (if provided)
if (-not [string]::IsNullOrEmpty($WebhookURL)) {
    Invoke-Json -Method PUT -Url "$ApiBase/api/workspace" `
        -Body @{ webhook_urls = @{ '_default' = $WebhookURL } } | Out-Null
    Write-Host "  webhook default set to $WebhookURL"
}

# 3. Pre-create (or reuse) one scenario per mode
Write-Host ''
Write-Host 'Creating scenarios...' -ForegroundColor Yellow
$existing = Invoke-Json -Method GET -Url "$ApiBase/api/scenarios"
$scenarioByMode = @{}
foreach ($existingScenario in $existing) {
    $scenarioByMode[$existingScenario.name] = $existingScenario.id
}
foreach ($mode in $Modes) {
    $name = "fmatrix-$mode"
    if ($scenarioByMode.ContainsKey($name)) {
        Write-Host ("  {0,-30} reusing scenario" -f $name)
        continue
    }
    $body = @{
        name             = $name
        description      = "failure-matrix: $mode"
        gateway          = 'agnostic'
        webhook_delay_ms = 0
        is_default       = $false
        steps            = @(@{ event = 'charge'; outcome = $mode })
    }
    $created = Invoke-Json -Method POST -Url "$ApiBase/api/scenarios" -Body $body
    $scenarioByMode[$name] = $created.id
    Write-Host ("  {0,-30} created {1}" -f $name, $created.id.Substring(0,8))
}

# 4. Matrix: for each mode, pin scenario via session, hit every gateway
$results = @()
foreach ($mode in $Modes) {
    $scName = "fmatrix-$mode"
    $scId   = $scenarioByMode[$scName]

    # Pin scenario with short TTL (plenty for the few calls we'll fire)
    Invoke-Json -Method POST -Url "$ApiBase/api/sessions" `
        -Body @{ scenario_id = $scId; ttl_seconds = 60 } | Out-Null

    Write-Host ''
    Write-Host "Mode: $mode" -ForegroundColor Cyan
    foreach ($g in $Gateways) {
        $path = $GatewayPath[$g]
        if (-not $path) {
            Write-Host ("  {0,-12} <unknown gateway path>" -f $g) -ForegroundColor DarkGray
            continue
        }

        $body = Build-RequestBody -Gateway $g -Mode $mode

        $status = $null
        $msg    = ''
        try {
            $resp = Invoke-WebRequest -Method POST -Uri "$ApiBase$path" `
                -Headers @{ Authorization = "Bearer $apiKey" } `
                -ContentType 'application/json' `
                -Body ($body | ConvertTo-Json -Compress -Depth 5) `
                -UseBasicParsing -ErrorAction Stop
            $status = [int]$resp.StatusCode
        } catch {
            if ($_.Exception.Response) {
                $status = [int]$_.Exception.Response.StatusCode
            }
            $msg = $_.Exception.Message
        }
        $marker = switch ($true) {
            ($status -ge 200 -and $status -lt 300) { 'OK'   }
            ($status -ge 400 -and $status -lt 500) { 'DECL' }
            ($status -ge 500)                      { 'SRVR' }
            default                                { '?'    }
        }
        Write-Host ("  {0,-12} {1,-4} http={2}  {3}" -f $g, $marker, $status, $msg)
        $results += [pscustomobject]@{
            Mode = $mode; Gateway = $g; Status = $status
        }
    }
}

# 5. Summary
Write-Host ''
Write-Host '=== Summary ===' -ForegroundColor Cyan
$grouped = $results | Group-Object Mode
foreach ($group in $grouped) {
    $statuses = ($group.Group | Select-Object -ExpandProperty Status) -join ' '
    Write-Host ("  {0,-22} statuses: {1}" -f $group.Name, $statuses)
}

Write-Host ''
Write-Host "Total requests fired: $($results.Count)"
Write-Host "Inspect webhooks at:  $WebhookURL" -ForegroundColor Cyan
Write-Host "Dashboard:            http://localhost:7701 (log in as $Email)" -ForegroundColor Cyan

# (Build-RequestBody helper moved to top so it's defined before use.)
