package sentiment

import (
	"math"
	"strings"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// ------------------------------------------------------------------
// Keyword-based sentiment scorer (offline, no LLM needed).
// When an LLM backend is configured the agent layer will use it
// instead; this package provides a deterministic fallback.
// ------------------------------------------------------------------

// bullish / bearish keyword dictionaries (lowercase, stemmed).
var bullishWords = map[string]float64{
	"bullish": 0.7, "rally": 0.6, "surge": 0.7, "upbeat": 0.5,
	"positive": 0.4, "growth": 0.4, "upgrade": 0.6, "outperform": 0.6,
	"buy": 0.5, "strong": 0.4, "recovery": 0.5, "breakout": 0.6,
	"record high": 0.7, "all-time high": 0.7, "beat": 0.5,
	"exceeds": 0.5, "beats estimate": 0.6, "expansion": 0.4,
	"profit": 0.3, "dividend": 0.4, "accumulate": 0.5,
}

var bearishWords = map[string]float64{
	"bearish": 0.7, "crash": 0.8, "plunge": 0.7, "slump": 0.6,
	"negative": 0.4, "downgrade": 0.6, "underperform": 0.6,
	"sell": 0.5, "weak": 0.4, "decline": 0.5, "loss": 0.4,
	"selloff": 0.7, "fall": 0.4, "correction": 0.5,
	"default": 0.7, "fraud": 0.8, "scam": 0.8, "investigation": 0.5,
	"cut": 0.3, "miss": 0.5, "warning": 0.5, "concern": 0.3,
}

// ScoreHeadline returns a sentiment score for a single headline.
// Score ranges from -1.0 (very bearish) to +1.0 (very bullish).
func ScoreHeadline(headline string) (score float64, confidence float64) {
	lower := strings.ToLower(headline)

	bullScore := 0.0
	bearScore := 0.0
	matches := 0

	for word, weight := range bullishWords {
		if strings.Contains(lower, word) {
			bullScore += weight
			matches++
		}
	}

	for word, weight := range bearishWords {
		if strings.Contains(lower, word) {
			bearScore += weight
			matches++
		}
	}

	if matches == 0 {
		return 0, 0.1 // no signal
	}

	total := bullScore + bearScore
	if total == 0 {
		return 0, 0.1
	}

	// Net score normalized to -1..+1.
	score = (bullScore - bearScore) / total

	// Confidence based on number of keyword matches.
	confidence = math.Min(float64(matches)*0.15+0.2, 0.85)

	return score, confidence
}

// ScoreArticle scores a news article and returns a SentimentScore.
func ScoreArticle(article models.NewsArticle) models.SentimentScore {
	text := article.Title
	if article.Summary != "" {
		text += " " + article.Summary
	}

	score, confidence := ScoreHeadline(text)

	return models.SentimentScore{
		Source:      article.Source,
		Headline:    article.Title,
		Score:       score,
		Confidence:  models.Confidence(confidence),
		URL:         article.URL,
		PublishedAt: article.PublishedAt,
	}
}

// AggregateSentiment computes a time-weighted aggregate sentiment from multiple scores.
func AggregateSentiment(ticker string, scores []models.SentimentScore) models.AggregatedSentiment {
	if len(scores) == 0 {
		return models.AggregatedSentiment{
			Ticker:    ticker,
			Label:     "Neutral",
			Timestamp: time.Now(),
		}
	}

	now := time.Now()
	weightedSum := 0.0
	totalWeight := 0.0
	confSum := 0.0

	for _, s := range scores {
		// Time decay: halve weight every 24 hours.
		age := now.Sub(s.PublishedAt).Hours()
		if age < 0 {
			age = 0
		}
		timeWeight := math.Exp(-0.693 * age / 24) // ln(2)/1 * t/24
		w := timeWeight * float64(s.Confidence)

		weightedSum += s.Score * w
		totalWeight += w
		confSum += float64(s.Confidence)
	}

	avgScore := 0.0
	if totalWeight > 0 {
		avgScore = weightedSum / totalWeight
	}

	avgConf := confSum / float64(len(scores))

	label := "Neutral"
	switch {
	case avgScore > 0.3:
		label = "Bullish"
	case avgScore > 0.1:
		label = "Slightly Bullish"
	case avgScore < -0.3:
		label = "Bearish"
	case avgScore < -0.1:
		label = "Slightly Bearish"
	}

	return models.AggregatedSentiment{
		Ticker:       ticker,
		Score:        avgScore,
		Confidence:   models.Confidence(avgConf),
		Label:        label,
		Sources:      scores,
		ArticleCount: len(scores),
		Timestamp:    now,
	}
}

// FullSentimentAnalysis runs end-to-end sentiment analysis on a set of news articles.
func FullSentimentAnalysis(ticker string, articles []models.NewsArticle) *models.AnalysisResult {
	if len(articles) == 0 {
		return nil
	}

	var scores []models.SentimentScore
	for _, a := range articles {
		scores = append(scores, ScoreArticle(a))
	}

	agg := AggregateSentiment(ticker, scores)

	// Convert to signal.
	var sigType models.SignalType
	switch {
	case agg.Score > 0.1:
		sigType = models.SignalBuy
	case agg.Score < -0.1:
		sigType = models.SignalSell
	default:
		sigType = models.SignalNeutral
	}

	sig := models.Signal{
		Source:     "Sentiment",
		Type:       sigType,
		Confidence: agg.Confidence,
		Reason:     agg.Label + " sentiment across " + string(rune('0'+len(articles))) + " articles",
	}
	if len(articles) >= 10 {
		sig.Reason = agg.Label + " sentiment across " + itoa(len(articles)) + " articles"
	}

	var rec models.Recommendation
	switch {
	case agg.Score > 0.3:
		rec = models.ModerateBuy
	case agg.Score > 0.1:
		rec = models.ModerateBuy
	case agg.Score < -0.3:
		rec = models.ModerateSell
	case agg.Score < -0.1:
		rec = models.ModerateSell
	default:
		rec = models.Hold
	}

	details := map[string]any{
		"aggregated_sentiment": agg,
		"article_count":        len(articles),
	}

	return &models.AnalysisResult{
		Ticker:         ticker,
		Type:           models.AnalysisSentiment,
		AgentName:      "sentiment-analysis",
		Signals:        []models.Signal{sig},
		Recommendation: rec,
		Confidence:     agg.Confidence,
		Summary:        agg.Label + " overall sentiment for " + ticker,
		Details:        details,
		Timestamp:      time.Now(),
	}
}

// itoa is a simple int-to-string helper to avoid importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
