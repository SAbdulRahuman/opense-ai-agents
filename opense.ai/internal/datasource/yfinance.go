package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
	"github.com/seenimoa/openseai/pkg/utils"
)

// YFinance implements the DataSource interface using Yahoo Finance API.
type YFinance struct {
	cache   *Cache
	limiter *RateLimiter
}

// NewYFinance creates a new Yahoo Finance data source.
func NewYFinance() *YFinance {
	return &YFinance{
		cache:   NewCache(5 * time.Minute),
		limiter: NewRateLimiter(5, time.Second), // 5 req/s
	}
}

// Name returns the data source name.
func (y *YFinance) Name() string { return "Yahoo Finance" }

// --- Yahoo Finance v8 API types ---

type yfQuoteResponse struct {
	QuoteResponse struct {
		Result []yfQuoteResult `json:"result"`
		Error  *yfError        `json:"error"`
	} `json:"quoteResponse"`
}

type yfQuoteResult struct {
	Symbol                     string  `json:"symbol"`
	ShortName                  string  `json:"shortName"`
	LongName                   string  `json:"longName"`
	RegularMarketPrice         float64 `json:"regularMarketPrice"`
	RegularMarketChange        float64 `json:"regularMarketChange"`
	RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
	RegularMarketOpen          float64 `json:"regularMarketOpen"`
	RegularMarketDayHigh       float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow        float64 `json:"regularMarketDayLow"`
	RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
	RegularMarketVolume        int64   `json:"regularMarketVolume"`
	MarketCap                  float64 `json:"marketCap"`
	FiftyTwoWeekHigh           float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow            float64 `json:"fiftyTwoWeekLow"`
	TrailingPE                 float64 `json:"trailingPE"`
	PriceToBook                float64 `json:"priceToBook"`
	DividendYield              float64 `json:"dividendYield"`
	RegularMarketTime          int64   `json:"regularMarketTime"`
}

type yfChartResponse struct {
	Chart struct {
		Result []yfChartResult `json:"result"`
		Error  *yfError        `json:"error"`
	} `json:"chart"`
}

type yfChartResult struct {
	Meta       yfChartMeta    `json:"meta"`
	Timestamp  []int64        `json:"timestamp"`
	Indicators yfIndicators   `json:"indicators"`
}

type yfChartMeta struct {
	Symbol             string  `json:"symbol"`
	Currency           string  `json:"currency"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
}

type yfIndicators struct {
	Quote    []yfOHLCV    `json:"quote"`
	AdjClose []yfAdjClose `json:"adjclose"`
}

type yfOHLCV struct {
	Open   []*float64 `json:"open"`
	High   []*float64 `json:"high"`
	Low    []*float64 `json:"low"`
	Close  []*float64 `json:"close"`
	Volume []*int64   `json:"volume"`
}

type yfAdjClose struct {
	AdjClose []*float64 `json:"adjclose"`
}

type yfFinancialsResponse struct {
	QuoteSummary struct {
		Result []yfFinancialResult `json:"result"`
		Error  *yfError            `json:"error"`
	} `json:"quoteSummary"`
}

type yfFinancialResult struct {
	IncomeStatementHistory         *yfStatementHistory `json:"incomeStatementHistory"`
	IncomeStatementHistoryQuarterly *yfStatementHistory `json:"incomeStatementHistoryQuarterly"`
	BalanceSheetHistory            *yfStatementHistory `json:"balanceSheetHistory"`
	BalanceSheetHistoryQuarterly   *yfStatementHistory `json:"balanceSheetHistoryQuarterly"`
	CashflowStatementHistory       *yfStatementHistory `json:"cashflowStatementHistory"`
	CashflowStatementHistoryQuarterly *yfStatementHistory `json:"cashflowStatementHistoryQuarterly"`
}

type yfStatementHistory struct {
	Statements []map[string]yfFinVal `json:"incomeStatementHistory,omitempty"`
}

type yfFinVal struct {
	Raw float64 `json:"raw"`
	Fmt string  `json:"fmt"`
}

type yfError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// --- Public methods ---

// GetQuote returns a real-time quote from Yahoo Finance.
func (y *YFinance) GetQuote(ctx context.Context, ticker string) (*models.Quote, error) {
	yfTicker := utils.ToYFinanceTicker(ticker)

	// Check cache.
	cacheKey := "quote:" + yfTicker
	if cached, ok := y.cache.Get(cacheKey); ok {
		return cached.(*models.Quote), nil
	}

	if err := y.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s", yfTicker)
	body, _, err := doGet(ctx, url, map[string]string{
		"Accept": "application/json",
	})
	if err != nil {
		return nil, fmt.Errorf("yfinance quote %s: %w", yfTicker, err)
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var resp yfQuoteResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse yfinance quote: %w", err)
	}

	if resp.QuoteResponse.Error != nil {
		return nil, fmt.Errorf("yfinance API error: %s", resp.QuoteResponse.Error.Description)
	}
	if len(resp.QuoteResponse.Result) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrTickerNotFound, ticker)
	}

	r := resp.QuoteResponse.Result[0]
	quote := &models.Quote{
		Ticker:        utils.FromYFinanceTicker(r.Symbol),
		Name:          coalesce(r.LongName, r.ShortName),
		LastPrice:     r.RegularMarketPrice,
		Change:        r.RegularMarketChange,
		ChangePct:     r.RegularMarketChangePercent,
		Open:          r.RegularMarketOpen,
		High:          r.RegularMarketDayHigh,
		Low:           r.RegularMarketDayLow,
		PrevClose:     r.RegularMarketPreviousClose,
		Volume:        r.RegularMarketVolume,
		WeekHigh52:    r.FiftyTwoWeekHigh,
		WeekLow52:     r.FiftyTwoWeekLow,
		MarketCap:     r.MarketCap,
		PE:            r.TrailingPE,
		PB:            r.PriceToBook,
		DividendYield: r.DividendYield * 100, // convert from ratio to percentage
		Timestamp:     time.Unix(r.RegularMarketTime, 0),
	}

	y.cache.Set(cacheKey, quote)
	return quote, nil
}

// GetHistoricalData returns OHLCV candles from Yahoo Finance chart API.
func (y *YFinance) GetHistoricalData(ctx context.Context, ticker string, from, to time.Time, tf models.Timeframe) ([]models.OHLCV, error) {
	yfTicker := utils.ToYFinanceTicker(ticker)

	cacheKey := fmt.Sprintf("hist:%s:%d:%d:%s", yfTicker, from.Unix(), to.Unix(), tf)
	if cached, ok := y.cache.Get(cacheKey); ok {
		return cached.([]models.OHLCV), nil
	}

	if err := y.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	interval := yfInterval(tf)
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=%s",
		yfTicker, from.Unix(), to.Unix(), interval,
	)

	body, _, err := doGet(ctx, url, map[string]string{
		"Accept": "application/json",
	})
	if err != nil {
		return nil, fmt.Errorf("yfinance chart %s: %w", yfTicker, err)
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var resp yfChartResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse yfinance chart: %w", err)
	}

	if resp.Chart.Error != nil {
		return nil, fmt.Errorf("yfinance chart error: %s", resp.Chart.Error.Description)
	}
	if len(resp.Chart.Result) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrTickerNotFound, ticker)
	}

	result := resp.Chart.Result[0]
	candles := parseYFCandles(result)

	y.cache.SetWithTTL(cacheKey, candles, 15*time.Minute)
	return candles, nil
}

// GetFinancials returns financial statements from Yahoo Finance.
func (y *YFinance) GetFinancials(ctx context.Context, ticker string) (*models.FinancialData, error) {
	yfTicker := utils.ToYFinanceTicker(ticker)

	cacheKey := "fin:" + yfTicker
	if cached, ok := y.cache.Get(cacheKey); ok {
		return cached.(*models.FinancialData), nil
	}

	if err := y.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	modules := "incomeStatementHistory,incomeStatementHistoryQuarterly,balanceSheetHistory,balanceSheetHistoryQuarterly,cashflowStatementHistory,cashflowStatementHistoryQuarterly"
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=%s",
		yfTicker, modules,
	)

	body, _, err := doGet(ctx, url, map[string]string{
		"Accept": "application/json",
	})
	if err != nil {
		return nil, fmt.Errorf("yfinance financials %s: %w", yfTicker, err)
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var resp yfFinancialsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse yfinance financials: %w", err)
	}

	if resp.QuoteSummary.Error != nil {
		return nil, fmt.Errorf("yfinance API error: %s", resp.QuoteSummary.Error.Description)
	}
	if len(resp.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrTickerNotFound, ticker)
	}

	fd := &models.FinancialData{
		Ticker: utils.FromYFinanceTicker(yfTicker),
	}
	// Financials parsing from the raw Yahoo Finance JSON is complex and
	// version-dependent. We store the raw response and parse on demand.
	// For now, return a skeleton.

	y.cache.SetWithTTL(cacheKey, fd, 1*time.Hour)
	return fd, nil
}

// GetOptionChain is not supported by the Yahoo Finance source for Indian markets.
func (y *YFinance) GetOptionChain(_ context.Context, _ string, _ string) (*models.OptionChain, error) {
	return nil, ErrNotSupported
}

// GetStockProfile assembles a stock profile from Yahoo Finance data.
func (y *YFinance) GetStockProfile(ctx context.Context, ticker string) (*models.StockProfile, error) {
	quote, err := y.GetQuote(ctx, ticker)
	if err != nil {
		return nil, err
	}

	profile := &models.StockProfile{
		Stock: models.Stock{
			Ticker:    utils.NormalizeTicker(ticker),
			NSETicker: utils.ToYFinanceTicker(ticker),
			Name:      quote.Name,
			Exchange:  "NSE",
			MarketCap: quote.MarketCap,
		},
		Quote:     quote,
		FetchedAt: time.Now(),
	}

	return profile, nil
}

// --- Helpers ---

func parseYFCandles(result yfChartResult) []models.OHLCV {
	if len(result.Indicators.Quote) == 0 {
		return nil
	}

	q := result.Indicators.Quote[0]
	var adjCloses []*float64
	if len(result.Indicators.AdjClose) > 0 {
		adjCloses = result.Indicators.AdjClose[0].AdjClose
	}

	candles := make([]models.OHLCV, 0, len(result.Timestamp))
	for i, ts := range result.Timestamp {
		c := models.OHLCV{
			Timestamp: time.Unix(ts, 0),
		}
		if i < len(q.Open) && q.Open[i] != nil {
			c.Open = *q.Open[i]
		}
		if i < len(q.High) && q.High[i] != nil {
			c.High = *q.High[i]
		}
		if i < len(q.Low) && q.Low[i] != nil {
			c.Low = *q.Low[i]
		}
		if i < len(q.Close) && q.Close[i] != nil {
			c.Close = *q.Close[i]
		}
		if i < len(q.Volume) && q.Volume[i] != nil {
			c.Volume = *q.Volume[i]
		}
		if i < len(adjCloses) && adjCloses[i] != nil {
			c.AdjClose = *adjCloses[i]
		}
		candles = append(candles, c)
	}
	return candles
}

func yfInterval(tf models.Timeframe) string {
	switch tf {
	case models.Timeframe1Min:
		return "1m"
	case models.Timeframe5Min:
		return "5m"
	case models.Timeframe15Min:
		return "15m"
	case models.Timeframe1Hour:
		return "1h"
	case models.Timeframe1Day:
		return "1d"
	case models.Timeframe1Week:
		return "1wk"
	case models.Timeframe1Mon:
		return "1mo"
	default:
		return "1d"
	}
}

func coalesce(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
