// Package federalreserve implements a Federal Reserve data provider.
// Data is sourced from the NY Fed Markets API (JSON), Fed Board Data Downloads (CSV),
// and Fed Board static files. No API key required.
package federalreserve

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/seenimoa/openseai/internal/infra"
	"github.com/seenimoa/openseai/internal/provider"
)

const (
	providerName = "federal_reserve"

	// NY Fed Markets API.
	baseNYFed = "https://markets.newyorkfed.org/api"

	// Fed Board data downloads.
	baseFedBoard = "https://www.federalreserve.gov"
)

// Provider is the Federal Reserve data provider.
type Provider struct {
	provider.BaseProvider
}

// New creates a new Federal Reserve provider and registers all fetchers.
func New() *Provider {
	p := &Provider{
		BaseProvider: provider.NewBaseProvider(
			providerName,
			"Federal Reserve — NY Fed Markets API & Fed Board data (free, no API key)",
			"https://www.federalreserve.gov",
			nil, // no credentials required
		),
	}

	// NY Fed JSON-based rates.
	p.RegisterFetcher(newFederalFundsRateFetcher())
	p.RegisterFetcher(newSOFRFetcher())
	p.RegisterFetcher(newOBFRFetcher())

	// NY Fed SOMA (central bank holdings).
	p.RegisterFetcher(newCentralBankHoldingsFetcher())

	// NY Fed primary dealer data.
	p.RegisterFetcher(newPrimaryDealerFailsFetcher())
	p.RegisterFetcher(newPrimaryDealerPositioningFetcher())

	// Fed Board CSV-based.
	p.RegisterFetcher(newTreasuryRatesFetcher())
	p.RegisterFetcher(newYieldCurveFetcher())
	p.RegisterFetcher(newMoneyMeasuresFetcher())

	// Fed Board static files.
	p.RegisterFetcher(newSvenssonYieldCurveFetcher())

	// FOMC documents.
	p.RegisterFetcher(newFomcDocumentsFetcher())

	// Inflation expectations (Philly Fed).
	p.RegisterFetcher(newInflationExpectationsFetcher())

	// TFP (SF Fed).
	p.RegisterFetcher(newTFPFetcher())

	return p
}

// Ping verifies connectivity to the NY Fed Markets API.
func (p *Provider) Ping(ctx context.Context) error {
	url := baseNYFed + "/rates/unsecured/effr/last/1.json"
	body, _, err := infra.DoGet(ctx, url, fedHeaders)
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// ---------------------------------------------------------------------------
// HTTP helpers.
// ---------------------------------------------------------------------------

var fedHeaders = map[string]string{
	"Accept":          "application/json, text/csv, */*",
	"Accept-Language": "en-US,en;q=0.9",
}

// fetchFedJSON fetches a NY Fed JSON endpoint and decodes into dst.
func fetchFedJSON(ctx context.Context, u string, dst any) error {
	body, _, err := infra.DoGet(ctx, u, fedHeaders)
	if err != nil {
		return err
	}
	defer body.Close()
	return json.NewDecoder(body).Decode(dst)
}

// fetchFedCSV fetches a Fed Board CSV endpoint, skips skipRows header rows,
// and returns parsed CSV records.
func fetchFedCSV(ctx context.Context, u string, skipRows int) ([][]string, error) {
	body, _, err := infra.DoGet(ctx, u, map[string]string{
		"Accept":          "text/csv, */*",
		"Accept-Language": "en-US,en;q=0.9",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	raw, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(raw), "\n")
	if len(lines) <= skipRows {
		return nil, fmt.Errorf("fed csv: too few lines (%d, expected >%d)", len(lines), skipRows)
	}

	reader := csv.NewReader(strings.NewReader(strings.Join(lines[skipRows:], "\n")))
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // variable fields

	return reader.ReadAll()
}

// fetchFedRaw fetches a URL and returns raw bytes.
func fetchFedRaw(ctx context.Context, u string) ([]byte, error) {
	body, _, err := infra.DoGet(ctx, u, fedHeaders)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return io.ReadAll(body)
}

// ---------------------------------------------------------------------------
// Utility functions.
// ---------------------------------------------------------------------------

// newResult creates a FetchResult with the current timestamp.
func newResult(data any) *provider.FetchResult {
	return &provider.FetchResult{
		Data:      data,
		FetchedAt: time.Now(),
	}
}

// defaultDate returns a default start date if none provided.
func defaultDate(params provider.QueryParams, key string, fallback string) string {
	if v := params[key]; v != "" {
		return v
	}
	return fallback
}

// parseDate parses an ISO date string.
func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

// parseFloat64 safely parses a string to float64.
func parseFloat64(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "ND" || s == "NA" || s == "''" {
		return 0
	}
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// nyfedDateParam formats a time for NY Fed API query params (MM/DD/YYYY).
func nyfedDateParam(isoDate string) string {
	t := parseDate(isoDate)
	if t.IsZero() {
		return isoDate
	}
	return t.Format("01/02/2006")
}

// buildNYFedRatesURL builds a NY Fed rates search URL.
func buildNYFedRatesURL(rateType, startDate, endDate string) string {
	return fmt.Sprintf("%s/rates/%s/search.json?startDate=%s&endDate=%s",
		baseNYFed, rateType, startDate, endDate)
}

// buildH15URL builds a Fed Board H.15 CSV download URL.
func buildH15URL() string {
	v := url.Values{}
	v.Set("rel", "H15")
	v.Set("series", "bf17364827e38702b42a58cf8eaa3f78")
	v.Set("lastobs", "")
	v.Set("from", "")
	v.Set("to", "")
	v.Set("filetype", "csv")
	v.Set("label", "include")
	v.Set("layout", "seriescolumn")
	v.Set("type", "package")
	return baseFedBoard + "/datadownload/Output.aspx?" + v.Encode()
}
