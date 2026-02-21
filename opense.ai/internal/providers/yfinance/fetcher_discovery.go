package yfinance

import (
	"context"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// YF screener predefined list IDs.
const (
	yfScreenerGainers = "day_gainers"
	yfScreenerLosers  = "day_losers"
	yfScreenerActive  = "most_actives"
)

// --- EquityGainers fetcher ---

type equityGainersFetcher struct {
	provider.BaseFetcher
}

func newEquityGainersFetcher() *equityGainersFetcher {
	return &equityGainersFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquityGainers,
			"Top gaining stocks from Yahoo Finance",
			nil,
			[]string{provider.ParamLimit},
			5*time.Minute, 3, time.Second,
		),
	}
}

func (f *equityGainersFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	return fetchScreenerList(ctx, f, yfScreenerGainers, params)
}

// --- EquityLosers fetcher ---

type equityLosersFetcher struct {
	provider.BaseFetcher
}

func newEquityLosersFetcher() *equityLosersFetcher {
	return &equityLosersFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquityLosers,
			"Top losing stocks from Yahoo Finance",
			nil,
			[]string{provider.ParamLimit},
			5*time.Minute, 3, time.Second,
		),
	}
}

func (f *equityLosersFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	return fetchScreenerList(ctx, f, yfScreenerLosers, params)
}

// --- EquityActive fetcher ---

type equityActiveFetcher struct {
	provider.BaseFetcher
}

func newEquityActiveFetcher() *equityActiveFetcher {
	return &equityActiveFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquityActive,
			"Most active stocks from Yahoo Finance",
			nil,
			[]string{provider.ParamLimit},
			5*time.Minute, 3, time.Second,
		),
	}
}

func (f *equityActiveFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	return fetchScreenerList(ctx, f, yfScreenerActive, params)
}

// --- CompanyNews fetcher ---

type companyNewsFetcher struct {
	provider.BaseFetcher
}

func newCompanyNewsFetcher() *companyNewsFetcher {
	return &companyNewsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCompanyNews,
			"Company-specific news from Yahoo Finance",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamLimit},
			10*time.Minute, 5, time.Second,
		),
	}
}

func (f *companyNewsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	yfTicker := toYFTicker(symbol)

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v1/finance/search?q=%s&quotesCount=0&newsCount=20",
		yfTicker,
	)

	var resp yfSearchResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("yfinance news %s: %w", yfTicker, err)
	}

	articles := make([]models.CompanyNewsArticle, 0, len(resp.News))
	for _, n := range resp.News {
		articles = append(articles, models.CompanyNewsArticle{
			Title:     n.Title,
			Source:    n.Publisher,
			URL:       n.Link,
		})
	}

	f.CacheSetTTL(cacheKey, articles, 10*time.Minute)
	return newResult(articles), nil
}

// --- Shared screener helper ---

type screenerFetcher interface {
	CacheGet(key string) (any, bool)
	CacheSetTTL(key string, value any, ttl time.Duration)
	RateLimit(ctx context.Context) error
	ModelType() provider.ModelType
}

func fetchScreenerList(ctx context.Context, f screenerFetcher, listID string, params provider.QueryParams) (*provider.FetchResult, error) {
	cacheKey := string(f.ModelType()) + ":" + listID
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v1/finance/screener/predefined/saved?formatted=false&scrIds=%s&count=25",
		listID,
	)

	var resp yfScreenerResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("yfinance screener %s: %w", listID, err)
	}
	if resp.Finance.Error != nil {
		return nil, fmt.Errorf("yfinance screener error: %s", resp.Finance.Error.Description)
	}

	movers := make([]models.MarketMover, 0)
	for _, result := range resp.Finance.Result {
		for _, q := range result.Quotes {
			movers = append(movers, models.MarketMover{
				Symbol:    q.Symbol,
				Name:      coalesce(q.LongName, q.ShortName),
				Price:     q.RegularMarketPrice,
				Change:    q.RegularMarketChange,
				ChangePct: q.RegularMarketChangePercent,
				Volume:    q.RegularMarketVolume,
				MarketCap: q.MarketCap,
			})
		}
	}

	f.CacheSetTTL(cacheKey, movers, 5*time.Minute)
	return newResult(movers), nil
}
