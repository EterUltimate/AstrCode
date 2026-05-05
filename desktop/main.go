//
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
//

package main

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// App struct
type App struct {
	ctx            context.Context
	runtimeProcess *exec.Cmd
	runtimeOutput  []string
	outputMutex    sync.Mutex
	running        bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetHealth returns the system health status
func (a *App) GetHealth() string {
	return `{
		"status": "ok",
		"version": "0.4.0",
		"features": 15,
		"skills_loaded": 3,
		"mcp_tools": 4,
		"docs_available": 22,
		"hooks_documented": 12,
		"uptime": "running",
		"astrbot_skill_version": "1.0.0"
	}`
}

// GetSkills returns the available skills based on AstrBot-Skill repository
func (a *App) GetSkills() string {
	return `{
		"skills": [
			{
				"name": "skill-astrbot-dev",
				"version": "1.0.0",
				"description": "AstrBot plugin development reference skill. Covers message models, platform adapters, plugin config, agent system, hooks, and decorators.",
				"tags": ["plugin-dev", "reference", "core"],
				"category": "development",
				"author": "AstrBot Community",
				"docs_path": "skill-astrbot-dev/references/"
			},
			{
				"name": "docs4agent",
				"version": "1.0.0",
				"description": "Technical documentation writing standards for AI-friendly docs. Guidelines for creating structured, searchable, and agent-compatible documentation.",
				"tags": ["documentation", "standards", "ai-friendly"],
				"category": "documentation",
				"author": "AstrBot Community",
				"docs_path": "data/skills/docs4agent/"
			},
			{
				"name": "mcp-astrbot",
				"version": "1.0.0",
				"description": "MCP server for AstrBot plugin development. Provides get_doc, list_docs, search_docs, and get_hook_inventory tools via Model Context Protocol.",
				"tags": ["mcp", "server", "tools"],
				"category": "integration",
				"author": "AstrBot Community",
				"docs_path": "mcp-astrbot/"
			}
		],
		"count": 3,
		"categories": ["development", "documentation", "integration"]
	}`
}

// GenerateCode handles code generation requests
func (a *App) GenerateCode(requirement string) string {
	return `{"status":"ready","message":"Generate endpoint ready"}`
}

// ReviewCode handles code review requests
func (a *App) ReviewCode(files string) string {
	return `{"status":"ready","message":"Review endpoint ready"}`
}

// DeployPlugin handles plugin deployment requests
func (a *App) DeployPlugin(pluginName, description string) string {
	return `{"status":"ready","message":"Deploy endpoint ready"}`
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return "0.4.0"
}

// MCP Tool Methods - Based on AstrBot-Skill MCP Server

// MCPListDocs returns the list of available documentation categories
func (a *App) MCPListDocs(category string) string {
	docsIndex := map[string]map[string]string{
		"agent": {
			"agent-runner":        "Agent Runner overview and configuration",
			"agent-related-hooks": "Agent related Hooks",
			"conversation":        "Conversation management",
			"context-compression": "Context compression",
			"cron":                "Scheduled tasks",
			"Invoke-llm":          "Invoke LLM",
			"persona-resolution":  "Persona resolution",
			"persona-sets":        "Persona sets",
			"registe tools":       "Tool registration",
			"sandbox":             "Sandbox environment",
			"subagents":           "Sub-agents",
		},
		"design_standards": {
			"architecture_overview": "Architecture overview",
			"best_practices":        "Best practices",
			"context_usage":         "Context usage",
			"core_concepts":         "Core concepts",
			"event_flow":            "Event flow",
			"sandbox":               "Sandbox design",
			"visual_utils":          "Visual utilities",
		},
		"messages": {
			"components": "Message components",
			"events":     "Message events",
			"model":      "Message model",
			"umo":        "UMO model",
		},
		"plugin_config": {
			"command_management": "Command management",
			"decorators":         "Decorators",
			"file_config":        "File configuration",
			"hooks":              "Hooks",
			"lifecycle":          "Lifecycle",
			"schema":             "Configuration Schema",
			"session_control":    "Session control",
		},
		"platform_adapters": {
			"adapter_interface":    "Adapter interface",
			"message_conversion":   "Message conversion",
			"telegram_media_group": "Telegram media group",
		},
	}

	if category != "" {
		docs, ok := docsIndex[category]
		if !ok {
			return `{"error":"Unknown category","available":["agent","design_standards","messages","plugin_config","platform_adapters"]}`
		}
		result := "## " + category + " Category Docs\n\n"
		for docID, desc := range docs {
			result += "- `" + docID + "`: " + desc + "\n"
		}
		return `{"category":"` + category + `","docs_count":` + fmt.Sprintf("%d", len(docs)) + `,"content":"` + result + `"}`
	}

	result := "# AstrBot Development Documentation Index\n\n"
	for cat, docs := range docsIndex {
		result += "## " + cat + "\n\n"
		for docID, desc := range docs {
			result += "- `" + docID + "`: " + desc + "\n"
		}
		result += "\n"
	}
	return `{"categories":5,"total_docs":22,"content":"` + result + `"}`
}

// MCPGetDoc returns a specific document content (simulated)
func (a *App) MCPGetDoc(category, docName string) string {
	docContents := map[string]map[string]string{
		"plugin_config": {
			"hooks": `# Event Hooks (Hooks)

Event hooks are used to intervene at key points in AstrBot's core execution flow.

## Core Hooks List

### 1) Lifecycle
- @filter.on_astrbot_loaded() - Triggered when AstrBot loading completes
- @filter.on_platform_loaded() - Triggered when platform loading completes

### 2) Before/After LLM Request
- @filter.on_waiting_llm_request() - Before acquiring session lock
- @filter.on_llm_request() - Before sending LLM request (can modify request)
- @filter.on_llm_response() - After LLM request completes

### 3) Before/After Tool Call
- @filter.on_using_llm_tool() - Before function tool call
- @filter.on_llm_tool_respond() - After function tool call

### 4) Before/After Sending Message
- @filter.on_decorating_result() - Before sending message
- @filter.after_message_sent() - After message sent to platform`,
			"decorators": `# Decorators

## Plugin Registration
- @register(id, author, description, version, repo_url): Register plugin

## Command Registration
- @filter.command(name, alias, priority): Register command
- @filter.command_group(name): Register command group

## Message Filtering
- @filter.event_message_type(type): Filter message types
- @filter.platform_adapter_type(type): Filter platform types
- @filter.permission_type(type): Check permissions
- @filter.regex(pattern): Regex matching

## Tool Registration
- @filter.llm_tool(name): Register as AI-callable tool
- @session_waiter(timeout, record_history_chains): Wait for next message`,
			"lifecycle": `# Plugin Lifecycle

## Startup Phase
1. Plugin class instantiation
2. __init__ initialization
3. on_astrbot_loaded hook

## Runtime Phase
4. Event listening and handling
5. Command execution
6. Tool calls

## Shutdown Phase
7. Cleanup and resource release`,
		},
		"design_standards": {
			"core_concepts": `# Core Concepts API List

## 1. Decorators
- @register(): Plugin registration
- @filter.command(): Command registration
- @filter.llm_tool(): AI tool registration

## 2. Message Components
- Plain(text): Plain text
- Image.fromFileSystem(path): Image
- Record.fromFileSystem(path): Voice
- Video.fromFileSystem(path): Video

## 3. Core Objects
- AstrMessageEvent: Message event object
- Context: Core hub with send_message, get_platform, add_llm_tools

## 4. Storage
- await self.get_kv_data(key, default)
- await self.put_kv_data(key, value)
- await self.html_render(html_text)

## 5. Agent
- context.tool_loop_agent(): Tool loop agent
- BaseAgentRunHooks: Agent hook interface`,
		},
		"agent": {
			"agent-runner": "# Agent Runner (v4.7.0+)\n\n## Overview\nAgent Runner is AstrBot's intelligent agent execution framework.\n\n## Components\n- Tools: Function tools for agent\n- Providers: LLM providers\n- Personas: Character configurations\n- Sandbox: Isolated execution environment\n- Subagents: Child intelligent agents\n\n## Usage\nUse FunctionTool dataclass to define tools and register with context.add_llm_tools()",
		},
	}

	catDocs, ok := docContents[category]
	if !ok {
		return `{"error":"Category not found","category":"` + category + `"}`
	}
	content, ok := catDocs[docName]
	if !ok {
		return `{"error":"Document not found","category":"` + category + `","doc":"` + docName + `"}`
	}
	return `{"category":"` + category + `","doc":"` + docName + `","content":"` + content + `"}`
}

// MCPSearchDocs searches documentation by keyword
func (a *App) MCPSearchDocs(query string) string {
	// Simulated search results based on AstrBot-Skill docs
	results := []map[string]string{}

	if query == "hook" || query == "hooks" {
		results = append(results, map[string]string{
			"category": "plugin_config",
			"doc":      "hooks",
			"snippet":  "Event hooks are used to intervene at key points... @filter.on_astrbot_loaded(), @filter.on_llm_request()...",
		})
	}
	if query == "decorator" || query == "decorators" {
		results = append(results, map[string]string{
			"category": "plugin_config",
			"doc":      "decorators",
			"snippet":  "@register(), @filter.command(), @filter.llm_tool()...",
		})
	}
	if query == "agent" {
		results = append(results, map[string]string{
			"category": "agent",
			"doc":      "agent-runner",
			"snippet":  "Agent Runner is AstrBot's intelligent agent execution framework...",
		})
	}
	if query == "message" {
		results = append(results, map[string]string{
			"category": "messages",
			"doc":      "model",
			"snippet":  "Message model and event flow...",
		})
	}

	if len(results) == 0 {
		return `{"query":"` + query + `","results":[],"message":"No matching documents found. Try: hooks, decorators, agent, message"}`
	}

	resultJSON, _ := json.Marshal(map[string]interface{}{
		"query":   query,
		"results": results,
		"count":   len(results),
	})
	return string(resultJSON)
}

// MCPGetHookInventory returns the complete hook inventory
func (a *App) MCPGetHookInventory() string {
	return `{
		"hooks": {
			"lifecycle": [
				{"name": "on_astrbot_loaded", "description": "Triggered when AstrBot loading completes", "type": "async def"},
				{"name": "on_platform_loaded", "description": "Triggered when platform loading completes", "type": "async def"}
			],
			"llm_request": [
				{"name": "on_waiting_llm_request", "description": "Before acquiring session lock", "type": "async def"},
				{"name": "on_llm_request", "description": "Before sending LLM request", "type": "async def"},
				{"name": "on_llm_response", "description": "After LLM request completes", "type": "async def"}
			],
			"tool_call": [
				{"name": "on_using_llm_tool", "description": "Before function tool call", "type": "async def"},
				{"name": "on_llm_tool_respond", "description": "After function tool call", "type": "async def"}
			],
			"message": [
				{"name": "on_decorating_result", "description": "Before sending message", "type": "async def"},
				{"name": "after_message_sent", "description": "After message sent to platform", "type": "async def"}
			]
		},
		"total": 8,
		"note": "For complete hook inventory, run: python scripts/generate_hook_inventory.py"
	}`
}

// RuntimeStart starts the AstrBot runtime
func (a *App) RuntimeStart(astrbotPath string) string {
	if a.running {
		return `{"success": false, "error": "Runtime is already running"}`
	}

	cmd := exec.Command("python", "main.py")
	cmd.Dir = astrbotPath

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Sprintf(`{"success": false, "error": "%s"}`, err.Error())
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Sprintf(`{"success": false, "error": "%s"}`, err.Error())
	}

	if err := cmd.Start(); err != nil {
		return fmt.Sprintf(`{"success": false, "error": "%s"}`, err.Error())
	}

	a.runtimeProcess = cmd
	a.running = true
	a.runtimeOutput = []string{}

	// Start goroutines to read output
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			a.outputMutex.Lock()
			a.runtimeOutput = append(a.runtimeOutput, line)
			// Keep only last 1000 lines
			if len(a.runtimeOutput) > 1000 {
				a.runtimeOutput = a.runtimeOutput[1:]
			}
			a.outputMutex.Unlock()
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := "[ERROR] " + scanner.Text()
			a.outputMutex.Lock()
			a.runtimeOutput = append(a.runtimeOutput, line)
			if len(a.runtimeOutput) > 1000 {
				a.runtimeOutput = a.runtimeOutput[1:]
			}
			a.outputMutex.Unlock()
		}
	}()

	// Wait for process in background
	go func() {
		cmd.Wait()
		a.running = false
	}()

	return `{"success": true, "message": "AstrBot runtime started"}`
}

// RuntimeStop stops the AstrBot runtime
func (a *App) RuntimeStop() string {
	if !a.running || a.runtimeProcess == nil {
		return `{"success": false, "error": "Runtime is not running"}`
	}

	if err := a.runtimeProcess.Process.Kill(); err != nil {
		return fmt.Sprintf(`{"success": false, "error": "%s"}`, err.Error())
	}

	a.running = false
	return `{"success": true, "message": "AstrBot runtime stopped"}`
}

// RuntimeGetStatus returns the runtime status
func (a *App) RuntimeGetStatus() string {
	status := map[string]interface{}{
		"running": a.running,
		"url":     "http://localhost:6185",
	}

	if a.running {
		status["pid"] = a.runtimeProcess.Process.Pid
	}

	data, _ := json.Marshal(status)
	return string(data)
}

// RuntimeGetOutput returns the recent output from the runtime
func (a *App) RuntimeGetOutput() string {
	a.outputMutex.Lock()
	defer a.outputMutex.Unlock()

	output := strings.Join(a.runtimeOutput, "\n")
	return output
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:             "AstrCode",
		Width:             1280,
		Height:            800,
		MinWidth:          800,
		MinHeight:         600,
		MaxWidth:          1920,
		MaxHeight:         1080,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 10, G: 10, B: 10, A: 1},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			Theme:                windows.Dark,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
