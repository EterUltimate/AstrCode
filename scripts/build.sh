#!/bin/bash
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
