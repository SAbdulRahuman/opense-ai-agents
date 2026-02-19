package utils

import (
	"testing"
	"time"
)

func TestNowIST(t *testing.T) {
	now := NowIST()
	if now.Location().String() != "Asia/Kolkata" && now.Location().String() != "IST" {
		t.Errorf("NowIST() location = %s, want Asia/Kolkata or IST", now.Location().String())
	}
}

func TestMarketOpenClose(t *testing.T) {
	date := time.Date(2026, 2, 19, 12, 0, 0, 0, IST)

	open := MarketOpenTime(date)
	if open.Hour() != 9 || open.Minute() != 15 {
		t.Errorf("MarketOpenTime = %v, want 09:15", open)
	}

	close := MarketCloseTime(date)
	if close.Hour() != 15 || close.Minute() != 30 {
		t.Errorf("MarketCloseTime = %v, want 15:30", close)
	}
}

func TestIsMarketOpenAt(t *testing.T) {
	// Wednesday at 10:00 AM IST — should be open
	weekday := time.Date(2026, 2, 18, 10, 0, 0, 0, IST)
	if !IsMarketOpenAt(weekday) {
		t.Error("Expected market to be open on Wednesday 10:00 AM")
	}

	// Saturday — should be closed
	saturday := time.Date(2026, 2, 21, 10, 0, 0, 0, IST)
	if IsMarketOpenAt(saturday) {
		t.Error("Expected market to be closed on Saturday")
	}

	// Wednesday at 8:00 AM — before market open
	earlyMorning := time.Date(2026, 2, 18, 8, 0, 0, 0, IST)
	if IsMarketOpenAt(earlyMorning) {
		t.Error("Expected market to be closed at 8:00 AM")
	}

	// Wednesday at 4:00 PM — after market close
	afterHours := time.Date(2026, 2, 18, 16, 0, 0, 0, IST)
	if IsMarketOpenAt(afterHours) {
		t.Error("Expected market to be closed at 4:00 PM")
	}
}

func TestIsTradingHoliday(t *testing.T) {
	// Republic Day 2026
	republicDay := time.Date(2026, 1, 26, 10, 0, 0, 0, IST)
	if !IsTradingHoliday(republicDay) {
		t.Error("Expected Republic Day to be a trading holiday")
	}

	// Regular trading day
	normalDay := time.Date(2026, 2, 18, 10, 0, 0, 0, IST)
	if IsTradingHoliday(normalDay) {
		t.Error("Expected Feb 18 to NOT be a trading holiday")
	}
}

func TestIsTradingDay(t *testing.T) {
	// Wednesday — trading day
	if !IsTradingDay(time.Date(2026, 2, 18, 0, 0, 0, 0, IST)) {
		t.Error("Expected Wednesday to be a trading day")
	}

	// Saturday — not a trading day
	if IsTradingDay(time.Date(2026, 2, 21, 0, 0, 0, 0, IST)) {
		t.Error("Expected Saturday to not be a trading day")
	}

	// Trading holiday — not a trading day
	if IsTradingDay(time.Date(2026, 1, 26, 0, 0, 0, 0, IST)) {
		t.Error("Expected Republic Day to not be a trading day")
	}
}

func TestNextPrevTradingDay(t *testing.T) {
	// Friday → next trading day should be Monday (assuming no holiday)
	friday := time.Date(2026, 2, 20, 0, 0, 0, 0, IST)
	next := NextTradingDay(friday)
	if next.Weekday() != time.Monday || next.Day() != 23 {
		t.Errorf("NextTradingDay(Friday Feb 20) = %v, want Monday Feb 23", next)
	}

	// Monday → prev trading day should be Friday
	monday := time.Date(2026, 2, 23, 0, 0, 0, 0, IST)
	prev := PrevTradingDay(monday)
	if prev.Weekday() != time.Friday || prev.Day() != 20 {
		t.Errorf("PrevTradingDay(Monday Feb 23) = %v, want Friday Feb 20", prev)
	}
}

func TestParseDateIST(t *testing.T) {
	d, err := ParseDateIST("2026-02-19")
	if err != nil {
		t.Fatalf("ParseDateIST failed: %v", err)
	}
	if d.Year() != 2026 || d.Month() != 2 || d.Day() != 19 {
		t.Errorf("ParseDateIST = %v, want 2026-02-19", d)
	}
}

func TestFormatDateIST(t *testing.T) {
	d := time.Date(2026, 2, 19, 10, 30, 0, 0, IST)
	result := FormatDateIST(d)
	if result != "2026-02-19" {
		t.Errorf("FormatDateIST = %s, want 2026-02-19", result)
	}
}

func TestMarketStatus(t *testing.T) {
	// Just verify it doesn't panic and returns a non-empty string
	status := MarketStatus()
	if status == "" {
		t.Error("MarketStatus() returned empty string")
	}
}
