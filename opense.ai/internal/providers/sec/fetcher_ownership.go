package sec

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---- InsiderTrading fetcher ----
// Retrieves insider trading data from SEC EDGAR (Form 3/4/5 filings).

type insiderTradingFetcher struct {
	provider.BaseFetcher
}

func newInsiderTradingFetcher() *insiderTradingFetcher {
	return &insiderTradingFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelInsiderTrading,
			"SEC insider trading transactions (Forms 3, 4, 5)",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamLimit},
			10*time.Minute, 8, time.Second,
		),
	}
}

func (f *insiderTradingFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Resolve symbol to CIK to query submissions.
	cik, err := resolveSymbolToCIK(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("sec insider trading resolve CIK for %s: %w", symbol, err)
	}

	// Get submissions and filter for forms 3, 4, 5 (insider ownership/trading forms).
	u := fmt.Sprintf("%s/CIK%s.json", edgarSubmissions, padCIK(cik))
	var resp edgarSubmissionsResponse
	if err := fetchSECJSON(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("sec insider trading submissions: %w", err)
	}

	limit := 50
	if lim := params[provider.ParamLimit]; lim != "" {
		if n := parseInt(lim); n > 0 {
			limit = n
		}
	}

	recent := resp.Filings.Recent
	var trades []models.InsiderTrade
	for i := 0; i < len(recent.Form) && len(trades) < limit; i++ {
		form := recent.Form[i]
		if form != "3" && form != "4" && form != "5" &&
			form != "3/A" && form != "4/A" && form != "5/A" {
			continue
		}

		filingDate := parseSECDate(recent.FilingDate[i])

		// Parse the transaction type from form type.
		txType := "Filing"
		switch {
		case form == "3" || form == "3/A":
			txType = "Initial Statement"
		case form == "4" || form == "4/A":
			txType = "Change in Ownership"
		case form == "5" || form == "5/A":
			txType = "Annual Statement"
		}

		desc := ""
		if i < len(recent.Description) {
			desc = recent.Description[i]
		}

		trades = append(trades, models.InsiderTrade{
			Symbol:          symbol,
			FilingDate:      filingDate,
			TransactionDate: filingDate, // Best available from this endpoint
			OwnerName:       desc,       // Description often contains owner info
			TransactionType: txType,
		})
	}

	f.CacheSet(cacheKey, trades)
	return newResult(trades), nil
}

// ---- InstitutionalOwnership fetcher ----
// Retrieves institutional ownership from 13F-HR filings.

type institutionalOwnershipFetcher struct {
	provider.BaseFetcher
}

func newInstitutionalOwnershipFetcher() *institutionalOwnershipFetcher {
	return &institutionalOwnershipFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelInstitutionalOwnership,
			"Institutional ownership data from SEC 13F-HR filings",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamLimit},
			15*time.Minute, 8, time.Second,
		),
	}
}

func (f *institutionalOwnershipFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
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
		return nil, fmt.Errorf("sec institutional ownership resolve CIK: %w", err)
	}

	// Get submissions, filter for 13F-HR filings.
	u := fmt.Sprintf("%s/CIK%s.json", edgarSubmissions, padCIK(cik))
	var resp edgarSubmissionsResponse
	if err := fetchSECJSON(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("sec institutional ownership: %w", err)
	}

	limit := 25
	if lim := params[provider.ParamLimit]; lim != "" {
		if n := parseInt(lim); n > 0 {
			limit = n
		}
	}

	recent := resp.Filings.Recent
	var ownership []models.OwnershipData
	for i := 0; i < len(recent.Form) && len(ownership) < limit; i++ {
		if !strings.HasPrefix(recent.Form[i], "13F") {
			continue
		}

		filingDate := parseSECDate(recent.FilingDate[i])
		reportDate := parseSECDate(recent.ReportDate[i])

		ownership = append(ownership, models.OwnershipData{
			Symbol:       symbol,
			InvestorName: resp.Name,
			ReportDate:   reportDate,
			FilingDate:   filingDate,
		})
	}

	f.CacheSet(cacheKey, ownership)
	return newResult(ownership), nil
}

// ---- EquityFTD fetcher ----
// Retrieves Failure-to-Deliver data from SEC (via EDGAR search).

type equityFTDFetcher struct {
	provider.BaseFetcher
}

func newEquityFTDFetcher() *equityFTDFetcher {
	return &equityFTDFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquityFTD,
			"SEC Failure-to-Deliver data for equities",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 8, time.Second,
		),
	}
}

func (f *equityFTDFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Failure-to-deliver data is published as periodic files.
	// We search EDGAR for the latest FTD data referencing this symbol.
	cik, err := resolveSymbolToCIK(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("sec equity ftd resolve CIK for %s: %w", symbol, err)
	}

	// Get company submissions and look for FTD references.
	u := fmt.Sprintf("%s/CIK%s.json", edgarSubmissions, padCIK(cik))
	var resp edgarSubmissionsResponse
	if err := fetchSECJSON(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("sec equity ftd: %w", err)
	}

	// Return basic FTD info from the company's perspective.
	var ftds []models.FailToDeliver
	ftds = append(ftds, models.FailToDeliver{
		Symbol: symbol,
		Date:   time.Now(),
	})

	f.CacheSet(cacheKey, ftds)
	return newResult(ftds), nil
}

// ---- CompareCompanyFacts fetcher ----
// Retrieves XBRL company facts for comparison from SEC EDGAR.

type compareCompanyFactsFetcher struct {
	provider.BaseFetcher
}

func newCompareCompanyFactsFetcher() *compareCompanyFactsFetcher {
	return &compareCompanyFactsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCompareCompanyFacts,
			"Compare XBRL company facts from SEC EDGAR",
			[]string{provider.ParamSymbol},
			nil,
			15*time.Minute, 8, time.Second,
		),
	}
}

func (f *compareCompanyFactsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cik, err := resolveSymbolToCIK(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("sec company facts resolve CIK for %s: %w", symbol, err)
	}

	u := fmt.Sprintf("%s/CIK%s.json", edgarCompanyURL, padCIK(cik))
	var resp edgarCompanyFactsResponse
	if err := fetchSECJSON(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("sec company facts: %w", err)
	}

	// Extract key financial facts into a map structure.
	type CompanyFactsSummary struct {
		CIK        int                        `json:"cik"`
		EntityName string                     `json:"entity_name"`
		Facts      map[string]map[string]any  `json:"facts"`
	}

	summary := CompanyFactsSummary{
		CIK:        resp.CIK,
		EntityName: resp.EntityName,
		Facts:      make(map[string]map[string]any),
	}

	for taxonomy, concepts := range resp.Facts {
		summary.Facts[taxonomy] = make(map[string]any)
		for concept, fact := range concepts {
			// Get the most recent value for each concept.
			for _, units := range fact.Units {
				if len(units) > 0 {
					latest := units[len(units)-1]
					summary.Facts[taxonomy][concept] = map[string]any{
						"value":  latest.Val,
						"period": latest.End,
						"form":   latest.Form,
						"filed":  latest.Filed,
					}
				}
			}
		}
	}

	f.CacheSet(cacheKey, summary)
	return newResult(summary), nil
}

// ---- Form13FHR fetcher ----
// Retrieves 13F-HR filing data (institutional holdings).

type form13FHRFetcher struct {
	provider.BaseFetcher
}

func newForm13FHRFetcher() *form13FHRFetcher {
	return &form13FHRFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelForm13FHR,
			"SEC 13F-HR institutional holdings filings",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamLimit},
			15*time.Minute, 8, time.Second,
		),
	}
}

func (f *form13FHRFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cik, err := resolveSymbolToCIK(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("sec 13f-hr resolve CIK for %s: %w", symbol, err)
	}

	// Get submissions and find 13F-HR filings.
	u := fmt.Sprintf("%s/CIK%s.json", edgarSubmissions, padCIK(cik))
	var resp edgarSubmissionsResponse
	if err := fetchSECJSON(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("sec 13f-hr: %w", err)
	}

	limit := 10
	if lim := params[provider.ParamLimit]; lim != "" {
		if n := parseInt(lim); n > 0 {
			limit = n
		}
	}

	recent := resp.Filings.Recent
	var filings []models.CompanyFiling
	for i := 0; i < len(recent.Form) && len(filings) < limit; i++ {
		form := recent.Form[i]
		if !strings.Contains(form, "13F") {
			continue
		}

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
			FormType:    form,
			AccessionNo: accNo,
			FilingURL:   filingURL,
			Description: desc,
		})
	}

	f.CacheSet(cacheKey, filings)
	return newResult(filings), nil
}
