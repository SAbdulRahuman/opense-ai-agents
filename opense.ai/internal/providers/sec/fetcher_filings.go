package sec

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---- CompanyFilings fetcher ----
// Queries EDGAR full-text search for company filings.

type companyFilingsFetcher struct {
	provider.BaseFetcher
}

func newCompanyFilingsFetcher() *companyFilingsFetcher {
	return &companyFilingsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCompanyFilings,
			"Search SEC EDGAR for company filings by ticker, CIK, or keyword",
			[]string{provider.ParamQuery},
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			10*time.Minute, 8, time.Second,
		),
	}
}

func (f *companyFilingsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	query := params[provider.ParamQuery]
	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("%s/search-index?q=%s&dateRange=custom", edgarBaseURL, url.QueryEscape(query))
	if sd := params[provider.ParamStartDate]; sd != "" {
		u += "&startdt=" + sd
	}
	if ed := params[provider.ParamEndDate]; ed != "" {
		u += "&enddt=" + ed
	}

	var resp edgarSearchResponse
	if err := fetchSECJSON(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("sec company filings search: %w", err)
	}

	var filings []models.CompanyFiling
	for _, hit := range resp.Hits.Hits {
		doc := hit.Source
		symbol := ""
		cik := ""
		if len(doc.Tickers) > 0 {
			symbol = doc.Tickers[0]
		}
		if len(doc.CIKs) > 0 {
			cik = doc.CIKs[0]
		}
		filings = append(filings, models.CompanyFiling{
			Date:        parseSECDate(doc.FiledAt),
			Symbol:      symbol,
			CIK:         cik,
			CompanyName: doc.EntityName,
			FormType:    doc.FormType,
			AccessionNo: hit.ID,
			Description: doc.FileDescription,
		})
	}

	f.CacheSet(cacheKey, filings)
	return newResult(filings), nil
}

// ---- SecFiling fetcher ----
// Retrieves a specific company's filing list from EDGAR submissions API.

type secFilingFetcher struct {
	provider.BaseFetcher
}

func newSecFilingFetcher() *secFilingFetcher {
	return &secFilingFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelSecFiling,
			"Get SEC filing details for a specific CIK or ticker",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamLimit},
			10*time.Minute, 8, time.Second,
		),
	}
}

func (f *secFilingFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Resolve symbol to CIK.
	cik, err := resolveSymbolToCIK(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("sec filing resolve CIK for %s: %w", symbol, err)
	}

	u := fmt.Sprintf("%s/CIK%s.json", edgarSubmissions, padCIK(cik))
	var resp edgarSubmissionsResponse
	if err := fetchSECJSON(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("sec filing submissions: %w", err)
	}

	limit := 100
	if lim := params[provider.ParamLimit]; lim != "" {
		if n := parseInt(lim); n > 0 {
			limit = n
		}
	}

	recent := resp.Filings.Recent
	n := len(recent.AccessionNumber)
	if n > limit {
		n = limit
	}

	var filings []models.CompanyFiling
	for i := 0; i < n; i++ {
		accNo := recent.AccessionNumber[i]
		accNoClean := strings.ReplaceAll(accNo, "-", "")
		filingURL := fmt.Sprintf("https://www.sec.gov/Archives/edgar/data/%s/%s/%s",
			resp.CIK, accNoClean, recent.PrimaryDocument[i])

		desc := ""
		if i < len(recent.Description) {
			desc = recent.Description[i]
		}

		filings = append(filings, models.CompanyFiling{
			Date:        parseSECDate(recent.FilingDate[i]),
			Symbol:      symbol,
			CIK:         resp.CIK,
			CompanyName: resp.Name,
			FormType:    recent.Form[i],
			AccessionNo: accNo,
			FilingURL:   filingURL,
			Description: desc,
		})
	}

	f.CacheSet(cacheKey, filings)
	return newResult(filings), nil
}

// ---- LatestFinancialReports fetcher ----
// Returns recent 10-K and 10-Q filings via EDGAR full-text search.

type latestFinancialReportsFetcher struct {
	provider.BaseFetcher
}

func newLatestFinancialReportsFetcher() *latestFinancialReportsFetcher {
	return &latestFinancialReportsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelLatestFinancialReports,
			"Latest 10-K and 10-Q financial reports from SEC EDGAR",
			nil,
			[]string{provider.ParamSymbol, provider.ParamLimit},
			10*time.Minute, 8, time.Second,
		),
	}
}

func (f *latestFinancialReportsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	query := `forms:"10-K,10-Q"`
	if sym := params[provider.ParamSymbol]; sym != "" {
		query = fmt.Sprintf(`%s AND tickers:"%s"`, query, sym)
	}

	u := fmt.Sprintf("%s/search-index?q=%s", edgarBaseURL, url.QueryEscape(query))

	var resp edgarSearchResponse
	if err := fetchSECJSON(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("sec latest financial reports: %w", err)
	}

	var filings []models.CompanyFiling
	for _, hit := range resp.Hits.Hits {
		doc := hit.Source
		symbol := ""
		cik := ""
		if len(doc.Tickers) > 0 {
			symbol = doc.Tickers[0]
		}
		if len(doc.CIKs) > 0 {
			cik = doc.CIKs[0]
		}
		filings = append(filings, models.CompanyFiling{
			Date:        parseSECDate(doc.FiledAt),
			Symbol:      symbol,
			CIK:         cik,
			CompanyName: doc.EntityName,
			FormType:    doc.FormType,
			AccessionNo: hit.ID,
			Description: doc.FileDescription,
		})
	}

	f.CacheSet(cacheKey, filings)
	return newResult(filings), nil
}

// ---- RssLitigation fetcher ----
// Returns SEC litigation releases (from RSS feed or EDGAR search).

type rssLitigationFetcher struct {
	provider.BaseFetcher
}

func newRssLitigationFetcher() *rssLitigationFetcher {
	return &rssLitigationFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelRssLitigation,
			"SEC litigation releases and administrative proceedings",
			nil,
			[]string{provider.ParamLimit},
			10*time.Minute, 8, time.Second,
		),
	}
}

func (f *rssLitigationFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Use EDGAR full-text search for litigation releases.
	u := fmt.Sprintf("%s/search-index?q=%s&forms=LIT",
		edgarBaseURL, url.QueryEscape("litigation release"))

	var resp edgarSearchResponse
	if err := fetchSECJSON(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("sec rss litigation: %w", err)
	}

	var filings []models.CompanyFiling
	for _, hit := range resp.Hits.Hits {
		doc := hit.Source
		cik := ""
		if len(doc.CIKs) > 0 {
			cik = doc.CIKs[0]
		}
		filings = append(filings, models.CompanyFiling{
			Date:        parseSECDate(doc.FiledAt),
			CIK:         cik,
			CompanyName: doc.EntityName,
			FormType:    doc.FormType,
			AccessionNo: hit.ID,
			Description: doc.FileDescription,
		})
	}

	f.CacheSet(cacheKey, filings)
	return newResult(filings), nil
}

// --- helpers ---

// resolveSymbolToCIK resolves a ticker symbol to a CIK number using SEC tickers JSON.
func resolveSymbolToCIK(ctx context.Context, symbol string) (string, error) {
	u := edgarDataURL + "/files/company_tickers.json"
	var tickers map[string]edgarTickerEntry
	if err := fetchSECJSON(ctx, u, &tickers); err != nil {
		return "", fmt.Errorf("fetch company tickers: %w", err)
	}

	sym := strings.ToUpper(strings.TrimSpace(symbol))
	for _, entry := range tickers {
		if strings.EqualFold(entry.Ticker, sym) {
			return entry.CIKStr, nil
		}
	}
	// If symbol looks like a CIK number already, return as-is.
	if isNumeric(sym) {
		return sym, nil
	}
	return "", fmt.Errorf("CIK not found for symbol %s", symbol)
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	return n
}
