package provider

import (
	"context"
	"time"

	"github.com/seenimoa/openseai/internal/infra"
)

// BaseFetcher provides common functionality for fetcher implementations.
// Embed this in concrete fetchers to get caching and rate limiting for free.
type BaseFetcher struct {
	model       ModelType
	description string
	required    []string
	optional    []string
	cache       *infra.Cache
	limiter     *infra.RateLimiter
}

// NewBaseFetcher creates a base fetcher with sensible defaults.
func NewBaseFetcher(model ModelType, desc string, required, optional []string) BaseFetcher {
	return BaseFetcher{
		model:       model,
		description: desc,
		required:    required,
		optional:    optional,
		cache:       infra.NewCache(5 * time.Minute),
		limiter:     infra.NewRateLimiter(10, time.Second),
	}
}

// NewBaseFetcherWithOpts creates a base fetcher with custom cache TTL and rate limit.
func NewBaseFetcherWithOpts(model ModelType, desc string, required, optional []string, cacheTTL time.Duration, rateLimit int, rateWindow time.Duration) BaseFetcher {
	return BaseFetcher{
		model:       model,
		description: desc,
		required:    required,
		optional:    optional,
		cache:       infra.NewCache(cacheTTL),
		limiter:     infra.NewRateLimiter(rateLimit, rateWindow),
	}
}

func (b *BaseFetcher) ModelType() ModelType   { return b.model }
func (b *BaseFetcher) Description() string    { return b.description }
func (b *BaseFetcher) RequiredParams() []string { return b.required }
func (b *BaseFetcher) OptionalParams() []string { return b.optional }

// CacheGet retrieves a value from the fetcher's cache.
func (b *BaseFetcher) CacheGet(key string) (any, bool) {
	return b.cache.Get(key)
}

// CacheSet stores a value in the fetcher's cache.
func (b *BaseFetcher) CacheSet(key string, value any) {
	b.cache.Set(key, value)
}

// CacheSetTTL stores a value with a custom TTL.
func (b *BaseFetcher) CacheSetTTL(key string, value any, ttl time.Duration) {
	b.cache.SetWithTTL(key, value, ttl)
}

// RateLimit waits until a request slot is available.
func (b *BaseFetcher) RateLimit(ctx context.Context) error {
	return b.limiter.Wait(ctx)
}

// CacheKey builds a cache key from model type and query parameters.
func CacheKey(model ModelType, params QueryParams) string {
	key := string(model)
	// Deterministic ordering of params for consistent cache keys.
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == ParamProvider {
			continue // Don't include provider in cache key.
		}
		keys = append(keys, k)
	}
	// Simple sort (no import needed for few keys).
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	for _, k := range keys {
		key += ":" + k + "=" + params[k]
	}
	return key
}

// BaseProvider provides common functionality for provider implementations.
// Embed this in concrete providers to simplify implementation.
type BaseProvider struct {
	info       ProviderInfo
	fetchers   map[ModelType]Fetcher
	credentials map[string]string
}

// NewBaseProvider creates a base provider.
func NewBaseProvider(name, description, website string, creds []ProviderCredential) BaseProvider {
	return BaseProvider{
		info: ProviderInfo{
			Name:        name,
			Description: description,
			Website:     website,
			Credentials: creds,
		},
		fetchers:    make(map[ModelType]Fetcher),
		credentials: make(map[string]string),
	}
}

func (bp *BaseProvider) Info() ProviderInfo { return bp.info }

func (bp *BaseProvider) Init(credentials map[string]string) error {
	// Validate required credentials.
	for _, cred := range bp.info.Credentials {
		if cred.Required {
			val, ok := credentials[cred.Name]
			if !ok || val == "" {
				return &ErrInvalidCredentials{
					Provider: bp.info.Name,
					Detail:   "missing required credential: " + cred.Name,
				}
			}
		}
	}
	bp.credentials = credentials
	return nil
}

func (bp *BaseProvider) Fetcher(model ModelType) Fetcher {
	return bp.fetchers[model]
}

func (bp *BaseProvider) SupportedModels() []ModelType {
	models := make([]ModelType, 0, len(bp.fetchers))
	for m := range bp.fetchers {
		models = append(models, m)
	}
	return models
}

func (bp *BaseProvider) Ping(ctx context.Context) error {
	return nil // Override in concrete providers.
}

// RegisterFetcher adds a fetcher to this provider.
func (bp *BaseProvider) RegisterFetcher(f Fetcher) {
	model := f.ModelType()
	bp.fetchers[model] = f
	// Update info models list.
	bp.info.Models = bp.SupportedModels()
}

// Credential returns a stored credential value.
func (bp *BaseProvider) Credential(name string) string {
	return bp.credentials[name]
}
