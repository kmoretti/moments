$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$backend = Join-Path $root 'backend'
$front = Join-Path $root 'front'
$goBin = 'E:\go\bin\go.exe'
$pnpmCommand = Get-Command 'pnpm' -ErrorAction SilentlyContinue

if (-not (Test-Path $backend)) {
  Write-Host '未找到 backend 目录，请确认脚本位于仓库根目录。' -ForegroundColor Red
  exit 1
}

if (-not (Test-Path $front)) {
  Write-Host '未找到 front 目录，请确认脚本位于仓库根目录。' -ForegroundColor Red
  exit 1
}

if (-not (Test-Path $goBin)) {
  Write-Host "未找到 go.exe：$goBin" -ForegroundColor Red
  Write-Host '请先安装 Go，或把脚本中的 $goBin 改成你本机的 go.exe 路径。' -ForegroundColor Yellow
  exit 1
}

if (-not $pnpmCommand) {
  Write-Host '未找到 pnpm 命令。' -ForegroundColor Red
  Write-Host '请先安装 pnpm，例如执行：npm install -g pnpm' -ForegroundColor Yellow
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
  "Set-Location '$backend'; `$env:PORT='37893'; `$env:CORS_ORIGIN='http://127.0.0.1:3000,http://localhost:3000,http://0.0.0.0:3000'; & '$goBin' run ."
)

Start-Sleep -Seconds 3

Start-Process powershell -ArgumentList @(
  '-NoExit',
  '-Command',
  "Set-Location '$front'; `$env:NUXT_DEV_PROXY_TARGET='http://127.0.0.1:37893'; & '$($pnpmCommand.Source)' dev"
)

Start-Sleep -Seconds 5
Start-Process 'http://127.0.0.1:3000/'

Write-Host '本地测试已启动：'
Write-Host '前端: http://127.0.0.1:3000/'
Write-Host '后端: http://127.0.0.1:37893/api'
