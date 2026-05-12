$ErrorActionPreference = 'Stop'

$rebuildRoot = Resolve-Path (Join-Path $PSScriptRoot '..')
$runDir = Join-Path $rebuildRoot '.run'
$pidsPath = Join-Path $runDir 'pids.json'

if (-not (Test-Path $pidsPath)) {
  Write-Host "[warn] PID file not found: $pidsPath" -ForegroundColor Yellow
  exit 0
}

$p = Get-Content $pidsPath -Raw | ConvertFrom-Json
$targets = @(
  @{ name = 'campus_frontend'; pid = [int]$p.pids.campus_frontend },
  @{ name = 'role_management'; pid = [int]$p.pids.role_management },
  @{ name = 'campus_emergency'; pid = [int]$p.pids.campus_emergency },
  @{ name = 'path_optimization'; pid = [int]$p.pids.path_optimization }
)

foreach ($t in $targets) {
  if ($t.pid -le 0) { continue }
  try {
    $proc = Get-Process -Id $t.pid -ErrorAction Stop
    Write-Host "[stop] $($t.name) pid=$($t.pid)" -ForegroundColor Cyan
    Stop-Process -Id $t.pid -Force
  } catch {
    Write-Host "[info] $($t.name) pid=$($t.pid) already stopped" -ForegroundColor DarkGray
  }
}

Remove-Item $pidsPath -Force
Write-Host '[ok] stopped' -ForegroundColor Green
