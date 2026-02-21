// Package providers initializes and registers all concrete data providers
// with the global provider registry.
package providers

import (
	"os"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/internal/providers/cboe"
	"github.com/seenimoa/openseai/internal/providers/fmp"
	"github.com/seenimoa/openseai/internal/providers/fred"
	"github.com/seenimoa/openseai/internal/providers/sec"
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

	// --- SEC EDGAR (free, no API key) ---
	sp := sec.New()
	if err := sp.Init(nil); err != nil {
		return err
	}
	if err := reg.Register(sp); err != nil {
		return err
	}

	// --- FRED (requires free API key) ---
	if apiKey := os.Getenv("FRED_API_KEY"); apiKey != "" {
		fp := fred.New()
		if err := fp.Init(map[string]string{"api_key": apiKey}); err != nil {
			return err
		}
		if err := reg.Register(fp); err != nil {
			return err
		}
	}

	// --- CBOE (free, no API key) ---
	cp := cboe.New()
	if err := cp.Init(nil); err != nil {
		return err
	}
	if err := reg.Register(cp); err != nil {
		return err
	}

	return nil
}
