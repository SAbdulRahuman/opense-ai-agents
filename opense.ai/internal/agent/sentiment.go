package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/agent/prompts"
	"github.com/seenimoa/openseai/internal/analysis/sentiment"
	"github.com/seenimoa/openseai/internal/datasource"
	"github.com/seenimoa/openseai/internal/llm"
	"github.com/seenimoa/openseai/pkg/models"
)

// SentimentAgent is the Sentiment Analyst specialized agent.
// It analyzes market news, social sentiment, and detects catalysts.
type SentimentAgent struct {
	*BaseAgent
	news *datasource.News
}

// NewSentimentAgent creates a Sentiment Analyst agent.
func NewSentimentAgent(provider llm.LLMProvider, news *datasource.News, opts *llm.ChatOptions) *SentimentAgent {
	agent := &SentimentAgent{news: news}

	tools := agent.buildTools()

	systemPrompt := prompts.SentimentSystemPrompt + prompts.IndianMarketPromptSuffix()

	agent.BaseAgent = NewBaseAgent(BaseAgentConfig{
		Name:         prompts.AgentSentiment,
		Role:         "Sentiment Analyst — News sentiment, social signals, catalyst detection",
		SystemPrompt: systemPrompt,
		Provider:     provider,
		Tools:        tools,
		ChatOptions:  opts,
		MemorySize:   30,
		MaxToolIter:  6,
	})

	return agent
}

func (a *SentimentAgent) buildTools() []llm.Tool {
	return []llm.Tool{
		{
			Name:        "get_stock_news",
			Description: "Fetch recent news articles for a specific stock from Indian financial news sources (Moneycontrol, Economic Times, Livemint, etc.)",
			Parameters: llm.ObjectSchema("Stock news parameters",
				map[string]*llm.JSONSchema{
					"ticker": llm.StringProp("NSE ticker symbol (e.g., RELIANCE, TCS)"),
					"limit":  llm.IntProp("Maximum number of articles to fetch (default: 20)"),
				},
				"ticker",
			),
			Handler: a.handleGetStockNews,
		},
		{
			Name:        "get_market_news",
			Description: "Fetch general market news from Indian financial news sources",
			Parameters: llm.ObjectSchema("Market news parameters",
				map[string]*llm.JSONSchema{
					"limit": llm.IntProp("Maximum number of articles (default: 20)"),
				},
			),
			Handler: a.handleGetMarketNews,
		},
		{
			Name:        "get_sector_news",
			Description: "Fetch news for a specific market sector (IT, Banking, Pharma, etc.)",
			Parameters: llm.ObjectSchema("Sector news parameters",
				map[string]*llm.JSONSchema{
					"sector": llm.StringProp("Sector name (IT, Banking, Pharma, Auto, Oil & Gas, Metal, FMCG, etc.)"),
					"limit":  llm.IntProp("Maximum number of articles (default: 15)"),
				},
				"sector",
			),
			Handler: a.handleGetSectorNews,
		},
		{
			Name:        "analyze_sentiment",
			Description: "Run full sentiment analysis: score all articles, aggregate sentiment, detect catalysts, and produce an overall sentiment rating",
			Parameters: llm.ObjectSchema("Sentiment analysis parameters",
				map[string]*llm.JSONSchema{
					"ticker": llm.StringProp("NSE ticker symbol"),
					"limit":  llm.IntProp("Number of articles to analyze (default: 20)"),
				},
				"ticker",
			),
			Handler: a.handleAnalyzeSentiment,
		},
		{
			Name:        "score_headline",
			Description: "Score the sentiment of a single news headline (-1.0 to +1.0)",
			Parameters: llm.ObjectSchema("Headline scoring parameters",
				map[string]*llm.JSONSchema{
					"headline": llm.StringProp("The news headline text to score"),
					"source":   llm.StringProp("The news source (optional, for credibility weighting)"),
				},
				"headline",
			),
			Handler: a.handleScoreHeadline,
		},
	}
}

// ── Tool Handlers ──

func (a *SentimentAgent) handleGetStockNews(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker string `json:"ticker"`
		Limit  int    `json:"limit"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	articles, err := a.news.GetStockNews(ctx, params.Ticker, params.Limit)
	if err != nil {
		return fmt.Sprintf("Could not fetch news for %s: %v", params.Ticker, err), nil
	}

	result := map[string]any{
		"ticker":   params.Ticker,
		"articles": len(articles),
		"items":    formatArticles(articles),
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

func (a *SentimentAgent) handleGetMarketNews(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Limit int `json:"limit"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	articles, err := a.news.GetMarketNews(ctx, params.Limit)
	if err != nil {
		return fmt.Sprintf("Could not fetch market news: %v", err), nil
	}

	result := map[string]any{
		"articles": len(articles),
		"items":    formatArticles(articles),
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

func (a *SentimentAgent) handleGetSectorNews(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Sector string `json:"sector"`
		Limit  int    `json:"limit"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	if params.Limit <= 0 {
		params.Limit = 15
	}

	articles, err := a.news.GetSectorNews(ctx, params.Sector, params.Limit)
	if err != nil {
		return fmt.Sprintf("Could not fetch news for %s sector: %v", params.Sector, err), nil
	}

	result := map[string]any{
		"sector":   params.Sector,
		"articles": len(articles),
		"items":    formatArticles(articles),
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

func (a *SentimentAgent) handleAnalyzeSentiment(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker string `json:"ticker"`
		Limit  int    `json:"limit"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	articles, err := a.news.GetStockNews(ctx, params.Ticker, params.Limit)
	if err != nil {
		return fmt.Sprintf("Could not fetch news for %s: %v", params.Ticker, err), nil
	}
	if len(articles) == 0 {
		return fmt.Sprintf("No news articles found for %s", params.Ticker), nil
	}

	analysisResult := sentiment.FullSentimentAnalysis(params.Ticker, articles)
	data, _ := json.MarshalIndent(analysisResult, "", "  ")
	return string(data), nil
}

func (a *SentimentAgent) handleScoreHeadline(_ context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Headline string `json:"headline"`
		Source   string `json:"source"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	score, confidence := sentiment.ScoreHeadline(params.Headline)
	result := map[string]any{
		"headline":   params.Headline,
		"score":      score,
		"confidence": confidence,
	}
	if params.Source != "" {
		result["source"] = params.Source
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

// formatArticles converts articles to a display-friendly format (limited fields).
func formatArticles(articles []models.NewsArticle) []map[string]string {
	items := make([]map[string]string, 0, len(articles))
	for _, a := range articles {
		item := map[string]string{
			"title":     a.Title,
			"source":    a.Source,
			"published": a.PublishedAt.Format("2006-01-02 15:04"),
		}
		if a.Summary != "" {
			item["summary"] = a.Summary
		}
		if a.URL != "" {
			item["url"] = a.URL
		}
		items = append(items, item)
	}
	return items
}

// Analyze runs a full sentiment analysis with chain-of-thought reasoning.
func (a *SentimentAgent) Analyze(ctx context.Context, ticker string) (*AgentResult, error) {
	task := fmt.Sprintf(
		"Analyze the current market sentiment for %s.\n\n%s\n\n"+
			"Assess overall sentiment, key catalysts (positive and negative), "+
			"news momentum, and how sentiment might affect the stock price in the near term.",
		ticker, prompts.FormatTickerPrompt(ticker),
	)
	return a.Process(ctx, task)
}

// AnalyzeWithTimestamp runs the analysis and attaches a typed result.
func (a *SentimentAgent) AnalyzeWithTimestamp(ctx context.Context, ticker string) (*AgentResult, error) {
	result, err := a.Analyze(ctx, ticker)
	if err != nil {
		return result, err
	}

	result.Analysis = ParseAnalysisResult(result.Content, models.AnalysisResult{
		Ticker:    ticker,
		Type:      models.AnalysisSentiment,
		AgentName: a.Name(),
		Timestamp: time.Now(),
	})

	return result, nil
}
