package datasource

import (
	"testing"

	"github.com/seenimoa/openseai/pkg/models"
)

func TestYfInterval(t *testing.T) {
	tests := []struct {
		tf   models.Timeframe
		want string
	}{
		{models.Timeframe1Min, "1m"},
		{models.Timeframe5Min, "5m"},
		{models.Timeframe15Min, "15m"},
		{models.Timeframe1Hour, "1h"},
		{models.Timeframe1Day, "1d"},
		{models.Timeframe1Week, "1wk"},
		{models.Timeframe1Mon, "1mo"},
		{models.Timeframe("unknown"), "1d"},
	}
	for _, tt := range tests {
		got := yfInterval(tt.tf)
		if got != tt.want {
			t.Errorf("yfInterval(%q) = %q, want %q", tt.tf, got, tt.want)
		}
	}
}

func TestParseYFCandlesEmpty(t *testing.T) {
	result := yfChartResult{}
	candles := parseYFCandles(result)
	if candles != nil {
		t.Fatalf("expected nil candles for empty result, got %d", len(candles))
	}
}

func TestParseYFCandles(t *testing.T) {
	open := 100.0
	high := 105.0
	low := 98.0
	close_ := 103.0
	vol := int64(1000)
	adj := 102.5

	result := yfChartResult{
		Timestamp: []int64{1700000000, 1700086400},
		Indicators: yfIndicators{
			Quote: []yfOHLCV{
				{
					Open:   []*float64{&open, &open},
					High:   []*float64{&high, &high},
					Low:    []*float64{&low, &low},
					Close:  []*float64{&close_, &close_},
					Volume: []*int64{&vol, &vol},
				},
			},
			AdjClose: []yfAdjClose{
				{AdjClose: []*float64{&adj, &adj}},
			},
		},
	}

	candles := parseYFCandles(result)
	if len(candles) != 2 {
		t.Fatalf("expected 2 candles, got %d", len(candles))
	}

	c := candles[0]
	if c.Open != 100.0 || c.High != 105.0 || c.Low != 98.0 || c.Close != 103.0 {
		t.Errorf("OHLC mismatch: %+v", c)
	}
	if c.Volume != 1000 {
		t.Errorf("volume = %d, want 1000", c.Volume)
	}
	if c.AdjClose != 102.5 {
		t.Errorf("adjclose = %f, want 102.5", c.AdjClose)
	}
}

func TestParseYFCandlesNilPointers(t *testing.T) {
	// Some entries may be nil (market holidays, etc.)
	open := 100.0
	result := yfChartResult{
		Timestamp: []int64{1700000000},
		Indicators: yfIndicators{
			Quote: []yfOHLCV{
				{
					Open:   []*float64{&open},
					High:   []*float64{nil},
					Low:    []*float64{nil},
					Close:  []*float64{nil},
					Volume: []*int64{nil},
				},
			},
		},
	}

	candles := parseYFCandles(result)
	if len(candles) != 1 {
		t.Fatalf("expected 1 candle, got %d", len(candles))
	}
	if candles[0].Open != 100.0 {
		t.Errorf("open = %f, want 100.0", candles[0].Open)
	}
	if candles[0].High != 0 || candles[0].Low != 0 || candles[0].Close != 0 {
		t.Error("expected zero for nil pointer fields")
	}
}

func TestYFinanceName(t *testing.T) {
	yf := NewYFinance()
	if yf.Name() != "Yahoo Finance" {
		t.Errorf("Name() = %q, want %q", yf.Name(), "Yahoo Finance")
	}
}
