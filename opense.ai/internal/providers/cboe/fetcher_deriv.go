package cboe

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// OptionsChains — Full options chain from CBOE.
// URL: https://cdn.cboe.com/api/global/delayed_quotes/options/{SYMBOL}.json
// ---------------------------------------------------------------------------

type optionsChainsFetcher struct {
	provider.BaseFetcher
	prov *Provider
}

func newOptionsChainsFetcher(p *Provider) *optionsChainsFetcher {
	return &optionsChainsFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelOptionsChains,
			"CBOE delayed options chain with Greeks",
			[]string{provider.ParamSymbol},
			nil,
		),
		prov: p,
	}
}

// optionSymbolRE parses CBOE option symbols like "AAPL240119C00150000".
// Format: TICKER + YYMMDD + C/P + STRIKE*1000 (8 digits, zero-padded).
var optionSymbolRE = regexp.MustCompile(`^([A-Z]+)(\d{6})([CP])(\d{8})$`)

func (f *optionsChainsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	symbol := strings.ToUpper(params[provider.ParamSymbol])
	if symbol == "" {
		return nil, fmt.Errorf("cboe: %s is required", provider.ParamSymbol)
	}
	symbol = strings.ReplaceAll(symbol, "^", "")

	cacheKey := provider.CacheKey(provider.ModelOptionsChains, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	_, _ = f.prov.getIndexDirectory(ctx)
	url := optionsURL(f.prov.symbolPath(symbol))

	var resp cboeOptionsResponse
	if err := fetchCBOEJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("cboe options chains: %w", err)
	}

	// Parse option records into OptionContractEx.
	expirySet := make(map[time.Time]bool)
	var contracts []models.OptionContractEx

	for _, opt := range resp.Data.Options {
		parts := optionSymbolRE.FindStringSubmatch(opt.Option)
		if parts == nil {
			continue
		}
		// parts: [full, ticker, YYMMDD, C/P, strikeStr]
		expStr := parts[2]
		optType := "call"
		if parts[3] == "P" {
			optType = "put"
		}
		strike, _ := strconv.ParseFloat(parts[4], 64)
		strike /= 1000 // CBOE encodes strike × 1000

		expDate, _ := time.Parse("060102", expStr)
		expirySet[expDate] = true

		contracts = append(contracts, models.OptionContractEx{
			Symbol:         symbol,
			Underlying:     symbol,
			ContractSymbol: opt.Option,
			StrikePrice:    strike,
			OptionType:     optType,
			ExpirationDate: expDate,
			Exchange:       "CBOE",
			LastPrice:      opt.LastTradePrice,
			BidPrice:       opt.Bid,
			AskPrice:       opt.Ask,
			BidSize:        opt.BidSize,
			AskSize:        opt.AskSize,
			Volume:         opt.Volume,
			OpenInterest:   opt.OpenInterest,
			Change:         opt.Change,
			ChangePct:      opt.PctChange / 100, // normalize
			IV:             opt.IV,
			Delta:          opt.Delta,
			Gamma:          opt.Gamma,
			Theta:          opt.Theta,
			Vega:           opt.Vega,
			Rho:            opt.Rho,
		})
	}

	// Collect unique expiries.
	var expiries []time.Time
	for exp := range expirySet {
		expiries = append(expiries, exp)
	}

	chain := models.OptionsChainData{
		Symbol:    symbol,
		Underlying: symbol,
		Exchange:  "CBOE",
		Expiries:  expiries,
		Contracts: contracts,
		FetchedAt: time.Now(),
	}

	result := newResult(chain)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// FuturesCurve — VIX futures term structure from CBOE.
// Uses delayed quotes for VIX futures contracts.
// ---------------------------------------------------------------------------

type futuresCurveFetcher struct {
	provider.BaseFetcher
}

func newFuturesCurveFetcher() *futuresCurveFetcher {
	return &futuresCurveFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelFuturesCurve,
			"CBOE VIX futures term structure",
			nil, // symbol defaults to VIX
			[]string{provider.ParamSymbol},
		),
	}
}

// VIX futures use month codes: F(Jan), G(Feb), H(Mar), J(Apr), K(May), M(Jun),
// N(Jul), Q(Aug), U(Sep), V(Oct), X(Nov), Z(Dec).
var vxMonthCodes = []string{"F", "G", "H", "J", "K", "M", "N", "Q", "U", "V", "X", "Z"}

func (f *futuresCurveFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelFuturesCurve, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	// Fetch the VX futures quotes from CBOE.
	// The settlement prices endpoint gives us current VIX futures data.
	url := "https://www.cboe.com/us/futures/market_statistics/settlement/csv"

	raw, err := fetchCBOERaw(ctx, url)
	if err != nil {
		// Fallback: try the JSON delayed quotes for individual VX contracts.
		return f.fetchViaDelayedQuotes(ctx, params, cacheKey)
	}

	points, err := parseVXSettlement(raw)
	if err != nil || len(points) == 0 {
		return f.fetchViaDelayedQuotes(ctx, params, cacheKey)
	}

	result := newResult(points)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// fetchViaDelayedQuotes tries to get VIX futures curve from JSON delayed quotes.
func (f *futuresCurveFetcher) fetchViaDelayedQuotes(ctx context.Context, params provider.QueryParams, cacheKey string) (*provider.FetchResult, error) {
	// Build VX contract symbols for next 9 months.
	now := time.Now()
	frontMonth := thirdWednesday(now)
	startMonth := now.Month()
	if now.Day() > frontMonth {
		startMonth = (startMonth % 12) + 1
	}

	var points []models.FuturesCurvePoint
	year := now.Year()
	month := int(startMonth)

	for i := 0; i < 9; i++ {
		if month > 12 {
			month = 1
			year++
		}
		// CBOE VX EOD symbols: UZ + month code (F-Z)
		monthIdx := month - 1
		sym := "UZ" + vxMonthCodes[monthIdx]

		url := quotesURL(sym)
		var resp cboeQuoteResponse
		if err := fetchCBOEJSON(ctx, url, &resp); err == nil && resp.Data.CurrentPrice > 0 {
			// Estimate expiration as third Wednesday of the month.
			expDate := time.Date(year, time.Month(month), thirdWednesdayOfMonth(year, time.Month(month)), 0, 0, 0, 0, time.UTC)
			points = append(points, models.FuturesCurvePoint{
				Symbol:     fmt.Sprintf("VX%d", i+1),
				Expiration: expDate,
				Price:      resp.Data.CurrentPrice,
				Volume:     resp.Data.Volume,
			})
		}
		month++
	}

	result := newResult(points)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// parseVXSettlement parses the CBOE VX settlement CSV.
func parseVXSettlement(raw []byte) ([]models.FuturesCurvePoint, error) {
	lines := strings.Split(string(raw), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("cboe: empty settlement CSV")
	}

	var points []models.FuturesCurvePoint
	for _, line := range lines[1:] { // skip header
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			continue
		}
		// Expected fields: Symbol, Expiration, Price/Settlement, ...
		sym := strings.TrimSpace(fields[0])
		if !strings.HasPrefix(sym, "VX") {
			continue
		}

		expStr := strings.TrimSpace(fields[1])
		expDate := parseCBOEDate(expStr)
		if expDate.IsZero() {
			// Try MM/DD/YYYY format.
			expDate, _ = time.Parse("01/02/2006", expStr)
		}

		priceStr := strings.TrimSpace(fields[2])
		price := parseFloatStr(priceStr)
		if price <= 0 {
			continue
		}

		points = append(points, models.FuturesCurvePoint{
			Symbol:     sym,
			Expiration: expDate,
			Price:      price,
		})
	}
	return points, nil
}

// parseFloatStr is a simple float parser that returns 0 on error.
func parseFloatStr(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

// thirdWednesday returns the day of the third Wednesday for the current month.
func thirdWednesday(t time.Time) int {
	return thirdWednesdayOfMonth(t.Year(), t.Month())
}

// thirdWednesdayOfMonth calculates the day number of the third Wednesday.
func thirdWednesdayOfMonth(year int, month time.Month) int {
	// Find the first day of the month.
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	// Find the first Wednesday.
	wd := first.Weekday()
	daysUntilWed := (time.Wednesday - wd + 7) % 7
	firstWed := 1 + int(daysUntilWed)
	// Third Wednesday = first Wednesday + 14.
	return firstWed + 14
}

// Ensure json import is used.
var _ = json.Unmarshal
