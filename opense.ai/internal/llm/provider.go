// Package llm provides a unified interface for multiple LLM providers
// (OpenAI, Ollama, Gemini, Anthropic) with tool/function calling support,
// streaming, and model routing with fallback.
package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Provider names for routing and configuration.
const (
	ProviderOpenAI    = "openai"
	ProviderOllama    = "ollama"
	ProviderGemini    = "gemini"
	ProviderAnthropic = "anthropic"
)

// Common errors returned by LLM providers.
var (
	ErrNoAPIKey       = errors.New("llm: API key not configured")
	ErrRateLimit      = errors.New("llm: rate limit exceeded")
	ErrContextLength  = errors.New("llm: context length exceeded")
	ErrProviderDown   = errors.New("llm: provider unavailable")
	ErrInvalidModel   = errors.New("llm: invalid model")
	ErrStreamClosed   = errors.New("llm: stream closed")
	ErrToolNotFound   = errors.New("llm: tool not found")
	ErrNoProviders    = errors.New("llm: no providers configured")
)

// Role represents the role of a message sender.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// FinishReason indicates why the model stopped generating.
type FinishReason string

const (
	FinishStop      FinishReason = "stop"
	FinishToolCalls FinishReason = "tool_calls"
	FinishLength    FinishReason = "length"
	FinishError     FinishReason = "error"
)

// Message represents a single message in a conversation.
type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"` // for tool result messages
	Name       string     `json:"name,omitempty"`         // for tool result messages
}

// ToolCall represents a function/tool call requested by the model.
type ToolCall struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"` // JSON-encoded arguments
}

// Response represents a complete response from the LLM.
type Response struct {
	Content      string       `json:"content"`
	ToolCalls    []ToolCall   `json:"tool_calls,omitempty"`
	FinishReason FinishReason `json:"finish_reason"`
	Usage        Usage        `json:"usage"`
	Model        string       `json:"model"`
	Provider     string       `json:"provider"`
	Latency      time.Duration `json:"latency"`
}

// Usage tracks token consumption for a request.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents a single chunk in a streaming response.
type StreamChunk struct {
	Content      string       `json:"content,omitempty"`
	ToolCalls    []ToolCall   `json:"tool_calls,omitempty"`
	FinishReason FinishReason `json:"finish_reason,omitempty"`
	Done         bool         `json:"done"`
	Err          error        `json:"-"`
}

// ChatOptions configures a single chat request.
type ChatOptions struct {
	Model       string  `json:"model,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

// LLMProvider is the interface that all LLM backends must implement.
type LLMProvider interface {
	// Name returns the provider identifier (e.g., "openai", "ollama").
	Name() string

	// Chat sends a conversation and returns a complete response.
	// tools may be nil if no tool calling is needed.
	Chat(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error)

	// ChatStream sends a conversation and returns a channel of streaming chunks.
	// The channel is closed when the response is complete.
	ChatStream(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (<-chan StreamChunk, error)

	// Models returns the list of available models for this provider.
	Models() []string

	// Ping checks if the provider is reachable and the API key is valid.
	Ping(ctx context.Context) error
}

// ProviderConfig holds common configuration for creating an LLM provider.
type ProviderConfig struct {
	APIKey      string        `json:"api_key,omitempty"`
	BaseURL     string        `json:"base_url,omitempty"`
	Model       string        `json:"model"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
	Timeout     time.Duration `json:"timeout"`
}

// DefaultProviderConfig returns sensible defaults for provider configuration.
func DefaultProviderConfig() ProviderConfig {
	return ProviderConfig{
		Model:       "gpt-4o",
		Temperature: 0.1,
		MaxTokens:   4096,
		Timeout:     120 * time.Second,
	}
}

// NewMessage creates a message with the given role and content.
func NewMessage(role Role, content string) Message {
	return Message{Role: role, Content: content}
}

// SystemMessage creates a system prompt message.
func SystemMessage(content string) Message {
	return NewMessage(RoleSystem, content)
}

// UserMessage creates a user message.
func UserMessage(content string) Message {
	return NewMessage(RoleUser, content)
}

// AssistantMessage creates an assistant message.
func AssistantMessage(content string) Message {
	return NewMessage(RoleAssistant, content)
}

// ToolResultMessage creates a tool result message.
func ToolResultMessage(toolCallID, name, content string) Message {
	return Message{
		Role:       RoleTool,
		Content:    content,
		ToolCallID: toolCallID,
		Name:       name,
	}
}

// AssistantToolCallMessage creates an assistant message that contains tool calls.
func AssistantToolCallMessage(toolCalls []ToolCall) Message {
	return Message{
		Role:      RoleAssistant,
		ToolCalls: toolCalls,
	}
}

// HasToolCalls returns true if the response contains tool calls.
func (r *Response) HasToolCalls() bool {
	return len(r.ToolCalls) > 0
}

// String returns a human-readable summary of the response.
func (r *Response) String() string {
	if r.HasToolCalls() {
		return fmt.Sprintf("[%s/%s] %d tool call(s), %d tokens, %v",
			r.Provider, r.Model, len(r.ToolCalls), r.Usage.TotalTokens, r.Latency.Round(time.Millisecond))
	}
	truncated := r.Content
	if len(truncated) > 100 {
		truncated = truncated[:100] + "..."
	}
	return fmt.Sprintf("[%s/%s] %q, %d tokens, %v",
		r.Provider, r.Model, truncated, r.Usage.TotalTokens, r.Latency.Round(time.Millisecond))
}
