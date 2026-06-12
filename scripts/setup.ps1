param(
    [string]$SdkRoot = $env:ANDROID_HOME
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $ProjectRoot

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go not found. Install from https://go.dev/dl/"
}

if (-not $SdkRoot) {
    $SdkRoot = $env:ANDROID_SDK_ROOT
}
if ($SdkRoot) {
    $env:PATH = "$SdkRoot\platform-tools;$SdkRoot\emulator;$env:PATH"
}

go mod tidy
New-Item -ItemType Directory -Force -Path bin, assets\qr, assets\clones | Out-Null
go build -o bin\riotscanner.exe .\cmd\riotscanner

Write-Host "Build complete: bin\riotscanner.exe"
