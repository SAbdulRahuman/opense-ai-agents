package datasource

import (
	"context"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
	"github.com/seenimoa/openseai/pkg/utils"
)

// ProviderAggregator extends Aggregator with provider-based data routing.
// It first tries the new provider registry, falling back to legacy data
// sources when a model type is not covered by any registered provider.
type ProviderAggregator struct {
	*Aggregator
	registry *provider.Registry
}

// NewProviderAggregator creates a new aggregator backed by both legacy sources
// and the provider registry. If registry is nil, the global registry is used.
func NewProviderAggregator(reg *provider.Registry) *ProviderAggregator {
	if reg == nil {
		reg = provider.Global()
	}
	return &ProviderAggregator{
		Aggregator: NewAggregator(),
		registry:   reg,
	}
}

// Registry returns the provider registry used by this aggregator.
func (pa *ProviderAggregator) Registry() *provider.Registry {
	return pa.registry
}

// FetchViaProvider fetches data through the provider registry for any model type.
// This is the primary entry point for provider-based data access.
func (pa *ProviderAggregator) FetchViaProvider(ctx context.Context, model provider.ModelType, params provider.QueryParams) (*provider.FetchResult, error) {
	return pa.registry.FetchWithFallback(ctx, model, params)
}

// FetchQuote fetches a stock quote, trying providers first, then legacy sources.
func (pa *ProviderAggregator) FetchQuote(ctx context.Context, ticker string) (*models.Quote, error) {
	symbol := utils.NormalizeTicker(ticker)

	// Try new provider registry first.
	params := provider.QueryParams{
		provider.ParamSymbol: symbol,
	}
	result, err := pa.registry.Fetch(ctx, provider.ModelEquityQuote, params)
	if err == nil {
		if quote, ok := result.Data.(*models.Quote); ok {
			return quote, nil
		}
	}

	// Fall back to legacy: YFinance → NSE.
	quote, err := pa.yfinance.GetQuote(ctx, symbol)
	if err == nil {
		return quote, nil
	}
	return pa.nse.GetQuote(ctx, symbol)
}

// FetchHistorical fetches historical OHLCV data, trying providers first.
func (pa *ProviderAggregator) FetchHistorical(ctx context.Context, ticker string, from, to time.Time, tf models.Timeframe) ([]models.OHLCV, error) {
	symbol := utils.NormalizeTicker(ticker)

	// Try new provider registry first.
	params := provider.QueryParams{
		provider.ParamSymbol:    symbol,
		provider.ParamStartDate: from.Format("2006-01-02"),
		provider.ParamEndDate:   to.Format("2006-01-02"),
		provider.ParamInterval:  string(tf),
	}
	result, err := pa.registry.Fetch(ctx, provider.ModelEquityHistorical, params)
	if err == nil {
		if ohlcv, ok := result.Data.([]models.OHLCV); ok {
			return ohlcv, nil
		}
	}

	// Fall back to legacy.
	return pa.FetchHistoricalData(ctx, ticker, from, to, tf)
}

// FetchFinancials fetches financial statements, trying providers first.
func (pa *ProviderAggregator) FetchFinancials(ctx context.Context, ticker string, period string) (*models.FinancialData, error) {
	symbol := utils.NormalizeTicker(ticker)

	// Try income statement via provider.
	params := provider.QueryParams{
		provider.ParamSymbol: symbol,
		provider.ParamPeriod: period,
	}
	result, err := pa.registry.Fetch(ctx, provider.ModelIncomeStatement, params)
	if err == nil {
		if fd, ok := result.Data.(*models.FinancialData); ok {
			return fd, nil
		}
	}

	// Fall back to legacy Screener.
	return pa.screener.GetFinancials(ctx, symbol)
}

// FetchBalanceSheet fetches balance sheet data via providers.
func (pa *ProviderAggregator) FetchBalanceSheet(ctx context.Context, ticker string, period string) ([]models.BalanceSheet, error) {
	params := provider.QueryParams{
		provider.ParamSymbol: utils.NormalizeTicker(ticker),
		provider.ParamPeriod: period,
	}
	result, err := pa.registry.FetchWithFallback(ctx, provider.ModelBalanceSheet, params)
	if err != nil {
		return nil, fmt.Errorf("balance sheet for %s: %w", ticker, err)
	}
	if bs, ok := result.Data.([]models.BalanceSheet); ok {
		return bs, nil
	}
	return nil, fmt.Errorf("unexpected data type for balance sheet")
}

// FetchCashFlow fetches cash flow statement data via providers.
func (pa *ProviderAggregator) FetchCashFlow(ctx context.Context, ticker string, period string) ([]models.CashFlow, error) {
	params := provider.QueryParams{
		provider.ParamSymbol: utils.NormalizeTicker(ticker),
		provider.ParamPeriod: period,
	}
	result, err := pa.registry.FetchWithFallback(ctx, provider.ModelCashFlowStatement, params)
	if err != nil {
		return nil, fmt.Errorf("cash flow for %s: %w", ticker, err)
	}
	if cf, ok := result.Data.([]models.CashFlow); ok {
		return cf, nil
	}
	return nil, fmt.Errorf("unexpected data type for cash flow")
}

// FetchOptions fetches option chain data, trying providers first.
func (pa *ProviderAggregator) FetchOptions(ctx context.Context, ticker string, expiry string) (*models.OptionChain, error) {
	symbol := utils.NormalizeTicker(ticker)

	// Try provider registry.
	params := provider.QueryParams{
		provider.ParamSymbol: symbol,
		provider.ParamExpiry: expiry,
	}
	result, err := pa.registry.Fetch(ctx, provider.ModelOptionsChains, params)
	if err == nil {
		if oc, ok := result.Data.(*models.OptionChain); ok {
			return oc, nil
		}
	}

	// Fall back to legacy NSE derivatives.
	return pa.derivatives.GetOptionChain(ctx, symbol, expiry)
}

// FetchETFInfo fetches ETF information via providers.
func (pa *ProviderAggregator) FetchETFInfo(ctx context.Context, symbol string) (*models.ETFInfo, error) {
	params := provider.QueryParams{
		provider.ParamSymbol: symbol,
	}
	result, err := pa.registry.FetchWithFallback(ctx, provider.ModelEtfInfo, params)
	if err != nil {
		return nil, fmt.Errorf("ETF info for %s: %w", symbol, err)
	}
	if info, ok := result.Data.(*models.ETFInfo); ok {
		return info, nil
	}
	return nil, fmt.Errorf("unexpected data type for ETF info")
}

// FetchIndexConstituents fetches index constituents via providers.
func (pa *ProviderAggregator) FetchIndexConstituents(ctx context.Context, symbol string) ([]models.IndexConstituent, error) {
	params := provider.QueryParams{
		provider.ParamSymbol: symbol,
	}
	result, err := pa.registry.FetchWithFallback(ctx, provider.ModelIndexConstituents, params)
	if err != nil {
		return nil, fmt.Errorf("index constituents for %s: %w", symbol, err)
	}
	if ic, ok := result.Data.([]models.IndexConstituent); ok {
		return ic, nil
	}
	return nil, fmt.Errorf("unexpected data type for index constituents")
}

// FetchEconomicCalendar fetches economic calendar events via providers.
func (pa *ProviderAggregator) FetchEconomicCalendar(ctx context.Context, from, to time.Time, country string) ([]models.EconomicCalendarEvent, error) {
	params := provider.QueryParams{
		provider.ParamStartDate: from.Format("2006-01-02"),
		provider.ParamEndDate:   to.Format("2006-01-02"),
	}
	if country != "" {
		params[provider.ParamCountry] = country
	}
	result, err := pa.registry.FetchWithFallback(ctx, provider.ModelEconomicCalendar, params)
	if err != nil {
		return nil, fmt.Errorf("economic calendar: %w", err)
	}
	if events, ok := result.Data.([]models.EconomicCalendarEvent); ok {
		return events, nil
	}
	return nil, fmt.Errorf("unexpected data type for economic calendar")
}

// FetchCompanyNews fetches company news, trying providers first, then legacy.
func (pa *ProviderAggregator) FetchCompanyNews(ctx context.Context, ticker string, limit int) ([]models.NewsArticle, error) {
	symbol := utils.NormalizeTicker(ticker)

	// Try provider registry for rich news data.
	params := provider.QueryParams{
		provider.ParamSymbol: symbol,
		provider.ParamLimit:  fmt.Sprintf("%d", limit),
	}
	result, err := pa.registry.Fetch(ctx, provider.ModelCompanyNews, params)
	if err == nil {
		// Convert from CompanyNewsArticle to legacy NewsArticle if needed.
		if articles, ok := result.Data.([]models.CompanyNewsArticle); ok {
			legacy := make([]models.NewsArticle, len(articles))
			for i, a := range articles {
				legacy[i] = models.NewsArticle{
					Title:       a.Title,
					URL:         a.URL,
					Source:      a.Source,
					Summary:     a.Summary,
					PublishedAt: a.PublishedAt,
					Tickers:     []string{a.Symbol},
				}
			}
			return legacy, nil
		}
		if articles, ok := result.Data.([]models.NewsArticle); ok {
			return articles, nil
		}
	}

	// Fall back to legacy RSS news.
	return pa.news.GetStockNews(ctx, ticker, limit)
}
