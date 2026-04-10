$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$buildDir = Join-Path $repoRoot "build"
$binaryPath = Join-Path $buildDir "codexass.exe"

New-Item -ItemType Directory -Force -Path $buildDir | Out-Null

go build -o $binaryPath ./cmd/codexass

Write-Host "Built: $binaryPath"
