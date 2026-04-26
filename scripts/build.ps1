# AstrCode Build Script for Windows

Write-Host "Building AstrCode..." -ForegroundColor Green

# Create output directory
New-Item -ItemType Directory -Force -Path "bin" | Out-Null

# Build Windows version
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o bin/astrcode.exe cmd/server/main.go

Write-Host "Build complete!" -ForegroundColor Green
Write-Host "  - bin/astrcode.exe"

# Optional: Build Linux version (requires cross-compilation setup)
# $env:GOOS = "linux"
# $env:GOARCH = "amd64"
# go build -o bin/astrcode-linux cmd/server/main.go
