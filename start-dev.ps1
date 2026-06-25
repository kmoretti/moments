$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new()
$OutputEncoding = [System.Text.UTF8Encoding]::new()
chcp 65001 > $null

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$backend = Join-Path $root 'backend'
$front = Join-Path $root 'front'
$goBin = 'E:\go\bin\go.exe'
$pnpmCommand = Get-Command 'pnpm' -ErrorAction SilentlyContinue

if (-not (Test-Path $backend)) {
  Write-Host 'backend directory not found. Please run this script from repo root.' -ForegroundColor Red
  exit 1
}

if (-not (Test-Path $front)) {
  Write-Host 'front directory not found. Please run this script from repo root.' -ForegroundColor Red
  exit 1
}

if (-not (Test-Path $goBin)) {
  Write-Host "go.exe not found: $goBin" -ForegroundColor Red
  Write-Host 'Please install Go first, or update $goBin in this script.' -ForegroundColor Yellow
  exit 1
}

if (-not $pnpmCommand) {
  Write-Host 'pnpm command not found.' -ForegroundColor Red
  Write-Host 'Please install pnpm first, for example: npm install -g pnpm' -ForegroundColor Yellow
  exit 1
}

Get-CimInstance Win32_Process |
  Where-Object {
    $_.CommandLine -like "*$front*nuxt*dev*" -or
    $_.CommandLine -like "*$front*pnpm.js*dev*" -or
    $_.CommandLine -like "*$backend*go run*" -or
    $_.CommandLine -like "*$backend*moments.exe*"
  } |
  ForEach-Object {
    Stop-Process -Id $_.ProcessId -Force -ErrorAction SilentlyContinue
  }

Start-Process powershell -ArgumentList @(
  '-NoExit',
  '-Command',
  "[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new(); chcp 65001 > `$null; Set-Location '$backend'; `$env:PORT='37893'; `$env:CORS_ORIGIN='http://127.0.0.1:3000,http://localhost:3000,http://0.0.0.0:3000'; & '$goBin' run ."
)

$backendReady = $false
for ($i = 0; $i -lt 10; $i++) {
  Start-Sleep -Seconds 1
  try {
    $resp = Invoke-WebRequest -Uri 'http://127.0.0.1:37893/api/sysConfig/get' -Method Post -UseBasicParsing
    if ($resp.StatusCode -eq 200) {
      $backendReady = $true
      break
    }
  } catch {
  }
}

if (-not $backendReady) {
  Write-Host 'Backend failed to start. Check the backend terminal window.' -ForegroundColor Red
  exit 1
}

Start-Process powershell -ArgumentList @(
  '-NoExit',
  '-Command',
  "[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new(); chcp 65001 > `$null; Set-Location '$front'; `$env:NUXT_DEV_PROXY_TARGET='http://127.0.0.1:37893'; & '$($pnpmCommand.Source)' dev"
)

Start-Sleep -Seconds 5
Start-Process 'http://127.0.0.1:3000/'

Write-Host 'Local dev started:'
Write-Host 'Front: http://127.0.0.1:3000/'
Write-Host 'Backend: http://127.0.0.1:37893/api'
