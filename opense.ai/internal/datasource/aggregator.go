package datasource

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/seenimoa/openseai/pkg/models"
	"github.com/seenimoa/openseai/pkg/utils"
)

// Aggregator fetches and merges data from multiple sources concurrently.
type Aggregator struct {
	yfinance    *YFinance
	nse         *NSE
	derivatives *NSEDerivatives
	screener    *Screener
	news        *News
	fiidii      *FIIDII
}

// NewAggregator creates a new data source aggregator with all default sources.
func NewAggregator() *Aggregator {
	nse := NewNSE()
	return &Aggregator{
		yfinance:    NewYFinance(),
		nse:         nse,
		derivatives: NewNSEDerivatives(nse),
		screener:    NewScreener(),
		news:        NewNews(),
		fiidii:      NewFIIDII(nse),
	}
}

// Sources returns all registered data sources.
func (a *Aggregator) Sources() []DataSource {
	return []DataSource{
		a.yfinance,
		a.nse,
		a.derivatives,
		a.screener,
		a.news,
		a.fiidii,
	}
}

// YFinance returns the Yahoo Finance source for direct access.
func (a *Aggregator) YFinance() *YFinance { return a.yfinance }

// NSE returns the NSE source for direct access.
func (a *Aggregator) NSE() *NSE { return a.nse }

// Derivatives returns the NSE derivatives source for direct access.
func (a *Aggregator) Derivatives() *NSEDerivatives { return a.derivatives }

// Screener returns the Screener.in source for direct access.
func (a *Aggregator) Screener() *Screener { return a.screener }

// NewsSource returns the news source for direct access.
func (a *Aggregator) NewsSource() *News { return a.news }

// FIIDII returns the FII/DII source for direct access.
func (a *Aggregator) FIIDII() *FIIDII { return a.fiidii }

// FetchProfile fetches a comprehensive stock profile by aggregating data
// from all available sources concurrently.
func (a *Aggregator) FetchProfile(ctx context.Context, ticker string) (*models.StockProfile, error) {
	symbol := utils.NormalizeTicker(ticker)

	profile := &models.StockProfile{
		Stock: models.Stock{
			Ticker:   symbol,
			Exchange: "NSE",
		},
		FetchedAt: utils.NowIST(),
	}

	var mu sync.Mutex
	var errs []error

	g, gctx := errgroup.WithContext(ctx)

	// 1. Quote from Yahoo Finance (primary) with NSE as fallback.
	g.Go(func() error {
		quote, err := a.yfinance.GetQuote(gctx, symbol)
		if err != nil {
			// Try NSE as fallback.
			quote, err = a.nse.GetQuote(gctx, symbol)
		}
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("quote: %w", err))
			mu.Unlock()
			return nil // non-fatal
		}
		mu.Lock()
		profile.Quote = quote
		profile.Stock.Name = quote.Name
		profile.Stock.MarketCap = quote.MarketCap
		mu.Unlock()
		return nil
	})

	// 2. Financials from Screener.in.
	g.Go(func() error {
		fd, err := a.screener.GetFinancials(gctx, symbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("financials: %w", err))
			mu.Unlock()
			return nil
		}
		mu.Lock()
		profile.Financials = fd
		mu.Unlock()
		return nil
	})

	// 3. Financial ratios from Screener.in.
	g.Go(func() error {
		ratios, err := a.screener.GetFinancialRatios(gctx, symbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("ratios: %w", err))
			mu.Unlock()
			return nil
		}
		mu.Lock()
		profile.Ratios = ratios
		mu.Unlock()
		return nil
	})

	// 4. Shareholding from NSE.
	g.Go(func() error {
		pd, err := a.nse.GetShareholding(gctx, symbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("shareholding: %w", err))
			mu.Unlock()
			return nil
		}
		mu.Lock()
		profile.Promoter = pd
		mu.Unlock()
		return nil
	})

	// Wait for all goroutines.
	if err := g.Wait(); err != nil {
		return profile, err
	}

	// If we got at least a quote, consider it a success.
	if profile.Quote == nil && len(errs) > 0 {
		return nil, fmt.Errorf("all sources failed for %s: %w", symbol, errors.Join(errs...))
	}

	return profile, nil
}

// FetchHistoricalData fetches OHLCV data, trying Yahoo Finance first, then NSE.
func (a *Aggregator) FetchHistoricalData(ctx context.Context, ticker string, from, to time.Time, tf models.Timeframe) ([]models.OHLCV, error) {
	// Try YFinance first (better historical data coverage).
	candles, err := a.yfinance.GetHistoricalData(ctx, ticker, from, to, tf)
	if err == nil && len(candles) > 0 {
		return candles, nil
	}

	// Fallback to NSE.
	candles, err = a.nse.GetHistoricalData(ctx, ticker, from, to, tf)
	if err != nil {
		return nil, fmt.Errorf("historical data unavailable for %s: %w", ticker, err)
	}
	return candles, nil
}

// FetchOptionChain fetches the option chain from NSE derivatives.
func (a *Aggregator) FetchOptionChain(ctx context.Context, ticker string, expiry string) (*models.OptionChain, error) {
	return a.derivatives.GetOptionChain(ctx, ticker, expiry)
}

// FetchMarketOverview returns a market overview with indices, VIX, and FII/DII data.
func (a *Aggregator) FetchMarketOverview(ctx context.Context) (*MarketOverview, error) {
	overview := &MarketOverview{
		FetchedAt: utils.NowIST(),
	}

	var mu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)

	// VIX.
	g.Go(func() error {
		vix, err := a.derivatives.GetIndiaVIX(gctx)
		if err != nil {
			return nil // non-fatal
		}
		mu.Lock()
		overview.IndiaVIX = vix
		mu.Unlock()
		return nil
	})

	// FII/DII.
	g.Go(func() error {
		fd, err := a.fiidii.GetFIIDIIActivity(gctx)
		if err != nil {
			return nil
		}
		mu.Lock()
		overview.FIIDII = fd
		mu.Unlock()
		return nil
	})

	// NIFTY 50 quote.
	g.Go(func() error {
		q, err := a.yfinance.GetQuote(gctx, "NIFTY50")
		if err != nil {
			return nil
		}
		mu.Lock()
		overview.Nifty50 = q
		mu.Unlock()
		return nil
	})

	// Bank NIFTY.
	g.Go(func() error {
		q, err := a.yfinance.GetQuote(gctx, "BANKNIFTY")
		if err != nil {
			return nil
		}
		mu.Lock()
		overview.BankNifty = q
		mu.Unlock()
		return nil
	})

	if err := g.Wait(); err != nil {
		return overview, err
	}

	return overview, nil
}

// FetchStockNews returns recent news for a ticker.
func (a *Aggregator) FetchStockNews(ctx context.Context, ticker string, limit int) ([]models.NewsArticle, error) {
	return a.news.GetStockNews(ctx, ticker, limit)
}

// MarketOverview holds a snapshot of the overall market state.
type MarketOverview struct {
	Nifty50   *models.Quote     `json:"nifty50,omitempty"`
	BankNifty *models.Quote     `json:"bank_nifty,omitempty"`
	IndiaVIX  *models.IndiaVIX  `json:"india_vix,omitempty"`
	FIIDII    *models.FIIDIIData `json:"fii_dii,omitempty"`
	FetchedAt time.Time          `json:"fetched_at"`
}
