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

# AstrCode MSI Package Builder
# Requires WiX Toolset 3.11+ installed

param(
    [string]$Version = "0.4.0",
    [string]$OutputDir = "dist"
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  AstrCode MSI Package Builder" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check WiX installation
$wixPath = $null
if (Test-Path "C:\Program Files (x86)\WiX Toolset v3.11\bin") {
    $wixPath = "C:\Program Files (x86)\WiX Toolset v3.11\bin"
} elseif (Test-Path "C:\Program Files\WiX Toolset v3.11\bin") {
    $wixPath = "C:\Program Files\WiX Toolset v3.11\bin"
} else {
    Write-Host "ERROR: WiX Toolset 3.11 not found!" -ForegroundColor Red
    Write-Host "Please install from: https://wixtoolset.org/releases/" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ WiX Toolset found at: $wixPath" -ForegroundColor Green
Write-Host ""

# Set paths
$candleExe = Join-Path $wixPath "candle.exe"
$lightExe = Join-Path $wixPath "light.exe"
$heatExe = Join-Path $wixPath "heat.exe"

# Create output directory
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

# Step 1: Build the binary
Write-Host "[1/5] Building AstrCode binary..." -ForegroundColor Yellow
go build -ldflags="-s -w -X main.version=$Version" -o bin\astrcode.exe cmd/server/main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Binary built successfully" -ForegroundColor Green
Write-Host ""

# Step 2: Generate RTF license file
Write-Host "[2/5] Converting LICENSE to RTF format..." -ForegroundColor Yellow
$licenseRtf = @"
{\rtf1\ansi\ansicpg1252\deff0\nouicompat\deflang1033{\fonttbl{\f0\fnil\fcharset0 Calibri;}}
{\*\generator Riched20 10.0.19041}\viewkind4\uc1 
\pard\sa200\sl276\slmult1\lang9\b\f0\fs28\lang2 GNU AFFERO GENERAL PUBLIC LICENSE\par
\pard\sa200\sl276\slmult1 Version 3, 19 November 2007\par
\pard\sa200\sl276\slmult1\par
Copyright (C) 2007 Free Software Foundation, Inc. <https://fsf.org/>\par
Everyone is permitted to copy and distribute verbatim copies of this license document, but changing it is not allowed.\par
\pard\sa200\sl276\slmult1\par
This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.\par
\pard\sa200\sl276\slmult1\par
This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License for more details.\par
\pard\sa200\sl276\slmult1\par
You should have received a copy of the GNU Affero General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.\par
}
"@
New-Item -ItemType Directory -Force -Path "installer\license" | Out-Null
Set-Content -Path "installer\license\LICENSE.rtf" -Value $licenseRtf -Encoding UTF8
Write-Host "✓ License converted" -ForegroundColor Green
Write-Host ""

# Step 3: Compile WiX source to object files
Write-Host "[3/5] Compiling WiX source..." -ForegroundColor Yellow
& $candleExe -dProductVersion=$Version -ext WixFirewallExtension installer\astrcode.wxs -out installer\astrcode.wixobj
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Candle compilation failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ WiX source compiled" -ForegroundColor Green
Write-Host ""

# Step 4: Link object files to create MSI
Write-Host "[4/5] Creating MSI package..." -ForegroundColor Yellow
$msiName = "AstrCode-$Version-x64.msi"
& $lightExe -ext WixUIExtension -ext WixFirewallExtension installer\astrcode.wixobj -out "$OutputDir\$msiName"
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Light linking failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ MSI created: $OutputDir\$msiName" -ForegroundColor Green
Write-Host ""

# Step 5: Generate checksums
Write-Host "[5/5] Generating checksums..." -ForegroundColor Yellow
cd $OutputDir
$hash = Get-FileHash $msiName -Algorithm SHA256
"$($hash.Hash)  $msiName" | Out-File -FilePath "SHA256SUMS.txt" -Encoding UTF8
Write-Host "✓ Checksum generated" -ForegroundColor Green
Write-Host ""

# Summary
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Build Complete!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "MSI Package: $OutputDir\$msiName" -ForegroundColor Green
Write-Host "Size: $([math]::Round((Get-Item "$OutputDir\$msiName").Length / 1MB, 2)) MB" -ForegroundColor Green
Write-Host "SHA256: $($hash.Hash.Substring(0, 16))..." -ForegroundColor Green
Write-Host ""
Write-Host "Installation instructions:" -ForegroundColor Yellow
Write-Host "  1. Double-click the MSI file" -ForegroundColor White
Write-Host "  2. Follow the installation wizard" -ForegroundColor White
Write-Host "  3. Launch from Start Menu or run start.bat" -ForegroundColor White
Write-Host "  4. Open http://localhost:8080 in browser" -ForegroundColor White
Write-Host ""
