package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
	"github.com/seenimoa/openseai/pkg/utils"
)

// FIIDII fetches FII/DII activity data from NSE and NSDL sources.
type FIIDII struct {
	nse   *NSE
	cache *Cache
}

// NewFIIDII creates a new FII/DII data source.
// It reuses the NSE client for cookie management.
func NewFIIDII(nse *NSE) *FIIDII {
	return &FIIDII{
		nse:   nse,
		cache: NewCache(15 * time.Minute),
	}
}

// Name returns the data source name.
func (f *FIIDII) Name() string { return "FII/DII Activity" }

// --- Response types ---

type nsdlFPIResponse struct {
	Data []nsdlFPIEntry `json:"data"`
}

type nsdlFPIEntry struct {
	Date        string  `json:"date"`
	Category    string  `json:"category"`
	BuyValue    float64 `json:"buyValue"`
	SellValue   float64 `json:"sellValue"`
	NetValue    float64 `json:"netValue"`
	AssetClass  string  `json:"assetClass"` // "Equity", "Debt", etc.
}

// --- Public methods ---

// GetFIIDIIActivity returns today's FII/DII cash market activity.
func (f *FIIDII) GetFIIDIIActivity(ctx context.Context) (*models.FIIDIIData, error) {
	cacheKey := "fiidii:today"
	if cached, ok := f.cache.Get(cacheKey); ok {
		return cached.(*models.FIIDIIData), nil
	}

	if err := f.nse.ensureCookies(ctx); err != nil {
		return nil, fmt.Errorf("cookie refresh: %w", err)
	}
	if err := f.nse.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/fiidiiTradeReact", nseAPIBase)
	data, err := f.nse.nseGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("FII/DII activity: %w", err)
	}

	var entries []struct {
		Category  string  `json:"category"`
		Date      string  `json:"date"`
		BuyValue  float64 `json:"buyValue"`
		SellValue float64 `json:"sellValue"`
		NetValue  float64 `json:"netValue"`
	}
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

	f.cache.Set(cacheKey, result)
	return result, nil
}

// GetHistoricalFIIDII returns historical FII/DII activity for a date range.
func (f *FIIDII) GetHistoricalFIIDII(ctx context.Context, from, to time.Time) ([]models.FIIDIIData, error) {
	cacheKey := fmt.Sprintf("fiidii:hist:%s:%s", from.Format("20060102"), to.Format("20060102"))
	if cached, ok := f.cache.Get(cacheKey); ok {
		return cached.([]models.FIIDIIData), nil
	}

	if err := f.nse.ensureCookies(ctx); err != nil {
		return nil, err
	}
	if err := f.nse.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	// NSE provides FII/DII data through participant-wise trading data.
	url := fmt.Sprintf(
		"%s/reports/fii-dii?startDate=%s&endDate=%s",
		nseAPIBase,
		from.Format("02-01-2006"),
		to.Format("02-01-2006"),
	)

	data, err := f.nse.nseGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("historical FII/DII: %w", err)
	}

	var raw []struct {
		Category  string  `json:"category"`
		Date      string  `json:"date"`
		BuyValue  float64 `json:"buyValue"`
		SellValue float64 `json:"sellValue"`
		NetValue  float64 `json:"netValue"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse historical FII/DII: %w", err)
	}

	// Group by date.
	dateMap := make(map[string]*models.FIIDIIData)
	for _, e := range raw {
		d, ok := dateMap[e.Date]
		if !ok {
			d = &models.FIIDIIData{
				Date: e.Date,
			}
			dateMap[e.Date] = d
		}
		switch {
		case contains(e.Category, "FII", "FPI"):
			d.FIIBuy = e.BuyValue
			d.FIISell = e.SellValue
			d.FIINet = e.NetValue
		case contains(e.Category, "DII"):
			d.DIIBuy = e.BuyValue
			d.DIISell = e.SellValue
			d.DIINet = e.NetValue
		}
	}

	var result []models.FIIDIIData
	for _, v := range dateMap {
		result = append(result, *v)
	}

	f.cache.SetWithTTL(cacheKey, result, 30*time.Minute)
	return result, nil
}

// --- DataSource interface stubs ---

// GetQuote is not supported.
func (f *FIIDII) GetQuote(_ context.Context, _ string) (*models.Quote, error) {
	return nil, ErrNotSupported
}

// GetHistoricalData is not supported.
func (f *FIIDII) GetHistoricalData(_ context.Context, _ string, _, _ time.Time, _ models.Timeframe) ([]models.OHLCV, error) {
	return nil, ErrNotSupported
}

// GetFinancials is not supported.
func (f *FIIDII) GetFinancials(_ context.Context, _ string) (*models.FinancialData, error) {
	return nil, ErrNotSupported
}

// GetOptionChain is not supported.
func (f *FIIDII) GetOptionChain(_ context.Context, _ string, _ string) (*models.OptionChain, error) {
	return nil, ErrNotSupported
}

// GetStockProfile is not supported.
func (f *FIIDII) GetStockProfile(_ context.Context, _ string) (*models.StockProfile, error) {
	return nil, ErrNotSupported
}
