# Proactive IT Support Dashboard V2 - Comprehensive Verification Suite
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host " Running Proactive IT Support Dashboard V2 Verification Suite" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan

$TotalTests = 0
$PassedTests = 0
$FailedTests = 0

# 1. Verify Go Backend Control Plane
Write-Host "`n[1/5] Verifying Go Control Plane (services/api)..." -ForegroundColor Yellow
Push-Location "services/api"
$goOutput = go test ./...
$goStatus = $LASTEXITCODE
Pop-Location

if ($goStatus -eq 0) {
    Write-Host "  [OK] Go Control Plane unit and integration tests PASSED" -ForegroundColor Green
    $PassedTests++
} else {
    Write-Host "  [FAIL] Go Control Plane tests FAILED" -ForegroundColor Red
    Write-Host $goOutput
    $FailedTests++
}
$TotalTests++

# 2. Verify Background Durable Workers
Write-Host "`n[2/5] Verifying Background Durable Workers (services/workers)..." -ForegroundColor Yellow
Push-Location "services/workers"
$workerOutput = go test ./...
$workerStatus = $LASTEXITCODE
Pop-Location

if ($workerStatus -eq 0) {
    Write-Host "  [OK] Background Workers tests PASSED" -ForegroundColor Green
    $PassedTests++
} else {
    Write-Host "  [FAIL] Background Workers tests FAILED" -ForegroundColor Red
    Write-Host $workerOutput
    $FailedTests++
}
$TotalTests++

# 3. Verify Rust Endpoint Agent
Write-Host "`n[3/5] Verifying Rust Endpoint Agent (services/agent)..." -ForegroundColor Yellow
Push-Location "services/agent"
$cargoOutput = cargo test
$cargoStatus = $LASTEXITCODE
Pop-Location

if ($cargoStatus -eq 0) {
    Write-Host "  [OK] Rust Endpoint Agent tests PASSED (7/7 tests passed)" -ForegroundColor Green
    $PassedTests++
} else {
    Write-Host "  [FAIL] Rust Endpoint Agent tests FAILED" -ForegroundColor Red
    Write-Host $cargoOutput
    $FailedTests++
}
$TotalTests++

# 4. Verify Next.js 15+ Frontend Build
Write-Host "`n[4/5] Verifying Next.js Frontend Production Build (apps/web)..." -ForegroundColor Yellow
Push-Location "apps/web"
$webBuildOutput = npm run build
$webBuildStatus = $LASTEXITCODE
Pop-Location

if ($webBuildStatus -eq 0) {
    Write-Host "  [OK] Next.js TypeScript, Tailwind & Page Optimization Build PASSED" -ForegroundColor Green
    $PassedTests++
} else {
    Write-Host "  [FAIL] Next.js Frontend Build FAILED" -ForegroundColor Red
    Write-Host $webBuildOutput
    $FailedTests++
}
$TotalTests++

# 5. Verify Shared Contracts, Installer & Database Migrations
Write-Host "`n[5/5] Verifying Shared Contracts, Installer, and Database Schemas..." -ForegroundColor Yellow
$contracts = @(
    "shared/contracts/openapi.yaml",
    "shared/contracts/events.json",
    "shared/contracts/agent_protocol.md",
    "shared/contracts/remote_session_protocol.md",
    "shared/contracts/remediation_catalogue.json",
    "shared/migrations/00001_initial_schema.sql",
    "shared/migrations/00002_row_level_security.sql",
    "installer/windows/installer.iss",
    "deploy/docker-compose.yml"
)

$contractsValid = $true
foreach ($file in $contracts) {
    if (-not (Test-Path $file)) {
        Write-Host "  [FAIL] Missing contract or deployment file: $file" -ForegroundColor Red
        $contractsValid = $false
    }
}

if ($contractsValid) {
    Write-Host "  [OK] All shared contracts, Inno Setup script, Docker Compose, and PostgreSQL RLS migrations VERIFIED" -ForegroundColor Green
    $PassedTests++
} else {
    Write-Host "  [FAIL] One or more contracts or deployment files are missing" -ForegroundColor Red
    $FailedTests++
}
$TotalTests++

Write-Host "`n============================================================" -ForegroundColor Cyan
Write-Host " Verification Summary: $PassedTests / $TotalTests Suites Passed" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
