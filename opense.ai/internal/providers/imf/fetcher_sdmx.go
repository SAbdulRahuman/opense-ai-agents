package imf

import (
	"context"
	"fmt"
	"net/url"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// AvailableIndicators — IMF metadata query (simplified: returns known dataflows).
// ---------------------------------------------------------------------------

type availableIndicatorsFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newAvailableIndicatorsFetcher(p *Provider) *availableIndicatorsFetcher {
	return &availableIndicatorsFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelAvailableIndicators,
			"IMF available economic indicator dataflows",
			nil,
			nil,
		),
		p: p,
	}
}

// knownIMFDataflows lists the most commonly used IMF SDMX dataflows.
var knownIMFDataflows = []models.AvailableEconomicIndicator{
	{ID: "CPI", Name: "Consumer Price Index", Category: "prices", Source: "IMF"},
	{ID: "IMTS", Name: "International Merchandise Trade Statistics", Category: "trade", Source: "IMF"},
	{ID: "WEO", Name: "World Economic Outlook", Category: "macro", Source: "IMF"},
	{ID: "BOP", Name: "Balance of Payments", Category: "external", Source: "IMF"},
	{ID: "IFS", Name: "International Financial Statistics", Category: "finance", Source: "IMF"},
	{ID: "GFS", Name: "Government Finance Statistics", Category: "fiscal", Source: "IMF"},
	{ID: "DOT", Name: "Direction of Trade Statistics", Category: "trade", Source: "IMF"},
	{ID: "GFSR", Name: "Global Financial Stability Report", Category: "finance", Source: "IMF"},
	{ID: "FSI", Name: "Financial Soundness Indicators", Category: "finance", Source: "IMF"},
	{ID: "COFER", Name: "Currency Composition of Official Foreign Exchange Reserves", Category: "reserves", Source: "IMF"},
	{ID: "FAS", Name: "Financial Access Survey", Category: "finance", Source: "IMF"},
	{ID: "MFS", Name: "Monetary and Financial Statistics", Category: "monetary", Source: "IMF"},
}

func (f *availableIndicatorsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	return newResult(knownIMFDataflows), nil
}

// ---------------------------------------------------------------------------
// ConsumerPriceIndex (CPI) — IMF SDMX
// Dataflow: CPI
// Key: {COUNTRY}.CPI._T.IX.M   (CPI, all items, index, monthly)
// ---------------------------------------------------------------------------

type imfCPIFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newIMFCPIFetcher(p *Provider) *imfCPIFetcher {
	return &imfCPIFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelConsumerPriceIndex,
			"IMF Consumer Price Index (global coverage, 196 countries)",
			nil,
			[]string{provider.ParamCountry, provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *imfCPIFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelConsumerPriceIndex, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	country := resolveCountry(params[provider.ParamCountry])
	if country == "*" {
		country = "USA"
	}

	key := country + ".CPI._T.IX.M"
	sdmxParams := make(map[string]string)
	if start := params[provider.ParamStartDate]; start != "" {
		if end := params[provider.ParamEndDate]; end != "" {
			sdmxParams["c[TIME_PERIOD]"] = "ge:" + start + "+le:" + end
		} else {
			sdmxParams["c[TIME_PERIOD]"] = "ge:" + start
		}
	}

	u := buildSDMXURL("IMF", "CPI", key, sdmxParams)

	var resp sdmxDataResponse
	if err := f.p.fetchJSON(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("IMF CPI: %w", err)
	}

	data := extractCPIFromSDMX(resp, country)
	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// extractCPIFromSDMX extracts CPI data from the SDMX JSON response.
func extractCPIFromSDMX(resp sdmxDataResponse, country string) []models.CPIData {
	var data []models.CPIData

	if len(resp.Data.DataSets) == 0 {
		return data
	}

	ds := resp.Data.DataSets[0]
	for _, series := range ds.Series {
		for timePeriod, obs := range series.Observations {
			if len(obs) == 0 {
				continue
			}
			val := parseAnyFloat(obs[0])
			if val == 0 {
				continue
			}
			data = append(data, models.CPIData{
				Date:    parseAnyDate(timePeriod),
				Country: country,
				Value:   val,
			})
		}
	}
	return data
}

// ---------------------------------------------------------------------------
// DirectionOfTrade — IMF SDMX (IMTS dataflow)
// Key: {COUNTRY}.{INDICATOR}.{FREQ}.{COUNTERPART}
// ---------------------------------------------------------------------------

type directionOfTradeFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newDirectionOfTradeFetcher(p *Provider) *directionOfTradeFetcher {
	return &directionOfTradeFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelDirectionOfTrade,
			"IMF Direction of Trade statistics (imports/exports between countries)",
			nil,
			[]string{provider.ParamCountry, provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *directionOfTradeFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelDirectionOfTrade, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	country := resolveCountry(params[provider.ParamCountry])
	if country == "*" {
		country = "USA"
	}

	// Default to exports (FOB) unless specified.
	indicator := "XG_FOB_USD"
	if d := params["direction"]; d != "" {
		switch d {
		case "imports":
			indicator = "MG_CIF_USD"
		case "balance":
			indicator = "TBG_USD"
		case "all":
			indicator = "*"
		}
	}

	counterpart := resolveCountry(params["counterpart"])
	if counterpart == "" {
		counterpart = "*"
	}

	freq := "A"
	if f2 := params["frequency"]; f2 == "quarterly" || f2 == "quarter" {
		freq = "Q"
	} else if f2 == "monthly" || f2 == "month" {
		freq = "M"
	}

	key := country + "." + indicator + "." + freq + "." + counterpart
	sdmxParams := make(map[string]string)
	if start := params[provider.ParamStartDate]; start != "" {
		if end := params[provider.ParamEndDate]; end != "" {
			sdmxParams["c[TIME_PERIOD]"] = "ge:" + start + "+le:" + end
		} else {
			sdmxParams["c[TIME_PERIOD]"] = "ge:" + start
		}
	}

	u := buildSDMXURL("IMF", "IMTS", key, sdmxParams)

	var resp sdmxDataResponse
	if err := f.p.fetchJSON(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("IMF direction of trade: %w", err)
	}

	data := extractTradeFromSDMX(resp, country)
	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

func extractTradeFromSDMX(resp sdmxDataResponse, country string) []models.BalanceOfPaymentsData {
	var data []models.BalanceOfPaymentsData

	if len(resp.Data.DataSets) == 0 {
		return data
	}

	ds := resp.Data.DataSets[0]
	for _, series := range ds.Series {
		for timePeriod, obs := range series.Observations {
			if len(obs) == 0 {
				continue
			}
			val := parseAnyFloat(obs[0])
			if val == 0 {
				continue
			}
			data = append(data, models.BalanceOfPaymentsData{
				Date:         parseAnyDate(timePeriod),
				Country:      country,
				TradeBalance: val,
				Currency:     "USD",
			})
		}
	}
	return data
}

// ---------------------------------------------------------------------------
// EconomicIndicators — generic IMF SDMX query by symbol.
// Symbol format: "{DATAFLOW}::{KEY}" e.g. "WEO::NGDP_RPCH"
// ---------------------------------------------------------------------------

type economicIndicatorsFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newEconomicIndicatorsFetcher(p *Provider) *economicIndicatorsFetcher {
	return &economicIndicatorsFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelEconomicIndicators,
			"IMF economic indicators (any dataflow, universal query)",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamCountry, provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *economicIndicatorsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelEconomicIndicators, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	// Parse symbol: DATAFLOW::INDICATOR
	symbol := params[provider.ParamSymbol]
	parts := splitSymbol(symbol)
	if len(parts) < 2 {
		return nil, fmt.Errorf("IMF economic indicators: symbol must be 'DATAFLOW::INDICATOR', got %q", symbol)
	}
	dataflow := parts[0]
	indicator := parts[1]

	country := resolveCountry(params[provider.ParamCountry])
	if country == "" {
		country = "*"
	}

	// Build a key; for WEO it's: {COUNTRY}.{INDICATOR}
	// For other dataflows, use indicator as-is in the key.
	key := country + "." + url.PathEscape(indicator)

	sdmxParams := make(map[string]string)
	if start := params[provider.ParamStartDate]; start != "" {
		if end := params[provider.ParamEndDate]; end != "" {
			sdmxParams["c[TIME_PERIOD]"] = "ge:" + start + "+le:" + end
		} else {
			sdmxParams["c[TIME_PERIOD]"] = "ge:" + start
		}
	}

	u := buildSDMXURL("IMF", dataflow, key, sdmxParams)

	var resp sdmxDataResponse
	if err := f.p.fetchJSON(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("IMF economic indicators (%s): %w", dataflow, err)
	}

	data := extractEconomicFromSDMX(resp, country)
	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

func extractEconomicFromSDMX(resp sdmxDataResponse, country string) []models.EconomicIndicatorData {
	var data []models.EconomicIndicatorData

	if len(resp.Data.DataSets) == 0 {
		return data
	}

	ds := resp.Data.DataSets[0]
	for _, series := range ds.Series {
		for timePeriod, obs := range series.Observations {
			if len(obs) == 0 {
				continue
			}
			val := parseAnyFloat(obs[0])
			if val == 0 {
				continue
			}
			data = append(data, models.EconomicIndicatorData{
				Date:    parseAnyDate(timePeriod),
				Country: country,
				Value:   val,
			})
		}
	}
	return data
}

// splitSymbol splits a symbol of form "DATAFLOW::INDICATOR" into parts.
func splitSymbol(s string) []string {
	idx := -1
	for i := 0; i < len(s)-1; i++ {
		if s[i] == ':' && s[i+1] == ':' {
			idx = i
			break
		}
	}
	if idx < 0 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+2:]}
}
