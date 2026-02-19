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

// geminiModels lists commonly available Gemini models.
var geminiModels = []string{
	"gemini-2.0-flash",
	"gemini-2.0-flash-lite",
	"gemini-1.5-pro",
	"gemini-1.5-flash",
	"gemini-1.5-flash-8b",
}

// GeminiProvider implements LLMProvider for Google's Gemini API.
type GeminiProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// GeminiOption configures the Gemini provider.
type GeminiOption func(*GeminiProvider)

// WithGeminiModel sets the default model.
func WithGeminiModel(model string) GeminiOption {
	return func(p *GeminiProvider) { p.model = model }
}

// WithGeminiHTTPClient sets a custom HTTP client.
func WithGeminiHTTPClient(client *http.Client) GeminiOption {
	return func(p *GeminiProvider) { p.client = client }
}

// NewGeminiProvider creates a Gemini provider.
func NewGeminiProvider(apiKey string, opts ...GeminiOption) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, ErrNoAPIKey
	}
	p := &GeminiProvider{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
		model:   "gemini-2.0-flash",
		client:  &http.Client{Timeout: 120 * time.Second},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p, nil
}

func (p *GeminiProvider) Name() string    { return ProviderGemini }
func (p *GeminiProvider) Models() []string { return geminiModels }

// Ping verifies the API key by listing models.
func (p *GeminiProvider) Ping(ctx context.Context) error {
	url := fmt.Sprintf("%s/models?key=%s", p.baseURL, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProviderDown, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%w: invalid API key", ErrNoAPIKey)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", ErrProviderDown, resp.StatusCode)
	}
	return nil
}

// Chat sends a generate content request to Gemini.
func (p *GeminiProvider) Chat(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
	start := time.Now()
	model := p.resolveModel(opts)

	body := p.buildRequest(messages, tools, opts)
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("gemini: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderDown, err)
	}
	defer resp.Body.Close()

	if err := p.checkError(resp); err != nil {
		return nil, err
	}

	var result geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("gemini: decode response: %w", err)
	}

	return p.parseResponse(&result, model, start), nil
}

// ChatStream sends a streaming generate content request to Gemini.
func (p *GeminiProvider) ChatStream(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (<-chan StreamChunk, error) {
	model := p.resolveModel(opts)

	body := p.buildRequest(messages, tools, opts)
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("gemini: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse&key=%s", p.baseURL, model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

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

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	Tools            []geminiToolDecl       `json:"tools,omitempty"`
	SystemInstruction *geminiContent        `json:"system_instruction,omitempty"`
	GenerationConfig *geminiGenerationConfig `json:"generation_config,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text             string              `json:"text,omitempty"`
	FunctionCall     *geminiFunctionCall `json:"functionCall,omitempty"`
	FunctionResponse *geminiFuncResponse `json:"functionResponse,omitempty"`
}

type geminiFunctionCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

type geminiFuncResponse struct {
	Name     string          `json:"name"`
	Response json.RawMessage `json:"response"`
}

type geminiToolDecl struct {
	FunctionDeclarations []geminiFunctionDecl `json:"function_declarations"`
}

type geminiFunctionDecl struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  *JSONSchema `json:"parameters"`
}

type geminiGenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate    `json:"candidates"`
	UsageMetadata geminiUsageMetadata  `json:"usageMetadata"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type geminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type geminiErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// ── Helpers ──

func (p *GeminiProvider) resolveModel(opts *ChatOptions) string {
	if opts != nil && opts.Model != "" {
		return opts.Model
	}
	return p.model
}

func (p *GeminiProvider) buildRequest(messages []Message, tools []Tool, opts *ChatOptions) geminiRequest {
	r := geminiRequest{}

	// Extract system prompt, convert rest to Gemini content format
	for _, m := range messages {
		switch m.Role {
		case RoleSystem:
			r.SystemInstruction = &geminiContent{
				Parts: []geminiPart{{Text: m.Content}},
			}
		case RoleUser:
			r.Contents = append(r.Contents, geminiContent{
				Role:  "user",
				Parts: []geminiPart{{Text: m.Content}},
			})
		case RoleAssistant:
			content := geminiContent{Role: "model"}
			if m.Content != "" {
				content.Parts = append(content.Parts, geminiPart{Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				content.Parts = append(content.Parts, geminiPart{
					FunctionCall: &geminiFunctionCall{
						Name: tc.Name,
						Args: tc.Arguments,
					},
				})
			}
			r.Contents = append(r.Contents, content)
		case RoleTool:
			// Wrap content in JSON object for Gemini's format
			respData := json.RawMessage(fmt.Sprintf(`{"result": %s}`, quoteIfNeeded(m.Content)))
			r.Contents = append(r.Contents, geminiContent{
				Role: "user",
				Parts: []geminiPart{{
					FunctionResponse: &geminiFuncResponse{
						Name:     m.Name,
						Response: respData,
					},
				}},
			})
		}
	}

	if len(tools) > 0 {
		decls := make([]geminiFunctionDecl, len(tools))
		for i, t := range tools {
			decls[i] = geminiFunctionDecl{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			}
		}
		r.Tools = []geminiToolDecl{{FunctionDeclarations: decls}}
	}

	if opts != nil {
		gc := &geminiGenerationConfig{}
		hasConfig := false
		if opts.Temperature > 0 {
			gc.Temperature = opts.Temperature
			hasConfig = true
		}
		if opts.MaxTokens > 0 {
			gc.MaxOutputTokens = opts.MaxTokens
			hasConfig = true
		}
		if opts.TopP > 0 {
			gc.TopP = opts.TopP
			hasConfig = true
		}
		if len(opts.Stop) > 0 {
			gc.StopSequences = opts.Stop
			hasConfig = true
		}
		if hasConfig {
			r.GenerationConfig = gc
		}
	}

	return r
}

func (p *GeminiProvider) checkError(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var apiErr geminiErrorResponse
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Message != "" {
		switch resp.StatusCode {
		case http.StatusForbidden, http.StatusUnauthorized:
			return fmt.Errorf("%w: %s", ErrNoAPIKey, apiErr.Error.Message)
		case http.StatusTooManyRequests:
			return fmt.Errorf("%w: %s", ErrRateLimit, apiErr.Error.Message)
		case http.StatusBadRequest:
			if strings.Contains(apiErr.Error.Message, "not found") {
				return fmt.Errorf("%w: %s", ErrInvalidModel, apiErr.Error.Message)
			}
		}
		return fmt.Errorf("gemini: API error (%d): %s", resp.StatusCode, apiErr.Error.Message)
	}
	return fmt.Errorf("gemini: HTTP %d: %s", resp.StatusCode, string(body))
}

func (p *GeminiProvider) parseResponse(raw *geminiResponse, model string, start time.Time) *Response {
	r := &Response{
		Model:    model,
		Provider: ProviderGemini,
		Latency:  time.Since(start),
		Usage: Usage{
			PromptTokens:     raw.UsageMetadata.PromptTokenCount,
			CompletionTokens: raw.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      raw.UsageMetadata.TotalTokenCount,
		},
		FinishReason: FinishStop,
	}

	if len(raw.Candidates) > 0 {
		candidate := raw.Candidates[0]
		r.FinishReason = mapGeminiFinishReason(candidate.FinishReason)

		var textParts []string
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				textParts = append(textParts, part.Text)
			}
			if part.FunctionCall != nil {
				r.ToolCalls = append(r.ToolCalls, ToolCall{
					ID:        fmt.Sprintf("call_%s_%d", part.FunctionCall.Name, len(r.ToolCalls)),
					Name:      part.FunctionCall.Name,
					Arguments: part.FunctionCall.Args,
				})
			}
		}
		r.Content = strings.Join(textParts, "")
		if len(r.ToolCalls) > 0 {
			r.FinishReason = FinishToolCalls
		}
	}

	return r
}

func (p *GeminiProvider) readStream(body io.ReadCloser, ch chan<- StreamChunk) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "" {
			continue
		}

		var chunk geminiResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			ch <- StreamChunk{Err: fmt.Errorf("gemini: stream parse: %w", err)}
			return
		}

		sc := StreamChunk{}
		if len(chunk.Candidates) > 0 {
			candidate := chunk.Candidates[0]
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					sc.Content += part.Text
				}
				if part.FunctionCall != nil {
					sc.ToolCalls = append(sc.ToolCalls, ToolCall{
						ID:        fmt.Sprintf("call_%s", part.FunctionCall.Name),
						Name:      part.FunctionCall.Name,
						Arguments: part.FunctionCall.Args,
					})
				}
			}
			if candidate.FinishReason == "STOP" {
				sc.FinishReason = FinishStop
				sc.Done = true
			}
		}
		ch <- sc
	}
	if err := scanner.Err(); err != nil {
		ch <- StreamChunk{Err: fmt.Errorf("gemini: stream read: %w", err)}
	}
}

func mapGeminiFinishReason(reason string) FinishReason {
	switch reason {
	case "STOP":
		return FinishStop
	case "MAX_TOKENS":
		return FinishLength
	default:
		return FinishReason(reason)
	}
}

// quoteIfNeeded wraps a string as a JSON string value if it's not already valid JSON.
func quoteIfNeeded(s string) string {
	if json.Valid([]byte(s)) {
		return s
	}
	b, _ := json.Marshal(s)
	return string(b)
}
