// Package imf implements an IMF data provider.
// Data is sourced from IMF SDMX REST API v3.0 and ArcGIS PortWatch FeatureServer.
// No API key required.
package imf

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
)

const (
	providerName = "imf"

	// IMF SDMX REST API v3.0
	baseSDMX = "https://api.imf.org/external/sdmx/3.0/data/dataflow"

	// IMF PortWatch ArcGIS FeatureServer
	baseArcGIS = "https://services9.arcgis.com/weJ1QsnbMYJlCHdG/arcgis/rest/services"
)

// Provider is the IMF data provider.
type Provider struct {
	provider.BaseProvider
	client *http.Client
}

// New creates a new IMF provider and registers all fetchers.
func New() *Provider {
	p := &Provider{
		BaseProvider: provider.NewBaseProvider(
			providerName,
			"IMF — SDMX API & PortWatch ArcGIS for economic data and maritime tracking (free, no API key)",
			"https://data.imf.org",
			nil,
		),
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
	}

	// SDMX-based endpoints.
	p.RegisterFetcher(newAvailableIndicatorsFetcher(p))
	p.RegisterFetcher(newIMFCPIFetcher(p))
	p.RegisterFetcher(newDirectionOfTradeFetcher(p))
	p.RegisterFetcher(newEconomicIndicatorsFetcher(p))

	// ArcGIS PortWatch endpoints.
	p.RegisterFetcher(newChokePointInfoFetcher(p))
	p.RegisterFetcher(newChokePointVolumeFetcher(p))
	p.RegisterFetcher(newPortInfoFetcher(p))
	p.RegisterFetcher(newPortVolumeFetcher(p))

	return p
}

// Ping verifies connectivity to the IMF SDMX API.
func (p *Provider) Ping(ctx context.Context) error {
	url := baseSDMX + "/IMF/CPI/+/USA.CPI._T.IX.M?lastNObservations=1&detail=full"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("imf ping: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("imf ping: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("imf ping: HTTP %d", resp.StatusCode)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// fetchJSON fetches JSON from the given URL and decodes into dst.
func (p *Provider) fetchJSON(ctx context.Context, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}

// fetchArcGISAll fetches all pages from an ArcGIS FeatureServer query.
func (p *Provider) fetchArcGISAll(ctx context.Context, service, where, fields string) ([]map[string]any, error) {
	var allFeatures []map[string]any
	offset := 0

	for {
		url := fmt.Sprintf(
			"%s/%s/FeatureServer/0/query?where=%s&outFields=%s&returnGeometry=false&resultOffset=%d&resultRecordCount=2000&f=json",
			baseArcGIS, service, where, fields, offset,
		)

		var resp arcGISResponse
		if err := p.fetchJSON(ctx, url, &resp); err != nil {
			return nil, err
		}

		for _, f := range resp.Features {
			allFeatures = append(allFeatures, f.Attributes)
		}

		if !resp.ExceededTransferLimit || len(resp.Features) == 0 {
			break
		}
		offset += len(resp.Features)
	}

	return allFeatures, nil
}

// arcGISResponse wraps standard ArcGIS query response.
type arcGISResponse struct {
	Features              []arcGISFeature `json:"features"`
	ExceededTransferLimit bool            `json:"exceededTransferLimit"`
}

type arcGISFeature struct {
	Attributes map[string]any `json:"attributes"`
}

// sdmxDataResponse wraps the IMF SDMX JSON response (simplified).
type sdmxDataResponse struct {
	Data sdmxData `json:"data"`
}

type sdmxData struct {
	DataSets []sdmxDataSet `json:"dataSets"`
}

type sdmxDataSet struct {
	Series map[string]sdmxSeries `json:"series"`
}

type sdmxSeries struct {
	Observations map[string][]any `json:"observations"`
}

// buildSDMXURL constructs an IMF SDMX URL.
func buildSDMXURL(agency, dataflow, key string, params map[string]string) string {
	u := fmt.Sprintf("%s/%s/%s/+/%s?dimensionAtObservation=TIME_PERIOD&detail=full&includeHistory=false",
		baseSDMX, agency, dataflow, key)
	for k, v := range params {
		u += "&" + k + "=" + v
	}
	return u
}

// parseFloat parses a string to float64.
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "NaN" || s == "NA" {
		return 0
	}
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// parseAnyFloat extracts a float64 from an any value.
func parseAnyFloat(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		return parseFloat(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	default:
		return 0
	}
}

// parseAnyString extracts a string from an any value.
func parseAnyString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// parseAnyDate parses a date from various formats.
func parseAnyDate(s string) time.Time {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"2006-01-02", "2006-01", "2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// newResult wraps data in a FetchResult.
func newResult(data any) *provider.FetchResult {
	return &provider.FetchResult{
		Data:      data,
		FetchedAt: time.Now(),
	}
}

// resolveCountry maps user-friendly country names to ISO3 codes.
func resolveCountry(input string) string {
	if input == "" || input == "all" || input == "*" {
		return "*"
	}
	input = strings.ToLower(strings.TrimSpace(input))
	if code, ok := countryToISO[input]; ok {
		return code
	}
	if len(input) == 3 {
		return strings.ToUpper(input)
	}
	return strings.ToUpper(input)
}

var countryToISO = map[string]string{
	"united_states": "USA", "united_kingdom": "GBR", "germany": "DEU",
	"france": "FRA", "japan": "JPN", "canada": "CAN", "italy": "ITA",
	"australia": "AUS", "brazil": "BRA", "china": "CHN", "india": "IND",
	"indonesia": "IDN", "mexico": "MEX", "south_korea": "KOR",
	"south_africa": "ZAF", "turkey": "TUR", "spain": "ESP",
	"world": "G001", "euro_area": "G163",
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
