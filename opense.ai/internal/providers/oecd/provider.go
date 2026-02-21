// Package oecd implements an OECD SDMX data provider.
// Data is sourced from https://sdmx.oecd.org/public/rest/data/ using CSV responses.
// No API key required.
package oecd

import (
	"context"
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
)

const (
	providerName = "oecd"
	baseURL      = "https://sdmx.oecd.org/public/rest/data"
)

// Provider is the OECD data provider.
type Provider struct {
	provider.BaseProvider
	client *http.Client
}

// New creates a new OECD provider and registers all fetchers.
func New() *Provider {
	// OECD needs legacy TLS support.
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS10, //nolint:gosec // OECD SDMX server requires legacy TLS
		},
	}

	p := &Provider{
		BaseProvider: provider.NewBaseProvider(
			providerName,
			"OECD — SDMX REST API for economic indicators (free, no API key)",
			"https://sdmx.oecd.org",
			nil,
		),
		client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}

	// 9 endpoints
	p.RegisterFetcher(newCLIFetcher(p))
	p.RegisterFetcher(newCPIFetcher(p))
	p.RegisterFetcher(newInterestRatesFetcher(p))
	p.RegisterFetcher(newGdpNominalFetcher(p))
	p.RegisterFetcher(newGdpRealFetcher(p))
	p.RegisterFetcher(newGdpForecastFetcher(p))
	p.RegisterFetcher(newHousePriceFetcher(p))
	p.RegisterFetcher(newSharePriceFetcher(p))
	p.RegisterFetcher(newUnemploymentFetcher(p))

	return p
}

// Ping verifies connectivity to the OECD SDMX API.
func (p *Provider) Ping(ctx context.Context) error {
	url := baseURL + "/OECD.SDD.STES,DSD_KEI@DF_KEI,4.0/USA.M.IR3TIB....?lastNObservations=1&detail=dataonly"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("oecd ping: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.sdmx.data+csv; charset=utf-8")
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("oecd ping: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("oecd ping: HTTP %d", resp.StatusCode)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// fetchCSV fetches SDMX CSV data from the OECD API and returns parsed records.
// The first row is the header.
func (p *Provider) fetchCSV(ctx context.Context, url string) ([][]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.sdmx.data+csv; charset=utf-8")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	reader := csv.NewReader(resp.Body)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	return reader.ReadAll()
}

// buildURL constructs an OECD SDMX URL.
func buildURL(dsd, key, startPeriod, endPeriod string, extraParams ...string) string {
	u := baseURL + "/" + dsd + "/" + key + "?"
	parts := []string{
		"dimensionAtObservation=TIME_PERIOD",
		"detail=dataonly",
	}
	if startPeriod != "" {
		parts = append(parts, "startPeriod="+startPeriod)
	}
	if endPeriod != "" {
		parts = append(parts, "endPeriod="+endPeriod)
	}
	parts = append(parts, extraParams...)
	return u + strings.Join(parts, "&")
}

// findColumn returns the index of a column name in the header, or -1.
func findColumn(header []string, name string) int {
	for i, h := range header {
		if strings.EqualFold(strings.TrimSpace(h), name) {
			return i
		}
	}
	return -1
}

// parseSDMXDate parses OECD date formats: "2023", "2023-05", "2023-Q1", "2023-01-15".
func parseSDMXDate(s string) time.Time {
	s = strings.TrimSpace(s)

	// "2023-Q1" format
	if strings.Contains(s, "-Q") {
		parts := strings.SplitN(s, "-Q", 2)
		if len(parts) == 2 {
			month := "01"
			switch parts[1] {
			case "2":
				month = "04"
			case "3":
				month = "07"
			case "4":
				month = "10"
			}
			t, _ := time.Parse("2006-01-02", parts[0]+"-"+month+"-01")
			return t
		}
	}

	// Try common formats.
	for _, layout := range []string{"2006-01-02", "2006-01", "2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// parseFloat parses a string to float64, returning 0 if invalid.
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "NaN" || s == "NA" {
		return 0
	}
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// newResult wraps data in a FetchResult.
func newResult(data any) *provider.FetchResult {
	return &provider.FetchResult{
		Data:      data,
		FetchedAt: time.Now(),
	}
}

// countryCodeToName maps ISO3 country codes to display names.
var countryCodeToName = map[string]string{
	"AUS": "Australia", "AUT": "Austria", "BEL": "Belgium", "BRA": "Brazil",
	"CAN": "Canada", "CHE": "Switzerland", "CHL": "Chile", "CHN": "China",
	"COL": "Colombia", "CRI": "Costa Rica", "CZE": "Czech Republic",
	"DEU": "Germany", "DNK": "Denmark", "ESP": "Spain", "EST": "Estonia",
	"FIN": "Finland", "FRA": "France", "GBR": "United Kingdom",
	"GRC": "Greece", "HUN": "Hungary", "IDN": "Indonesia", "IND": "India",
	"IRL": "Ireland", "ISL": "Iceland", "ISR": "Israel", "ITA": "Italy",
	"JPN": "Japan", "KOR": "South Korea", "LTU": "Lithuania", "LUX": "Luxembourg",
	"LVA": "Latvia", "MEX": "Mexico", "NLD": "Netherlands", "NOR": "Norway",
	"NZL": "New Zealand", "POL": "Poland", "PRT": "Portugal", "ROU": "Romania",
	"RUS": "Russia", "SAU": "Saudi Arabia", "SVK": "Slovakia", "SVN": "Slovenia",
	"SWE": "Sweden", "TUR": "Turkey", "USA": "United States", "ZAF": "South Africa",
	"G7": "G7", "G20": "G20", "OECD": "OECD", "EA19": "Euro Area",
}

func countryName(code string) string {
	if name, ok := countryCodeToName[code]; ok {
		return name
	}
	return code
}

// inputCountryToISO maps user-friendly country names to ISO3 codes for OECD URLs.
var inputCountryToISO = map[string]string{
	"united_states": "USA", "united_kingdom": "GBR", "germany": "DEU",
	"france": "FRA", "japan": "JPN", "canada": "CAN", "italy": "ITA",
	"australia": "AUS", "brazil": "BRA", "china": "CHN", "india": "IND",
	"indonesia": "IDN", "mexico": "MEX", "south_korea": "KOR",
	"south_africa": "ZAF", "turkey": "TUR", "spain": "ESP",
	"g20": "G20", "g7": "G7", "all": "",
}

func resolveCountry(input string) string {
	if input == "" || input == "all" {
		return ""
	}
	input = strings.ToLower(strings.TrimSpace(input))
	if code, ok := inputCountryToISO[input]; ok {
		return code
	}
	// If already an ISO code, use as-is.
	if len(input) == 3 {
		return strings.ToUpper(input)
	}
	return strings.ToUpper(input)
}

// datePeriod extracts start/end period from params.
func datePeriod(params provider.QueryParams) (string, string) {
	start := params[provider.ParamStartDate]
	end := params[provider.ParamEndDate]
	// Truncate to YYYY-MM for OECD API.
	if len(start) > 7 {
		start = start[:7]
	}
	if len(end) > 7 {
		end = end[:7]
	}
	return start, end
}

// io import helper — reads full body (used by XML endpoints if needed).
var _ = io.ReadAll
