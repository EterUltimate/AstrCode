#!/bin/bash

set -e

echo "Building AstrCode..."

# 创建输出目录
mkdir -p bin

# 构建 Linux 版本
GOOS=linux GOARCH=amd64 go build -o bin/astrcode-linux cmd/server/main.go

# 构建 Windows 版本
GOOS=windows GOARCH=amd64 go build -o bin/astrcode.exe cmd/server/main.go

# 构建 macOS 版本
GOOS=darwin GOARCH=amd64 go build -o bin/astrcode-darwin cmd/server/main.go

echo "Build complete!"
echo "  - bin/astrcode-linux"
echo "  - bin/astrcode.exe"
echo "  - bin/astrcode-darwin"
