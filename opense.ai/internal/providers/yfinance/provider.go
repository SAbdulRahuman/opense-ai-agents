// Package yfinance implements the Yahoo Finance data provider.
// It wraps Yahoo Finance's public APIs (v7 quote, v8 chart, v10 quoteSummary,
// v1 screener) into the standard provider/fetcher framework.
//
// Yahoo Finance is a free, no-API-key provider that covers equities,
// ETFs, indices, crypto, currencies, futures, and options worldwide.
package yfinance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/seenimoa/openseai/internal/infra"
	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/utils"
)

const providerName = "yfinance"

// Provider implements provider.Provider for Yahoo Finance.
type Provider struct {
	provider.BaseProvider
}

// New creates a new YFinance provider and registers all fetchers.
func New() *Provider {
	p := &Provider{
		BaseProvider: provider.NewBaseProvider(
			providerName,
			"Yahoo Finance - free global financial data",
			"https://finance.yahoo.com",
			nil, // no credentials required
		),
	}

	// --- Equity / Price ---
	p.RegisterFetcher(newEquityHistoricalFetcher())
	p.RegisterFetcher(newEquityQuoteFetcher())
	p.RegisterFetcher(newEquityInfoFetcher())
	p.RegisterFetcher(newEquitySearchFetcher())

	// --- Equity / Fundamentals ---
	p.RegisterFetcher(newBalanceSheetFetcher())
	p.RegisterFetcher(newIncomeStatementFetcher())
	p.RegisterFetcher(newCashFlowStatementFetcher())
	p.RegisterFetcher(newKeyMetricsFetcher())
	p.RegisterFetcher(newHistoricalDividendsFetcher())
	p.RegisterFetcher(newShareStatisticsFetcher())

	// --- Equity / Discovery ---
	p.RegisterFetcher(newEquityGainersFetcher())
	p.RegisterFetcher(newEquityLosersFetcher())
	p.RegisterFetcher(newEquityActiveFetcher())

	// --- ETF ---
	p.RegisterFetcher(newEtfHistoricalFetcher())
	p.RegisterFetcher(newEtfInfoFetcher())

	// --- Index ---
	p.RegisterFetcher(newIndexHistoricalFetcher())

	// --- Options ---
	p.RegisterFetcher(newOptionsChainsFetcher())

	// --- Crypto ---
	p.RegisterFetcher(newCryptoHistoricalFetcher())

	// --- Currency ---
	p.RegisterFetcher(newCurrencyHistoricalFetcher())

	// --- News ---
	p.RegisterFetcher(newCompanyNewsFetcher())

	return p
}

// Ping checks connectivity to Yahoo Finance.
func (p *Provider) Ping(ctx context.Context) error {
	url := "https://query1.finance.yahoo.com/v7/finance/quote?symbols=AAPL"
	body, _, err := infra.DoGet(ctx, url, jsonHeaders())
	if err != nil {
		return fmt.Errorf("yfinance ping: %w", err)
	}
	body.Close()
	return nil
}

// --- Shared helpers ---

func jsonHeaders() map[string]string {
	return map[string]string{"Accept": "application/json"}
}

// fetchJSON performs a GET request and decodes the response into dest.
func fetchJSON(ctx context.Context, url string, dest any) error {
	body, _, err := infra.DoGet(ctx, url, jsonHeaders())
	if err != nil {
		return err
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}
	return nil
}

// toYFTicker converts a symbol to Yahoo Finance format.
func toYFTicker(symbol string) string {
	// If it already looks like a global YF ticker (has . or ^), leave it.
	if strings.Contains(symbol, ".") || strings.HasPrefix(symbol, "^") {
		return symbol
	}
	return utils.ToYFinanceTicker(symbol)
}

// fromYFTicker converts a Yahoo Finance ticker back to canonical form.
func fromYFTicker(yfTicker string) string {
	return utils.FromYFinanceTicker(yfTicker)
}

// coalesce returns the first non-empty string.
func coalesce(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// newResult creates a FetchResult with the current timestamp.
func newResult(data any) *provider.FetchResult {
	return &provider.FetchResult{
		Data:      data,
		FetchedAt: time.Now(),
	}
}

// newCachedResult creates a FetchResult marked as cached.
func newCachedResult(data any) *provider.FetchResult {
	return &provider.FetchResult{
		Data:      data,
		FetchedAt: time.Now(),
		Cached:    true,
	}
}
