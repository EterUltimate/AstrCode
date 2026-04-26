@echo off
REM AstrCode - Agent orchestration engine for AstrBot
REM Copyright (C) 2026 EterUltimate
REM
REM This program is free software: you can redistribute it and/or modify
REM it under the terms of the GNU Affero General Public License as published by
REM the Free Software Foundation, either version 3 of the License, or
REM (at your option) any later version.
REM
REM This program is distributed in the hope that it will be useful,
REM but WITHOUT ANY WARRANTY; without even the implied warranty of
REM MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
REM GNU Affero General Public License for more details.
REM
REM You should have received a copy of the GNU Affero General Public License
REM along with this program. If not, see <https://www.gnu.org/licenses/>.

echo ========================================
echo   Starting AstrCode Server...
echo ========================================
echo.

cd /d "%~dp0.."

REM Check if astrcode.exe exists
if not exist "bin\astrcode.exe" (
    echo ERROR: astrcode.exe not found!
    echo Please reinstall AstrCode.
    pause
    exit /b 1
)

REM Start the server
echo Starting AstrCode on port 8080...
echo Dashboard: http://localhost:8080
echo Press Ctrl+C to stop the server
echo.

bin\astrcode.exe -addr :8080 -static-dir web

pause
