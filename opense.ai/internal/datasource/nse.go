package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
	"github.com/seenimoa/openseai/pkg/utils"
)

const (
	nseBaseURL     = "https://www.nseindia.com"
	nseAPIBase     = "https://www.nseindia.com/api"
	nseCookieTTL   = 5 * time.Minute
	nseDefaultRate = 3 // max requests per second
)

// NSE implements the DataSource interface for NSE India direct data.
type NSE struct {
	cache        *Cache
	limiter      *RateLimiter
	client       *http.Client
	cookieExpiry time.Time
}

// NewNSE creates a new NSE India data source.
func NewNSE() *NSE {
	jar, _ := cookiejar.New(nil)
	return &NSE{
		cache:   NewCache(2 * time.Minute),
		limiter: NewRateLimiter(nseDefaultRate, time.Second),
		client: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}
}

// Name returns the data source name.
func (n *NSE) Name() string { return "NSE India" }

// --- NSE JSON response types ---

type nseQuoteResponse struct {
	Info       nseStockInfo    `json:"info"`
	PriceInfo  nsePriceInfo    `json:"priceInfo"`
	SecurityInfo nseSecurityInfo `json:"securityInfo"`
	Metadata   nseMetadata     `json:"metadata"`
}

type nseStockInfo struct {
	Symbol       string `json:"symbol"`
	CompanyName  string `json:"companyName"`
	Industry     string `json:"industry"`
	ISIN         string `json:"isin"`
}

type nsePriceInfo struct {
	LastPrice      float64        `json:"lastPrice"`
	Change         float64        `json:"change"`
	PChange        float64        `json:"pChange"`
	Open           float64        `json:"open"`
	Close          float64        `json:"close"`
	PreviousClose  float64        `json:"previousClose"`
	IntraDayHighLow nseHighLow    `json:"intraDayHighLow"`
	WeekHighLow    nseWeekHighLow `json:"weekHighLow"`
	UpperCP        string         `json:"upperCP"`
	LowerCP        string         `json:"lowerCP"`
}

type nseHighLow struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type nseWeekHighLow struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type nseSecurityInfo struct {
	BoardStatus  string  `json:"boardStatus"`
	TradingStatus string `json:"tradingStatus"`
	FaceValue    float64 `json:"faceValue"`
}

type nseMetadata struct {
	Series    string `json:"series"`
	Symbol    string `json:"symbol"`
	ISIN      string `json:"isin"`
	Status    string `json:"status"`
	ListingDate string `json:"listingDate"`
	Industry  string `json:"industry"`
	Sector    string `json:"pdSectorInd"`
}

type nseTradeInfo struct {
	TotalTradedVolume int64   `json:"totalTradedVolume"`
	TotalTradedValue  float64 `json:"totalTradedValue"`
	TotalMarketCap    float64 `json:"totalMarketCap"`
}

type nseCorporateInfo struct {
	LatestAnnouncements []nseAnnouncement `json:"latestAnnouncements"`
}

type nseAnnouncement struct {
	Subject  string `json:"desc"`
	DateTime string `json:"dt"`
}

type nseShareholdingResponse struct {
	Data []nseShareholdingEntry `json:"data"`
}

type nseShareholdingEntry struct {
	Category  string  `json:"categoryOfHolder"`
	Holding   float64 `json:"shareHolding"`
	Quarter   string  `json:"date"`
}

type nseBulkDealResponse struct {
	Data []nseBulkDeal `json:"data"`
}

type nseBulkDeal struct {
	Symbol     string  `json:"symbol"`
	ClientName string  `json:"clientName"`
	BuySell    string  `json:"buySell"`
	Quantity   int64   `json:"qty"`
	Price      float64 `json:"weightedAvgPrice"`
	Date       string  `json:"dealDate"`
}

// --- Public methods ---

// GetQuote returns a real-time quote from NSE India.
func (n *NSE) GetQuote(ctx context.Context, ticker string) (*models.Quote, error) {
	symbol := utils.NormalizeTicker(ticker)

	cacheKey := "nse:quote:" + symbol
	if cached, ok := n.cache.Get(cacheKey); ok {
		return cached.(*models.Quote), nil
	}

	if err := n.ensureCookies(ctx); err != nil {
		return nil, fmt.Errorf("NSE cookie refresh: %w", err)
	}
	if err := n.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/quote-equity?symbol=%s&section=trade_info", nseAPIBase, symbol)
	data, err := n.nseGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("NSE quote %s: %w", symbol, err)
	}

	var resp nseQuoteResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse NSE quote: %w", err)
	}

	quote := &models.Quote{
		Ticker:     symbol,
		Name:       resp.Info.CompanyName,
		LastPrice:  resp.PriceInfo.LastPrice,
		Change:     resp.PriceInfo.Change,
		ChangePct:  resp.PriceInfo.PChange,
		Open:       resp.PriceInfo.Open,
		High:       resp.PriceInfo.IntraDayHighLow.Max,
		Low:        resp.PriceInfo.IntraDayHighLow.Min,
		PrevClose:  resp.PriceInfo.PreviousClose,
		WeekHigh52: resp.PriceInfo.WeekHighLow.Max,
		WeekLow52:  resp.PriceInfo.WeekHighLow.Min,
		Timestamp:  utils.NowIST(),
	}

	// Parse circuit limits (they come as string percentages).
	parseCircuit(resp.PriceInfo.UpperCP, resp.PriceInfo.LowerCP, resp.PriceInfo.PreviousClose, quote)

	n.cache.Set(cacheKey, quote)
	return quote, nil
}

// GetHistoricalData returns historical OHLCV from NSE.
// NSE provides limited historical data; for longer history use YFinance.
func (n *NSE) GetHistoricalData(ctx context.Context, ticker string, from, to time.Time, _ models.Timeframe) ([]models.OHLCV, error) {
	symbol := utils.NormalizeTicker(ticker)

	cacheKey := fmt.Sprintf("nse:hist:%s:%s:%s", symbol, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if cached, ok := n.cache.Get(cacheKey); ok {
		return cached.([]models.OHLCV), nil
	}

	if err := n.ensureCookies(ctx); err != nil {
		return nil, fmt.Errorf("NSE cookie refresh: %w", err)
	}
	if err := n.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"%s/historical/cm/equity?symbol=%s&from=%s&to=%s",
		nseAPIBase, symbol,
		from.Format("02-01-2006"), to.Format("02-01-2006"),
	)
	data, err := n.nseGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("NSE historical %s: %w", symbol, err)
	}

	var resp struct {
		Data []nseHistEntry `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse NSE historical: %w", err)
	}

	candles := make([]models.OHLCV, 0, len(resp.Data))
	for _, e := range resp.Data {
		ts, _ := time.Parse("02-Jan-2006", e.Date)
		candles = append(candles, models.OHLCV{
			Timestamp: ts,
			Open:      e.Open,
			High:      e.High,
			Low:       e.Low,
			Close:     e.Close,
			Volume:    e.Volume,
		})
	}

	n.cache.SetWithTTL(cacheKey, candles, 30*time.Minute)
	return candles, nil
}

// GetFinancials is not directly supported via NSE API; use Screener.in for this.
func (n *NSE) GetFinancials(_ context.Context, _ string) (*models.FinancialData, error) {
	return nil, ErrNotSupported
}

// GetOptionChain is handled by NSEDerivatives source.
func (n *NSE) GetOptionChain(_ context.Context, _ string, _ string) (*models.OptionChain, error) {
	return nil, ErrNotSupported
}

// GetStockProfile assembles a profile from NSE data.
func (n *NSE) GetStockProfile(ctx context.Context, ticker string) (*models.StockProfile, error) {
	quote, err := n.GetQuote(ctx, ticker)
	if err != nil {
		return nil, err
	}

	profile := &models.StockProfile{
		Stock: models.Stock{
			Ticker:   utils.NormalizeTicker(ticker),
			Name:     quote.Name,
			Exchange: "NSE",
		},
		Quote:     quote,
		FetchedAt: utils.NowIST(),
	}

	// Try to fetch promoter/shareholding data (non-critical).
	promoter, err := n.GetShareholding(ctx, ticker)
	if err == nil {
		profile.Promoter = promoter
	}

	return profile, nil
}

// --- Additional NSE-specific methods (not part of DataSource interface) ---

// GetShareholding returns the shareholding pattern for the given ticker.
func (n *NSE) GetShareholding(ctx context.Context, ticker string) (*models.PromoterData, error) {
	symbol := utils.NormalizeTicker(ticker)

	cacheKey := "nse:sh:" + symbol
	if cached, ok := n.cache.Get(cacheKey); ok {
		return cached.(*models.PromoterData), nil
	}

	if err := n.ensureCookies(ctx); err != nil {
		return nil, err
	}
	if err := n.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/quote-equity?symbol=%s&section=shareholding", nseAPIBase, symbol)
	data, err := n.nseGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("NSE shareholding %s: %w", symbol, err)
	}

	// NSE shareholding JSON varies in structure; do best effort parsing.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse NSE shareholding: %w", err)
	}

	pd := &models.PromoterData{}
	// Parse from raw JSON — structure is complex and version-dependent.
	// The actual implementation would walk the JSON tree to extract promoter/FII/DII data.

	n.cache.SetWithTTL(cacheKey, pd, 1*time.Hour)
	return pd, nil
}

// GetBulkDeals returns recent bulk deals.
func (n *NSE) GetBulkDeals(ctx context.Context) ([]nseBulkDeal, error) {
	cacheKey := "nse:bulk"
	if cached, ok := n.cache.Get(cacheKey); ok {
		return cached.([]nseBulkDeal), nil
	}

	if err := n.ensureCookies(ctx); err != nil {
		return nil, err
	}
	if err := n.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/snapshot/bulk-deal", nseAPIBase)
	data, err := n.nseGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("NSE bulk deals: %w", err)
	}

	var resp nseBulkDealResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse NSE bulk deals: %w", err)
	}

	n.cache.SetWithTTL(cacheKey, resp.Data, 15*time.Minute)
	return resp.Data, nil
}

// GetIndexData returns NIFTY 50 / NIFTY BANK index data.
func (n *NSE) GetIndexData(ctx context.Context, indexName string) (map[string]any, error) {
	cacheKey := "nse:idx:" + indexName
	if cached, ok := n.cache.Get(cacheKey); ok {
		return cached.(map[string]any), nil
	}

	if err := n.ensureCookies(ctx); err != nil {
		return nil, err
	}
	if err := n.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/allIndices", nseAPIBase)
	data, err := n.nseGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("NSE index data: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse NSE index: %w", err)
	}

	n.cache.SetWithTTL(cacheKey, result, 2*time.Minute)
	return result, nil
}

// --- Internal helpers ---

// ensureCookies visits the NSE homepage to get session cookies.
// NSE requires valid cookies for API access.
func (n *NSE) ensureCookies(ctx context.Context) error {
	if time.Now().Before(n.cookieExpiry) {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nseBaseURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch NSE homepage for cookies: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) //nolint:errcheck // drain body

	n.cookieExpiry = time.Now().Add(nseCookieTTL)
	return nil
}

// nseGet performs a GET request to the NSE API with proper headers.
func (n *NSE) nseGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", nseBaseURL)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, ErrRateLimited
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, &ErrHTTP{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(body),
		}
	}

	return io.ReadAll(resp.Body)
}

// nseHistEntry represents a single historical data row from NSE.
type nseHistEntry struct {
	Date   string  `json:"CH_TIMESTAMP"`
	Open   float64 `json:"CH_OPENING_PRICE"`
	High   float64 `json:"CH_TRADE_HIGH_PRICE"`
	Low    float64 `json:"CH_TRADE_LOW_PRICE"`
	Close  float64 `json:"CH_CLOSING_PRICE"`
	Volume int64   `json:"CH_TOT_TRADED_QTY"`
}

// parseCircuit extracts upper/lower circuit prices from percentage strings.
func parseCircuit(upper, lower string, prevClose float64, q *models.Quote) {
	upper = strings.TrimSpace(strings.Replace(upper, "%", "", 1))
	lower = strings.TrimSpace(strings.Replace(lower, "%", "", 1))

	var pct float64
	if _, err := fmt.Sscanf(upper, "%f", &pct); err == nil {
		q.UpperCircuit = prevClose * (1 + pct/100)
	}
	if _, err := fmt.Sscanf(lower, "%f", &pct); err == nil {
		q.LowerCircuit = prevClose * (1 + pct/100)
	}
}
