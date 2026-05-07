package extension

import (
	"context"
	"fmt"
	"testing"
)

func TestRegistry_RegisterCustomTool(t *testing.T) {
	reg := NewRegistry()

	tool := &CustomTool{
		Name:        "test_tool",
		Description: "A test tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "result", nil
		},
	}

	err := reg.RegisterCustomTool(tool)
	if err != nil {
		t.Fatalf("RegisterCustomTool failed: %v", err)
	}

	// 验证工具已注册
	retrieved, exists := reg.GetCustomTool("test_tool")
	if !exists {
		t.Fatal("Tool should exist")
	}

	if retrieved.Name != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", retrieved.Name)
	}
}

func TestRegistry_RegisterCustomTool_Duplicate(t *testing.T) {
	reg := NewRegistry()

	tool1 := &CustomTool{Name: "tool1"}
	tool2 := &CustomTool{Name: "tool1"}

	reg.RegisterCustomTool(tool1)
	err := reg.RegisterCustomTool(tool2)

	if err == nil {
		t.Error("Expected error for duplicate registration")
	}
}

func TestRegistry_ListCustomTools(t *testing.T) {
	reg := NewRegistry()

	reg.RegisterCustomTool(&CustomTool{Name: "tool1"})
	reg.RegisterCustomTool(&CustomTool{Name: "tool2"})
	reg.RegisterCustomTool(&CustomTool{Name: "tool3"})

	tools := reg.ListCustomTools()
	if len(tools) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(tools))
	}
}

func TestRegistry_UnregisterCustomTool(t *testing.T) {
	reg := NewRegistry()

	reg.RegisterCustomTool(&CustomTool{Name: "tool1"})

	err := reg.UnregisterCustomTool("tool1")
	if err != nil {
		t.Fatalf("UnregisterCustomTool failed: %v", err)
	}

	_, exists := reg.GetCustomTool("tool1")
	if exists {
		t.Error("Tool should not exist after unregister")
	}
}

func TestRegistry_RegisterSlashCommand(t *testing.T) {
	reg := NewRegistry()

	cmd := &SlashCommand{
		Command:     "/help",
		Description: "Show help",
		Handler: func(ctx context.Context, args []string) (string, error) {
			return "Help content", nil
		},
	}

	err := reg.RegisterSlashCommand(cmd)
	if err != nil {
		t.Fatalf("RegisterSlashCommand failed: %v", err)
	}

	// 验证命令已注册（应该自动添加 / 前缀）
	retrieved, exists := reg.GetSlashCommand("/help")
	if !exists {
		t.Fatal("Command should exist")
	}

	if retrieved.Command != "/help" {
		t.Errorf("Expected command '/help', got '%s'", retrieved.Command)
	}
}

func TestRegistry_RegisterSlashCommand_AutoPrefix(t *testing.T) {
	reg := NewRegistry()

	cmd := &SlashCommand{Command: "test"}
	reg.RegisterSlashCommand(cmd)

	// 验证自动添加了 / 前缀
	_, exists := reg.GetSlashCommand("/test")
	if !exists {
		t.Error("Command should have auto-added / prefix")
	}
}

func TestRegistry_ListSlashCommands(t *testing.T) {
	reg := NewRegistry()

	reg.RegisterSlashCommand(&SlashCommand{Command: "/cmd1"})
	reg.RegisterSlashCommand(&SlashCommand{Command: "/cmd2"})
	reg.RegisterSlashCommand(&SlashCommand{Command: "/cmd3"})

	commands := reg.ListSlashCommands()
	if len(commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(commands))
	}
}

func TestRegistry_RegisterMiddleware(t *testing.T) {
	reg := NewRegistry()

	mw1 := &Middleware{Name: "mw1", Priority: 2}
	mw2 := &Middleware{Name: "mw2", Priority: 1}
	mw3 := &Middleware{Name: "mw3", Priority: 3}

	reg.RegisterMiddleware(mw1)
	reg.RegisterMiddleware(mw2)
	reg.RegisterMiddleware(mw3)

	middlewares := reg.GetMiddlewares()

	// 验证按优先级排序
	if middlewares[0].Name != "mw2" {
		t.Errorf("Expected first middleware 'mw2' (priority 1), got '%s'", middlewares[0].Name)
	}

	if middlewares[1].Name != "mw1" {
		t.Errorf("Expected second middleware 'mw1' (priority 2), got '%s'", middlewares[1].Name)
	}

	if middlewares[2].Name != "mw3" {
		t.Errorf("Expected third middleware 'mw3' (priority 3), got '%s'", middlewares[2].Name)
	}
}

func TestRegistry_ExecuteMiddlewareChain(t *testing.T) {
	reg := NewRegistry()

	// 注册中间件链
	reg.RegisterMiddleware(&Middleware{
		Name: "uppercase",
		Handler: func(ctx context.Context, data interface{}) (interface{}, error) {
			if str, ok := data.(string); ok {
				return str + " (uppercased)", nil
			}
			return data, nil
		},
		Priority: 1,
	})

	reg.RegisterMiddleware(&Middleware{
		Name: "add_timestamp",
		Handler: func(ctx context.Context, data interface{}) (interface{}, error) {
			if str, ok := data.(string); ok {
				return str + " [timestamp]", nil
			}
			return data, nil
		},
		Priority: 2,
	})

	result, err := reg.ExecuteMiddlewareChain(context.Background(), "hello")
	if err != nil {
		t.Fatalf("ExecuteMiddlewareChain failed: %v", err)
	}

	expected := "hello (uppercased) [timestamp]"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestRegistry_GetStats(t *testing.T) {
	reg := NewRegistry()

	reg.RegisterCustomTool(&CustomTool{Name: "tool1"})
	reg.RegisterCustomTool(&CustomTool{Name: "tool2"})
	reg.RegisterSlashCommand(&SlashCommand{Command: "/cmd1"})
	reg.RegisterMiddleware(&Middleware{Name: "mw1"})

	stats := reg.GetStats()

	if stats["custom_tools"] != 2 {
		t.Errorf("Expected 2 custom tools, got %d", stats["custom_tools"])
	}

	if stats["slash_commands"] != 1 {
		t.Errorf("Expected 1 slash command, got %d", stats["slash_commands"])
	}

	if stats["middlewares"] != 1 {
		t.Errorf("Expected 1 middleware, got %d", stats["middlewares"])
	}
}

func TestRegistry_CustomTool_Handler(t *testing.T) {
	reg := NewRegistry()

	tool := &CustomTool{
		Name: "calculator",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			return a + b, nil
		},
	}

	reg.RegisterCustomTool(tool)

	retrieved, _ := reg.GetCustomTool("calculator")
	result, err := retrieved.Handler(context.Background(), map[string]interface{}{
		"a": 10.0,
		"b": 20.0,
	})

	if err != nil {
		t.Fatalf("Tool handler failed: %v", err)
	}

	if result != 30.0 {
		t.Errorf("Expected 30.0, got %v", result)
	}
}

func TestRegistry_SlashCommand_Handler(t *testing.T) {
	reg := NewRegistry()

	cmd := &SlashCommand{
		Command: "/echo",
		Handler: func(ctx context.Context, args []string) (string, error) {
			return fmt.Sprintf("Echo: %v", args), nil
		},
	}

	reg.RegisterSlashCommand(cmd)

	retrieved, _ := reg.GetSlashCommand("/echo")
	result, err := retrieved.Handler(context.Background(), []string{"hello", "world"})

	if err != nil {
		t.Fatalf("Command handler failed: %v", err)
	}

	expected := "Echo: [hello world]"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}
