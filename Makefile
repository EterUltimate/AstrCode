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

.PHONY: build run test clean docker fmt lint vet deps coverage ci

BINARY=bin/astrcode
VERSION?=0.4.0
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"

# 构建
build:
	go build $(LDFLAGS) -o $(BINARY) cmd/server/main.go

# 运行
run:
	go run cmd/server/main.go -addr :8080 -static-dir ./web

# 测试
test:
	go test -v -race ./...

# 测试覆盖率
coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -func=coverage.txt

# 清理
clean:
	rm -rf bin/ dist/ coverage.txt

# Docker 构建
docker:
	docker build -t astrcode:$(VERSION) .

# 格式化代码
fmt:
	gofmt -w .
	goimports -w . 2>/dev/null || true

# 代码检查
lint:
	go vet ./...
	golangci-lint run --timeout=5m 2>/dev/null || go vet ./...

# 仅 vet
vet:
	go vet ./...

# 检查格式
fmt-check:
	@output=$$(gofmt -l .); if [ -n "$$output" ]; then echo "Files not formatted:"; echo "$$output"; exit 1; fi

# 下载依赖
deps:
	go mod tidy
	go mod download

# CI 完整流程
ci: fmt-check vet test build
	@echo "✅ CI pipeline passed"

# 多平台构建
build-all:
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/astrcode-linux-amd64 cmd/server/main.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/astrcode-linux-arm64 cmd/server/main.go
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/astrcode-windows-amd64.exe cmd/server/main.go
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/astrcode-darwin-amd64 cmd/server/main.go
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/astrcode-darwin-arm64 cmd/server/main.go
	@echo "✅ All platforms built in dist/"
