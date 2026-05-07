package extension

import (
	"context"
	"fmt"
	"sync"
)

// ExtensionType 扩展类型
type ExtensionType string

const (
	ExtensionTypeCustomTool   ExtensionType = "custom_tool"
	ExtensionTypeSlashCommand ExtensionType = "slash_command"
	ExtensionTypeMiddleware   ExtensionType = "middleware"
)

// CustomTool 自定义工具
type CustomTool struct {
	Name        string
	Description string
	Handler     func(ctx context.Context, args map[string]interface{}) (interface{}, error)
	Metadata    map[string]string
}

// SlashCommand Slash 命令
type SlashCommand struct {
	Command     string
	Description string
	Handler     func(ctx context.Context, args []string) (string, error)
	Usage       string
	Metadata    map[string]string
}

// Middleware 中间件
type Middleware struct {
	Name     string
	Handler  func(ctx context.Context, data interface{}) (interface{}, error)
	Priority int // 优先级（越小越先执行）
}

// Registry 扩展注册表
type Registry struct {
	mu            sync.RWMutex
	customTools   map[string]*CustomTool
	slashCommands map[string]*SlashCommand
	middlewares   []*Middleware
}

// NewRegistry 创建新的扩展注册表
func NewRegistry() *Registry {
	return &Registry{
		customTools:   make(map[string]*CustomTool),
		slashCommands: make(map[string]*SlashCommand),
		middlewares:   make([]*Middleware, 0),
	}
}

// RegisterCustomTool 注册自定义工具
func (r *Registry) RegisterCustomTool(tool *CustomTool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if _, exists := r.customTools[tool.Name]; exists {
		return fmt.Errorf("tool '%s' already registered", tool.Name)
	}

	r.customTools[tool.Name] = tool
	return nil
}

// GetCustomTool 获取自定义工具
func (r *Registry) GetCustomTool(name string) (*CustomTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.customTools[name]
	return tool, exists
}

// ListCustomTools 列出所有自定义工具
func (r *Registry) ListCustomTools() []*CustomTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]*CustomTool, 0, len(r.customTools))
	for _, tool := range r.customTools {
		tools = append(tools, tool)
	}
	return tools
}

// UnregisterCustomTool 注销自定义工具
func (r *Registry) UnregisterCustomTool(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.customTools[name]; !exists {
		return fmt.Errorf("tool '%s' not found", name)
	}

	delete(r.customTools, name)
	return nil
}

// RegisterSlashCommand 注册 Slash 命令
func (r *Registry) RegisterSlashCommand(cmd *SlashCommand) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cmd.Command == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// 确保命令以 / 开头
	if cmd.Command[0] != '/' {
		cmd.Command = "/" + cmd.Command
	}

	if _, exists := r.slashCommands[cmd.Command]; exists {
		return fmt.Errorf("command '%s' already registered", cmd.Command)
	}

	r.slashCommands[cmd.Command] = cmd
	return nil
}

// GetSlashCommand 获取 Slash 命令
func (r *Registry) GetSlashCommand(command string) (*SlashCommand, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 确保命令以 / 开头
	if command[0] != '/' {
		command = "/" + command
	}

	cmd, exists := r.slashCommands[command]
	return cmd, exists
}

// ListSlashCommands 列出所有 Slash 命令
func (r *Registry) ListSlashCommands() []*SlashCommand {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]*SlashCommand, 0, len(r.slashCommands))
	for _, cmd := range r.slashCommands {
		commands = append(commands, cmd)
	}
	return commands
}

// UnregisterSlashCommand 注销 Slash 命令
func (r *Registry) UnregisterSlashCommand(command string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if command[0] != '/' {
		command = "/" + command
	}

	if _, exists := r.slashCommands[command]; !exists {
		return fmt.Errorf("command '%s' not found", command)
	}

	delete(r.slashCommands, command)
	return nil
}

// RegisterMiddleware 注册中间件
func (r *Registry) RegisterMiddleware(mw *Middleware) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if mw.Name == "" {
		return fmt.Errorf("middleware name cannot be empty")
	}

	r.middlewares = append(r.middlewares, mw)

	// 按优先级排序
	r.sortMiddlewares()

	return nil
}

// GetMiddlewares 获取所有中间件（已排序）
func (r *Registry) GetMiddlewares() []*Middleware {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Middleware, len(r.middlewares))
	copy(result, r.middlewares)
	return result
}

// sortMiddlewares 按优先级排序中间件
func (r *Registry) sortMiddlewares() {
	for i := 0; i < len(r.middlewares)-1; i++ {
		for j := 0; j < len(r.middlewares)-i-1; j++ {
			if r.middlewares[j].Priority > r.middlewares[j+1].Priority {
				r.middlewares[j], r.middlewares[j+1] = r.middlewares[j+1], r.middlewares[j]
			}
		}
	}
}

// ExecuteMiddlewareChain 执行中间件链
func (r *Registry) ExecuteMiddlewareChain(ctx context.Context, data interface{}) (interface{}, error) {
	middlewares := r.GetMiddlewares()

	current := data
	var err error

	for _, mw := range middlewares {
		current, err = mw.Handler(ctx, current)
		if err != nil {
			return nil, fmt.Errorf("middleware '%s' failed: %w", mw.Name, err)
		}
	}

	return current, nil
}

// GetStats 获取注册表统计信息
func (r *Registry) GetStats() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]int{
		"custom_tools":   len(r.customTools),
		"slash_commands": len(r.slashCommands),
		"middlewares":    len(r.middlewares),
	}
}
