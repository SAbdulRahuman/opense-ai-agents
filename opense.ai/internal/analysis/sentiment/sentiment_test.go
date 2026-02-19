package sentiment

import (
	"testing"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

func TestScoreHeadlineBullish(t *testing.T) {
	score, conf := ScoreHeadline("Reliance shares rally 5% on strong growth and positive results")
	if score <= 0 {
		t.Errorf("expected positive score for bullish headline, got %.4f", score)
	}
	if conf <= 0 {
		t.Errorf("expected positive confidence, got %.4f", conf)
	}
}

func TestScoreHeadlineBearish(t *testing.T) {
	score, conf := ScoreHeadline("Market crash: stocks plunge amid fraud investigation concerns")
	if score >= 0 {
		t.Errorf("expected negative score for bearish headline, got %.4f", score)
	}
	if conf <= 0 {
		t.Errorf("expected positive confidence, got %.4f", conf)
	}
}

func TestScoreHeadlineNeutral(t *testing.T) {
	score, conf := ScoreHeadline("Company announces new office location in Bengaluru")
	if score != 0 {
		t.Errorf("expected zero score for neutral headline, got %.4f", score)
	}
	if conf > 0.2 {
		t.Errorf("expected low confidence for neutral, got %.4f", conf)
	}
}

func TestScoreArticle(t *testing.T) {
	article := models.NewsArticle{
		Title:       "NIFTY surges to record high on bullish momentum",
		Source:      "Moneycontrol",
		URL:         "https://example.com/article1",
		PublishedAt: time.Now(),
	}
	ss := ScoreArticle(article)
	if ss.Score <= 0 {
		t.Errorf("expected positive score, got %.4f", ss.Score)
	}
	if ss.Source != "Moneycontrol" {
		t.Errorf("expected source Moneycontrol, got %s", ss.Source)
	}
}

func TestAggregateSentiment(t *testing.T) {
	now := time.Now()
	scores := []models.SentimentScore{
		{Source: "MC", Score: 0.5, Confidence: 0.7, PublishedAt: now},
		{Source: "ET", Score: 0.3, Confidence: 0.6, PublishedAt: now.Add(-12 * time.Hour)},
		{Source: "LM", Score: -0.1, Confidence: 0.5, PublishedAt: now.Add(-36 * time.Hour)},
	}

	agg := AggregateSentiment("RELIANCE", scores)
	if agg.Ticker != "RELIANCE" {
		t.Errorf("expected RELIANCE, got %s", agg.Ticker)
	}
	if agg.Score <= 0 {
		t.Errorf("expected positive aggregate score, got %.4f", agg.Score)
	}
	if agg.ArticleCount != 3 {
		t.Errorf("expected 3 articles, got %d", agg.ArticleCount)
	}
	if agg.Label == "" {
		t.Error("expected non-empty label")
	}
}

func TestAggregateSentimentEmpty(t *testing.T) {
	agg := AggregateSentiment("RELIANCE", nil)
	if agg.Label != "Neutral" {
		t.Errorf("expected Neutral, got %s", agg.Label)
	}
}

func TestFullSentimentAnalysis(t *testing.T) {
	now := time.Now()
	articles := []models.NewsArticle{
		{Title: "Stock surges on strong earnings beat", Source: "MC", PublishedAt: now},
		{Title: "Positive growth outlook for Q4", Source: "ET", PublishedAt: now.Add(-6 * time.Hour)},
		{Title: "Investors bullish on expansion plans", Source: "LM", PublishedAt: now.Add(-12 * time.Hour)},
	}

	result := FullSentimentAnalysis("RELIANCE", articles)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Type != models.AnalysisSentiment {
		t.Errorf("expected sentiment type, got %s", result.Type)
	}
	if len(result.Signals) == 0 {
		t.Error("expected at least one signal")
	}
	if result.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestFullSentimentAnalysisEmpty(t *testing.T) {
	result := FullSentimentAnalysis("RELIANCE", nil)
	if result != nil {
		t.Error("expected nil for empty articles")
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{42, "42"},
		{-5, "-5"},
		{100, "100"},
	}
	for _, tt := range tests {
		got := itoa(tt.n)
		if got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
