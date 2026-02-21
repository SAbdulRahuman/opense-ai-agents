package oecd

import (
	"context"
	"fmt"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// CompositeLeadingIndicator (CLI)
// DSD: OECD.SDD.STES,DSD_STES@DF_CLI,4.1
// Key: {COUNTRY}.M.LI...AA.IX..H   (amplitude-adjusted, index)
// ---------------------------------------------------------------------------

type cliFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newCLIFetcher(p *Provider) *cliFetcher {
	return &cliFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelCompositeLeadingIndicator,
			"OECD Composite Leading Indicator (CLI)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *cliFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelCompositeLeadingIndicator, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	country := resolveCountry(params[provider.ParamCountry])
	if country == "" {
		country = "G20"
	}
	start, end := datePeriod(params)

	dsd := "OECD.SDD.STES,DSD_STES@DF_CLI,4.1"
	key := country + ".M.LI...AA.IX..H"
	url := buildURL(dsd, key, start, end, "format=csvfile")

	records, err := f.p.fetchCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("CLI: %w", err)
	}

	data := parseEconomicCSV(records, "OECD CLI")
	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// ConsumerPriceIndex (CPI)
// DSD: OECD.SDD.TPS,DSD_PRICES@DF_PRICES_ALL,1.0
// Key: {COUNTRY}.M.N.CPI.IX._T.N.
// ---------------------------------------------------------------------------

type cpiFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newCPIFetcher(p *Provider) *cpiFetcher {
	return &cpiFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelConsumerPriceIndex,
			"OECD Consumer Price Index (CPI)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *cpiFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelConsumerPriceIndex, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	country := resolveCountry(params[provider.ParamCountry])
	if country == "" {
		country = "USA"
	}
	start, end := datePeriod(params)

	dsd := "OECD.SDD.TPS,DSD_PRICES@DF_PRICES_ALL,1.0"
	key := country + ".M.N.CPI.IX._T.N."
	url := buildURL(dsd, key, start, end)

	records, err := f.p.fetchCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("CPI: %w", err)
	}

	data := parseCPICSV(records)
	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// CountryInterestRates
// DSD: OECD.SDD.STES,DSD_KEI@DF_KEI,4.0
// Key: {COUNTRY}.M.IR3TIB....   (short-term rates by default)
// ---------------------------------------------------------------------------

type interestRatesFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newInterestRatesFetcher(p *Provider) *interestRatesFetcher {
	return &interestRatesFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelCountryInterestRates,
			"OECD country interest rates (short/long term)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *interestRatesFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelCountryInterestRates, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	country := resolveCountry(params[provider.ParamCountry])
	if country == "" {
		country = "USA"
	}
	start, end := datePeriod(params)

	// Default to short-term (3-month interbank) rates.
	duration := "IR3TIB"
	if d := params["duration"]; d == "long" {
		duration = "IRLT"
	} else if d == "immediate" {
		duration = "IRSTCI"
	}

	dsd := "OECD.SDD.STES,DSD_KEI@DF_KEI,4.0"
	key := country + ".M." + duration + "...."
	url := buildURL(dsd, key, start, end)

	records, err := f.p.fetchCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("interest rates: %w", err)
	}

	data := parseInterestRateCSV(records, duration)
	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// CSV parsing helpers
// ---------------------------------------------------------------------------

// parseEconomicCSV extracts economic indicator data from SDMX CSV.
// Expected columns: REF_AREA, TIME_PERIOD, OBS_VALUE
func parseEconomicCSV(records [][]string, _ string) []models.EconomicIndicatorData {
	if len(records) < 2 {
		return nil
	}

	header := records[0]
	refIdx := findColumn(header, "REF_AREA")
	timeIdx := findColumn(header, "TIME_PERIOD")
	obsIdx := findColumn(header, "OBS_VALUE")

	if refIdx < 0 || timeIdx < 0 || obsIdx < 0 {
		return nil
	}

	var data []models.EconomicIndicatorData
	for _, row := range records[1:] {
		if len(row) <= obsIdx || len(row) <= timeIdx || len(row) <= refIdx {
			continue
		}
		val := parseFloat(row[obsIdx])
		if val == 0 {
			continue
		}
		data = append(data, models.EconomicIndicatorData{
			Date:    parseSDMXDate(row[timeIdx]),
			Country: countryName(row[refIdx]),
			Value:   val,
		})
	}
	return data
}

// parseCPICSV parses CPI-specific CSV.
func parseCPICSV(records [][]string) []models.CPIData {
	if len(records) < 2 {
		return nil
	}

	header := records[0]
	refIdx := findColumn(header, "REF_AREA")
	timeIdx := findColumn(header, "TIME_PERIOD")
	obsIdx := findColumn(header, "OBS_VALUE")

	if refIdx < 0 || timeIdx < 0 || obsIdx < 0 {
		return nil
	}

	var data []models.CPIData
	for _, row := range records[1:] {
		if len(row) <= obsIdx || len(row) <= timeIdx || len(row) <= refIdx {
			continue
		}
		val := parseFloat(row[obsIdx])
		if val == 0 {
			continue
		}
		data = append(data, models.CPIData{
			Date:    parseSDMXDate(row[timeIdx]),
			Country: countryName(row[refIdx]),
			Value:   val,
		})
	}
	return data
}

// parseInterestRateCSV parses interest rate CSV, dividing by 100.
func parseInterestRateCSV(records [][]string, rateType string) []models.InterestRateData {
	if len(records) < 2 {
		return nil
	}

	header := records[0]
	timeIdx := findColumn(header, "TIME_PERIOD")
	obsIdx := findColumn(header, "OBS_VALUE")

	if timeIdx < 0 || obsIdx < 0 {
		return nil
	}

	var data []models.InterestRateData
	for _, row := range records[1:] {
		if len(row) <= obsIdx || len(row) <= timeIdx {
			continue
		}
		val := parseFloat(row[obsIdx])
		if val == 0 {
			continue
		}
		data = append(data, models.InterestRateData{
			Date:     parseSDMXDate(row[timeIdx]),
			Rate:     val / 100, // normalize percentage
			RateType: rateType,
		})
	}
	return data
}
