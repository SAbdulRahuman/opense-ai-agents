package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
	"github.com/seenimoa/openseai/pkg/utils"
)

// NSEDerivatives implements derivatives data fetching from NSE India.
type NSEDerivatives struct {
	nse *NSE // reuse NSE client for cookie management and rate limiting
}

// NewNSEDerivatives creates a new NSE derivatives data source.
// It shares the underlying NSE HTTP client and cookies.
func NewNSEDerivatives(nse *NSE) *NSEDerivatives {
	return &NSEDerivatives{nse: nse}
}

// Name returns the data source name.
func (d *NSEDerivatives) Name() string { return "NSE Derivatives" }

// --- NSE derivatives response types ---

type nseOptionChainResponse struct {
	Records    nseOCRecords    `json:"records"`
	Filtered   nseOCRecords    `json:"filtered"`
}

type nseOCRecords struct {
	ExpiryDates   []string       `json:"expiryDates"`
	StrikePrices  []float64      `json:"strikePrices"`
	Data          []nseOCEntry   `json:"data"`
	Timestamp     string         `json:"timestamp"`
	UnderlyingValue float64      `json:"underlyingValue"`
	TotalCEOI     int64          `json:"-"` // computed
	TotalPEOI     int64          `json:"-"` // computed
}

type nseOCEntry struct {
	StrikePrice float64       `json:"strikePrice"`
	ExpiryDate  string        `json:"expiryDate"`
	CE          *nseOCLeg     `json:"CE"`
	PE          *nseOCLeg     `json:"PE"`
}

type nseOCLeg struct {
	StrikePrice      float64 `json:"strikePrice"`
	ExpiryDate       string  `json:"expiryDate"`
	Underlying       string  `json:"underlying"`
	Identifier       string  `json:"identifier"`
	OpenInterest     int64   `json:"openInterest"`
	ChangeinOI       int64   `json:"changeinOpenInterest"`
	PChangeinOI      float64 `json:"pchangeinOpenInterest"`
	TotalTradedVolume int64  `json:"totalTradedVolume"`
	ImpliedVolatility float64 `json:"impliedVolatility"`
	LastPrice        float64 `json:"lastPrice"`
	Change           float64 `json:"change"`
	PChange          float64 `json:"pChange"`
	TotalBuyQuantity int64   `json:"totalBuyQuantity"`
	TotalSellQuantity int64  `json:"totalSellQuantity"`
	BidQty           int64   `json:"bidQty"`
	BidPrice         float64 `json:"bidprice"`
	AskQty           int64   `json:"askQty"`
	AskPrice         float64 `json:"askPrice"`
	UnderlyingValue  float64 `json:"underlyingValue"`
}

type nseVIXResponse struct {
	CurrentVIXValue float64 `json:"currentVixSnapShot"`
	Data            []struct {
		Time     string  `json:"TIMESTAMP"`
		VIXClose float64 `json:"CLOSE"`
		VIXHigh  float64 `json:"HIGH"`
		VIXLow   float64 `json:"LOW"`
		VIXOpen  float64 `json:"OPEN"`
	} `json:"data"`
}

type nseFIIDIIResponse struct {
	Data []nseFIIDIIEntry `json:"data"`
}

type nseFIIDIIEntry struct {
	Category   string  `json:"category"`    // "FII/FPI" or "DII"
	Date       string  `json:"date"`
	BuyValue   float64 `json:"buyValue"`
	SellValue  float64 `json:"sellValue"`
	NetValue   float64 `json:"netValue"`
}

// --- Public methods ---

// GetOptionChain returns the full option chain for a ticker from NSE.
func (d *NSEDerivatives) GetOptionChain(ctx context.Context, ticker string, expiry string) (*models.OptionChain, error) {
	symbol := utils.NormalizeTicker(ticker)

	cacheKey := fmt.Sprintf("nse:oc:%s:%s", symbol, expiry)
	if cached, ok := d.nse.cache.Get(cacheKey); ok {
		return cached.(*models.OptionChain), nil
	}

	if err := d.nse.ensureCookies(ctx); err != nil {
		return nil, fmt.Errorf("NSE cookie refresh: %w", err)
	}
	if err := d.nse.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/option-chain-equities?symbol=%s", nseAPIBase, symbol)
	if utils.IsIndex(symbol) {
		url = fmt.Sprintf("%s/option-chain-indices?symbol=%s", nseAPIBase, symbol)
	}

	data, err := d.nse.nseGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("NSE option chain %s: %w", symbol, err)
	}

	var resp nseOptionChainResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse NSE option chain: %w", err)
	}

	oc := d.buildOptionChain(symbol, expiry, &resp)

	d.nse.cache.Set(cacheKey, oc)
	return oc, nil
}

// GetIndiaVIX returns the current India VIX value and recent history.
func (d *NSEDerivatives) GetIndiaVIX(ctx context.Context) (*models.IndiaVIX, error) {
	cacheKey := "nse:vix"
	if cached, ok := d.nse.cache.Get(cacheKey); ok {
		return cached.(*models.IndiaVIX), nil
	}

	if err := d.nse.ensureCookies(ctx); err != nil {
		return nil, err
	}
	if err := d.nse.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/allIndices", nseAPIBase)
	data, err := d.nse.nseGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("NSE VIX: %w", err)
	}

	// Parse the allIndices response to find India VIX.
	var indices struct {
		Data []struct {
			IndexSymbol string  `json:"indexSymbol"`
			Last        float64 `json:"last"`
			Change      float64 `json:"change"`
			PChange     float64 `json:"percentChange"`
			Open        float64 `json:"open"`
			High        float64 `json:"high"`
			Low         float64 `json:"low"`
			PrevClose   float64 `json:"previousClose"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &indices); err != nil {
		return nil, fmt.Errorf("parse VIX response: %w", err)
	}

	vix := &models.IndiaVIX{
		Timestamp: utils.NowIST(),
	}
	for _, idx := range indices.Data {
		if idx.IndexSymbol == "India VIX" || idx.IndexSymbol == "INDIA VIX" {
			vix.Value = idx.Last
			vix.Change = idx.Change
			vix.ChangePct = idx.PChange
			vix.High = idx.High
			vix.Low = idx.Low
			break
		}
	}

	d.nse.cache.Set(cacheKey, vix)
	return vix, nil
}

// GetFIIDIIData returns FII/DII trading activity.
func (d *NSEDerivatives) GetFIIDIIData(ctx context.Context) (*models.FIIDIIData, error) {
	cacheKey := "nse:fiidii"
	if cached, ok := d.nse.cache.Get(cacheKey); ok {
		return cached.(*models.FIIDIIData), nil
	}

	if err := d.nse.ensureCookies(ctx); err != nil {
		return nil, err
	}
	if err := d.nse.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/fiidiiTradeReact", nseAPIBase)
	data, err := d.nse.nseGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("NSE FII/DII: %w", err)
	}

	var entries []nseFIIDIIEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse FII/DII: %w", err)
	}

	result := &models.FIIDIIData{
		Date: utils.NowIST().Format("2006-01-02"),
	}
	for _, e := range entries {
		switch {
		case contains(e.Category, "FII", "FPI"):
			result.FIIBuy = e.BuyValue
			result.FIISell = e.SellValue
			result.FIINet = e.NetValue
		case contains(e.Category, "DII"):
			result.DIIBuy = e.BuyValue
			result.DIISell = e.SellValue
			result.DIINet = e.NetValue
		}
	}

	d.nse.cache.SetWithTTL(cacheKey, result, 10*time.Minute)
	return result, nil
}

// GetFuturesData returns futures chain data for a symbol.
func (d *NSEDerivatives) GetFuturesData(ctx context.Context, ticker string) ([]models.FuturesContract, error) {
	symbol := utils.NormalizeTicker(ticker)

	cacheKey := "nse:fut:" + symbol
	if cached, ok := d.nse.cache.Get(cacheKey); ok {
		return cached.([]models.FuturesContract), nil
	}

	if err := d.nse.ensureCookies(ctx); err != nil {
		return nil, err
	}
	if err := d.nse.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/quote-derivative?symbol=%s", nseAPIBase, symbol)
	data, err := d.nse.nseGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("NSE futures %s: %w", symbol, err)
	}

	var resp struct {
		Stocks []struct {
			Metadata struct {
				InstrumentType string  `json:"instrumentType"`
				ExpiryDate     string  `json:"expiryDate"`
				StrikePrice    float64 `json:"strikePrice"`
			} `json:"metadata"`
			MarketDeptOrderBook struct {
				TradeInfo struct {
					TradedVolume int64   `json:"tradedVolume"`
					OI           int64   `json:"openInterest"`
					ChangeInOI   int64   `json:"changeinOpenInterest"`
				} `json:"tradeInfo"`
				OtherInfo struct {
					SettlementPrice float64 `json:"settlementPrice"`
					DailyVolatility float64 `json:"dailyvolatility"`
				} `json:"otherInfo"`
			} `json:"marketDeptOrderBook"`
			Underlying struct {
				UnderlyingValue float64 `json:"underlyingValue"`
			} `json:"underlyingValue"`
		} `json:"stocks"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse NSE futures: %w", err)
	}

	var futures []models.FuturesContract
	for _, s := range resp.Stocks {
		if s.Metadata.InstrumentType != "Stock Futures" && s.Metadata.InstrumentType != "Index Futures" {
			continue
		}
		spot := s.Underlying.UnderlyingValue
		ltp := s.MarketDeptOrderBook.OtherInfo.SettlementPrice
		basis := ltp - spot
		basisPct := 0.0
		if spot > 0 {
			basisPct = (basis / spot) * 100
		}

		futures = append(futures, models.FuturesContract{
			Ticker:     symbol,
			ExpiryDate: s.Metadata.ExpiryDate,
			LTP:        ltp,
			Volume:     s.MarketDeptOrderBook.TradeInfo.TradedVolume,
			OI:         s.MarketDeptOrderBook.TradeInfo.OI,
			OIChange:   s.MarketDeptOrderBook.TradeInfo.ChangeInOI,
			Basis:      basis,
			BasisPct:   basisPct,
			FetchedAt:  utils.NowIST(),
		})
	}

	d.nse.cache.Set(cacheKey, futures)
	return futures, nil
}

// --- DataSource interface stubs (partial implementation) ---

// GetQuote delegates to the underlying NSE source.
func (d *NSEDerivatives) GetQuote(ctx context.Context, ticker string) (*models.Quote, error) {
	return d.nse.GetQuote(ctx, ticker)
}

// GetHistoricalData is not supported by derivatives source.
func (d *NSEDerivatives) GetHistoricalData(_ context.Context, _ string, _, _ time.Time, _ models.Timeframe) ([]models.OHLCV, error) {
	return nil, ErrNotSupported
}

// GetFinancials is not supported by derivatives source.
func (d *NSEDerivatives) GetFinancials(_ context.Context, _ string) (*models.FinancialData, error) {
	return nil, ErrNotSupported
}

// GetStockProfile is not applicable for the derivatives source.
func (d *NSEDerivatives) GetStockProfile(_ context.Context, _ string) (*models.StockProfile, error) {
	return nil, ErrNotSupported
}

// --- Internal helpers ---

// buildOptionChain converts NSE API response into our OptionChain model.
func (d *NSEDerivatives) buildOptionChain(symbol, expiry string, resp *nseOptionChainResponse) *models.OptionChain {
	records := resp.Filtered
	if expiry == "" && len(records.ExpiryDates) > 0 {
		expiry = records.ExpiryDates[0] // nearest expiry
	}

	oc := &models.OptionChain{
		Ticker:    symbol,
		SpotPrice: records.UnderlyingValue,
		ExpiryDate: expiry,
		Expiries:  records.ExpiryDates,
		FetchedAt: utils.NowIST(),
	}

	var totalCEOI, totalPEOI int64
	for _, entry := range records.Data {
		// Filter by expiry if specified.
		if expiry != "" && entry.ExpiryDate != expiry {
			continue
		}

		if entry.CE != nil {
			totalCEOI += entry.CE.OpenInterest
			oc.Contracts = append(oc.Contracts, models.OptionContract{
				StrikePrice: entry.StrikePrice,
				OptionType:  "CE",
				ExpiryDate:  entry.ExpiryDate,
				LTP:         entry.CE.LastPrice,
				Change:      entry.CE.Change,
				ChangePct:   entry.CE.PChange,
				Volume:      entry.CE.TotalTradedVolume,
				OI:          entry.CE.OpenInterest,
				OIChange:    entry.CE.ChangeinOI,
				OIChangePct: entry.CE.PChangeinOI,
				BidPrice:    entry.CE.BidPrice,
				AskPrice:    entry.CE.AskPrice,
				BidQty:      entry.CE.BidQty,
				AskQty:      entry.CE.AskQty,
				IV:          entry.CE.ImpliedVolatility,
			})
		}

		if entry.PE != nil {
			totalPEOI += entry.PE.OpenInterest
			oc.Contracts = append(oc.Contracts, models.OptionContract{
				StrikePrice: entry.StrikePrice,
				OptionType:  "PE",
				ExpiryDate:  entry.ExpiryDate,
				LTP:         entry.PE.LastPrice,
				Change:      entry.PE.Change,
				ChangePct:   entry.PE.PChange,
				Volume:      entry.PE.TotalTradedVolume,
				OI:          entry.PE.OpenInterest,
				OIChange:    entry.PE.ChangeinOI,
				OIChangePct: entry.PE.PChangeinOI,
				BidPrice:    entry.PE.BidPrice,
				AskPrice:    entry.PE.AskPrice,
				BidQty:      entry.PE.BidQty,
				AskQty:      entry.PE.AskQty,
				IV:          entry.PE.ImpliedVolatility,
			})
		}
	}

	oc.TotalCEOI = totalCEOI
	oc.TotalPEOI = totalPEOI
	if totalCEOI > 0 {
		oc.PCR = float64(totalPEOI) / float64(totalCEOI)
	}
	oc.MaxPain = d.calculateMaxPain(oc)

	return oc
}

// calculateMaxPain computes the max pain strike from the option chain.
// Max pain is the strike price where option writers (sellers) face the least loss.
func (d *NSEDerivatives) calculateMaxPain(oc *models.OptionChain) float64 {
	if len(oc.Contracts) == 0 {
		return 0
	}

	// Collect unique strikes.
	strikeMap := make(map[float64]struct{})
	for _, c := range oc.Contracts {
		strikeMap[c.StrikePrice] = struct{}{}
	}

	// Build OI lookup by strike+type.
	ceOI := make(map[float64]int64)
	peOI := make(map[float64]int64)
	for _, c := range oc.Contracts {
		if c.OptionType == "CE" {
			ceOI[c.StrikePrice] = c.OI
		} else {
			peOI[c.StrikePrice] = c.OI
		}
	}

	var minPain float64
	minPainValue := math.MaxFloat64

	for strike := range strikeMap {
		totalPain := 0.0

		// CE pain: for each strike k, if spot (=strike being tested) > k, CE is ITM.
		for k, oi := range ceOI {
			if strike > k {
				totalPain += float64(oi) * (strike - k)
			}
		}
		// PE pain: for each strike k, if spot < k, PE is ITM.
		for k, oi := range peOI {
			if strike < k {
				totalPain += float64(oi) * (k - strike)
			}
		}

		if totalPain < minPainValue {
			minPainValue = totalPain
			minPain = strike
		}
	}

	return minPain
}

// contains checks if s contains any of the substrings (case-insensitive).
func contains(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(s) >= len(sub) {
			found := false
			for i := 0; i <= len(s)-len(sub); i++ {
				match := true
				for j := 0; j < len(sub); j++ {
					c1 := s[i+j]
					c2 := sub[j]
					if c1 >= 'A' && c1 <= 'Z' {
						c1 += 32
					}
					if c2 >= 'A' && c2 <= 'Z' {
						c2 += 32
					}
					if c1 != c2 {
						match = false
						break
					}
				}
				if match {
					found = true
					break
				}
			}
			if found {
				return true
			}
		}
	}
	return false
}
