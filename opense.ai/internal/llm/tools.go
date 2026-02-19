package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Tool represents a function/tool that can be called by the LLM.
type Tool struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Parameters  *JSONSchema      `json:"parameters"`
	Handler     ToolHandler      `json:"-"` // excluded from JSON serialization
}

// ToolHandler is a function that executes a tool call and returns a string result.
type ToolHandler func(ctx context.Context, args json.RawMessage) (string, error)

// JSONSchema represents a JSON Schema definition for tool parameters.
type JSONSchema struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]*JSONSchema `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Enum        []string               `json:"enum,omitempty"`
	Items       *JSONSchema            `json:"items,omitempty"` // for array type
	Default     any                    `json:"default,omitempty"`
}

// ObjectSchema creates a JSON Schema for an object with the given properties.
func ObjectSchema(desc string, props map[string]*JSONSchema, required ...string) *JSONSchema {
	return &JSONSchema{
		Type:        "object",
		Description: desc,
		Properties:  props,
		Required:    required,
	}
}

// StringProp creates a JSON Schema for a string property.
func StringProp(desc string) *JSONSchema {
	return &JSONSchema{Type: "string", Description: desc}
}

// NumberProp creates a JSON Schema for a number property.
func NumberProp(desc string) *JSONSchema {
	return &JSONSchema{Type: "number", Description: desc}
}

// IntProp creates a JSON Schema for an integer property.
func IntProp(desc string) *JSONSchema {
	return &JSONSchema{Type: "integer", Description: desc}
}

// BoolProp creates a JSON Schema for a boolean property.
func BoolProp(desc string) *JSONSchema {
	return &JSONSchema{Type: "boolean", Description: desc}
}

// EnumProp creates a JSON Schema for a string enum property.
func EnumProp(desc string, values ...string) *JSONSchema {
	return &JSONSchema{Type: "string", Description: desc, Enum: values}
}

// ArrayProp creates a JSON Schema for an array property.
func ArrayProp(desc string, items *JSONSchema) *JSONSchema {
	return &JSONSchema{Type: "array", Description: desc, Items: items}
}

// ToolRegistry manages available tools and executes tool calls.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewToolRegistry creates an empty tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry. Overwrites if already exists.
func (r *ToolRegistry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = tool
}

// RegisterFunc is a convenience method to register a tool with inline handler.
func (r *ToolRegistry) RegisterFunc(name, desc string, params *JSONSchema, handler ToolHandler) {
	r.Register(Tool{
		Name:        name,
		Description: desc,
		Parameters:  params,
		Handler:     handler,
	})
}

// Get retrieves a tool by name.
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// List returns all registered tools as a slice.
func (r *ToolRegistry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

// Names returns the names of all registered tools.
func (r *ToolRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered tools.
func (r *ToolRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// Execute runs a tool call and returns the string result.
func (r *ToolRegistry) Execute(ctx context.Context, call ToolCall) (string, error) {
	tool, ok := r.Get(call.Name)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrToolNotFound, call.Name)
	}
	if tool.Handler == nil {
		return "", fmt.Errorf("llm: tool %q has no handler", call.Name)
	}
	return tool.Handler(ctx, call.Arguments)
}

// ExecuteAll runs all tool calls concurrently and returns results in order.
func (r *ToolRegistry) ExecuteAll(ctx context.Context, calls []ToolCall) []ToolResult {
	results := make([]ToolResult, len(calls))
	var wg sync.WaitGroup
	for i, call := range calls {
		wg.Add(1)
		go func(idx int, c ToolCall) {
			defer wg.Done()
			output, err := r.Execute(ctx, c)
			results[idx] = ToolResult{
				ToolCallID: c.ID,
				Name:       c.Name,
				Content:    output,
				Err:        err,
			}
		}(i, call)
	}
	wg.Wait()
	return results
}

// ToolResult represents the result of executing a tool.
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	Err        error  `json:"error,omitempty"`
}

// ToMessage converts a ToolResult to a Message for feeding back to the LLM.
func (tr ToolResult) ToMessage() Message {
	content := tr.Content
	if tr.Err != nil {
		content = fmt.Sprintf("Error executing tool %s: %v", tr.Name, tr.Err)
	}
	return ToolResultMessage(tr.ToolCallID, tr.Name, content)
}

// RunToolLoop executes the LLM tool-calling loop:
// 1. Send messages to LLM
// 2. If LLM returns tool calls, execute them
// 3. Append tool results to messages
// 4. Repeat until LLM returns a text response or maxIterations is reached
func RunToolLoop(ctx context.Context, provider LLMProvider, registry *ToolRegistry,
	messages []Message, tools []Tool, opts *ChatOptions, maxIterations int) (*Response, []Message, error) {

	if maxIterations <= 0 {
		maxIterations = 10
	}

	// Work with a copy of messages to avoid mutating the caller's slice
	msgs := make([]Message, len(messages))
	copy(msgs, messages)

	for i := 0; i < maxIterations; i++ {
		resp, err := provider.Chat(ctx, msgs, tools, opts)
		if err != nil {
			return nil, msgs, err
		}

		// If no tool calls, we're done
		if !resp.HasToolCalls() {
			return resp, msgs, nil
		}

		// Append the assistant message with tool calls
		msgs = append(msgs, AssistantToolCallMessage(resp.ToolCalls))

		// Execute all tool calls
		results := registry.ExecuteAll(ctx, resp.ToolCalls)

		// Append tool results as messages
		for _, result := range results {
			msgs = append(msgs, result.ToMessage())
		}
	}

	return nil, msgs, fmt.Errorf("llm: tool loop exceeded %d iterations", maxIterations)
}
