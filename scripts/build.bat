@echo off
setlocal

where go >nul 2>&1
if errorlevel 1 (
    echo Go is not installed. Download from https://go.dev/dl/
    exit /b 1
)

cd /d "%~dp0\.."
go mod tidy
go build -o bin\riotscanner.exe .\cmd\riotscanner

if errorlevel 1 (
    echo Build failed.
    exit /b 1
)

echo Built: bin\riotscanner.exe
endlocal
