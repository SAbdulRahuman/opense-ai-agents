package providers

import (
	"testing"

	"github.com/seenimoa/openseai/internal/provider"
)

func TestRegisterAllTo(t *testing.T) {
	reg := provider.NewRegistry()
	if err := RegisterAllTo(reg); err != nil {
		t.Fatalf("RegisterAllTo: %v", err)
	}

	// YFinance should always be registered (no key needed).
	yf, err := reg.Get("yfinance")
	if err != nil {
		t.Fatalf("YFinance not registered: %v", err)
	}
	if yf.Info().Name != "yfinance" {
		t.Error("wrong yfinance provider name")
	}

	// SEC should always be registered (no key needed).
	secProv, err := reg.Get("sec")
	if err != nil {
		t.Fatalf("SEC not registered: %v", err)
	}
	if secProv.Info().Name != "sec" {
		t.Error("wrong sec provider name")
	}

	// FMP should only be registered if FMP_API_KEY is set.
	_, err = reg.Get("fmp")
	if err == nil {
		t.Log("FMP registered (FMP_API_KEY env var is set)")
	} else {
		t.Log("FMP not registered (no FMP_API_KEY)")
	}

	// FRED should only be registered if FRED_API_KEY is set.
	_, err = reg.Get("fred")
	if err == nil {
		t.Log("FRED registered (FRED_API_KEY env var is set)")
	} else {
		t.Log("FRED not registered (no FRED_API_KEY)")
	}
}

func TestRegisterAllToWithModelCoverage(t *testing.T) {
	reg := provider.NewRegistry()
	if err := RegisterAllTo(reg); err != nil {
		t.Fatalf("RegisterAllTo: %v", err)
	}

	// Verify key models have providers.
	keyModels := []provider.ModelType{
		provider.ModelEquityHistorical,
		provider.ModelEquityQuote,
		provider.ModelEquityInfo,
		provider.ModelEquitySearch,
		provider.ModelBalanceSheet,
		provider.ModelIncomeStatement,
		provider.ModelCashFlowStatement,
		provider.ModelKeyMetrics,
		provider.ModelHistoricalDividends,
		provider.ModelEtfHistorical,
		provider.ModelOptionsChains,
		provider.ModelCryptoHistorical,
		provider.ModelCurrencyHistorical,
		provider.ModelCompanyNews,
		// SEC models (always available)
		provider.ModelCompanyFilings,
		provider.ModelSecFiling,
		provider.ModelInsiderTrading,
		provider.ModelCikMap,
	}

	coverage := reg.ModelCoverage()
	for _, m := range keyModels {
		provs, ok := coverage[m]
		if !ok || len(provs) == 0 {
			t.Errorf("no providers for model %s", m)
		}
	}
}

func TestRegisterAllIdempotent(t *testing.T) {
	reg := provider.NewRegistry()
	if err := RegisterAllTo(reg); err != nil {
		t.Fatalf("first RegisterAllTo: %v", err)
	}
	// Registering again should overwrite without error.
	if err := RegisterAllTo(reg); err != nil {
		t.Fatalf("second RegisterAllTo: %v", err)
	}

	// Still exactly one yfinance provider.
	list := reg.List()
	yfCount := 0
	for _, info := range list {
		if info.Name == "yfinance" {
			yfCount++
		}
	}
	if yfCount != 1 {
		t.Errorf("expected 1 yfinance, got %d", yfCount)
	}
}
