@echo off
REM AstrCode LLM 配置测试脚本
REM 用于快速测试不同的 LLM 提供商配置

echo ====================================
echo AstrCode LLM Configuration Test
echo ====================================
echo.

REM 检查二进制文件是否存在
if not exist "bin\astrcode-server.exe" (
    echo Error: bin\astrcode-server.exe not found
    echo Please run: go build -o bin/astrcode-server.exe ./cmd/server
    pause
    exit /b 1
)

echo Available LLM Providers:
echo   1. OpenAI (openai)
echo   2. Google Gemini (gemini)
echo   3. Anthropic Claude (claude)
echo.

set /p PROVIDER="Enter provider (openai/gemini/claude): "
set /p BASE_URL="Enter base URL: "
set /p API_KEY="Enter API Key: "
set /p MODEL="Enter model name: "

echo.
echo Starting AstrCode server with configuration:
echo   Provider: %PROVIDER%
echo   Base URL: %BASE_URL%
echo   Model: %MODEL%
echo.
echo Press Ctrl+C to stop the server
echo.

bin\astrcode-server.exe ^
  --llm-provider %PROVIDER% ^
  --llm-url %BASE_URL% ^
  --llm-key %API_KEY% ^
  --llm-model %MODEL%

pause
