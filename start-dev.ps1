$ErrorActionPreference = 'Stop'

$env:GOROOT = 'E:\go'
$env:GOPATH = 'E:\go-work\gopath'
$env:GOCACHE = 'E:\go-work\gocache'
$env:GOMODCACHE = 'E:\go-work\gomodcache'
$env:GOBIN = 'E:\go-work\gobin'
$env:PATH = "E:\go\bin;E:\go-work\gobin;" + $env:PATH

Write-Host '[moments] Go environment configured:' -ForegroundColor Green
Write-Host "  GOROOT=$env:GOROOT"
Write-Host "  GOPATH=$env:GOPATH"
Write-Host "  GOCACHE=$env:GOCACHE"
Write-Host "  GOMODCACHE=$env:GOMODCACHE"
Write-Host "  GOBIN=$env:GOBIN"

function Start-MomentsBackend {
  param(
    [string]$WorkDir = 'E:\kmoretti-github\moments\backend'
  )

  Push-Location $WorkDir
  try {
    go run .
  }
  finally {
    Pop-Location
  }
}

function Start-MomentsFrontend {
  param(
    [string]$WorkDir = 'E:\kmoretti-github\moments\front'
  )

  Push-Location $WorkDir
  try {
    pnpm run dev
  }
  finally {
    Pop-Location
  }
}

function Build-MomentsBackend {
  param(
    [string]$WorkDir = 'E:\kmoretti-github\moments\backend'
  )

  Push-Location $WorkDir
  try {
    go build ./...
  }
  finally {
    Pop-Location
  }
}

function Install-MomentsBackendDeps {
  param(
    [string]$WorkDir = 'E:\kmoretti-github\moments\backend'
  )

  Push-Location $WorkDir
  try {
    go mod download
  }
  finally {
    Pop-Location
  }
}

function Install-MomentsFrontendDeps {
  param(
    [string]$WorkDir = 'E:\kmoretti-github\moments\front'
  )

  Push-Location $WorkDir
  try {
    pnpm install
  }
  finally {
    Pop-Location
  }
}

Write-Host ''
Write-Host 'Available commands:' -ForegroundColor Cyan
Write-Host '  Start-MomentsBackend'
Write-Host '  Start-MomentsFrontend'
Write-Host '  Build-MomentsBackend'
Write-Host '  Install-MomentsBackendDeps'
Write-Host '  Install-MomentsFrontendDeps'
