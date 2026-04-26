# AstrCode - Agent orchestration engine for AstrBot
# Copyright (C) 2026 EterUltimate
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program. If not, see <https://www.gnu.org/licenses/>.

# Test MSI build prerequisites

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  AstrCode MSI Build Prerequisites Check" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$allPassed = $true

# Check 1: Go installation
Write-Host "[1/4] Checking Go installation..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✓ Go installed: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Go not found! Please install Go 1.21+" -ForegroundColor Red
    $allPassed = $false
}
Write-Host ""

# Check 2: WiX Toolset
Write-Host "[2/4] Checking WiX Toolset..." -ForegroundColor Yellow
$wixFound = $false
$wixPaths = @(
    "C:\Program Files (x86)\WiX Toolset v3.11\bin",
    "C:\Program Files\WiX Toolset v3.11\bin",
    "C:\Program Files (x86)\WiX Toolset v3.14\bin",
    "C:\Program Files\WiX Toolset v3.14\bin"
)

foreach ($path in $wixPaths) {
    if (Test-Path $path) {
        Write-Host "✓ WiX Toolset found at: $path" -ForegroundColor Green
        $wixFound = $true
        
        # Check specific tools
        $candle = Join-Path $path "candle.exe"
        $light = Join-Path $path "light.exe"
        
        if (Test-Path $candle) {
            Write-Host "  ✓ candle.exe exists" -ForegroundColor Green
        } else {
            Write-Host "  ✗ candle.exe missing" -ForegroundColor Red
            $allPassed = $false
        }
        
        if (Test-Path $light) {
            Write-Host "  ✓ light.exe exists" -ForegroundColor Green
        } else {
            Write-Host "  ✗ light.exe missing" -ForegroundColor Red
            $allPassed = $false
        }
        
        break
    }
}

if (-not $wixFound) {
    Write-Host "✗ WiX Toolset not found!" -ForegroundColor Red
    Write-Host "  Download from: https://wixtoolset.org/releases/" -ForegroundColor Yellow
    $allPassed = $false
}
Write-Host ""

# Check 3: Source files
Write-Host "[3/4] Checking source files..." -ForegroundColor Yellow
$requiredFiles = @(
    "cmd\server\main.go",
    "web\index.html",
    "configs\config.yaml",
    "installer\astrcode.wxs",
    "installer\scripts\start.bat",
    "installer\scripts\stop.bat"
)

foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "  ✓ $file" -ForegroundColor Green
    } else {
        Write-Host "  ✗ $file (missing)" -ForegroundColor Red
        $allPassed = $false
    }
}
Write-Host ""

# Check 4: Directory structure
Write-Host "[4/4] Checking directory structure..." -ForegroundColor Yellow
$requiredDirs = @("bin", "dist", "installer")

foreach ($dir in $requiredDirs) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir | Out-Null
        Write-Host "  ✓ Created $dir\" -ForegroundColor Yellow
    } else {
        Write-Host "  ✓ $dir\" -ForegroundColor Green
    }
}
Write-Host ""

# Summary
Write-Host "========================================" -ForegroundColor Cyan
if ($allPassed) {
    Write-Host "  ✅ All checks passed!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "You can now build the MSI package:" -ForegroundColor White
    Write-Host "  .\scripts\build-msi.ps1" -ForegroundColor Cyan
    Write-Host ""
} else {
    Write-Host "  ❌ Some checks failed!" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Please fix the issues above before building." -ForegroundColor Yellow
    Write-Host ""
    exit 1
}
