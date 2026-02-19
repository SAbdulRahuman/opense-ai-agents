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

// ollamaModels lists commonly used Ollama models.
var ollamaModels = []string{
	"qwen2.5:32b",
	"qwen2.5:14b",
	"qwen2.5:7b",
	"llama3.3:70b",
	"llama3.1:8b",
	"mistral:7b",
	"deepseek-r1:14b",
	"deepseek-r1:32b",
	"codestral:22b",
	"phi4:14b",
	"gemma2:27b",
}

// OllamaProvider implements LLMProvider for local Ollama instances.
type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

// OllamaOption configures the Ollama provider.
type OllamaOption func(*OllamaProvider)

// WithOllamaModel sets the default model.
func WithOllamaModel(model string) OllamaOption {
	return func(p *OllamaProvider) { p.model = model }
}

// WithOllamaHTTPClient sets a custom HTTP client.
func WithOllamaHTTPClient(client *http.Client) OllamaOption {
	return func(p *OllamaProvider) { p.client = client }
}

// NewOllamaProvider creates an Ollama provider.
// baseURL is the Ollama server URL (e.g., "http://localhost:11434").
func NewOllamaProvider(baseURL string, opts ...OllamaOption) (*OllamaProvider, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	p := &OllamaProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   "qwen2.5:7b",
		client:  &http.Client{Timeout: 300 * time.Second}, // longer timeout for local models
	}
	for _, opt := range opts {
		opt(p)
	}
	return p, nil
}

func (p *OllamaProvider) Name() string    { return ProviderOllama }
func (p *OllamaProvider) Models() []string { return ollamaModels }

// Ping checks if the Ollama server is reachable.
func (p *OllamaProvider) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/api/tags", nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProviderDown, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", ErrProviderDown, resp.StatusCode)
	}
	return nil
}

// Chat sends a chat request to Ollama using the /api/chat endpoint.
func (p *OllamaProvider) Chat(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
	start := time.Now()
	model := p.resolveModel(opts)

	body := p.buildRequest(messages, tools, model, opts, false)
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/chat", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderDown, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("ollama: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ollama: decode response: %w", err)
	}

	return p.parseResponse(&result, model, start), nil
}

// ChatStream sends a streaming chat request to Ollama.
func (p *OllamaProvider) ChatStream(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (<-chan StreamChunk, error) {
	model := p.resolveModel(opts)

	body := p.buildRequest(messages, tools, model, opts, true)
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/chat", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderDown, err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		return nil, fmt.Errorf("ollama: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	ch := make(chan StreamChunk, 64)
	go p.readStream(resp.Body, ch)
	return ch, nil
}

// ── Internal Types ──

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Tools    []ollamaTool    `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
	Options  *ollamaOptions  `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

type ollamaTool struct {
	Type     string             `json:"type"`
	Function ollamaFunctionDef  `json:"function"`
}

type ollamaFunctionDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  *JSONSchema `json:"parameters"`
}

type ollamaToolCall struct {
	Function ollamaFunctionCall `json:"function"`
}

type ollamaFunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ollamaOptions struct {
	Temperature float64  `json:"temperature,omitempty"`
	NumPredict  int      `json:"num_predict,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

type ollamaChatResponse struct {
	Model              string        `json:"model"`
	Message            ollamaMessage `json:"message"`
	Done               bool          `json:"done"`
	TotalDuration      int64         `json:"total_duration"`
	PromptEvalCount    int           `json:"prompt_eval_count"`
	EvalCount          int           `json:"eval_count"`
}

// ── Helpers ──

func (p *OllamaProvider) resolveModel(opts *ChatOptions) string {
	if opts != nil && opts.Model != "" {
		return opts.Model
	}
	return p.model
}

func (p *OllamaProvider) buildRequest(messages []Message, tools []Tool, model string, opts *ChatOptions, stream bool) ollamaChatRequest {
	r := ollamaChatRequest{
		Model:    model,
		Messages: convertToOllamaMessages(messages),
		Stream:   stream,
	}
	if len(tools) > 0 {
		r.Tools = convertToOllamaTools(tools)
	}
	if opts != nil {
		o := &ollamaOptions{}
		hasOpts := false
		if opts.Temperature > 0 {
			o.Temperature = opts.Temperature
			hasOpts = true
		}
		if opts.MaxTokens > 0 {
			o.NumPredict = opts.MaxTokens
			hasOpts = true
		}
		if opts.TopP > 0 {
			o.TopP = opts.TopP
			hasOpts = true
		}
		if len(opts.Stop) > 0 {
			o.Stop = opts.Stop
			hasOpts = true
		}
		if hasOpts {
			r.Options = o
		}
	}
	return r
}

func (p *OllamaProvider) parseResponse(raw *ollamaChatResponse, model string, start time.Time) *Response {
	r := &Response{
		Model:    raw.Model,
		Provider: ProviderOllama,
		Latency:  time.Since(start),
		Content:  raw.Message.Content,
		Usage: Usage{
			PromptTokens:     raw.PromptEvalCount,
			CompletionTokens: raw.EvalCount,
			TotalTokens:      raw.PromptEvalCount + raw.EvalCount,
		},
		FinishReason: FinishStop,
	}
	if len(raw.Message.ToolCalls) > 0 {
		r.FinishReason = FinishToolCalls
		for i, tc := range raw.Message.ToolCalls {
			r.ToolCalls = append(r.ToolCalls, ToolCall{
				ID:        fmt.Sprintf("call_%d", i),
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			})
		}
	}
	return r
}

func (p *OllamaProvider) readStream(body io.ReadCloser, ch chan<- StreamChunk) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	// Ollama may return large lines; increase buffer
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		var chunk ollamaChatResponse
		if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
			ch <- StreamChunk{Err: fmt.Errorf("ollama: stream parse: %w", err)}
			return
		}

		sc := StreamChunk{
			Content: chunk.Message.Content,
			Done:    chunk.Done,
		}
		for i, tc := range chunk.Message.ToolCalls {
			sc.ToolCalls = append(sc.ToolCalls, ToolCall{
				ID:        fmt.Sprintf("call_%d", i),
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			})
		}
		if chunk.Done {
			sc.FinishReason = FinishStop
		}
		ch <- sc
		if chunk.Done {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		ch <- StreamChunk{Err: fmt.Errorf("ollama: stream read: %w", err)}
	}
}

// ── Conversion Helpers ──

func convertToOllamaMessages(messages []Message) []ollamaMessage {
	out := make([]ollamaMessage, 0, len(messages))
	for _, m := range messages {
		msg := ollamaMessage{
			Role:    string(m.Role),
			Content: m.Content,
		}
		// Ollama uses "tool" role but maps tool results slightly differently
		if m.Role == RoleTool {
			msg.Role = "tool"
		}
		for _, tc := range m.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, ollamaToolCall{
				Function: ollamaFunctionCall{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			})
		}
		out = append(out, msg)
	}
	return out
}

func convertToOllamaTools(tools []Tool) []ollamaTool {
	out := make([]ollamaTool, len(tools))
	for i, t := range tools {
		out[i] = ollamaTool{
			Type: "function",
			Function: ollamaFunctionDef{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		}
	}
	return out
}
