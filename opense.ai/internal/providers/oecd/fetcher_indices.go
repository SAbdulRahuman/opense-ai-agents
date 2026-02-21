package oecd

import (
	"context"
	"fmt"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// HousePriceIndex
// DSD: OECD.SDD.TPS,DSD_RHPI_TARGET@DF_RHPI_TARGET,1.0
// Key: COU.{COUNTRY}.Q.RHPI.IX....
// ---------------------------------------------------------------------------

type housePriceFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newHousePriceFetcher(p *Provider) *housePriceFetcher {
	return &housePriceFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelHousePriceIndex,
			"OECD Real House Price Index",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *housePriceFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelHousePriceIndex, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	country := resolveCountry(params[provider.ParamCountry])
	if country == "" {
		country = "USA"
	}
	start, end := datePeriod(params)

	freq := "Q"
	if params["frequency"] == "monthly" {
		freq = "M"
	} else if params["frequency"] == "annual" {
		freq = "A"
	}

	transform := "IX" // index
	if t := params["transform"]; t == "yoy" {
		transform = "PA"
	} else if t == "period" {
		transform = "PC"
	}

	dsd := "OECD.SDD.TPS,DSD_RHPI_TARGET@DF_RHPI_TARGET,1.0"
	key := "COU." + country + "." + freq + ".RHPI." + transform + "...."
	url := buildURL(dsd, key, start, end)

	records, err := f.p.fetchCSV(ctx, url)
	if err != nil {
		// Fallback: if monthly returns 404, try quarterly.
		if freq == "M" {
			key = "COU." + country + ".Q.RHPI." + transform + "...."
			url = buildURL(dsd, key, start, end)
			records, err = f.p.fetchCSV(ctx, url)
		}
		if err != nil {
			return nil, fmt.Errorf("house price index: %w", err)
		}
	}

	data := parseEconomicCSV(records, "OECD House Price Index")
	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// SharePriceIndex
// DSD: OECD.SDD.STES,DSD_STES@DF_FINMARK,4.0
// Key: {COUNTRY}.M.SHARE......
// ---------------------------------------------------------------------------

type sharePriceFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newSharePriceFetcher(p *Provider) *sharePriceFetcher {
	return &sharePriceFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelSharePriceIndex,
			"OECD Share Price Index",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *sharePriceFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelSharePriceIndex, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	country := resolveCountry(params[provider.ParamCountry])
	if country == "" {
		country = "USA"
	}
	start, end := datePeriod(params)

	freq := "M"
	if params["frequency"] == "quarterly" || params["frequency"] == "quarter" {
		freq = "Q"
	} else if params["frequency"] == "annual" {
		freq = "A"
	}

	dsd := "OECD.SDD.STES,DSD_STES@DF_FINMARK,4.0"
	key := country + "." + freq + ".SHARE......"
	url := buildURL(dsd, key, start, end)

	records, err := f.p.fetchCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("share price index: %w", err)
	}

	data := parseEconomicCSV(records, "OECD Share Price Index")
	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// Unemployment
// DSD: OECD.SDD.TPS,DSD_LFS@DF_IALFS_UNE_M,1.0
// Key: {COUNTRY}..._Z.N._T.Y_GE15..M
// ---------------------------------------------------------------------------

type unemploymentFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newUnemploymentFetcher(p *Provider) *unemploymentFetcher {
	return &unemploymentFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelUnemployment,
			"OECD unemployment rate",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *unemploymentFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelUnemployment, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	country := resolveCountry(params[provider.ParamCountry])
	if country == "" {
		country = "USA"
	}
	start, end := datePeriod(params)

	freq := "M"
	if params["frequency"] == "quarterly" || params["frequency"] == "quarter" {
		freq = "Q"
	} else if params["frequency"] == "annual" {
		freq = "A"
	}

	sex := "_T"
	if params["sex"] == "male" {
		sex = "M"
	} else if params["sex"] == "female" {
		sex = "F"
	}

	age := "Y_GE15"
	if params["age"] == "15-24" {
		age = "Y15T24"
	} else if params["age"] == "25+" {
		age = "Y_GE25"
	}

	seasonal := "N"
	if params["seasonal_adjustment"] == "true" {
		seasonal = "Y"
	}

	dsd := "OECD.SDD.TPS,DSD_LFS@DF_IALFS_UNE_M,1.0"
	key := country + "..._Z." + seasonal + "." + sex + "." + age + ".." + freq
	url := buildURL(dsd, key, start, end)

	records, err := f.p.fetchCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("unemployment: %w", err)
	}

	data := parseUnemploymentCSV(records)
	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// parseUnemploymentCSV parses unemployment CSV, dividing by 100.
func parseUnemploymentCSV(records [][]string) []models.UnemploymentData {
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

	var data []models.UnemploymentData
	for _, row := range records[1:] {
		if len(row) <= obsIdx || len(row) <= timeIdx || len(row) <= refIdx {
			continue
		}
		val := parseFloat(row[obsIdx])
		if val == 0 {
			continue
		}
		data = append(data, models.UnemploymentData{
			Date:    parseSDMXDate(row[timeIdx]),
			Country: countryName(row[refIdx]),
			Value:   val / 100, // normalize from percentage
		})
	}
	return data
}
