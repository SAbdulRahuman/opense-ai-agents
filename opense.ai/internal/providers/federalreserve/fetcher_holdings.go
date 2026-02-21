package federalreserve

import (
	"context"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// CentralBankHoldings — SOMA Holdings (System Open Market Account).
// URL: https://markets.newyorkfed.org/api/soma/summary.json
//      https://markets.newyorkfed.org/api/soma/asofdates/list.json
//      https://markets.newyorkfed.org/api/soma/tsy/get/asof/{date}.json
// ---------------------------------------------------------------------------

type centralBankHoldingsFetcher struct {
	provider.BaseFetcher
}

func newCentralBankHoldingsFetcher() *centralBankHoldingsFetcher {
	return &centralBankHoldingsFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelCentralBankHoldings,
			"Federal Reserve SOMA holdings (treasuries, agencies, MBS)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
	}
}

func (f *centralBankHoldingsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelCentralBankHoldings, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	asOf := defaultDate(params, provider.ParamEndDate, time.Now().Format("2006-01-02"))

	// Fetch treasury holdings for the given date.
	url := baseNYFed + "/api/soma/tsy/get/asof/" + asOf + ".json"

	var resp nyfedSomaResponse
	if err := fetchFedJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("central bank holdings (treasury): %w", err)
	}

	var holdings []models.EconomicIndicatorData
	for _, entry := range resp.Soma.Holdings {
		dt, _ := time.Parse("01/02/2006", entry.AsOfDate)
		if dt.IsZero() {
			dt = parseDate(entry.AsOfDate)
		}
		faceVal := entry.ParValue
		if faceVal == 0 {
			faceVal = entry.CurrentFaceValue
		}

		holdings = append(holdings, models.EconomicIndicatorData{
			Date:    dt,
			Country: "US",
			Value:   faceVal,
		})
	}

	// Also fetch agency holdings.
	urlAgency := baseNYFed + "/api/soma/agency/get/asof/" + asOf + ".json"
	var agencyResp nyfedSomaResponse
	if err := fetchFedJSON(ctx, urlAgency, &agencyResp); err == nil {
		for _, entry := range agencyResp.Soma.Holdings {
			dt, _ := time.Parse("01/02/2006", entry.AsOfDate)
			if dt.IsZero() {
				dt = parseDate(entry.AsOfDate)
			}
			faceVal := entry.ParValue
			if faceVal == 0 {
				faceVal = entry.CurrentFaceValue
			}

			holdings = append(holdings, models.EconomicIndicatorData{
				Date:    dt,
				Country: "US",
				Value:   faceVal,
			})
		}
	}

	result := newResult(holdings)
	f.CacheSet(cacheKey, result)
	return result, nil
}
