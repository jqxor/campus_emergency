$ErrorActionPreference = 'Stop'

$rebuildRoot = Resolve-Path (Join-Path $PSScriptRoot '..')

function Invoke-GoBuild([string]$serviceDir, [string]$exeName) {
  Write-Host "[go build] $serviceDir" -ForegroundColor Cyan
  Push-Location $serviceDir
  go build -o $exeName main.go
  Pop-Location
}

function Invoke-FrontendBuild([string]$frontendDir) {
  Write-Host "[npm] $frontendDir" -ForegroundColor Cyan
  Push-Location $frontendDir
  if (-not (Test-Path 'node_modules')) {
    cmd /c "npm install"
  }
  cmd /c "npm run build"
  Pop-Location
}

Invoke-GoBuild (Join-Path $rebuildRoot 'path_optimization') 'app.exe'
Invoke-GoBuild (Join-Path $rebuildRoot 'campus_emergency') 'app.exe'
Invoke-GoBuild (Join-Path $rebuildRoot 'role_management') 'app.exe'
Invoke-FrontendBuild (Join-Path $rebuildRoot 'campus_frontend')

Write-Host 'Build done.' -ForegroundColor Green
