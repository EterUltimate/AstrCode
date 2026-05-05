# AstrCode API 测试脚本
# 使用方法: .\test-api.ps1

$BASE_URL = "http://localhost:8080"

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  🧪 AstrCode API 测试" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# 测试 1: 健康检查
Write-Host "1️⃣  测试健康检查..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/api/health" -Method Get
    Write-Host "   ✅ 状态: $($response.status)" -ForegroundColor Green
    Write-Host "   📦 版本: $($response.version)" -ForegroundColor Gray
    Write-Host "   🔧 特性数: $($response.features.Count)" -ForegroundColor Gray
} catch {
    Write-Host "   ❌ 失败: $_" -ForegroundColor Red
}
Write-Host ""

# 测试 2: 插件生成
Write-Host "2️⃣  测试插件生成..." -ForegroundColor Yellow
$generateBody = @{
    requirement = "创建一个天气查询插件，支持城市名称查询当前天气"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/api/generate" -Method Post -Body $generateBody -ContentType "application/json"
    Write-Host "   ✅ 状态: $($response.status)" -ForegroundColor Green
    Write-Host "   💬 消息: $($response.message)" -ForegroundColor Gray
    Write-Host "   📝 需求: $($response.requirement)" -ForegroundColor Gray
} catch {
    Write-Host "   ❌ 失败: $_" -ForegroundColor Red
}
Write-Host ""

# 测试 3: 代码审查
Write-Host "3️⃣  测试代码审查..." -ForegroundColor Yellow
$reviewBody = @{
    files = @{
        "main.py" = @"
import asyncio

class WeatherPlugin:
    def __init__(self):
        self.api_key = "sk-test123"
    
    async def handle_weather(self, event):
        city = event.message
        # TODO: 调用天气 API
        return f"Weather in {city}"
"@
    }
} | ConvertTo-Json -Depth 10

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/api/review" -Method Post -Body $reviewBody -ContentType "application/json"
    Write-Host "   ✅ 状态: $($response.status)" -ForegroundColor Green
    Write-Host "   💬 消息: $($response.message)" -ForegroundColor Gray
    Write-Host "   📄 文件数: $($response.files_count)" -ForegroundColor Gray
} catch {
    Write-Host "   ❌ 失败: $_" -ForegroundColor Red
}
Write-Host ""

# 测试 4: 插件部署
Write-Host "4️⃣  测试插件部署..." -ForegroundColor Yellow
$deployBody = @{
    plugin_name = "weather-plugin"
    files = @{
        "plugin.yaml" = @"
name: weather-plugin
version: "1.0.0"
description: 天气查询插件
author: Test User
"@
        "main.py" = @"
class WeatherPlugin:
    pass
"@
    }
} | ConvertTo-Json -Depth 10

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/api/deploy" -Method Post -Body $deployBody -ContentType "application/json"
    Write-Host "   ✅ 状态: $($response.status)" -ForegroundColor Green
    Write-Host "   💬 消息: $($response.message)" -ForegroundColor Gray
    Write-Host "   🔌 插件名: $($response.plugin_name)" -ForegroundColor Gray
} catch {
    Write-Host "   ❌ 失败: $_" -ForegroundColor Red
}
Write-Host ""

# 测试 5: 技能列表
Write-Host "5️⃣  测试技能列表..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/api/skills" -Method Get
    Write-Host "   ✅ 状态: ok" -ForegroundColor Green
    Write-Host "   ⭐ Stars: $($response.stars.Count)" -ForegroundColor Gray
    Write-Host "   🎯 Skills: $($response.skills.Count)" -ForegroundColor Gray
    Write-Host "   📊 总数: $($response.count)" -ForegroundColor Gray
} catch {
    Write-Host "   ❌ 失败: $_" -ForegroundColor Red
}
Write-Host ""

# 测试 6: WebSocket 连接测试
Write-Host "6️⃣  测试 WebSocket 端点..." -ForegroundColor Yellow
try {
    $wsResponse = Invoke-WebRequest -Uri "$BASE_URL/ws" -Method Get -ErrorAction Stop
    Write-Host "   ⚠️  WebSocket 需要客户端连接" -ForegroundColor Yellow
} catch {
    if ($_.Exception.Response.StatusCode -eq "BadRequest") {
        Write-Host "   ✅ WebSocket 端点可用 (需要 Upgrade 头)" -ForegroundColor Green
    } else {
        Write-Host "   ℹ️  WebSocket 端点: $BASE_URL/ws" -ForegroundColor Gray
    }
}
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  ✨ API 测试完成！" -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Cyan

Write-Host "💡 提示:" -ForegroundColor Yellow
Write-Host "  • 浏览器访问: http://localhost:8080" -ForegroundColor Gray
Write-Host "  • 查看文档: 阅读 README.md" -ForegroundColor Gray
Write-Host "  • 停止服务器: 按 Ctrl+C`n" -ForegroundColor Gray
