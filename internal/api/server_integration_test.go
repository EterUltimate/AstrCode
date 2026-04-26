// AstrCode - Agent orchestration engine for AstrBot
// Copyright (C) 2026 EterUltimate
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestServer_handleGeneratePlugin 测试插件生成端点
func TestServer_handleGeneratePlugin(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
	}{
		{
			name:   "valid request",
			method: http.MethodPost,
			body: map[string]string{
				"requirement": "创建一个天气查询插件",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "empty requirement",
			method: http.MethodPost,
			body: map[string]string{
				"requirement": "",
			},
			expectedStatus: http.StatusOK, // 应该返回占位响应
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if tt.body != nil {
				var err error
				reqBody, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("Failed to marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/generate", bytes.NewReader(reqBody))
			w := httptest.NewRecorder()

			server.handleGeneratePlugin(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// 验证响应是 JSON
			contentType := resp.Header.Get("Content-Type")
			if tt.expectedStatus == http.StatusOK && contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}
		})
	}
}

// TestServer_handleReviewCode 测试代码审查端点
func TestServer_handleReviewCode(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
	}{
		{
			name:   "valid request",
			method: http.MethodPost,
			body: map[string]interface{}{
				"files": map[string]string{
					"main.py": "print('hello')",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "empty files",
			method: http.MethodPost,
			body: map[string]interface{}{
				"files": map[string]string{},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if tt.body != nil {
				var err error
				switch v := tt.body.(type) {
				case string:
					reqBody = []byte(v)
				default:
					reqBody, err = json.Marshal(v)
					if err != nil {
						t.Fatalf("Failed to marshal body: %v", err)
					}
				}
			}

			req := httptest.NewRequest(tt.method, "/api/review", bytes.NewReader(reqBody))
			w := httptest.NewRecorder()

			server.handleReviewCode(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestServer_handleDeployPlugin 测试插件部署端点
func TestServer_handleDeployPlugin(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
	}{
		{
			name:   "valid deployment",
			method: http.MethodPost,
			body: map[string]interface{}{
				"plugin_name": "test-plugin",
				"files": map[string]string{
					"plugin.yaml": "name: test",
					"main.py":     "pass",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "missing plugin name",
			method: http.MethodPost,
			body: map[string]interface{}{
				"plugin_name": "",
				"files":       map[string]string{},
			},
			expectedStatus: http.StatusOK, // 占位响应
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("Failed to marshal body: %v", err)
			}

			req := httptest.NewRequest(tt.method, "/api/deploy", bytes.NewReader(reqBody))
			w := httptest.NewRecorder()

			server.handleDeployPlugin(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// 解析响应体
			var response map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
				if status, ok := response["status"]; ok {
					t.Logf("Response status: %v", status)
				}
			}
		})
	}
}

// TestServer_handleHealth 测试健康检查端点
func TestServer_handleHealth(t *testing.T) {
	server := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	// 验证必要字段
	requiredFields := []string{"status", "version", "features"}
	for _, field := range requiredFields {
		if _, ok := health[field]; !ok {
			t.Errorf("Missing required field: %s", field)
		}
	}

	if health["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", health["status"])
	}

	t.Logf("Health check passed - Version: %v, Features: %d",
		health["version"], len(health["features"].([]interface{})))
}

// TestServer_handleSkills 测试技能列表端点
func TestServer_handleSkills(t *testing.T) {
	t.Skip("Skipping test that requires full Agent initialization")
	server := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/skills", nil)
	w := httptest.NewRecorder()

	server.handleSkills(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var skills map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&skills); err != nil {
		t.Fatalf("Failed to decode skills response: %v", err)
	}

	// 验证响应结构
	if _, ok := skills["stars"]; !ok {
		t.Error("Missing 'stars' field")
	}
	if _, ok := skills["skills"]; !ok {
		t.Error("Missing 'skills' field")
	}
	if _, ok := skills["count"]; !ok {
		t.Error("Missing 'count' field")
	}

	t.Logf("Skills endpoint - Count: %v", skills["count"])
}

// setupTestServer 创建测试服务器
func setupTestServer(t *testing.T) *Server {
	// 创建一个最小化的 Hub 实例
	hub := NewHub()

	// 创建 Server 实例
	return &Server{
		hub: hub,
		// 在实际测试中，这里应该注入 mock 的依赖
	}
}

// Integration Tests - 需要完整初始化的测试

// TestIntegration_FullWorkflow 测试完整工作流（需要真实依赖）
func TestIntegration_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// TODO: 实现完整的集成测试
	// 1. 启动服务器
	// 2. 发送生成请求
	// 3. 发送审查请求
	// 4. 发送部署请求
	// 5. 验证整个流程

	t.Log("Integration test placeholder - requires full server initialization")
}
