// Package agent implements the multi-agent AI system for NSE stock analysis.
// It provides a base Agent interface, specialized analyst agents, and an
// orchestrator that coordinates single-agent and multi-agent workflows.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/seenimoa/openseai/internal/llm"
	"github.com/seenimoa/openseai/pkg/models"
)

// ── Agent Interface ──

// Agent defines the interface that all specialized analysis agents must implement.
type Agent interface {
	// Name returns the agent's unique identifier (e.g., "fundamental_analyst").
	Name() string

	// Role returns a human-readable description of the agent's role.
	Role() string

	// SystemPrompt returns the system prompt that configures this agent's behavior.
	SystemPrompt() string

	// Tools returns the set of LLM tools this agent can invoke.
	Tools() []llm.Tool

	// Process executes a task and returns an AgentResult.
	Process(ctx context.Context, task string) (*AgentResult, error)

	// ProcessWithMessages processes a task with existing conversation context.
	ProcessWithMessages(ctx context.Context, task string, history []llm.Message) (*AgentResult, error)
}

// ── AgentResult ──

// AgentResult holds the output from an agent's processing.
type AgentResult struct {
	AgentName  string         `json:"agent_name"`
	Role       string         `json:"role"`
	Content    string         `json:"content"`     // LLM-generated analysis text
	Analysis   *models.AnalysisResult `json:"analysis,omitempty"`
	ToolCalls  int            `json:"tool_calls"`  // number of tool calls made
	Tokens     int            `json:"tokens"`      // total tokens consumed
	Duration   time.Duration  `json:"duration"`
	Messages   []llm.Message  `json:"messages"`    // full conversation history
	Error      string         `json:"error,omitempty"`
}

// ── Memory ──

// Memory manages conversation history with a sliding window and optional summary.
type Memory struct {
	mu         sync.RWMutex
	messages   []llm.Message
	maxSize    int           // max messages before summarization
	summary    string        // compressed summary of older messages
}

// NewMemory creates a conversation memory with the given window size.
func NewMemory(maxSize int) *Memory {
	if maxSize <= 0 {
		maxSize = 50
	}
	return &Memory{
		maxSize:  maxSize,
		messages: make([]llm.Message, 0, maxSize),
	}
}

// Add appends a message to the memory.
func (m *Memory) Add(msg llm.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

// AddAll appends multiple messages to the memory.
func (m *Memory) AddAll(msgs []llm.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msgs...)
}

// Messages returns all messages in the memory, potentially with a summary prefix.
func (m *Memory) Messages() []llm.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.summary == "" {
		result := make([]llm.Message, len(m.messages))
		copy(result, m.messages)
		return result
	}

	// Prepend summary as a system message
	result := make([]llm.Message, 0, len(m.messages)+1)
	result = append(result, llm.SystemMessage(fmt.Sprintf("Previous conversation summary: %s", m.summary)))
	result = append(result, m.messages...)
	return result
}

// Size returns the number of messages currently in memory.
func (m *Memory) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.messages)
}

// NeedsSummarization returns true if the memory exceeds its max size.
func (m *Memory) NeedsSummarization() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.messages) > m.maxSize
}

// Summarize compresses older messages into a summary, keeping the most recent N messages.
// The summarizer function is typically an LLM call.
func (m *Memory) Summarize(ctx context.Context, keepRecent int, summarizer func(ctx context.Context, messages []llm.Message) (string, error)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.messages) <= keepRecent {
		return nil
	}

	// Messages to summarize (older ones)
	toSummarize := m.messages[:len(m.messages)-keepRecent]

	// Include any existing summary context
	if m.summary != "" {
		toSummarize = append([]llm.Message{
			llm.SystemMessage("Previous summary: " + m.summary),
		}, toSummarize...)
	}

	m.mu.Unlock()
	summary, err := summarizer(ctx, toSummarize)
	m.mu.Lock()

	if err != nil {
		return fmt.Errorf("summarize memory: %w", err)
	}

	m.summary = summary
	m.messages = m.messages[len(m.messages)-keepRecent:]
	return nil
}

// Clear resets the memory completely.
func (m *Memory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = m.messages[:0]
	m.summary = ""
}

// ── BaseAgent ──

// BaseAgent provides a reusable base implementation for specialized agents.
// Agents embed this struct and configure their role-specific prompts and tools.
type BaseAgent struct {
	name         string
	role         string
	systemPrompt string
	tools        []llm.Tool
	registry     *llm.ToolRegistry
	provider     llm.LLMProvider
	memory       *Memory
	opts         *llm.ChatOptions
	maxToolIter  int // max tool-call loop iterations
}

// BaseAgentConfig configures a BaseAgent.
type BaseAgentConfig struct {
	Name         string
	Role         string
	SystemPrompt string
	Provider     llm.LLMProvider
	Tools        []llm.Tool
	ChatOptions  *llm.ChatOptions
	MemorySize   int
	MaxToolIter  int
}

// NewBaseAgent creates a new BaseAgent from the given configuration.
func NewBaseAgent(cfg BaseAgentConfig) *BaseAgent {
	if cfg.MaxToolIter <= 0 {
		cfg.MaxToolIter = 10
	}
	if cfg.MemorySize <= 0 {
		cfg.MemorySize = 50
	}

	reg := llm.NewToolRegistry()
	for _, t := range cfg.Tools {
		reg.Register(t)
	}

	return &BaseAgent{
		name:         cfg.Name,
		role:         cfg.Role,
		systemPrompt: cfg.SystemPrompt,
		tools:        cfg.Tools,
		registry:     reg,
		provider:     cfg.Provider,
		memory:       NewMemory(cfg.MemorySize),
		opts:         cfg.ChatOptions,
		maxToolIter:  cfg.MaxToolIter,
	}
}

// Name returns the agent's identifier.
func (a *BaseAgent) Name() string { return a.name }

// Role returns the agent's role description.
func (a *BaseAgent) Role() string { return a.role }

// SystemPrompt returns the agent's system prompt.
func (a *BaseAgent) SystemPrompt() string { return a.systemPrompt }

// Tools returns the agent's available tools.
func (a *BaseAgent) Tools() []llm.Tool { return a.tools }

// Provider returns the agent's LLM provider.
func (a *BaseAgent) Provider() llm.LLMProvider { return a.provider }

// Memory returns the agent's conversation memory.
func (a *BaseAgent) Memory() *Memory { return a.memory }

// Process executes a task with a fresh conversation (system prompt + user message).
func (a *BaseAgent) Process(ctx context.Context, task string) (*AgentResult, error) {
	return a.ProcessWithMessages(ctx, task, nil)
}

// ProcessWithMessages processes a task with optional existing conversation history.
func (a *BaseAgent) ProcessWithMessages(ctx context.Context, task string, history []llm.Message) (*AgentResult, error) {
	start := time.Now()

	// Build message list: system prompt + history + user task
	messages := make([]llm.Message, 0, len(history)+2)
	messages = append(messages, llm.SystemMessage(a.systemPrompt))
	if len(history) > 0 {
		messages = append(messages, history...)
	}
	messages = append(messages, llm.UserMessage(task))

	// Run tool-calling loop
	resp, finalMsgs, err := llm.RunToolLoop(ctx, a.provider, a.registry, messages, a.tools, a.opts, a.maxToolIter)
	if err != nil {
		return &AgentResult{
			AgentName: a.name,
			Role:      a.role,
			Error:     err.Error(),
			Duration:  time.Since(start),
			Messages:  finalMsgs,
		}, err
	}

	// Count tool calls from the conversation
	toolCallCount := 0
	for _, msg := range finalMsgs {
		toolCallCount += len(msg.ToolCalls)
	}

	// Store in memory
	a.memory.AddAll(finalMsgs[1:]) // skip system prompt from memory

	result := &AgentResult{
		AgentName: a.name,
		Role:      a.role,
		Content:   resp.Content,
		ToolCalls: toolCallCount,
		Tokens:    resp.Usage.TotalTokens,
		Duration:  time.Since(start),
		Messages:  finalMsgs,
	}

	return result, nil
}

// ── Helper: Parse structured analysis from LLM response ──

// ParseAnalysisResult attempts to extract a structured AnalysisResult from LLM content.
// The LLM is expected to include a JSON block in its response.
func ParseAnalysisResult(content string, defaults models.AnalysisResult) *models.AnalysisResult {
	result := defaults

	// Try to find JSON block in content
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		jsonStr := content[start : end+1]
		var parsed models.AnalysisResult
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
			// Merge parsed fields with defaults
			if parsed.Ticker != "" {
				result.Ticker = parsed.Ticker
			}
			if parsed.Recommendation != "" {
				result.Recommendation = parsed.Recommendation
			}
			if parsed.Confidence > 0 {
				result.Confidence = parsed.Confidence
			}
			if parsed.Summary != "" {
				result.Summary = parsed.Summary
			}
			if len(parsed.Signals) > 0 {
				result.Signals = parsed.Signals
			}
			if parsed.Details != nil {
				result.Details = parsed.Details
			}
		}
	}

	// Fallback: use the full content as summary if not extracted
	if result.Summary == "" {
		result.Summary = content
	}

	result.Timestamp = time.Now()
	return &result
}

// ── Agent Registry ──

// Registry holds a collection of named agents for the orchestrator.
type Registry struct {
	mu     sync.RWMutex
	agents map[string]Agent
}

// NewRegistry creates an empty agent registry.
func NewRegistry() *Registry {
	return &Registry{agents: make(map[string]Agent)}
}

// Register adds an agent to the registry.
func (r *Registry) Register(agent Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[agent.Name()] = agent
}

// Get retrieves an agent by name.
func (r *Registry) Get(name string) (Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.agents[name]
	return a, ok
}

// List returns all registered agents.
func (r *Registry) List() []Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Agent, 0, len(r.agents))
	for _, a := range r.agents {
		result = append(result, a)
	}
	return result
}

// Names returns the names of all registered agents.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.agents))
	for name := range r.agents {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered agents.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}
