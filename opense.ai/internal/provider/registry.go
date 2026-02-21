package provider

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Registry is a thread-safe registry of data providers.
// It maps provider names to Provider instances and maintains an index
// of which providers support which standard model types.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider           // name → provider
	modelIdx  map[ModelType][]string        // model → provider names (priority order)
	defaults  map[ModelType]string          // model → default provider name
}

// NewRegistry creates a new empty provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
		modelIdx:  make(map[ModelType][]string),
		defaults:  make(map[ModelType]string),
	}
}

// Register adds a provider to the registry. If the provider requires
// credentials, they should be set via Init() before calling Register.
// Duplicate registrations overwrite the previous entry.
func (r *Registry) Register(p Provider) error {
	info := p.Info()
	if info.Name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers[info.Name] = p

	// Index the provider's supported models.
	for _, model := range p.SupportedModels() {
		existing := r.modelIdx[model]
		// Avoid duplicates.
		found := false
		for _, name := range existing {
			if name == info.Name {
				found = true
				break
			}
		}
		if !found {
			r.modelIdx[model] = append(existing, info.Name)
		}
		// Set as default if no default exists for this model.
		if _, ok := r.defaults[model]; !ok {
			r.defaults[model] = info.Name
		}
	}

	return nil
}

// Unregister removes a provider from the registry.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.providers, name)

	// Clean up model index.
	for model, names := range r.modelIdx {
		filtered := names[:0]
		for _, n := range names {
			if n != name {
				filtered = append(filtered, n)
			}
		}
		if len(filtered) == 0 {
			delete(r.modelIdx, model)
			delete(r.defaults, model)
		} else {
			r.modelIdx[model] = filtered
			if r.defaults[model] == name {
				r.defaults[model] = filtered[0]
			}
		}
	}
}

// Get returns a provider by name, or an error if not found.
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[name]
	if !ok {
		return nil, &ErrProviderNotFound{Name: name}
	}
	return p, nil
}

// List returns info about all registered providers, sorted by name.
func (r *Registry) List() []ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]ProviderInfo, 0, len(r.providers))
	for _, p := range r.providers {
		infos = append(infos, p.Info())
	}
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})
	return infos
}

// ProvidersFor returns the names of providers that support the given model type,
// in priority order (first = default).
func (r *Registry) ProvidersFor(model ModelType) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := r.modelIdx[model]
	result := make([]string, len(names))
	copy(result, names)
	return result
}

// DefaultProvider returns the default provider name for a model type.
func (r *Registry) DefaultProvider(model ModelType) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	name, ok := r.defaults[model]
	return name, ok
}

// SetDefault sets the default provider for a model type.
func (r *Registry) SetDefault(model ModelType, providerName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verify the provider exists and supports this model.
	p, ok := r.providers[providerName]
	if !ok {
		return &ErrProviderNotFound{Name: providerName}
	}

	fetcher := p.Fetcher(model)
	if fetcher == nil {
		return &ErrModelNotSupported{Provider: providerName, Model: model}
	}

	r.defaults[model] = providerName
	return nil
}

// Fetch retrieves data for the given model type using the specified provider
// (or the default if providerName is empty).
func (r *Registry) Fetch(ctx context.Context, model ModelType, params QueryParams) (*FetchResult, error) {
	providerName := params[ParamProvider]

	r.mu.RLock()
	if providerName == "" {
		providerName = r.defaults[model]
	}
	p, ok := r.providers[providerName]
	r.mu.RUnlock()

	if !ok || providerName == "" {
		return nil, &ErrProviderNotFound{Name: providerName}
	}

	fetcher := p.Fetcher(model)
	if fetcher == nil {
		return nil, &ErrModelNotSupported{Provider: providerName, Model: model}
	}

	// Validate required params.
	if err := ValidateParams(params, fetcher.RequiredParams()); err != nil {
		return nil, err
	}

	result, err := fetcher.Fetch(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("provider %q fetch %s: %w", providerName, model, err)
	}

	result.Provider = providerName
	result.Model = model
	if result.FetchedAt.IsZero() {
		result.FetchedAt = time.Now()
	}

	return result, nil
}

// FetchWithFallback tries the preferred provider first, then falls back to
// other providers that support the model, in priority order.
func (r *Registry) FetchWithFallback(ctx context.Context, model ModelType, params QueryParams) (*FetchResult, error) {
	// Try preferred provider first.
	result, err := r.Fetch(ctx, model, params)
	if err == nil {
		return result, nil
	}

	// Get all providers for this model.
	providers := r.ProvidersFor(model)
	preferred := params[ParamProvider]

	for _, name := range providers {
		if name == preferred {
			continue // Already tried.
		}
		fallbackParams := make(QueryParams, len(params))
		for k, v := range params {
			fallbackParams[k] = v
		}
		fallbackParams[ParamProvider] = name

		result, err = r.Fetch(ctx, model, fallbackParams)
		if err == nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("all providers failed for model %s: %w", model, err)
}

// ModelCoverage returns a map of model types to the list of providers that support them.
func (r *Registry) ModelCoverage() map[ModelType][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	coverage := make(map[ModelType][]string, len(r.modelIdx))
	for model, names := range r.modelIdx {
		cp := make([]string, len(names))
		copy(cp, names)
		coverage[model] = cp
	}
	return coverage
}

// global is the default global registry.
var global = NewRegistry()

// Global returns the default global provider registry.
func Global() *Registry {
	return global
}

// Register adds a provider to the global registry.
func RegisterProvider(p Provider) error {
	return global.Register(p)
}
