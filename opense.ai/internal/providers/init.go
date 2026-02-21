// Package providers initializes and registers all concrete data providers
// with the global provider registry.
package providers

import (
	"os"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/internal/providers/fmp"
	"github.com/seenimoa/openseai/internal/providers/yfinance"
)

// RegisterAll creates and registers all available providers with the
// global registry. Providers that require API keys will only be registered
// if their environment variable is set.
func RegisterAll() error {
	return RegisterAllTo(provider.Global())
}

// RegisterAllTo registers all available providers to the given registry.
func RegisterAllTo(reg *provider.Registry) error {
	// --- YFinance (free, no API key) ---
	yf := yfinance.New()
	if err := yf.Init(nil); err != nil {
		return err
	}
	if err := reg.Register(yf); err != nil {
		return err
	}

	// --- FMP (requires API key) ---
	if apiKey := os.Getenv("FMP_API_KEY"); apiKey != "" {
		fp := fmp.New()
		if err := fp.Init(map[string]string{"api_key": apiKey}); err != nil {
			return err
		}
		if err := reg.Register(fp); err != nil {
			return err
		}
	}

	return nil
}
