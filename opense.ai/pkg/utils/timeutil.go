package utils

import (
	"time"
)

// IST is the Indian Standard Time location (UTC+5:30).
var IST *time.Location

func init() {
	var err error
	IST, err = time.LoadLocation("Asia/Kolkata")
	if err != nil {
		// Fallback: create fixed zone if tz database is not available
		IST = time.FixedZone("IST", 5*60*60+30*60)
	}
}

// NowIST returns the current time in IST.
func NowIST() time.Time {
	return time.Now().In(IST)
}

// ToIST converts a time.Time to IST.
func ToIST(t time.Time) time.Time {
	return t.In(IST)
}

// MarketOpenTime returns the NSE market opening time (9:15 AM IST) for a given date.
func MarketOpenTime(date time.Time) time.Time {
	d := date.In(IST)
	return time.Date(d.Year(), d.Month(), d.Day(), 9, 15, 0, 0, IST)
}

// MarketCloseTime returns the NSE market closing time (3:30 PM IST) for a given date.
func MarketCloseTime(date time.Time) time.Time {
	d := date.In(IST)
	return time.Date(d.Year(), d.Month(), d.Day(), 15, 30, 0, 0, IST)
}

// PreOpenStart returns the pre-open session start time (9:00 AM IST).
func PreOpenStart(date time.Time) time.Time {
	d := date.In(IST)
	return time.Date(d.Year(), d.Month(), d.Day(), 9, 0, 0, 0, IST)
}

// IsMarketOpen checks if the NSE market is currently open.
func IsMarketOpen() bool {
	return IsMarketOpenAt(NowIST())
}

// IsMarketOpenAt checks if the NSE market would be open at the given time.
func IsMarketOpenAt(t time.Time) bool {
	t = t.In(IST)

	// Check if it's a weekend
	if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		return false
	}

	// Check if it's a trading holiday
	if IsTradingHoliday(t) {
		return false
	}

	// Check if within market hours (9:15 AM - 3:30 PM IST)
	open := MarketOpenTime(t)
	close := MarketCloseTime(t)

	return !t.Before(open) && !t.After(close)
}

// NextTradingDay returns the next trading day from the given date.
// If the given date is a trading day, it returns the next one.
func NextTradingDay(from time.Time) time.Time {
	next := from.In(IST).AddDate(0, 0, 1)
	for !IsTradingDay(next) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// PrevTradingDay returns the previous trading day from the given date.
func PrevTradingDay(from time.Time) time.Time {
	prev := from.In(IST).AddDate(0, 0, -1)
	for !IsTradingDay(prev) {
		prev = prev.AddDate(0, 0, -1)
	}
	return prev
}

// IsTradingDay checks if the given date is a trading day (not weekend, not holiday).
func IsTradingDay(t time.Time) bool {
	t = t.In(IST)
	if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		return false
	}
	return !IsTradingHoliday(t)
}

// TradingDaysBetween returns the number of trading days between two dates (exclusive of end).
func TradingDaysBetween(start, end time.Time) int {
	start = start.In(IST)
	end = end.In(IST)
	count := 0
	current := start
	for current.Before(end) {
		if IsTradingDay(current) {
			count++
		}
		current = current.AddDate(0, 0, 1)
	}
	return count
}

// IsTradingHoliday checks if the given date is an NSE trading holiday.
// This list should be updated annually.
func IsTradingHoliday(t time.Time) bool {
	t = t.In(IST)
	dateStr := t.Format("2006-01-02")

	_, isHoliday := nseHolidays2026[dateStr]
	return isHoliday
}

// NSE Trading Holidays for 2026 (update annually).
// Source: NSE India circular.
var nseHolidays2026 = map[string]string{
	"2026-01-26": "Republic Day",
	"2026-02-17": "Mahashivratri",
	"2026-03-10": "Holi",
	"2026-03-30": "Id-ul-Fitr (Ramadan)",
	"2026-04-02": "Ram Navami",
	"2026-04-03": "Good Friday",
	"2026-04-14": "Dr. Ambedkar Jayanti",
	"2026-05-01": "Maharashtra Day",
	"2026-05-25": "Buddha Purnima",
	"2026-06-05": "Id-ul-Zuha (Bakri Id)",
	"2026-07-06": "Muharram",
	"2026-08-15": "Independence Day",
	"2026-08-18": "Parsi New Year",
	"2026-09-04": "Milad-un-Nabi",
	"2026-10-02": "Mahatma Gandhi Jayanti",
	"2026-10-20": "Dussehra",
	"2026-11-09": "Diwali (Laxmi Pujan)",
	"2026-11-10": "Diwali (Balipratipada)",
	"2026-11-30": "Guru Nanak Jayanti",
	"2026-12-25": "Christmas",
}

// GetTradingHolidays returns all trading holidays for the current year.
func GetTradingHolidays() map[string]string {
	return nseHolidays2026
}

// ParseDateIST parses a date string in "2006-01-02" format and returns it in IST.
func ParseDateIST(dateStr string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", dateStr, IST)
}

// FormatDateIST formats a time.Time to "2006-01-02" in IST.
func FormatDateIST(t time.Time) string {
	return t.In(IST).Format("2006-01-02")
}

// FormatDateTimeIST formats a time.Time to "2006-01-02 15:04:05 IST".
func FormatDateTimeIST(t time.Time) string {
	return t.In(IST).Format("2006-01-02 15:04:05 IST")
}

// MarketStatus returns the current market status string.
func MarketStatus() string {
	now := NowIST()

	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return "CLOSED (Weekend)"
	}

	if IsTradingHoliday(now) {
		holiday := nseHolidays2026[now.Format("2006-01-02")]
		return "CLOSED (" + holiday + ")"
	}

	open := MarketOpenTime(now)
	close := MarketCloseTime(now)
	preOpen := PreOpenStart(now)

	switch {
	case now.Before(preOpen):
		return "PRE-MARKET"
	case now.Before(open):
		return "PRE-OPEN SESSION"
	case !now.After(close):
		return "OPEN"
	default:
		return "CLOSED"
	}
}
