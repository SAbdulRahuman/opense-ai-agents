package federalreserve

import (
	"context"
	"fmt"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// InflationExpectations — Philadelphia Fed Survey of Professional Forecasters.
// Source: https://www.philadelphiafed.org/surveys-and-data/real-time-data-research/inflation-forecasts
// NOTE: The source data is in XLSX format. This is a simplified implementation
// that uses a known URL pattern. Full XLSX parsing requires an Excel library.
// ---------------------------------------------------------------------------

type inflationExpectationsFetcher struct {
	provider.BaseFetcher
}

func newInflationExpectationsFetcher() *inflationExpectationsFetcher {
	return &inflationExpectationsFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelInflationExpectations,
			"Philadelphia Fed Survey of Professional Forecasters inflation expectations",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
	}
}

func (f *inflationExpectationsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelInflationExpectations, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	// The Philadelphia Fed provides CSV exports that are simpler to parse.
	// We use the mean CPI forecast from SPF.
	url := "https://www.philadelphiafed.org/-/media/frbp/assets/surveys-and-data/survey-of-professional-forecasters/data-files/files/meanLevel_CPI_Level.csv"

	records, err := fetchFedCSV(ctx, url, 1) // skip header row
	if err != nil {
		// Fallback: return empty with informational note.
		return newResult([]models.InflationExpectationData{}), nil
	}

	startDate := params[provider.ParamStartDate]
	endDate := params[provider.ParamEndDate]

	var expectations []models.InflationExpectationData
	for _, row := range records {
		if len(row) < 3 {
			continue
		}
		// Columns: YEAR, QUARTER, value columns for different horizons.
		year := row[0]
		quarter := row[1]
		if !isDateLike(year) {
			continue
		}

		// Convert year+quarter to a date string.
		qMonth := "01"
		switch quarter {
		case "2":
			qMonth = "04"
		case "3":
			qMonth = "07"
		case "4":
			qMonth = "10"
		}
		dateStr := year + "-" + qMonth + "-01"

		if startDate != "" && dateStr < startDate {
			continue
		}
		if endDate != "" && dateStr > endDate {
			continue
		}

		dt := parseDate(dateStr)

		// CPI1 = 1-quarter ahead, CPI2 = 2-quarter ahead, etc.
		if len(row) > 2 {
			v := parseFloat64(row[2])
			if v != 0 {
				expectations = append(expectations, models.InflationExpectationData{
					Date:    dt,
					Horizon: "1Q",
					Value:   v,
					Source:  "Philadelphia Fed SPF",
				})
			}
		}
		if len(row) > 3 {
			v := parseFloat64(row[3])
			if v != 0 {
				expectations = append(expectations, models.InflationExpectationData{
					Date:    dt,
					Horizon: "2Q",
					Value:   v,
					Source:  "Philadelphia Fed SPF",
				})
			}
		}
		if len(row) > 5 {
			v := parseFloat64(row[5])
			if v != 0 {
				expectations = append(expectations, models.InflationExpectationData{
					Date:    dt,
					Horizon: "1Y",
					Value:   v,
					Source:  "Philadelphia Fed SPF",
				})
			}
		}
	}

	result := newResult(expectations)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// TFP — Total Factor Productivity from San Francisco Fed.
// Source: https://www.frbsf.org/research-and-insights/data-and-indicators/total-factor-productivity/
// NOTE: The source data is in XLSX format. This simplified implementation
// uses the annual CSV if available, or returns a stub.
// ---------------------------------------------------------------------------

type tfpFetcher struct {
	provider.BaseFetcher
}

func newTFPFetcher() *tfpFetcher {
	return &tfpFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelTotalFactorProductivity,
			"San Francisco Fed Total Factor Productivity",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
	}
}

func (f *tfpFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelTotalFactorProductivity, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	// SF Fed publishes quarterly TFP data.
	// The direct data download requires XLSX parsing.
	// Try the CSV endpoint; if not available, return an informational error.
	url := "https://www.frbsf.org/wp-content/uploads/quarterly_tfp.csv"

	records, err := fetchFedCSV(ctx, url, 1) // skip header
	if err != nil {
		return nil, fmt.Errorf("TFP: data source requires XLSX parsing (not yet supported): %w", err)
	}

	startDate := params[provider.ParamStartDate]
	endDate := params[provider.ParamEndDate]

	var data []models.EconomicIndicatorData
	for _, row := range records {
		if len(row) < 3 {
			continue
		}
		date := row[0]
		if !isDateLike(date) {
			continue
		}
		if startDate != "" && date < startDate {
			continue
		}
		if endDate != "" && date > endDate {
			continue
		}

		// Try to parse TFP growth column (typically column index 2 or later).
		tfpGrowth := parseFloat64(row[2])
		if tfpGrowth != 0 {
			data = append(data, models.EconomicIndicatorData{
				Date:    parseDate(date),
				Country: "US",
				Value:   tfpGrowth,
			})
		}
	}

	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}
