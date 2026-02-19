package llm

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/seenimoa/openseai/internal/config"
)

// TaskComplexity indicates how complex a query is, used for routing.
type TaskComplexity int

const (
	TaskSimple   TaskComplexity = iota // Quick lookups, single indicators
	TaskModerate                        // Multi-step analysis, comparisons
	TaskComplex                         // Deep analysis, multi-agent synthesis
)

// Router routes LLM requests to the appropriate provider based on
// task complexity, model availability, and fallback configuration.
type Router struct {
	mu          sync.RWMutex
	providers   map[string]LLMProvider
	primary     string
	fallbacks   []string
	modelMap    map[TaskComplexity]string // complexity → model override
	maxRetries  int
	retryDelay  time.Duration
}

// RouterOption configures the router.
type RouterOption func(*Router)

// WithFallbacks sets the fallback provider chain.
func WithFallbacks(providers ...string) RouterOption {
	return func(r *Router) { r.fallbacks = providers }
}

// WithModelMap configures model selection by task complexity.
func WithModelMap(m map[TaskComplexity]string) RouterOption {
	return func(r *Router) { r.modelMap = m }
}

// WithMaxRetries sets the maximum number of retry attempts per provider.
func WithMaxRetries(n int) RouterOption {
	return func(r *Router) { r.maxRetries = n }
}

// WithRetryDelay sets the base delay between retries.
func WithRetryDelay(d time.Duration) RouterOption {
	return func(r *Router) { r.retryDelay = d }
}

// NewRouter creates a new LLM router with the given primary provider.
func NewRouter(primary string, opts ...RouterOption) *Router {
	r := &Router{
		providers:  make(map[string]LLMProvider),
		primary:    primary,
		modelMap:   make(map[TaskComplexity]string),
		maxRetries: 2,
		retryDelay: 1 * time.Second,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// RegisterProvider adds a provider to the router.
func (r *Router) RegisterProvider(provider LLMProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[provider.Name()] = provider
}

// GetProvider returns a registered provider by name.
func (r *Router) GetProvider(name string) (LLMProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

// Primary returns the primary provider.
func (r *Router) Primary() (LLMProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[r.primary]
	if !ok {
		return nil, fmt.Errorf("%w: primary provider %q not registered", ErrNoProviders, r.primary)
	}
	return p, nil
}

// Chat routes a chat request through the provider chain with fallback.
// It tries the primary provider first, then falls back in order.
func (r *Router) Chat(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
	chain := r.providerChain()
	if len(chain) == 0 {
		return nil, ErrNoProviders
	}

	var lastErr error
	for _, providerName := range chain {
		provider, ok := r.GetProvider(providerName)
		if !ok {
			continue
		}

		resp, err := r.chatWithRetry(ctx, provider, messages, tools, opts)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		log.Printf("llm/router: provider %s failed: %v, trying next", providerName, err)

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Don't fallback on certain errors
		if isNonRetryable(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("llm/router: all providers failed, last error: %w", lastErr)
}

// ChatStream routes a streaming request using the same fallback chain.
func (r *Router) ChatStream(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (<-chan StreamChunk, error) {
	chain := r.providerChain()
	if len(chain) == 0 {
		return nil, ErrNoProviders
	}

	var lastErr error
	for _, providerName := range chain {
		provider, ok := r.GetProvider(providerName)
		if !ok {
			continue
		}

		ch, err := provider.ChatStream(ctx, messages, tools, opts)
		if err == nil {
			return ch, nil
		}

		lastErr = err
		log.Printf("llm/router: stream provider %s failed: %v, trying next", providerName, err)

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if isNonRetryable(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("llm/router: all stream providers failed, last error: %w", lastErr)
}

// ChatWithComplexity routes based on task complexity, selecting the appropriate model.
func (r *Router) ChatWithComplexity(ctx context.Context, complexity TaskComplexity,
	messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {

	// Apply model override if configured for this complexity level
	if model, ok := r.modelMap[complexity]; ok {
		if opts == nil {
			opts = &ChatOptions{}
		}
		if opts.Model == "" {
			opts.Model = model
		}
	}

	return r.Chat(ctx, messages, tools, opts)
}

// HealthCheck pings all registered providers and returns their status.
func (r *Router) HealthCheck(ctx context.Context) map[string]error {
	r.mu.RLock()
	providers := make(map[string]LLMProvider, len(r.providers))
	for k, v := range r.providers {
		providers[k] = v
	}
	r.mu.RUnlock()

	results := make(map[string]error, len(providers))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, provider := range providers {
		wg.Add(1)
		go func(n string, p LLMProvider) {
			defer wg.Done()
			pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			err := p.Ping(pingCtx)
			mu.Lock()
			results[n] = err
			mu.Unlock()
		}(name, provider)
	}

	wg.Wait()
	return results
}

// Name returns the name of the primary provider (satisfies LLMProvider).
func (r *Router) Name() string {
	return "router/" + r.primary
}

// Models returns the union of models from all registered providers (satisfies LLMProvider).
func (r *Router) Models() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var all []string
	seen := make(map[string]bool)
	for _, p := range r.providers {
		for _, m := range p.Models() {
			if !seen[m] {
				seen[m] = true
				all = append(all, m)
			}
		}
	}
	return all
}

// Ping checks the primary provider's health (satisfies LLMProvider).
func (r *Router) Ping(ctx context.Context) error {
	p, err := r.Primary()
	if err != nil {
		return err
	}
	return p.Ping(ctx)
}

// ProviderNames returns the names of all registered providers.
func (r *Router) ProviderNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// ── Internal Helpers ──

func (r *Router) providerChain() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	chain := []string{r.primary}
	for _, fb := range r.fallbacks {
		if fb != r.primary {
			chain = append(chain, fb)
		}
	}
	return chain
}

func (r *Router) chatWithRetry(ctx context.Context, provider LLMProvider,
	messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {

	var lastErr error
	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			delay := r.retryDelay * time.Duration(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, err := provider.Chat(ctx, messages, tools, opts)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		// Don't retry non-retryable errors
		if isNonRetryable(err) {
			return nil, err
		}
	}
	return nil, lastErr
}

func isNonRetryable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Don't retry auth errors, invalid model, or context length issues
	return strings.Contains(msg, "API key") ||
		strings.Contains(msg, ErrNoAPIKey.Error()) ||
		strings.Contains(msg, ErrInvalidModel.Error()) ||
		strings.Contains(msg, ErrContextLength.Error())
}

// NewRouterFromConfig creates a fully configured Router from the application config.
// It instantiates the appropriate providers based on available API keys.
func NewRouterFromConfig(cfg *config.Config) (*Router, error) {
	router := NewRouter(cfg.LLM.Primary,
		WithMaxRetries(2),
		WithRetryDelay(time.Second),
	)

	// Default model map: simple→mini, moderate→default, complex→best
	router.modelMap = map[TaskComplexity]string{
		TaskSimple:   selectSimpleModel(cfg.LLM.Primary),
		TaskModerate: cfg.LLM.Model,
		TaskComplex:  cfg.LLM.Model,
	}

	var fallbacks []string
	registered := 0

	// Register OpenAI if key is available
	if cfg.LLM.OpenAIKey != "" {
		p, err := NewOpenAIProvider(cfg.LLM.OpenAIKey,
			WithOpenAIModel(cfg.LLM.Model),
		)
		if err == nil {
			router.RegisterProvider(p)
			registered++
			if cfg.LLM.Primary != ProviderOpenAI {
				fallbacks = append(fallbacks, ProviderOpenAI)
			}
		}
	}

	// Register Ollama (no key needed, just URL)
	if cfg.LLM.OllamaURL != "" {
		model := cfg.LLM.Model
		if cfg.LLM.Primary != ProviderOllama {
			model = "qwen2.5:7b" // default local model
		}
		p, err := NewOllamaProvider(cfg.LLM.OllamaURL,
			WithOllamaModel(model),
		)
		if err == nil {
			router.RegisterProvider(p)
			registered++
			if cfg.LLM.Primary != ProviderOllama {
				fallbacks = append(fallbacks, ProviderOllama)
			}
		}
	}

	// Register Gemini if key is available
	if cfg.LLM.GeminiKey != "" {
		p, err := NewGeminiProvider(cfg.LLM.GeminiKey,
			WithGeminiModel(defaultGeminiModel(cfg.LLM.Model)),
		)
		if err == nil {
			router.RegisterProvider(p)
			registered++
			if cfg.LLM.Primary != ProviderGemini {
				fallbacks = append(fallbacks, ProviderGemini)
			}
		}
	}

	// Register Anthropic if key is available
	if cfg.LLM.AnthropicKey != "" {
		p, err := NewAnthropicProvider(cfg.LLM.AnthropicKey,
			WithAnthropicModel(defaultAnthropicModel(cfg.LLM.Model)),
		)
		if err == nil {
			router.RegisterProvider(p)
			registered++
			if cfg.LLM.Primary != ProviderAnthropic {
				fallbacks = append(fallbacks, ProviderAnthropic)
			}
		}
	}

	if registered == 0 {
		return nil, ErrNoProviders
	}

	router.fallbacks = fallbacks
	return router, nil
}

// selectSimpleModel returns a cheaper/faster model variant for simple tasks.
func selectSimpleModel(provider string) string {
	switch provider {
	case ProviderOpenAI:
		return "gpt-4o-mini"
	case ProviderGemini:
		return "gemini-2.0-flash-lite"
	case ProviderAnthropic:
		return "claude-3-5-haiku-20241022"
	default:
		return "" // use default
	}
}

func defaultGeminiModel(model string) string {
	if strings.HasPrefix(model, "gemini") {
		return model
	}
	return "gemini-2.0-flash"
}

func defaultAnthropicModel(model string) string {
	if strings.HasPrefix(model, "claude") {
		return model
	}
	return "claude-sonnet-4-20250514"
}
