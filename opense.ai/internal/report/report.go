package report

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
	"github.com/seenimoa/openseai/pkg/utils"
)

// ════════════════════════════════════════════════════════════════════
// Report Generator — Orchestrates chart + template rendering
// ════════════════════════════════════════════════════════════════════

// ReportFormat specifies the output format.
type ReportFormat string

const (
	FormatHTML ReportFormat = "html"
	FormatPDF  ReportFormat = "pdf"
	FormatText ReportFormat = "text"
)

// ReportSection identifies a section to include/exclude.
type ReportSection string

const (
	SectionSummary      ReportSection = "summary"
	SectionFundamental  ReportSection = "fundamental"
	SectionTechnical    ReportSection = "technical"
	SectionDerivatives  ReportSection = "derivatives"
	SectionSentiment    ReportSection = "sentiment"
	SectionRisk         ReportSection = "risk"
	SectionRecommend    ReportSection = "recommendation"
)

// AllSections returns all report sections in display order.
func AllSections() []ReportSection {
	return []ReportSection{
		SectionSummary,
		SectionFundamental,
		SectionTechnical,
		SectionDerivatives,
		SectionSentiment,
		SectionRisk,
		SectionRecommend,
	}
}

// ReportConfig controls report generation behaviour.
type ReportConfig struct {
	Format   ReportFormat    // output format (default: HTML)
	Sections []ReportSection // sections to include (default: all)
	Title    string          // custom report title (optional)
	Author   string          // author name (optional, default: "OpeNSE.ai Agent")
	Logo     string          // SVG or base64 logo (optional)
	ChartCfg ChartConfig     // chart rendering config
}

// DefaultReportConfig returns sensible defaults.
func DefaultReportConfig() ReportConfig {
	return ReportConfig{
		Format:   FormatHTML,
		Sections: AllSections(),
		Author:   "OpeNSE.ai Agent",
		ChartCfg: DefaultChartConfig(),
	}
}

// hasSection returns true if the section is included in the config.
func (rc ReportConfig) hasSection(s ReportSection) bool {
	for _, sec := range rc.Sections {
		if sec == s {
			return true
		}
	}
	return false
}

// ════════════════════════════════════════════════════════════════════
// Report Data — Flattened for template rendering
// ════════════════════════════════════════════════════════════════════

// ReportData is the template model passed to HTML templates.
type ReportData struct {
	// Header
	Title       string
	Ticker      string
	CompanyName string
	Exchange    string
	Sector      string
	Industry    string
	Author      string
	GeneratedAt string // IST formatted
	LogoSVG     string

	// Quote
	LastPrice     string
	Change        string
	ChangePct     string
	DayHigh       string
	DayLow        string
	WeekHigh52    string
	WeekLow52     string
	Volume        string
	MarketCap     string
	PE            string
	PB            string
	DividendYield string

	// Recommendation
	Recommendation     string
	RecommendationClass string // CSS class: strong-buy, buy, hold, sell, strong-sell
	Confidence         string
	ConfidenceValue    float64
	Summary            string
	EntryPrice         string
	TargetPrice        string
	StopLoss           string
	RiskReward         string
	Timeframe          string

	// Analysis sections
	TechnicalSummary    string
	TechnicalSignals    []SignalRow
	FundamentalSummary  string
	FundamentalSignals  []SignalRow
	DerivativesSummary  string
	DerivativesSignals  []SignalRow
	SentimentSummary    string
	SentimentSignals    []SignalRow
	RiskSummary         string
	RiskSignals         []SignalRow

	// Charts (embedded SVG strings)
	PriceChart         template.HTML
	PerformanceChart   template.HTML
	PayoffChart        template.HTML
	GaugeChart         template.HTML
	RatioChart         template.HTML

	// Financials
	FinancialRatios    []RatioRow
	IncomeStatements   []FinancialRow
	BalanceSheetItems  []FinancialRow

	// Section visibility flags
	ShowFundamental bool
	ShowTechnical   bool
	ShowDerivatives bool
	ShowSentiment   bool
	ShowRisk        bool
	ShowRecommend   bool

	// Option strategy
	OptionStrategy string
	MaxProfit      string
	MaxLoss        string
	Breakevens     string
}

// SignalRow is a flattened signal for template rendering.
type SignalRow struct {
	Source     string
	Type       string // "BUY", "SELL", "NEUTRAL"
	TypeClass  string // CSS class: buy, sell, neutral
	Confidence string
	Reason     string
}

// RatioRow represents a key-value financial ratio row.
type RatioRow struct {
	Label string
	Value string
}

// FinancialRow represents a row in the income/bs table.
type FinancialRow struct {
	Period string
	Values []string
}

// ════════════════════════════════════════════════════════════════════
// Generate Report
// ════════════════════════════════════════════════════════════════════

// GenerateHTML generates an HTML research report from CompositeAnalysis.
func GenerateHTML(analysis *models.CompositeAnalysis, cfg ReportConfig) (string, error) {
	if analysis == nil {
		return "", fmt.Errorf("analysis is nil")
	}

	data := buildReportData(analysis, cfg)

	tmpl, err := template.New("report").Parse(ReportTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

// GenerateText generates a plain-text research report (terminal / CLI friendly).
func GenerateText(analysis *models.CompositeAnalysis, cfg ReportConfig) (string, error) {
	if analysis == nil {
		return "", fmt.Errorf("analysis is nil")
	}

	data := buildReportData(analysis, cfg)
	return renderTextReport(data), nil
}

// ════════════════════════════════════════════════════════════════════
// Internal — Build template data
// ════════════════════════════════════════════════════════════════════

func buildReportData(a *models.CompositeAnalysis, cfg ReportConfig) ReportData {
	now := utils.NowIST()
	profile := a.StockProfile

	data := ReportData{
		Title:       cfg.Title,
		Ticker:      a.Ticker,
		CompanyName: profile.Stock.Name,
		Exchange:    profile.Stock.Exchange,
		Sector:      profile.Stock.Sector,
		Industry:    profile.Stock.Industry,
		Author:      cfg.Author,
		GeneratedAt: now.Format("02 Jan 2006, 03:04 PM IST"),
		LogoSVG:     cfg.Logo,

		// Recommendation
		Recommendation:      formatRecommendation(a.Recommendation),
		RecommendationClass: recommendationClass(a.Recommendation),
		Confidence:          fmt.Sprintf("%.0f%%", float64(a.Confidence)*100),
		ConfidenceValue:     float64(a.Confidence) * 100,
		Summary:             a.Summary,
		Timeframe:           a.Timeframe,

		// Section visibility
		ShowFundamental: cfg.hasSection(SectionFundamental) && a.Fundamental != nil,
		ShowTechnical:   cfg.hasSection(SectionTechnical) && a.Technical != nil,
		ShowDerivatives: cfg.hasSection(SectionDerivatives) && a.Derivatives != nil,
		ShowSentiment:   cfg.hasSection(SectionSentiment) && a.Sentiment != nil,
		ShowRisk:        cfg.hasSection(SectionRisk) && a.Risk != nil,
		ShowRecommend:   cfg.hasSection(SectionRecommend),
	}

	if data.Title == "" {
		data.Title = fmt.Sprintf("%s — Research Report", a.Ticker)
	}

	// Quote info
	if profile.Quote != nil {
		q := profile.Quote
		data.LastPrice = utils.FormatINR(q.LastPrice)
		data.Change = utils.FormatINR(q.Change)
		data.ChangePct = utils.FormatPct(q.ChangePct)
		data.DayHigh = utils.FormatINR(q.High)
		data.DayLow = utils.FormatINR(q.Low)
		data.WeekHigh52 = utils.FormatINR(q.WeekHigh52)
		data.WeekLow52 = utils.FormatINR(q.WeekLow52)
		data.Volume = fmt.Sprintf("%d", q.Volume)
		data.MarketCap = utils.FormatINRCompact(q.MarketCap)
		data.PE = fmt.Sprintf("%.2f", q.PE)
		data.PB = fmt.Sprintf("%.2f", q.PB)
		data.DividendYield = fmt.Sprintf("%.2f%%", q.DividendYield)
	}

	// Entry / Target / SL
	if a.EntryPrice > 0 {
		data.EntryPrice = utils.FormatINR(a.EntryPrice)
	}
	if a.TargetPrice > 0 {
		data.TargetPrice = utils.FormatINR(a.TargetPrice)
	}
	if a.StopLoss > 0 {
		data.StopLoss = utils.FormatINR(a.StopLoss)
	}
	if a.RiskRewardRatio > 0 {
		data.RiskReward = fmt.Sprintf("1:%.1f", a.RiskRewardRatio)
	}

	// Analysis sections
	if a.Technical != nil {
		data.TechnicalSummary = a.Technical.Summary
		data.TechnicalSignals = flattenSignals(a.Technical.Signals)
	}
	if a.Fundamental != nil {
		data.FundamentalSummary = a.Fundamental.Summary
		data.FundamentalSignals = flattenSignals(a.Fundamental.Signals)
	}
	if a.Derivatives != nil {
		data.DerivativesSummary = a.Derivatives.Summary
		data.DerivativesSignals = flattenSignals(a.Derivatives.Signals)
	}
	if a.Sentiment != nil {
		data.SentimentSummary = a.Sentiment.Summary
		data.SentimentSignals = flattenSignals(a.Sentiment.Signals)
	}
	if a.Risk != nil {
		data.RiskSummary = a.Risk.Summary
		data.RiskSignals = flattenSignals(a.Risk.Signals)
	}

	// Financial ratios
	if profile.Ratios != nil {
		data.FinancialRatios = buildRatioRows(profile.Ratios)
	}

	// Charts
	data.GaugeChart = template.HTML(GaugeChart(data.ConfidenceValue, "Confidence", 180))

	// Price chart from historical data
	if len(profile.Historical) > 0 {
		chartCfg := cfg.ChartCfg
		chartCfg.Title = fmt.Sprintf("%s Price Chart", a.Ticker)
		overlays := buildOverlaysFromDetails(a.Technical)
		data.PriceChart = template.HTML(CandlestickChart(profile.Historical, overlays, chartCfg))
	}

	// Option payoff chart
	if a.Derivatives != nil && a.Derivatives.Details != nil {
		if strategy, ok := a.Derivatives.Details["strategy"]; ok {
			if strat, ok := strategy.(*models.OptionStrategy); ok {
				data.OptionStrategy = strat.Name
				data.MaxProfit = utils.FormatINR(strat.MaxProfit)
				data.MaxLoss = utils.FormatINR(strat.MaxLoss)
				bes := make([]string, len(strat.Breakevens))
				for i, b := range strat.Breakevens {
					bes[i] = utils.FormatINR(b)
				}
				data.Breakevens = strings.Join(bes, ", ")
				if len(strat.Payoff) > 0 {
					chartCfg := cfg.ChartCfg
					chartCfg.Title = strat.Name + " Payoff"
					data.PayoffChart = template.HTML(OptionPayoffChart(strat.Payoff, strat.Name, chartCfg))
				}
			}
		}
	}

	return data
}

func flattenSignals(signals []models.Signal) []SignalRow {
	rows := make([]SignalRow, len(signals))
	for i, s := range signals {
		rows[i] = SignalRow{
			Source:     s.Source,
			Type:       string(s.Type),
			TypeClass:  signalClass(s.Type),
			Confidence: fmt.Sprintf("%.0f%%", float64(s.Confidence)*100),
			Reason:     s.Reason,
		}
	}
	return rows
}

func signalClass(t models.SignalType) string {
	switch t {
	case models.SignalBuy:
		return "buy"
	case models.SignalSell:
		return "sell"
	default:
		return "neutral"
	}
}

func formatRecommendation(r models.Recommendation) string {
	switch r {
	case models.StrongBuy:
		return "Strong Buy"
	case models.ModerateBuy:
		return "Buy"
	case models.Hold:
		return "Hold"
	case models.ModerateSell:
		return "Sell"
	case models.StrongSell:
		return "Strong Sell"
	default:
		return string(r)
	}
}

func recommendationClass(r models.Recommendation) string {
	switch r {
	case models.StrongBuy:
		return "strong-buy"
	case models.ModerateBuy:
		return "buy"
	case models.Hold:
		return "hold"
	case models.ModerateSell:
		return "sell"
	case models.StrongSell:
		return "strong-sell"
	default:
		return "neutral"
	}
}

func buildRatioRows(r *models.FinancialRatios) []RatioRow {
	return []RatioRow{
		{Label: "P/E Ratio", Value: fmt.Sprintf("%.2f", r.PE)},
		{Label: "P/B Ratio", Value: fmt.Sprintf("%.2f", r.PB)},
		{Label: "EV/EBITDA", Value: fmt.Sprintf("%.2f", r.EVBITDA)},
		{Label: "ROE", Value: utils.FormatPct(r.ROE)},
		{Label: "ROCE", Value: utils.FormatPct(r.ROCE)},
		{Label: "Debt/Equity", Value: fmt.Sprintf("%.2f", r.DebtEquity)},
		{Label: "Current Ratio", Value: fmt.Sprintf("%.2f", r.CurrentRatio)},
		{Label: "Interest Coverage", Value: fmt.Sprintf("%.2f", r.InterestCoverage)},
		{Label: "Dividend Yield", Value: fmt.Sprintf("%.2f%%", r.DividendYield)},
		{Label: "EPS", Value: utils.FormatINR(r.EPS)},
		{Label: "Book Value", Value: utils.FormatINR(r.BookValue)},
		{Label: "PEG Ratio", Value: fmt.Sprintf("%.2f", r.PEGRatio)},
		{Label: "Graham Number", Value: utils.FormatINR(r.GrahamNumber)},
	}
}

func buildOverlaysFromDetails(tech *models.AnalysisResult) map[string][]float64 {
	if tech == nil || tech.Details == nil {
		return nil
	}
	overlays := make(map[string][]float64)
	if sma, ok := tech.Details["sma_20"]; ok {
		if vals, ok := sma.([]float64); ok {
			overlays["SMA 20"] = vals
		}
	}
	if sma, ok := tech.Details["sma_50"]; ok {
		if vals, ok := sma.([]float64); ok {
			overlays["SMA 50"] = vals
		}
	}
	if ema, ok := tech.Details["ema_20"]; ok {
		if vals, ok := ema.([]float64); ok {
			overlays["EMA 20"] = vals
		}
	}
	return overlays
}

// ════════════════════════════════════════════════════════════════════
// Plain-text renderer
// ════════════════════════════════════════════════════════════════════

func renderTextReport(d ReportData) string {
	var sb strings.Builder
	line := strings.Repeat("═", 60)
	thinLine := strings.Repeat("─", 60)

	sb.WriteString("\n" + line + "\n")
	sb.WriteString(fmt.Sprintf("  %s\n", d.Title))
	sb.WriteString(fmt.Sprintf("  Generated: %s | Author: %s\n", d.GeneratedAt, d.Author))
	sb.WriteString(line + "\n\n")

	// Company info
	sb.WriteString(fmt.Sprintf("  %s (%s) — %s\n", d.CompanyName, d.Ticker, d.Exchange))
	sb.WriteString(fmt.Sprintf("  Sector: %s | Industry: %s\n", d.Sector, d.Industry))
	sb.WriteString(thinLine + "\n")

	// Quote
	if d.LastPrice != "" {
		sb.WriteString(fmt.Sprintf("  Price: %s (%s, %s)\n", d.LastPrice, d.Change, d.ChangePct))
		sb.WriteString(fmt.Sprintf("  Day: %s — %s | 52W: %s — %s\n", d.DayLow, d.DayHigh, d.WeekLow52, d.WeekHigh52))
		sb.WriteString(fmt.Sprintf("  Volume: %s | Market Cap: %s\n", d.Volume, d.MarketCap))
		sb.WriteString(thinLine + "\n")
	}

	// Recommendation
	if d.ShowRecommend {
		sb.WriteString("\n  ★ RECOMMENDATION\n")
		sb.WriteString(fmt.Sprintf("  %s (Confidence: %s)\n", d.Recommendation, d.Confidence))
		if d.EntryPrice != "" {
			sb.WriteString(fmt.Sprintf("  Entry: %s | Target: %s | Stop Loss: %s\n", d.EntryPrice, d.TargetPrice, d.StopLoss))
		}
		if d.RiskReward != "" {
			sb.WriteString(fmt.Sprintf("  Risk/Reward: %s | Timeframe: %s\n", d.RiskReward, d.Timeframe))
		}
		sb.WriteString(fmt.Sprintf("\n  %s\n", d.Summary))
		sb.WriteString(thinLine + "\n")
	}

	// Analysis sections
	writeSection := func(title string, show bool, summary string, signals []SignalRow) {
		if !show {
			return
		}
		sb.WriteString(fmt.Sprintf("\n  ■ %s\n", title))
		sb.WriteString(fmt.Sprintf("  %s\n", summary))
		for _, s := range signals {
			sb.WriteString(fmt.Sprintf("    [%s] %s — %s (Conf: %s)\n", s.Type, s.Source, s.Reason, s.Confidence))
		}
		sb.WriteString(thinLine + "\n")
	}

	writeSection("FUNDAMENTAL ANALYSIS", d.ShowFundamental, d.FundamentalSummary, d.FundamentalSignals)
	writeSection("TECHNICAL ANALYSIS", d.ShowTechnical, d.TechnicalSummary, d.TechnicalSignals)
	writeSection("DERIVATIVES VIEW", d.ShowDerivatives, d.DerivativesSummary, d.DerivativesSignals)
	writeSection("SENTIMENT ANALYSIS", d.ShowSentiment, d.SentimentSummary, d.SentimentSignals)
	writeSection("RISK ASSESSMENT", d.ShowRisk, d.RiskSummary, d.RiskSignals)

	// Key ratios
	if len(d.FinancialRatios) > 0 {
		sb.WriteString("\n  ■ KEY FINANCIAL RATIOS\n")
		for _, r := range d.FinancialRatios {
			sb.WriteString(fmt.Sprintf("    %-20s %s\n", r.Label, r.Value))
		}
		sb.WriteString(thinLine + "\n")
	}

	// Option strategy
	if d.OptionStrategy != "" {
		sb.WriteString(fmt.Sprintf("\n  ■ OPTION STRATEGY: %s\n", d.OptionStrategy))
		sb.WriteString(fmt.Sprintf("    Max Profit: %s | Max Loss: %s\n", d.MaxProfit, d.MaxLoss))
		sb.WriteString(fmt.Sprintf("    Breakevens: %s\n", d.Breakevens))
		sb.WriteString(thinLine + "\n")
	}

	sb.WriteString("\n" + line + "\n")
	sb.WriteString("  Disclaimer: This report is AI-generated for educational purposes.\n")
	sb.WriteString("  Not financial advice. Always consult a SEBI-registered advisor.\n")
	sb.WriteString(line + "\n")

	return sb.String()
}

// ════════════════════════════════════════════════════════════════════
// Utility: Timestamp
// ════════════════════════════════════════════════════════════════════

// ReportTimestamp returns current IST time formatted for report headers.
func ReportTimestamp() string {
	return utils.NowIST().Format("02 Jan 2006, 03:04 PM IST")
}

// FormatDuration formats a duration for display.
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}
