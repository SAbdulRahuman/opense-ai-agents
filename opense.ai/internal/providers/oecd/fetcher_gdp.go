package oecd

import (
	"context"
	"fmt"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// GdpNominal
// DSD: OECD.SDD.NAD,DSD_NAMAIN1@DF_QNA_EXPENDITURE_USD,1.1
// Key: Q..{COUNTRY}.S1..B1GQ.....V..
// ---------------------------------------------------------------------------

type gdpNominalFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newGdpNominalFetcher(p *Provider) *gdpNominalFetcher {
	return &gdpNominalFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelGdpNominal,
			"OECD nominal GDP (USD, quarterly/annual)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *gdpNominalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelGdpNominal, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	country := resolveCountry(params[provider.ParamCountry])
	if country == "" {
		country = "USA"
	}
	start, end := datePeriod(params)

	freq := "Q"
	if params["frequency"] == "annual" {
		freq = "A"
	}

	dsd := "OECD.SDD.NAD,DSD_NAMAIN1@DF_QNA_EXPENDITURE_USD,1.1"
	key := freq + ".." + country + ".S1..B1GQ.....V.."
	url := buildURL(dsd, key, start, end, "format=csvfile")

	records, err := f.p.fetchCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("GDP nominal: %w", err)
	}

	data := parseGDPCSV(records, "nominal", 1_000_000)
	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// GdpReal
// DSD: OECD.SDD.NAD,DSD_NAMAIN1@DF_QNA,1.1
// Key: Q..{COUNTRY}.S1..B1GQ._Z...USD_PPP.LR.LA.T0102
// ---------------------------------------------------------------------------

type gdpRealFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newGdpRealFetcher(p *Provider) *gdpRealFetcher {
	return &gdpRealFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelGdpReal,
			"OECD real GDP (PPP, chain-linked volume)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *gdpRealFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelGdpReal, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	country := resolveCountry(params[provider.ParamCountry])
	if country == "" {
		country = "USA"
	}
	start, end := datePeriod(params)

	freq := "Q"
	if params["frequency"] == "annual" {
		freq = "A"
	}

	dsd := "OECD.SDD.NAD,DSD_NAMAIN1@DF_QNA,1.1"
	key := freq + ".." + country + ".S1..B1GQ._Z...USD_PPP.LR.LA.T0102"
	url := buildURL(dsd, key, start, end, "format=csvfile")

	records, err := f.p.fetchCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("GDP real: %w", err)
	}

	data := parseGDPCSV(records, "real", 1_000_000)
	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// GdpForecast
// DSD: OECD.ECO.MAD,DSD_EO@DF_EO,1.1
// Key: {COUNTRY}.GDPV_USD.A
// ---------------------------------------------------------------------------

type gdpForecastFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newGdpForecastFetcher(p *Provider) *gdpForecastFetcher {
	return &gdpForecastFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelGdpForecast,
			"OECD GDP forecast (Economic Outlook)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *gdpForecastFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelGdpForecast, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	country := resolveCountry(params[provider.ParamCountry])
	if country == "" {
		country = "USA"
	}
	start, end := datePeriod(params)

	// Default to volume GDP in USD (annual).
	measure := "GDPV_USD"
	if m := params["measure"]; m != "" {
		switch m {
		case "growth":
			measure = "GDPV_ANNPCT"
		case "capita":
			measure = "GDPVD_CAP"
		case "deflator":
			measure = "PGDP"
		}
	}

	dsd := "OECD.ECO.MAD,DSD_EO@DF_EO,1.1"
	key := country + "." + measure + ".A"
	url := buildURL(dsd, key, start, end, "format=csvfile")

	records, err := f.p.fetchCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("GDP forecast: %w", err)
	}

	multiplier := float64(1)
	if measure != "GDPV_ANNPCT" && measure != "PGDP" {
		multiplier = 1
	}
	data := parseGDPCSV(records, "forecast", multiplier)

	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// GDP CSV parser
// ---------------------------------------------------------------------------

func parseGDPCSV(records [][]string, gdpType string, multiplier float64) []models.GDPData {
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

	var data []models.GDPData
	for _, row := range records[1:] {
		if len(row) <= obsIdx || len(row) <= timeIdx || len(row) <= refIdx {
			continue
		}
		val := parseFloat(row[obsIdx])
		if val == 0 {
			continue
		}
		data = append(data, models.GDPData{
			Date:     parseSDMXDate(row[timeIdx]),
			Country:  countryName(row[refIdx]),
			Value:    val * multiplier,
			Currency: "USD",
			Type:     gdpType,
		})
	}
	return data
}
