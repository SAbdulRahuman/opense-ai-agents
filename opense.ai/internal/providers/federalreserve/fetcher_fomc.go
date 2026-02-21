package federalreserve

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// FomcDocuments — List of FOMC meeting dates and document links.
// We build structured data from the known calendar rather than scraping HTML.
// Uses the FOMC press release URL pattern.
// ---------------------------------------------------------------------------

type fomcDocumentsFetcher struct {
	provider.BaseFetcher
}

func newFomcDocumentsFetcher() *fomcDocumentsFetcher {
	return &fomcDocumentsFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelFomcDocuments,
			"Federal Reserve FOMC meeting documents and press releases",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
	}
}

func (f *fomcDocumentsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelFomcDocuments, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	// Scrape the FOMC calendars page for meeting dates.
	url := baseFedBoard + "/monetarypolicy/fomccalendars.htm"
	body, err := fetchFedRaw(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fomc documents: %w", err)
	}

	docs := parseFomcCalendar(string(body), params)

	result := newResult(docs)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// fomcDateRe matches FOMC meeting date patterns in the calendar HTML.
// Pattern: "January 28-29" or "March 18-19*" in the context of a year.
var fomcDateRe = regexp.MustCompile(`(?i)(\d{4}).*?(?:January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d{1,2})(?:-(\d{1,2}))?\*?`)

// fomcMeetingRe is a simpler regex to find year/month/day patterns in FOMC calendar links.
var fomcMeetingRe = regexp.MustCompile(`/monetarypolicy/fomcpresconf(\d{8})\.htm`)

func parseFomcCalendar(html string, params provider.QueryParams) []models.FOMCDocument {
	startDate := params[provider.ParamStartDate]
	endDate := params[provider.ParamEndDate]

	var docs []models.FOMCDocument
	seen := make(map[string]bool)

	// Try to extract meeting dates from press conference links.
	matches := fomcMeetingRe.FindAllStringSubmatch(html, -1)
	for _, m := range matches {
		dateStr := m[1] // YYYYMMDD
		dt, err := time.Parse("20060102", dateStr)
		if err != nil {
			continue
		}
		ymd := dt.Format("2006-01-02")
		if seen[ymd] {
			continue
		}
		seen[ymd] = true

		if startDate != "" && ymd < startDate {
			continue
		}
		if endDate != "" && ymd > endDate {
			continue
		}

		docs = append(docs, models.FOMCDocument{
			Date:  dt,
			Title: fmt.Sprintf("FOMC Meeting - %s", dt.Format("January 2, 2006")),
			Type:  "meeting",
			URL:   baseFedBoard + "/monetarypolicy/fomcpresconf" + dateStr + ".htm",
		})
	}

	// Also look for statement links.
	stmtRe := regexp.MustCompile(`/newsevents/pressreleases/monetary(\d{8})a\.htm`)
	stmtMatches := stmtRe.FindAllStringSubmatch(html, -1)
	for _, m := range stmtMatches {
		dateStr := m[1]
		dt, err := time.Parse("20060102", dateStr)
		if err != nil {
			continue
		}
		ymd := dt.Format("2006-01-02")
		key := ymd + "-statement"
		if seen[key] {
			continue
		}
		seen[key] = true

		if startDate != "" && ymd < startDate {
			continue
		}
		if endDate != "" && ymd > endDate {
			continue
		}

		docs = append(docs, models.FOMCDocument{
			Date:  dt,
			Title: fmt.Sprintf("FOMC Statement - %s", dt.Format("January 2, 2006")),
			Type:  "statement",
			URL:   baseFedBoard + "/newsevents/pressreleases/monetary" + dateStr + "a.htm",
		})
	}

	// Also look for minutes links.
	minutesRe := regexp.MustCompile(`/monetarypolicy/fomcminutes(\d{8})\.htm`)
	minutesMatches := minutesRe.FindAllStringSubmatch(html, -1)
	for _, m := range minutesMatches {
		dateStr := m[1]
		dt, err := time.Parse("20060102", dateStr)
		if err != nil {
			continue
		}
		ymd := dt.Format("2006-01-02")
		key := ymd + "-minutes"
		if seen[key] {
			continue
		}
		seen[key] = true

		if startDate != "" && ymd < startDate {
			continue
		}
		if endDate != "" && ymd > endDate {
			continue
		}

		docs = append(docs, models.FOMCDocument{
			Date:  dt,
			Title: fmt.Sprintf("FOMC Minutes - %s", dt.Format("January 2, 2006")),
			Type:  "minutes",
			URL:   baseFedBoard + "/monetarypolicy/fomcminutes" + dateStr + ".htm",
		})
	}

	return docs
}

// compileFOMCStatementURL constructs the expected FOMC statement URL.
func compileFOMCStatementURL(date string) string {
	d := strings.ReplaceAll(date, "-", "")
	return baseFedBoard + "/newsevents/pressreleases/monetary" + d + "a.htm"
}
