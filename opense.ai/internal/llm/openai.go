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

// openAIModels lists commonly available OpenAI models.
var openAIModels = []string{
	"gpt-4o",
	"gpt-4o-mini",
	"gpt-4-turbo",
	"gpt-4",
	"gpt-3.5-turbo",
	"o1",
	"o1-mini",
	"o3-mini",
}

// OpenAIProvider implements LLMProvider for OpenAI's Chat Completions API.
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// OpenAIOption configures the OpenAI provider.
type OpenAIOption func(*OpenAIProvider)

// WithOpenAIBaseURL sets a custom base URL (e.g., for Azure OpenAI or proxies).
func WithOpenAIBaseURL(url string) OpenAIOption {
	return func(p *OpenAIProvider) { p.baseURL = strings.TrimRight(url, "/") }
}

// WithOpenAIModel sets the default model.
func WithOpenAIModel(model string) OpenAIOption {
	return func(p *OpenAIProvider) { p.model = model }
}

// WithOpenAIHTTPClient sets a custom HTTP client.
func WithOpenAIHTTPClient(client *http.Client) OpenAIOption {
	return func(p *OpenAIProvider) { p.client = client }
}

// NewOpenAIProvider creates an OpenAI provider.
func NewOpenAIProvider(apiKey string, opts ...OpenAIOption) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, ErrNoAPIKey
	}
	p := &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		model:   "gpt-4o",
		client:  &http.Client{Timeout: 120 * time.Second},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p, nil
}

func (p *OpenAIProvider) Name() string      { return ProviderOpenAI }
func (p *OpenAIProvider) Models() []string   { return openAIModels }

// Ping verifies the API key by listing models.
func (p *OpenAIProvider) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProviderDown, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("%w: invalid API key", ErrNoAPIKey)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", ErrProviderDown, resp.StatusCode)
	}
	return nil
}

// Chat sends a chat completion request to OpenAI.
func (p *OpenAIProvider) Chat(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
	start := time.Now()
	model := p.resolveModel(opts)

	body := p.buildRequest(messages, tools, model, opts, false)
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("openai: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(data))
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

	var result openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("openai: decode response: %w", err)
	}

	return p.parseResponse(&result, model, start), nil
}

// ChatStream sends a streaming chat completion request.
func (p *OpenAIProvider) ChatStream(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (<-chan StreamChunk, error) {
	model := p.resolveModel(opts)

	body := p.buildRequest(messages, tools, model, opts, true)
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("openai: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(data))
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

type openAIChatRequest struct {
	Model       string            `json:"model"`
	Messages    []openAIMessage   `json:"messages"`
	Tools       []openAITool      `json:"tools,omitempty"`
	Stream      bool              `json:"stream,omitempty"`
	Temperature *float64          `json:"temperature,omitempty"`
	MaxTokens   *int              `json:"max_tokens,omitempty"`
	TopP        *float64          `json:"top_p,omitempty"`
	Stop        []string          `json:"stop,omitempty"`
}

type openAIMessage struct {
	Role       string            `json:"role"`
	Content    string            `json:"content,omitempty"`
	ToolCalls  []openAIToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
	Name       string            `json:"name,omitempty"`
}

type openAITool struct {
	Type     string             `json:"type"`
	Function openAIFunctionDef  `json:"function"`
}

type openAIFunctionDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  *JSONSchema `json:"parameters"`
}

type openAIToolCall struct {
	ID       string                `json:"id"`
	Type     string                `json:"type"`
	Function openAIFunctionCall    `json:"function"`
}

type openAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIChatResponse struct {
	ID      string            `json:"id"`
	Choices []openAIChoice    `json:"choices"`
	Usage   openAIUsage       `json:"usage"`
	Model   string            `json:"model"`
}

type openAIChoice struct {
	Index        int               `json:"index"`
	Message      openAIMessage     `json:"message"`
	Delta        openAIMessage     `json:"delta"` // for streaming
	FinishReason string            `json:"finish_reason"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// ── Helpers ──

func (p *OpenAIProvider) resolveModel(opts *ChatOptions) string {
	if opts != nil && opts.Model != "" {
		return opts.Model
	}
	return p.model
}

func (p *OpenAIProvider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
}

func (p *OpenAIProvider) buildRequest(messages []Message, tools []Tool, model string, opts *ChatOptions, stream bool) openAIChatRequest {
	r := openAIChatRequest{
		Model:    model,
		Messages: convertToOpenAIMessages(messages),
		Stream:   stream,
	}
	if len(tools) > 0 {
		r.Tools = convertToOpenAITools(tools)
	}
	if opts != nil {
		if opts.Temperature > 0 {
			r.Temperature = &opts.Temperature
		}
		if opts.MaxTokens > 0 {
			r.MaxTokens = &opts.MaxTokens
		}
		if opts.TopP > 0 {
			r.TopP = &opts.TopP
		}
		r.Stop = opts.Stop
	}
	return r
}

func (p *OpenAIProvider) checkError(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var apiErr openAIErrorResponse
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Message != "" {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("%w: %s", ErrNoAPIKey, apiErr.Error.Message)
		case http.StatusTooManyRequests, 529:
			return fmt.Errorf("%w: %s", ErrRateLimit, apiErr.Error.Message)
		case http.StatusBadRequest:
			if strings.Contains(apiErr.Error.Code, "context_length") {
				return fmt.Errorf("%w: %s", ErrContextLength, apiErr.Error.Message)
			}
			if strings.Contains(apiErr.Error.Code, "model_not_found") {
				return fmt.Errorf("%w: %s", ErrInvalidModel, apiErr.Error.Message)
			}
		}
		return fmt.Errorf("openai: API error (%d): %s", resp.StatusCode, apiErr.Error.Message)
	}
	return fmt.Errorf("openai: HTTP %d: %s", resp.StatusCode, string(body))
}

func (p *OpenAIProvider) parseResponse(raw *openAIChatResponse, model string, start time.Time) *Response {
	r := &Response{
		Model:    raw.Model,
		Provider: ProviderOpenAI,
		Latency:  time.Since(start),
		Usage: Usage{
			PromptTokens:     raw.Usage.PromptTokens,
			CompletionTokens: raw.Usage.CompletionTokens,
			TotalTokens:      raw.Usage.TotalTokens,
		},
	}
	if len(raw.Choices) > 0 {
		choice := raw.Choices[0]
		r.Content = choice.Message.Content
		r.FinishReason = mapFinishReason(choice.FinishReason)
		for _, tc := range choice.Message.ToolCalls {
			r.ToolCalls = append(r.ToolCalls, ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: json.RawMessage(tc.Function.Arguments),
			})
		}
	}
	return r
}

func (p *OpenAIProvider) readStream(body io.ReadCloser, ch chan<- StreamChunk) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			ch <- StreamChunk{Done: true}
			return
		}

		var chunk openAIChatResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			ch <- StreamChunk{Err: fmt.Errorf("openai: stream parse: %w", err)}
			return
		}
		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta
		sc := StreamChunk{Content: delta.Content}
		for _, tc := range delta.ToolCalls {
			sc.ToolCalls = append(sc.ToolCalls, ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: json.RawMessage(tc.Function.Arguments),
			})
		}
		if fr := chunk.Choices[0].FinishReason; fr != "" {
			sc.FinishReason = mapFinishReason(fr)
			if fr == "stop" {
				sc.Done = true
			}
		}
		ch <- sc
	}
	if err := scanner.Err(); err != nil {
		ch <- StreamChunk{Err: fmt.Errorf("openai: stream read: %w", err)}
	}
}

// ── Conversion Helpers ──

func convertToOpenAIMessages(messages []Message) []openAIMessage {
	out := make([]openAIMessage, len(messages))
	for i, m := range messages {
		msg := openAIMessage{
			Role:       string(m.Role),
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
			Name:       m.Name,
		}
		for _, tc := range m.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, openAIToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: openAIFunctionCall{
					Name:      tc.Name,
					Arguments: string(tc.Arguments),
				},
			})
		}
		out[i] = msg
	}
	return out
}

func convertToOpenAITools(tools []Tool) []openAITool {
	out := make([]openAITool, len(tools))
	for i, t := range tools {
		out[i] = openAITool{
			Type: "function",
			Function: openAIFunctionDef{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		}
	}
	return out
}

func mapFinishReason(reason string) FinishReason {
	switch reason {
	case "stop":
		return FinishStop
	case "tool_calls":
		return FinishToolCalls
	case "length":
		return FinishLength
	default:
		return FinishReason(reason)
	}
}
