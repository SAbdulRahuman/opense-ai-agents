package federalreserve

import (
	"context"
	"fmt"
	"strings"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// H.15 maturity column indices (after header row).
// The CSV has: Date, then 11 maturity columns.
var h15Maturities = []string{
	"1M", "3M", "6M", "1Y", "2Y", "3Y", "5Y", "7Y", "10Y", "20Y", "30Y",
}

// ---------------------------------------------------------------------------
// TreasuryRates — H.15 Release daily rates.
// URL: Fed Board CSV download (H.15 series).
// ---------------------------------------------------------------------------

type treasuryRatesFetcher struct {
	provider.BaseFetcher
}

func newTreasuryRatesFetcher() *treasuryRatesFetcher {
	return &treasuryRatesFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelTreasuryRates,
			"Federal Reserve H.15 Treasury rates (all maturities)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
	}
}

func (f *treasuryRatesFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelTreasuryRates, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	records, err := fetchH15Data(ctx)
	if err != nil {
		return nil, fmt.Errorf("treasury rates: %w", err)
	}

	startDate := params[provider.ParamStartDate]
	endDate := params[provider.ParamEndDate]

	var rates []models.TreasuryRate
	for _, row := range records {
		if len(row) < 12 {
			continue
		}
		date := strings.TrimSpace(row[0])
		if date == "" || date == "Series Description:" {
			continue
		}
		if startDate != "" && date < startDate {
			continue
		}
		if endDate != "" && date > endDate {
			continue
		}

		rateMap := make(map[string]float64)
		hasAny := false
		for i, mat := range h15Maturities {
			v := parseFloat64(row[i+1])
			if v != 0 {
				rateMap[mat] = v / 100 // normalize from percentage
				hasAny = true
			}
		}
		if !hasAny {
			continue
		}

		rates = append(rates, models.TreasuryRate{
			Date:  parseDate(date),
			Rates: rateMap,
		})
	}

	result := newResult(rates)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// YieldCurve — H.15 rates unpivoted to individual maturity points.
// Same data source as TreasuryRates but different output shape.
// ---------------------------------------------------------------------------

type yieldCurveFetcher struct {
	provider.BaseFetcher
}

func newYieldCurveFetcher() *yieldCurveFetcher {
	return &yieldCurveFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelYieldCurve,
			"Federal Reserve US Treasury yield curve (H.15)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
	}
}

func (f *yieldCurveFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelYieldCurve, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	records, err := fetchH15Data(ctx)
	if err != nil {
		return nil, fmt.Errorf("yield curve: %w", err)
	}

	startDate := params[provider.ParamStartDate]
	endDate := params[provider.ParamEndDate]

	var points []models.YieldCurvePoint
	for _, row := range records {
		if len(row) < 12 {
			continue
		}
		date := strings.TrimSpace(row[0])
		if date == "" || date == "Series Description:" {
			continue
		}
		if startDate != "" && date < startDate {
			continue
		}
		if endDate != "" && date > endDate {
			continue
		}

		dt := parseDate(date)
		for i, mat := range h15Maturities {
			v := parseFloat64(row[i+1])
			if v != 0 {
				points = append(points, models.YieldCurvePoint{
					Date:     dt,
					Maturity: mat,
					Rate:     v / 100,
				})
			}
		}
	}

	result := newResult(points)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// fetchH15Data downloads the H.15 CSV and returns parsed records.
// Skips the first 5 header rows.
func fetchH15Data(ctx context.Context) ([][]string, error) {
	return fetchFedCSV(ctx, buildH15URL(), 5)
}

// ---------------------------------------------------------------------------
// SvenssonYieldCurve — Fed Board static CSV (feds200628.csv).
// URL: https://www.federalreserve.gov/data/yield-curve-tables/feds200628.csv
// ---------------------------------------------------------------------------

type svenssonYieldCurveFetcher struct {
	provider.BaseFetcher
}

func newSvenssonYieldCurveFetcher() *svenssonYieldCurveFetcher {
	return &svenssonYieldCurveFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelSvenssonYieldCurve,
			"Federal Reserve Svensson zero-coupon yield curve parameters",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
	}
}

func (f *svenssonYieldCurveFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelSvenssonYieldCurve, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	url := baseFedBoard + "/data/yield-curve-tables/feds200628.csv"
	raw, err := fetchFedRaw(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("svensson yield curve: %w", err)
	}

	// Find the data start (line beginning with "Date,").
	lines := strings.Split(string(raw), "\n")
	dataStart := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "Date,") {
			dataStart = i
			break
		}
	}
	if dataStart < 0 {
		return nil, fmt.Errorf("svensson: could not find data header")
	}

	// Parse the header to find SVENY columns (zero-coupon yields).
	header := strings.Split(strings.TrimSpace(lines[dataStart]), ",")
	svenyIdx := make(map[int]string) // column index → maturity label
	for i, col := range header {
		col = strings.TrimSpace(col)
		if strings.HasPrefix(col, "SVENY") {
			mat := strings.TrimPrefix(col, "SVENY")
			if len(mat) == 2 && mat[0] == '0' {
				mat = mat[1:] // "01" → "1"
			}
			svenyIdx[i] = mat + "Y"
		}
	}

	startDate := params[provider.ParamStartDate]
	endDate := params[provider.ParamEndDate]

	var points []models.YieldCurvePoint
	for _, line := range lines[dataStart+1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 2 {
			continue
		}
		date := strings.TrimSpace(fields[0])
		if startDate != "" && date < startDate {
			continue
		}
		if endDate != "" && date > endDate {
			continue
		}

		dt := parseDate(date)
		for idx, mat := range svenyIdx {
			if idx >= len(fields) {
				continue
			}
			v := parseFloat64(fields[idx])
			if v != 0 && v != -999.99 {
				points = append(points, models.YieldCurvePoint{
					Date:     dt,
					Maturity: mat,
					Rate:     v / 100, // normalize
				})
			}
		}
	}

	result := newResult(points)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// MoneyMeasures — H.6 Release (M1, M2, etc.).
// URL: Fed Board CSV download (H.6 series).
// ---------------------------------------------------------------------------

type moneyMeasuresFetcher struct {
	provider.BaseFetcher
}

func newMoneyMeasuresFetcher() *moneyMeasuresFetcher {
	return &moneyMeasuresFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelMoneyMeasures,
			"Federal Reserve H.6 money supply measures (M1, M2)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
	}
}

func (f *moneyMeasuresFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelMoneyMeasures, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	url := baseFedBoard + "/datadownload/Output.aspx?rel=H6&series=798e2796917702a5f8423426ba7e6b42&lastobs=&from=&to=&filetype=csv&label=include&layout=seriescolumn&type=package"

	records, err := fetchFedCSV(ctx, url, 5) // skip 5 header rows
	if err != nil {
		return nil, fmt.Errorf("money measures: %w", err)
	}

	startDate := params[provider.ParamStartDate]
	endDate := params[provider.ParamEndDate]

	// We parse date (first column) + look for M1 and M2 columns.
	// H.6 columns are labeled by series IDs. We'll take the first two numeric columns as M1 and M2.
	var measures []models.MoneyMeasureData
	for _, row := range records {
		if len(row) < 3 {
			continue
		}
		date := strings.TrimSpace(row[0])
		if date == "" || !isDateLike(date) {
			continue
		}
		if startDate != "" && date < startDate {
			continue
		}
		if endDate != "" && date > endDate {
			continue
		}

		m1 := parseFloat64(row[1])
		m2 := float64(0)
		if len(row) > 2 {
			m2 = parseFloat64(row[2])
		}

		if m1 != 0 {
			measures = append(measures, models.MoneyMeasureData{
				Date:    parseDate(date),
				Country: "US",
				Measure: "M1",
				Value:   m1,
			})
		}
		if m2 != 0 {
			measures = append(measures, models.MoneyMeasureData{
				Date:    parseDate(date),
				Country: "US",
				Measure: "M2",
				Value:   m2,
			})
		}
	}

	result := newResult(measures)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// isDateLike checks if a string looks like a date (starts with 4 digits).
func isDateLike(s string) bool {
	if len(s) < 4 {
		return false
	}
	for _, c := range s[:4] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
