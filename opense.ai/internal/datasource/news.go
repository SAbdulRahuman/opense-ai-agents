package datasource

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"

	"github.com/seenimoa/openseai/pkg/models"
	"github.com/seenimoa/openseai/pkg/utils"
)

// NewsSource represents Indian financial news source configuration.
type NewsSource struct {
	Name    string
	RSSURL  string
	BaseURL string
}

// DefaultNewsSources lists the configured Indian financial news RSS feeds.
var DefaultNewsSources = []NewsSource{
	{
		Name:    "Moneycontrol",
		RSSURL:  "https://www.moneycontrol.com/rss/marketreports.xml",
		BaseURL: "https://www.moneycontrol.com",
	},
	{
		Name:    "Economic Times Markets",
		RSSURL:  "https://economictimes.indiatimes.com/markets/rssfeeds/1977021501.cms",
		BaseURL: "https://economictimes.indiatimes.com",
	},
	{
		Name:    "LiveMint Markets",
		RSSURL:  "https://www.livemint.com/rss/markets",
		BaseURL: "https://www.livemint.com",
	},
	{
		Name:    "Business Standard Markets",
		RSSURL:  "https://www.business-standard.com/rss/markets-106.rss",
		BaseURL: "https://www.business-standard.com",
	},
}

// News implements financial news fetching from Indian sources.
type News struct {
	sources []NewsSource
	cache   *Cache
	limiter *RateLimiter
	parser  *gofeed.Parser
}

// NewNews creates a new news data source with default Indian sources.
func NewNews() *News {
	return &News{
		sources: DefaultNewsSources,
		cache:   NewCache(10 * time.Minute),
		limiter: NewRateLimiter(2, time.Second), // conservative: 2 req/s
		parser:  gofeed.NewParser(),
	}
}

// NewNewsWithSources creates a news data source with custom sources.
func NewNewsWithSources(sources []NewsSource) *News {
	return &News{
		sources: sources,
		cache:   NewCache(10 * time.Minute),
		limiter: NewRateLimiter(2, time.Second),
		parser:  gofeed.NewParser(),
	}
}

// Name returns the data source name.
func (n *News) Name() string { return "Indian News" }

// --- Public methods ---

// GetMarketNews returns recent market news from all configured sources.
func (n *News) GetMarketNews(ctx context.Context, limit int) ([]models.NewsArticle, error) {
	cacheKey := fmt.Sprintf("news:market:%d", limit)
	if cached, ok := n.cache.Get(cacheKey); ok {
		return cached.([]models.NewsArticle), nil
	}

	var allArticles []models.NewsArticle
	for _, src := range n.sources {
		articles, err := n.fetchRSS(ctx, src)
		if err != nil {
			// Non-critical: skip failed sources.
			continue
		}
		allArticles = append(allArticles, articles...)
	}

	// Sort by published date (newest first) — already roughly sorted per source.
	sortArticlesByDate(allArticles)

	if limit > 0 && len(allArticles) > limit {
		allArticles = allArticles[:limit]
	}

	n.cache.Set(cacheKey, allArticles)
	return allArticles, nil
}

// GetStockNews returns news articles related to a specific ticker.
func (n *News) GetStockNews(ctx context.Context, ticker string, limit int) ([]models.NewsArticle, error) {
	symbol := utils.NormalizeTicker(ticker)

	cacheKey := fmt.Sprintf("news:stock:%s:%d", symbol, limit)
	if cached, ok := n.cache.Get(cacheKey); ok {
		return cached.([]models.NewsArticle), nil
	}

	// First get all market news, then filter by ticker mention.
	allNews, err := n.GetMarketNews(ctx, 0)
	if err != nil {
		return nil, err
	}

	var filtered []models.NewsArticle
	keywords := tickerKeywords(symbol)
	for _, a := range allNews {
		if matchesAny(a.Title+" "+a.Summary, keywords) {
			filtered = append(filtered, a)
		}
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	n.cache.Set(cacheKey, filtered)
	return filtered, nil
}

// GetSectorNews returns news related to a market sector.
func (n *News) GetSectorNews(ctx context.Context, sector string, limit int) ([]models.NewsArticle, error) {
	cacheKey := fmt.Sprintf("news:sector:%s:%d", sector, limit)
	if cached, ok := n.cache.Get(cacheKey); ok {
		return cached.([]models.NewsArticle), nil
	}

	allNews, err := n.GetMarketNews(ctx, 0)
	if err != nil {
		return nil, err
	}

	sectorLower := strings.ToLower(sector)
	var filtered []models.NewsArticle
	for _, a := range allNews {
		content := strings.ToLower(a.Title + " " + a.Summary)
		if strings.Contains(content, sectorLower) {
			filtered = append(filtered, a)
		}
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	n.cache.Set(cacheKey, filtered)
	return filtered, nil
}

// --- DataSource interface (partial) ---

// GetQuote is not supported by the news source.
func (n *News) GetQuote(_ context.Context, _ string) (*models.Quote, error) {
	return nil, ErrNotSupported
}

// GetHistoricalData is not supported by the news source.
func (n *News) GetHistoricalData(_ context.Context, _ string, _, _ time.Time, _ models.Timeframe) ([]models.OHLCV, error) {
	return nil, ErrNotSupported
}

// GetFinancials is not supported by the news source.
func (n *News) GetFinancials(_ context.Context, _ string) (*models.FinancialData, error) {
	return nil, ErrNotSupported
}

// GetOptionChain is not supported by the news source.
func (n *News) GetOptionChain(_ context.Context, _ string, _ string) (*models.OptionChain, error) {
	return nil, ErrNotSupported
}

// GetStockProfile is not supported by the news source.
func (n *News) GetStockProfile(_ context.Context, _ string) (*models.StockProfile, error) {
	return nil, ErrNotSupported
}

// --- Internal helpers ---

// fetchRSS parses an RSS feed and returns articles.
func (n *News) fetchRSS(ctx context.Context, src NewsSource) ([]models.NewsArticle, error) {
	if err := n.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	feed, err := n.parser.ParseURLWithContext(src.RSSURL, ctx)
	if err != nil {
		return nil, fmt.Errorf("parse RSS %s: %w", src.Name, err)
	}

	articles := make([]models.NewsArticle, 0, len(feed.Items))
	for _, item := range feed.Items {
		a := models.NewsArticle{
			Title:   item.Title,
			URL:     item.Link,
			Source:  src.Name,
			Summary: cleanHTML(item.Description),
		}
		if item.PublishedParsed != nil {
			a.PublishedAt = *item.PublishedParsed
		}
		articles = append(articles, a)
	}

	return articles, nil
}

// cleanHTML strips HTML tags from a string using goquery.
func cleanHTML(s string) string {
	if s == "" {
		return ""
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<body>" + s + "</body>"))
	if err != nil {
		return s
	}
	return strings.TrimSpace(doc.Text())
}

// tickerKeywords returns search keywords for a ticker.
// For example, "RELIANCE" → ["reliance", "reliance industries", "ril"].
func tickerKeywords(ticker string) []string {
	t := strings.ToLower(ticker)
	keywords := []string{t}

	// Add common name mappings.
	nameMap := map[string][]string{
		"reliance":    {"reliance industries", "ril", "mukesh ambani"},
		"tcs":         {"tata consultancy", "tcs"},
		"hdfcbank":    {"hdfc bank"},
		"infy":        {"infosys"},
		"icicibank":   {"icici bank"},
		"hindunilvr":  {"hindustan unilever", "hul"},
		"sbin":        {"sbi", "state bank"},
		"bhartiartl":  {"bharti airtel", "airtel"},
		"kotakbank":   {"kotak mahindra", "kotak bank"},
		"lt":          {"larsen", "l&t"},
		"bajfinance":  {"bajaj finance"},
		"axisbank":    {"axis bank"},
		"maruti":      {"maruti suzuki"},
		"tatamotors":  {"tata motors"},
		"tatasteel":   {"tata steel"},
		"wipro":       {"wipro"},
		"hcltech":     {"hcl tech", "hcl technologies"},
		"asianpaint":  {"asian paints"},
		"sunpharma":   {"sun pharma", "sun pharmaceutical"},
		"ongc":        {"ongc", "oil and natural gas"},
	}

	if extra, ok := nameMap[t]; ok {
		keywords = append(keywords, extra...)
	}

	return keywords
}

// matchesAny checks if text contains any of the keywords (case-insensitive).
func matchesAny(text string, keywords []string) bool {
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// sortArticlesByDate sorts articles by published date (newest first).
// Simple insertion sort — fine for small slices.
func sortArticlesByDate(articles []models.NewsArticle) {
	for i := 1; i < len(articles); i++ {
		key := articles[i]
		j := i - 1
		for j >= 0 && articles[j].PublishedAt.Before(key.PublishedAt) {
			articles[j+1] = articles[j]
			j--
		}
		articles[j+1] = key
	}
}
