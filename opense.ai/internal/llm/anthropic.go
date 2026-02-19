package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// anthropicModels lists commonly available Anthropic models.
var anthropicModels = []string{
	"claude-sonnet-4-20250514",
	"claude-opus-4-20250514",
	"claude-3-7-sonnet-20250219",
	"claude-3-5-sonnet-20241022",
	"claude-3-5-haiku-20241022",
	"claude-3-haiku-20240307",
}

// AnthropicProvider implements LLMProvider for Anthropic's Messages API.
type AnthropicProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// AnthropicOption configures the Anthropic provider.
type AnthropicOption func(*AnthropicProvider)

// WithAnthropicModel sets the default model.
func WithAnthropicModel(model string) AnthropicOption {
	return func(p *AnthropicProvider) { p.model = model }
}

// WithAnthropicBaseURL sets a custom base URL.
func WithAnthropicBaseURL(url string) AnthropicOption {
	return func(p *AnthropicProvider) { p.baseURL = strings.TrimRight(url, "/") }
}

// WithAnthropicHTTPClient sets a custom HTTP client.
func WithAnthropicHTTPClient(client *http.Client) AnthropicOption {
	return func(p *AnthropicProvider) { p.client = client }
}

// NewAnthropicProvider creates an Anthropic provider.
func NewAnthropicProvider(apiKey string, opts ...AnthropicOption) (*AnthropicProvider, error) {
	if apiKey == "" {
		return nil, ErrNoAPIKey
	}
	p := &AnthropicProvider{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1",
		model:   "claude-sonnet-4-20250514",
		client:  &http.Client{Timeout: 120 * time.Second},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p, nil
}

func (p *AnthropicProvider) Name() string    { return ProviderAnthropic }
func (p *AnthropicProvider) Models() []string { return anthropicModels }

// Ping verifies the API key is valid.
func (p *AnthropicProvider) Ping(ctx context.Context) error {
	// Anthropic doesn't have a lightweight ping endpoint;
	// send a minimal messages request to verify the key.
	body := anthropicRequest{
		Model:     p.model,
		MaxTokens: 1,
		Messages:  []anthropicMessage{{Role: "user", Content: []anthropicContentBlock{{Type: "text", Text: "hi"}}}},
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/messages", bytes.NewReader(data))
	if err != nil {
		return err
	}
	p.setHeaders(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProviderDown, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("%w: invalid API key", ErrNoAPIKey)
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("%w: status %d: %s", ErrProviderDown, resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// Chat sends a messages request to Anthropic.
func (p *AnthropicProvider) Chat(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
	start := time.Now()
	model := p.resolveModel(opts)

	body := p.buildRequest(messages, tools, model, opts)
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("anthropic: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/messages", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	p.setHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderDown, err)
	}
	defer resp.Body.Close()

	if err := p.checkError(resp); err != nil {
		return nil, err
	}

	var result anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("anthropic: decode response: %w", err)
	}

	return p.parseResponse(&result, model, start), nil
}

// ChatStream sends a streaming messages request to Anthropic.
func (p *AnthropicProvider) ChatStream(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (<-chan StreamChunk, error) {
	model := p.resolveModel(opts)

	body := p.buildRequest(messages, tools, model, opts)
	body.Stream = true
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("anthropic: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/messages", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	p.setHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderDown, err)
	}

	if err := p.checkError(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}

	ch := make(chan StreamChunk, 64)
	go p.readStream(resp.Body, ch)
	return ch, nil
}

// ── Internal Types ──

type anthropicRequest struct {
	Model     string              `json:"model"`
	Messages  []anthropicMessage  `json:"messages"`
	System    string              `json:"system,omitempty"`
	Tools     []anthropicTool     `json:"tools,omitempty"`
	MaxTokens int                 `json:"max_tokens"`
	Stream    bool                `json:"stream,omitempty"`
	Temperature *float64          `json:"temperature,omitempty"`
	TopP      *float64            `json:"top_p,omitempty"`
	StopSequences []string        `json:"stop_sequences,omitempty"`
}

type anthropicMessage struct {
	Role    string                   `json:"role"`
	Content []anthropicContentBlock  `json:"content"`
}

type anthropicContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`         // for tool_use
	Name      string          `json:"name,omitempty"`       // for tool_use
	Input     json.RawMessage `json:"input,omitempty"`      // for tool_use
	ToolUseID string          `json:"tool_use_id,omitempty"` // for tool_result
	Content   string          `json:"content,omitempty"`     // for tool_result (when nested)
}

type anthropicTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema *JSONSchema `json:"input_schema"`
}

type anthropicResponse struct {
	ID           string                   `json:"id"`
	Type         string                   `json:"type"`
	Role         string                   `json:"role"`
	Content      []anthropicContentBlock  `json:"content"`
	Model        string                   `json:"model"`
	StopReason   string                   `json:"stop_reason"`
	StopSequence *string                  `json:"stop_sequence"`
	Usage        anthropicUsage           `json:"usage"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Streaming event types
type anthropicStreamEvent struct {
	Type         string                  `json:"type"`
	Index        int                     `json:"index,omitempty"`
	ContentBlock *anthropicContentBlock  `json:"content_block,omitempty"`
	Delta        *anthropicStreamDelta   `json:"delta,omitempty"`
	Message      *anthropicResponse      `json:"message,omitempty"`
	Usage        *anthropicUsage         `json:"usage,omitempty"`
}

type anthropicStreamDelta struct {
	Type         string          `json:"type"`
	Text         string          `json:"text,omitempty"`
	PartialJSON  string          `json:"partial_json,omitempty"`
	StopReason   string          `json:"stop_reason,omitempty"`
}

// ── Helpers ──

func (p *AnthropicProvider) resolveModel(opts *ChatOptions) string {
	if opts != nil && opts.Model != "" {
		return opts.Model
	}
	return p.model
}

func (p *AnthropicProvider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
}

func (p *AnthropicProvider) buildRequest(messages []Message, tools []Tool, model string, opts *ChatOptions) anthropicRequest {
	maxTokens := 4096
	if opts != nil && opts.MaxTokens > 0 {
		maxTokens = opts.MaxTokens
	}

	r := anthropicRequest{
		Model:     model,
		MaxTokens: maxTokens,
	}

	// Extract system prompt, convert messages
	for _, m := range messages {
		if m.Role == RoleSystem {
			r.System = m.Content
			continue
		}
	}
	r.Messages = convertToAnthropicMessages(messages)

	if len(tools) > 0 {
		r.Tools = convertToAnthropicTools(tools)
	}

	if opts != nil {
		if opts.Temperature > 0 {
			r.Temperature = &opts.Temperature
		}
		if opts.TopP > 0 {
			r.TopP = &opts.TopP
		}
		r.StopSequences = opts.Stop
	}

	return r
}

func (p *AnthropicProvider) checkError(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var apiErr anthropicErrorResponse
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Message != "" {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("%w: %s", ErrNoAPIKey, apiErr.Error.Message)
		case http.StatusTooManyRequests:
			return fmt.Errorf("%w: %s", ErrRateLimit, apiErr.Error.Message)
		case http.StatusBadRequest:
			if strings.Contains(apiErr.Error.Type, "invalid_request") {
				return fmt.Errorf("anthropic: %s", apiErr.Error.Message)
			}
		}
		return fmt.Errorf("anthropic: API error (%d): %s", resp.StatusCode, apiErr.Error.Message)
	}
	return fmt.Errorf("anthropic: HTTP %d: %s", resp.StatusCode, string(body))
}

func (p *AnthropicProvider) parseResponse(raw *anthropicResponse, model string, start time.Time) *Response {
	r := &Response{
		Model:    raw.Model,
		Provider: ProviderAnthropic,
		Latency:  time.Since(start),
		Usage: Usage{
			PromptTokens:     raw.Usage.InputTokens,
			CompletionTokens: raw.Usage.OutputTokens,
			TotalTokens:      raw.Usage.InputTokens + raw.Usage.OutputTokens,
		},
		FinishReason: mapAnthropicStopReason(raw.StopReason),
	}

	var textParts []string
	for _, block := range raw.Content {
		switch block.Type {
		case "text":
			textParts = append(textParts, block.Text)
		case "tool_use":
			r.ToolCalls = append(r.ToolCalls, ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: block.Input,
			})
		}
	}
	r.Content = strings.Join(textParts, "")

	return r
}

func (p *AnthropicProvider) readStream(body io.ReadCloser, ch chan<- StreamChunk) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	// Track current tool call being built
	var currentToolID, currentToolName string
	var toolArgsBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "" {
			continue
		}

		var event anthropicStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			ch <- StreamChunk{Err: fmt.Errorf("anthropic: stream parse: %w", err)}
			return
		}

		switch event.Type {
		case "content_block_start":
			if event.ContentBlock != nil && event.ContentBlock.Type == "tool_use" {
				currentToolID = event.ContentBlock.ID
				currentToolName = event.ContentBlock.Name
				toolArgsBuilder.Reset()
			}

		case "content_block_delta":
			if event.Delta == nil {
				continue
			}
			switch event.Delta.Type {
			case "text_delta":
				ch <- StreamChunk{Content: event.Delta.Text}
			case "input_json_delta":
				toolArgsBuilder.WriteString(event.Delta.PartialJSON)
			}

		case "content_block_stop":
			if currentToolName != "" {
				ch <- StreamChunk{
					ToolCalls: []ToolCall{{
						ID:        currentToolID,
						Name:      currentToolName,
						Arguments: json.RawMessage(toolArgsBuilder.String()),
					}},
				}
				currentToolID = ""
				currentToolName = ""
				toolArgsBuilder.Reset()
			}

		case "message_delta":
			if event.Delta != nil && event.Delta.StopReason != "" {
				ch <- StreamChunk{
					FinishReason: mapAnthropicStopReason(event.Delta.StopReason),
					Done:         true,
				}
				return
			}

		case "message_stop":
			ch <- StreamChunk{Done: true}
			return
		}
	}
	if err := scanner.Err(); err != nil {
		ch <- StreamChunk{Err: fmt.Errorf("anthropic: stream read: %w", err)}
	}
}

// ── Conversion Helpers ──

func convertToAnthropicMessages(messages []Message) []anthropicMessage {
	var out []anthropicMessage
	for _, m := range messages {
		if m.Role == RoleSystem {
			continue // handled separately
		}

		switch m.Role {
		case RoleUser:
			out = append(out, anthropicMessage{
				Role: "user",
				Content: []anthropicContentBlock{{Type: "text", Text: m.Content}},
			})

		case RoleAssistant:
			msg := anthropicMessage{Role: "assistant"}
			if m.Content != "" {
				msg.Content = append(msg.Content, anthropicContentBlock{
					Type: "text",
					Text: m.Content,
				})
			}
			for _, tc := range m.ToolCalls {
				msg.Content = append(msg.Content, anthropicContentBlock{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Name,
					Input: tc.Arguments,
				})
			}
			out = append(out, msg)

		case RoleTool:
			out = append(out, anthropicMessage{
				Role: "user",
				Content: []anthropicContentBlock{{
					Type:      "tool_result",
					ToolUseID: m.ToolCallID,
					Content:   m.Content,
				}},
			})
		}
	}
	return out
}

func convertToAnthropicTools(tools []Tool) []anthropicTool {
	out := make([]anthropicTool, len(tools))
	for i, t := range tools {
		out[i] = anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Parameters,
		}
	}
	return out
}

func mapAnthropicStopReason(reason string) FinishReason {
	switch reason {
	case "end_turn":
		return FinishStop
	case "tool_use":
		return FinishToolCalls
	case "max_tokens":
		return FinishLength
	default:
		return FinishReason(reason)
	}
}
