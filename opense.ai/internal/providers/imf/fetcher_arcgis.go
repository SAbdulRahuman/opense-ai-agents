package imf

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ArcGIS service names on IMF PortWatch.
const (
	svcChokePointDB = "PortWatch_chokepoints_database"
	svcChokePointTS = "Daily_Chokepoints_Data"
	svcPortDB       = "PortWatch_ports_database"
	svcPortTS       = "Daily_Trade_Data"
)

// ---------------------------------------------------------------------------
// MaritimeChokePointInfo — static list of 24 global chokepoints.
// ---------------------------------------------------------------------------

type chokePointInfoFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newChokePointInfoFetcher(p *Provider) *chokePointInfoFetcher {
	return &chokePointInfoFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelMaritimeChokePointInfo,
			"IMF PortWatch maritime chokepoint metadata (24 global chokepoints)",
			nil,
			nil,
		),
		p: p,
	}
}

func (f *chokePointInfoFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelMaritimeChokePointInfo, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	features, err := f.p.fetchArcGISAll(ctx, svcChokePointDB, url.QueryEscape("1=1"), "*")
	if err != nil {
		return nil, fmt.Errorf("chokepoint info: %w", err)
	}

	var data []models.PortData
	for _, attrs := range features {
		data = append(data, models.PortData{
			Port:     parseAnyString(attrs["portname"]),
			Country:  "", // chokepoints are international
			Volume:   parseAnyFloat(attrs["vessel_count_total"]),
			Unit:     "vessels",
			Category: "chokepoint",
		})
	}

	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// MaritimeChokePointVolume — daily vessel traffic through chokepoints.
// ---------------------------------------------------------------------------

type chokePointVolumeFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newChokePointVolumeFetcher(p *Provider) *chokePointVolumeFetcher {
	return &chokePointVolumeFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelMaritimeChokePointVolume,
			"IMF PortWatch daily chokepoint traffic volume",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *chokePointVolumeFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelMaritimeChokePointVolume, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	where := buildDateWhere(params, "date")
	if cp := params["chokepoint"]; cp != "" {
		where = fmt.Sprintf("portid = '%s' AND %s", cp, where)
	}

	features, err := f.p.fetchArcGISAll(ctx, svcChokePointTS, url.QueryEscape(where), "*")
	if err != nil {
		return nil, fmt.Errorf("chokepoint volume: %w", err)
	}

	var data []models.PortData
	for _, attrs := range features {
		dt := constructDate(attrs)
		data = append(data, models.PortData{
			Port:     parseAnyString(attrs["portname"]),
			Date:     dt,
			Volume:   parseAnyFloat(attrs["n_total"]),
			Unit:     "vessels",
			Category: "chokepoint",
		})
	}

	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// PortInfo — static list of global ports.
// ---------------------------------------------------------------------------

type portInfoFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newPortInfoFetcher(p *Provider) *portInfoFetcher {
	return &portInfoFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelPortInfo,
			"IMF PortWatch global port metadata",
			nil,
			[]string{provider.ParamCountry},
		),
		p: p,
	}
}

func (f *portInfoFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelPortInfo, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	where := "1=1"
	if c := params[provider.ParamCountry]; c != "" {
		iso3 := resolveCountry(c)
		if iso3 != "*" {
			where = fmt.Sprintf("ISO3 = '%s'", iso3)
		}
	}

	features, err := f.p.fetchArcGISAll(ctx, svcPortDB, url.QueryEscape(where), "*")
	if err != nil {
		return nil, fmt.Errorf("port info: %w", err)
	}

	var data []models.PortData
	for _, attrs := range features {
		data = append(data, models.PortData{
			Port:     parseAnyString(attrs["portname"]),
			Country:  parseAnyString(attrs["ISO3"]),
			Volume:   parseAnyFloat(attrs["vessel_count_total"]),
			Unit:     "vessels",
			Category: "port",
		})
	}

	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// PortVolume — daily trade volume for a port.
// ---------------------------------------------------------------------------

type portVolumeFetcher struct {
	provider.BaseFetcher
	p *Provider
}

func newPortVolumeFetcher(p *Provider) *portVolumeFetcher {
	return &portVolumeFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelPortVolume,
			"IMF PortWatch daily port trade volume",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
		p: p,
	}
}

func (f *portVolumeFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}
	cacheKey := provider.CacheKey(provider.ModelPortVolume, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	portCode := params["port_code"]
	if portCode == "" {
		portCode = "port1114" // Shanghai default
	}

	where := fmt.Sprintf("portid = '%s'", portCode)
	dateFilter := buildDateWhere(params, "date")
	if dateFilter != "1=1" {
		where += " AND " + dateFilter
	}

	features, err := f.p.fetchArcGISAll(ctx, svcPortTS, url.QueryEscape(where), "*")
	if err != nil {
		return nil, fmt.Errorf("port volume: %w", err)
	}

	var data []models.PortData
	for _, attrs := range features {
		dt := constructDate(attrs)
		data = append(data, models.PortData{
			Port:     parseAnyString(attrs["portname"]),
			Country:  parseAnyString(attrs["ISO3"]),
			Date:     dt,
			Volume:   parseAnyFloat(attrs["import"]) + parseAnyFloat(attrs["export"]),
			Unit:     "metric_tons",
			Category: "port",
		})
	}

	result := newResult(data)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// Shared helpers for ArcGIS fetchers
// ---------------------------------------------------------------------------

// buildDateWhere constructs a SQL WHERE clause for ArcGIS date filtering.
func buildDateWhere(params provider.QueryParams, dateField string) string {
	start := params[provider.ParamStartDate]
	end := params[provider.ParamEndDate]

	if start == "" && end == "" {
		return "1=1"
	}

	var clauses []string
	if start != "" {
		clauses = append(clauses, fmt.Sprintf("%s >= TIMESTAMP '%s 00:00:00'", dateField, start))
	}
	if end != "" {
		clauses = append(clauses, fmt.Sprintf("%s <= TIMESTAMP '%s 00:00:00'", dateField, end))
	}

	result := clauses[0]
	if len(clauses) > 1 {
		result += " AND " + clauses[1]
	}
	return result
}

// constructDate builds a time.Time from year/month/day fields in ArcGIS attributes.
func constructDate(attrs map[string]any) time.Time {
	year := int(parseAnyFloat(attrs["year"]))
	month := int(parseAnyFloat(attrs["month"]))
	day := int(parseAnyFloat(attrs["day"]))

	if year > 0 && month > 0 && day > 0 {
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	}

	// Try the "date" field as Unix timestamp (ArcGIS often uses epoch millis).
	if d := parseAnyFloat(attrs["date"]); d > 0 {
		if d > 1e12 {
			return time.UnixMilli(int64(d)).UTC()
		}
		return time.Unix(int64(d), 0).UTC()
	}

	return time.Time{}
}
