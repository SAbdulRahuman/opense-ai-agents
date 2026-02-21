// Package provider implements the OpenBB-inspired provider abstraction layer.
// It defines a Provider interface, a Fetcher interface, and a central registry
// that routes data requests to the appropriate provider based on model type.
package provider

import (
	"context"
	"fmt"
	"time"
)

// ProviderCredential describes a required credential for a provider.
type ProviderCredential struct {
	Name        string `json:"name"`        // e.g., "api_key"
	Description string `json:"description"` // e.g., "FMP API key from financialmodelingprep.com"
	Required    bool   `json:"required"`
	EnvVar      string `json:"env_var"` // environment variable name, e.g., "FMP_API_KEY"
}

// ProviderInfo holds metadata about a registered provider.
type ProviderInfo struct {
	Name        string               `json:"name"`        // e.g., "fmp", "yfinance"
	Description string               `json:"description"` // human-readable description
	Website     string               `json:"website"`     // e.g., "https://financialmodelingprep.com"
	Credentials []ProviderCredential `json:"credentials"`
	Models      []ModelType          `json:"models"` // supported standard models
}

// Provider is the interface that all data providers must implement.
// Each provider registers one or more Fetcher implementations for specific
// standard model types (e.g., EquityHistorical, BalanceSheet, OptionsChains).
type Provider interface {
	// Info returns metadata about this provider.
	Info() ProviderInfo

	// Init initializes the provider with credentials and configuration.
	// Called once during registration. Returns an error if required credentials
	// are missing or invalid.
	Init(credentials map[string]string) error

	// Fetcher returns the fetcher for the given model type, or nil if unsupported.
	Fetcher(model ModelType) Fetcher

	// SupportedModels returns all model types this provider can fetch.
	SupportedModels() []ModelType

	// Ping verifies the provider's connectivity and credentials.
	Ping(ctx context.Context) error
}

// QueryParams is the generic query parameter map passed to fetchers.
// Common keys include:
//   - "symbol"        : ticker symbol (e.g., "AAPL", "RELIANCE.NS")
//   - "start_date"    : start date (RFC3339 or YYYY-MM-DD)
//   - "end_date"      : end date
//   - "interval"      : timeframe (e.g., "1d", "1h", "5m")
//   - "limit"         : max results
//   - "period"        : reporting period ("annual", "quarterly")
//   - "provider"      : override provider name
//
// Each fetcher defines which keys it requires/supports.
type QueryParams map[string]string

// QueryParamKey constants for commonly used query parameters.
const (
	ParamSymbol    = "symbol"
	ParamStartDate = "start_date"
	ParamEndDate   = "end_date"
	ParamInterval  = "interval"
	ParamLimit     = "limit"
	ParamPeriod    = "period"
	ParamExchange  = "exchange"
	ParamExpiry    = "expiry"
	ParamCountry   = "country"
	ParamCurrency  = "currency"
	ParamQuery     = "query"
	ParamSortBy    = "sort_by"
	ParamOrder     = "order"
	ParamProvider  = "provider"
)

// FetchResult wraps a fetcher result with metadata.
type FetchResult struct {
	Provider  string    `json:"provider"`   // which provider returned this data
	Model     ModelType `json:"model"`      // the standard model type
	Data      any       `json:"data"`       // the fetched data (typed per model)
	FetchedAt time.Time `json:"fetched_at"` // when the data was fetched
	Cached    bool      `json:"cached"`     // whether this came from cache
}

// Fetcher is the interface for fetching a specific data type.
// Each Fetcher handles a single standard model type (e.g., EquityHistorical).
type Fetcher interface {
	// ModelType returns the standard model type this fetcher handles.
	ModelType() ModelType

	// Description returns a human-readable description of what this fetcher does.
	Description() string

	// RequiredParams returns the parameter keys this fetcher requires.
	RequiredParams() []string

	// OptionalParams returns the parameter keys this fetcher optionally accepts.
	OptionalParams() []string

	// Fetch retrieves data for the given query parameters.
	// The returned data type depends on the standard model:
	//   - EquityHistorical → []models.OHLCV
	//   - BalanceSheet     → []models.BalanceSheet
	//   - OptionsChains    → *models.OptionChain
	//   etc.
	Fetch(ctx context.Context, params QueryParams) (*FetchResult, error)
}

// ErrProviderNotFound is returned when a requested provider is not registered.
type ErrProviderNotFound struct {
	Name string
}

func (e *ErrProviderNotFound) Error() string {
	return fmt.Sprintf("provider %q not found", e.Name)
}

// ErrModelNotSupported is returned when a provider doesn't support a model type.
type ErrModelNotSupported struct {
	Provider string
	Model    ModelType
}

func (e *ErrModelNotSupported) Error() string {
	return fmt.Sprintf("provider %q does not support model %q", e.Provider, e.Model)
}

// ErrMissingParam is returned when a required query parameter is missing.
type ErrMissingParam struct {
	Param string
}

func (e *ErrMissingParam) Error() string {
	return fmt.Sprintf("missing required parameter %q", e.Param)
}

// ErrInvalidCredentials is returned when provider credentials are invalid.
type ErrInvalidCredentials struct {
	Provider string
	Detail   string
}

func (e *ErrInvalidCredentials) Error() string {
	return fmt.Sprintf("invalid credentials for provider %q: %s", e.Provider, e.Detail)
}

// ValidateParams checks that all required parameters are present in params.
func ValidateParams(params QueryParams, required []string) error {
	for _, key := range required {
		if v, ok := params[key]; !ok || v == "" {
			return &ErrMissingParam{Param: key}
		}
	}
	return nil
}
