$ErrorActionPreference = 'Stop'

$rebuildRoot = Resolve-Path (Join-Path $PSScriptRoot '..')
$runDir = Join-Path $rebuildRoot '.run'
if (-not (Test-Path $runDir)) { New-Item -ItemType Directory -Path $runDir | Out-Null }
$pidsPath = Join-Path $runDir 'pids.json'

function Start-ServiceExe([string]$name, [string]$workDir, [string]$exePath, [string]$port, [string]$dbPath) {
  Write-Host "[start] $name (PORT=$port, DB_PATH=$dbPath)" -ForegroundColor Cyan

  $oldPort = $env:PORT
  $oldDb = $env:DB_PATH
  $env:PORT = $port
  $env:DB_PATH = $dbPath

  try {
    $p = Start-Process -FilePath $exePath -WorkingDirectory $workDir -WindowStyle Normal -PassThru
    return $p
  } finally {
    $env:PORT = $oldPort
    $env:DB_PATH = $oldDb
  }
}

function Start-Frontend([string]$workDir, [string]$port) {
  Write-Host "[start] campus_frontend (port=$port)" -ForegroundColor Cyan

  $cmd = "if not exist node_modules (npm install) else (echo node_modules ok) && npm run dev -- --port $port"
  $p = Start-Process -FilePath 'cmd.exe' -ArgumentList '/c', $cmd -WorkingDirectory $workDir -WindowStyle Normal -PassThru
  return $p
}

function Get-LatestBackendSourceTime([string]$serviceDir) {
  $latest = Get-ChildItem -Path $serviceDir -Recurse -File |
    Where-Object { $_.Name -like '*.go' -or $_.Name -eq 'go.mod' -or $_.Name -eq 'go.sum' } |
    Sort-Object LastWriteTime -Descending |
    Select-Object -First 1
  return $latest
}

function Ensure-GoBackendBuilt([string]$name, [string]$serviceDir, [string]$exePath) {
  $needsBuild = -not (Test-Path $exePath)
  if (-not $needsBuild) {
    $exeTime = (Get-Item $exePath).LastWriteTime
    $latestSrc = Get-LatestBackendSourceTime $serviceDir
    if ($latestSrc -and $latestSrc.LastWriteTime -gt $exeTime) {
      $needsBuild = $true
    }
  }

  if ($needsBuild) {
    Write-Host "[info] Building $name..." -ForegroundColor Yellow
    Push-Location $serviceDir
    go build -o (Split-Path -Leaf $exePath) main.go
    Pop-Location
  }
}

Ensure-GoBackendBuilt 'path_optimization' (Join-Path $rebuildRoot 'path_optimization') (Join-Path $rebuildRoot 'path_optimization\app.exe')
Ensure-GoBackendBuilt 'campus_emergency' (Join-Path $rebuildRoot 'campus_emergency') (Join-Path $rebuildRoot 'campus_emergency\app.exe')
Ensure-GoBackendBuilt 'role_management' (Join-Path $rebuildRoot 'role_management') (Join-Path $rebuildRoot 'role_management\app.exe')

$procs = @{}
$procs.path_optimization = Start-ServiceExe 'path_optimization' (Join-Path $rebuildRoot 'path_optimization') (Join-Path $rebuildRoot 'path_optimization\app.exe') '8080' 'path_optimization.db'
$procs.campus_emergency = Start-ServiceExe 'campus_emergency' (Join-Path $rebuildRoot 'campus_emergency') (Join-Path $rebuildRoot 'campus_emergency\app.exe') '8081' 'campus_emergency.db'
$procs.role_management = Start-ServiceExe 'role_management' (Join-Path $rebuildRoot 'role_management') (Join-Path $rebuildRoot 'role_management\app.exe') '8082' 'role_management.db'
$procs.campus_frontend = Start-Frontend (Join-Path $rebuildRoot 'campus_frontend') '5173'

$payload = [ordered]@{
  started_at = (Get-Date).ToString('s')
  pids = [ordered]@{
    path_optimization = $procs.path_optimization.Id
    campus_emergency = $procs.campus_emergency.Id
    role_management = $procs.role_management.Id
    campus_frontend = $procs.campus_frontend.Id
  }
}

$payload | ConvertTo-Json -Depth 5 | Out-File -Encoding utf8 $pidsPath
Write-Host "[ok] PID file: $pidsPath" -ForegroundColor Green
Write-Host 'Frontend: http://localhost:5173/' -ForegroundColor Green
