// Package report provides research report generation capabilities for OpeNSE.ai.
// It generates SVG charts, HTML research reports, and optional PDF exports for
// comprehensive stock analysis with Indian-market formatting.
package report

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
	"github.com/seenimoa/openseai/pkg/utils"
)

// ════════════════════════════════════════════════════════════════════
// SVG Chart Generator — Pure Go, Zero Dependencies
// ════════════════════════════════════════════════════════════════════

// ChartConfig holds rendering parameters for SVG charts.
type ChartConfig struct {
	Width      int    // SVG width in pixels (default: 800)
	Height     int    // SVG height in pixels (default: 400)
	MarginTop  int    // top margin (default: 30)
	MarginRight int   // right margin (default: 60)
	MarginBottom int  // bottom margin (default: 50)
	MarginLeft int    // left margin (default: 70)
	BgColor    string // background color (default: "#ffffff")
	GridColor  string // grid line color (default: "#e0e0e0")
	TextColor  string // axis label color (default: "#333333")
	FontSize   int    // axis label font size (default: 11)
	Title      string // chart title
}

// DefaultChartConfig returns sensible defaults for chart rendering.
func DefaultChartConfig() ChartConfig {
	return ChartConfig{
		Width:        800,
		Height:       400,
		MarginTop:    40,
		MarginRight:  60,
		MarginBottom: 50,
		MarginLeft:   70,
		BgColor:      "#ffffff",
		GridColor:    "#e8e8e8",
		TextColor:    "#333333",
		FontSize:     11,
	}
}

// plotArea returns the usable drawing area dimensions.
func (c ChartConfig) plotArea() (x, y, w, h int) {
	return c.MarginLeft, c.MarginTop,
		c.Width - c.MarginLeft - c.MarginRight,
		c.Height - c.MarginTop - c.MarginBottom
}

// ════════════════════════════════════════════════════════════════════
// Candlestick Chart
// ════════════════════════════════════════════════════════════════════

// CandlestickChart generates an SVG candlestick chart from OHLCV data,
// optionally overlaying SMA/EMA lines and volume bars.
func CandlestickChart(bars []models.OHLCV, overlays map[string][]float64, cfg ChartConfig) string {
	if len(bars) == 0 {
		return emptySVG(cfg, "No data available")
	}

	if cfg.Width == 0 {
		cfg = DefaultChartConfig()
	}
	if cfg.Title == "" {
		cfg.Title = "Price Chart"
	}

	px, py, pw, ph := cfg.plotArea()

	// Compute price range
	minPrice, maxPrice := bars[0].Low, bars[0].High
	for _, b := range bars {
		if b.Low < minPrice {
			minPrice = b.Low
		}
		if b.High > maxPrice {
			maxPrice = b.High
		}
	}
	// Add 5% padding
	priceRange := maxPrice - minPrice
	if priceRange < 0.01 {
		priceRange = 1
	}
	minPrice -= priceRange * 0.05
	maxPrice += priceRange * 0.05
	priceRange = maxPrice - minPrice

	// Volume range
	var maxVol int64
	for _, b := range bars {
		if b.Volume > maxVol {
			maxVol = b.Volume
		}
	}

	n := len(bars)
	candleWidth := float64(pw) / float64(n)
	if candleWidth > 12 {
		candleWidth = 12
	}
	bodyWidth := candleWidth * 0.7
	volHeight := float64(ph) * 0.2 // bottom 20% for volume

	var sb strings.Builder
	sb.WriteString(svgHeader(cfg))

	// Background
	sb.WriteString(fmt.Sprintf(`<rect x="0" y="0" width="%d" height="%d" fill="%s"/>`,
		cfg.Width, cfg.Height, cfg.BgColor))

	// Title
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="20" font-size="14" font-weight="bold" fill="%s" text-anchor="middle">%s</text>`,
		cfg.Width/2, cfg.TextColor, escapeXML(cfg.Title)))

	// Y-axis grid lines and labels (price)
	gridLines := 6
	for i := 0; i <= gridLines; i++ {
		price := minPrice + priceRange*float64(i)/float64(gridLines)
		y := py + ph - int(float64(ph-int(volHeight))*float64(i)/float64(gridLines)) - int(volHeight)
		sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-dasharray="3,3"/>`,
			px, y, px+pw, y, cfg.GridColor))
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="%d" fill="%s" text-anchor="end">%s</text>`,
			px-5, y+4, cfg.FontSize, cfg.TextColor, utils.FormatINR(price)))
	}

	// Helper: price to Y coordinate
	priceToY := func(p float64) int {
		ratio := (p - minPrice) / priceRange
		return py + ph - int(volHeight) - int(ratio*float64(ph-int(volHeight)))
	}

	// Draw volume bars
	if maxVol > 0 {
		for i, b := range bars {
			cx := float64(px) + float64(i)*float64(pw)/float64(n) + float64(pw)/float64(n)/2
			vRatio := float64(b.Volume) / float64(maxVol)
			vh := vRatio * volHeight
			vy := float64(py+ph) - vh
			color := "#c8e6c9" // green
			if b.Close < b.Open {
				color = "#ffcdd2" // red
			}
			sb.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" opacity="0.6"/>`,
				cx-bodyWidth/2, vy, bodyWidth, vh, color))
		}
	}

	// Draw candles
	for i, b := range bars {
		cx := float64(px) + float64(i)*float64(pw)/float64(n) + float64(pw)/float64(n)/2

		wickColor := "#26a69a"
		bodyColor := "#26a69a" // green (bullish)
		if b.Close < b.Open {
			wickColor = "#ef5350"
			bodyColor = "#ef5350" // red (bearish)
		}

		// Wick (high to low)
		sb.WriteString(fmt.Sprintf(`<line x1="%.1f" y1="%d" x2="%.1f" y2="%d" stroke="%s" stroke-width="1"/>`,
			cx, priceToY(b.High), cx, priceToY(b.Low), wickColor))

		// Body (open to close)
		openY := priceToY(b.Open)
		closeY := priceToY(b.Close)
		bodyTop := openY
		bodyH := closeY - openY
		if bodyH < 0 {
			bodyTop = closeY
			bodyH = -bodyH
		}
		if bodyH < 1 {
			bodyH = 1
		}
		sb.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%d" width="%.1f" height="%d" fill="%s"/>`,
			cx-bodyWidth/2, bodyTop, bodyWidth, bodyH, bodyColor))
	}

	// Draw overlay lines (SMA, EMA, etc.)
	colors := []string{"#ff9800", "#2196f3", "#9c27b0", "#4caf50"}
	colorIdx := 0
	for name, values := range overlays {
		if len(values) != n {
			continue
		}
		color := colors[colorIdx%len(colors)]
		colorIdx++

		var pathParts []string
		for i, v := range values {
			if v == 0 || math.IsNaN(v) {
				continue
			}
			cx := float64(px) + float64(i)*float64(pw)/float64(n) + float64(pw)/float64(n)/2
			y := priceToY(v)
			cmd := "L"
			if len(pathParts) == 0 {
				cmd = "M"
			}
			pathParts = append(pathParts, fmt.Sprintf("%s%.1f,%d", cmd, cx, y))
		}
		if len(pathParts) > 1 {
			sb.WriteString(fmt.Sprintf(`<path d="%s" fill="none" stroke="%s" stroke-width="1.5" opacity="0.8"/>`,
				strings.Join(pathParts, " "), color))
			// Legend
			ly := py + 15 + colorIdx*16
			sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="2"/>`,
				px+10, ly, px+30, ly, color))
			sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="10" fill="%s">%s</text>`,
				px+35, ly+4, cfg.TextColor, escapeXML(name)))
		}
	}

	// X-axis date labels
	labelInterval := n / 6
	if labelInterval < 1 {
		labelInterval = 1
	}
	for i := 0; i < n; i += labelInterval {
		cx := float64(px) + float64(i)*float64(pw)/float64(n) + float64(pw)/float64(n)/2
		label := bars[i].Timestamp.Format("02 Jan")
		sb.WriteString(fmt.Sprintf(`<text x="%.1f" y="%d" font-size="%d" fill="%s" text-anchor="middle" transform="rotate(-45,%.1f,%d)">%s</text>`,
			cx, py+ph+15, cfg.FontSize-1, cfg.TextColor, cx, py+ph+15, label))
	}

	sb.WriteString("</svg>")
	return sb.String()
}

// ════════════════════════════════════════════════════════════════════
// Line Chart
// ════════════════════════════════════════════════════════════════════

// LineChartSeries represents a named data series for line charts.
type LineChartSeries struct {
	Name   string
	Values []float64
	Color  string // hex color (optional, auto-assigned if empty)
}

// LineChart generates an SVG line chart with one or more series.
// Labels are optional X-axis labels corresponding to data points.
func LineChart(series []LineChartSeries, labels []string, cfg ChartConfig) string {
	if len(series) == 0 {
		return emptySVG(cfg, "No data")
	}

	if cfg.Width == 0 {
		cfg = DefaultChartConfig()
	}
	if cfg.Title == "" {
		cfg.Title = "Line Chart"
	}

	px, py, pw, ph := cfg.plotArea()

	// Find global min/max
	minVal, maxVal := math.MaxFloat64, -math.MaxFloat64
	maxLen := 0
	for _, s := range series {
		if len(s.Values) > maxLen {
			maxLen = len(s.Values)
		}
		for _, v := range s.Values {
			if !math.IsNaN(v) && v < minVal {
				minVal = v
			}
			if !math.IsNaN(v) && v > maxVal {
				maxVal = v
			}
		}
	}
	if maxLen == 0 {
		return emptySVG(cfg, "No data points")
	}

	vRange := maxVal - minVal
	if vRange < 0.001 {
		vRange = 1
	}
	minVal -= vRange * 0.05
	maxVal += vRange * 0.05
	vRange = maxVal - minVal

	var sb strings.Builder
	sb.WriteString(svgHeader(cfg))
	sb.WriteString(fmt.Sprintf(`<rect x="0" y="0" width="%d" height="%d" fill="%s"/>`,
		cfg.Width, cfg.Height, cfg.BgColor))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="20" font-size="14" font-weight="bold" fill="%s" text-anchor="middle">%s</text>`,
		cfg.Width/2, cfg.TextColor, escapeXML(cfg.Title)))

	// Y-axis grid
	gridLines := 5
	for i := 0; i <= gridLines; i++ {
		val := minVal + vRange*float64(i)/float64(gridLines)
		y := py + ph - int(float64(ph)*float64(i)/float64(gridLines))
		sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-dasharray="3,3"/>`,
			px, y, px+pw, y, cfg.GridColor))
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="%d" fill="%s" text-anchor="end">%.1f</text>`,
			px-5, y+4, cfg.FontSize, cfg.TextColor, val))
	}

	// Draw series
	defaultColors := []string{"#2196f3", "#ff9800", "#4caf50", "#e91e63", "#9c27b0", "#00bcd4"}
	for si, s := range series {
		color := s.Color
		if color == "" {
			color = defaultColors[si%len(defaultColors)]
		}

		var pathParts []string
		for i, v := range s.Values {
			if math.IsNaN(v) {
				continue
			}
			cx := float64(px) + float64(i)*float64(pw)/float64(maxLen-1)
			ratio := (v - minVal) / vRange
			cy := float64(py+ph) - ratio*float64(ph)
			cmd := "L"
			if len(pathParts) == 0 {
				cmd = "M"
			}
			pathParts = append(pathParts, fmt.Sprintf("%s%.1f,%.1f", cmd, cx, cy))
		}
		if len(pathParts) > 1 {
			sb.WriteString(fmt.Sprintf(`<path d="%s" fill="none" stroke="%s" stroke-width="2"/>`,
				strings.Join(pathParts, " "), color))
		}

		// Legend
		ly := py + 10 + si*16
		sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="2"/>`,
			px+10, ly, px+30, ly, color))
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="10" fill="%s">%s</text>`,
			px+35, ly+4, cfg.TextColor, escapeXML(s.Name)))
	}

	// X-axis labels
	if len(labels) > 0 {
		interval := maxLen / 6
		if interval < 1 {
			interval = 1
		}
		for i := 0; i < len(labels) && i < maxLen; i += interval {
			cx := float64(px) + float64(i)*float64(pw)/float64(maxLen-1)
			sb.WriteString(fmt.Sprintf(`<text x="%.1f" y="%d" font-size="%d" fill="%s" text-anchor="middle">%s</text>`,
				cx, py+ph+18, cfg.FontSize-1, cfg.TextColor, escapeXML(labels[i])))
		}
	}

	sb.WriteString("</svg>")
	return sb.String()
}

// ════════════════════════════════════════════════════════════════════
// Bar Chart (Horizontal)
// ════════════════════════════════════════════════════════════════════

// BarItem represents a single bar in a horizontal bar chart.
type BarItem struct {
	Label string
	Value float64
	Color string // optional
}

// HorizontalBarChart generates an SVG horizontal bar chart.
// Useful for peer comparison, ratio comparison, etc.
func HorizontalBarChart(items []BarItem, cfg ChartConfig) string {
	if len(items) == 0 {
		return emptySVG(cfg, "No data")
	}

	if cfg.Width == 0 {
		cfg = DefaultChartConfig()
	}
	cfg.MarginLeft = 120 // wider for labels
	if cfg.Title == "" {
		cfg.Title = "Comparison"
	}

	px, py, pw, ph := cfg.plotArea()

	maxVal := 0.0
	minVal := 0.0
	for _, item := range items {
		if item.Value > maxVal {
			maxVal = item.Value
		}
		if item.Value < minVal {
			minVal = item.Value
		}
	}

	hasNegative := minVal < 0
	valRange := maxVal - minVal
	if valRange < 0.001 {
		valRange = 1
	}

	barH := float64(ph) / float64(len(items)) * 0.7
	if barH > 30 {
		barH = 30
	}
	gap := (float64(ph) - barH*float64(len(items))) / float64(len(items)+1)

	var sb strings.Builder
	sb.WriteString(svgHeader(cfg))
	sb.WriteString(fmt.Sprintf(`<rect x="0" y="0" width="%d" height="%d" fill="%s"/>`,
		cfg.Width, cfg.Height, cfg.BgColor))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="20" font-size="14" font-weight="bold" fill="%s" text-anchor="middle">%s</text>`,
		cfg.Width/2, cfg.TextColor, escapeXML(cfg.Title)))

	// Zero line for mixed positive/negative
	zeroX := float64(px)
	if hasNegative {
		zeroX = float64(px) + (-minVal/valRange)*float64(pw)
		sb.WriteString(fmt.Sprintf(`<line x1="%.1f" y1="%d" x2="%.1f" y2="%d" stroke="#999" stroke-width="1"/>`,
			zeroX, py, zeroX, py+ph))
	}

	defaultColors := []string{"#2196f3", "#4caf50", "#ff9800", "#e91e63", "#9c27b0", "#00bcd4"}
	for i, item := range items {
		by := float64(py) + gap + float64(i)*(barH+gap)
		color := item.Color
		if color == "" {
			if item.Value >= 0 {
				color = "#4caf50"
			} else {
				color = "#ef5350"
			}
		}

		var bx, bw float64
		if hasNegative {
			if item.Value >= 0 {
				bx = zeroX
				bw = (item.Value / valRange) * float64(pw)
			} else {
				bw = (-item.Value / valRange) * float64(pw)
				bx = zeroX - bw
			}
		} else {
			bx = float64(px)
			bw = (item.Value / maxVal) * float64(pw)
		}

		_ = defaultColors // suppress unused warning for future use
		sb.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" rx="2"/>`,
			bx, by, bw, barH, color))

		// Label
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%.1f" font-size="%d" fill="%s" text-anchor="end">%s</text>`,
			px-5, by+barH/2+4, cfg.FontSize, cfg.TextColor, escapeXML(item.Label)))

		// Value
		sb.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" font-size="%d" fill="%s">%.1f</text>`,
			bx+bw+5, by+barH/2+4, cfg.FontSize, cfg.TextColor, item.Value))
	}

	sb.WriteString("</svg>")
	return sb.String()
}

// ════════════════════════════════════════════════════════════════════
// PE/PB Historical Band Chart
// ════════════════════════════════════════════════════════════════════

// BandDataPoint represents a point in a historical valuation band.
type BandDataPoint struct {
	Date  time.Time
	Value float64
	Price float64
}

// ValuationBandChart generates an SVG chart showing price plotted against
// historical PE or PB bands (showing where the stock trades relative to
// its historical valuation percentiles).
func ValuationBandChart(data []BandDataPoint, bandName string, cfg ChartConfig) string {
	if len(data) == 0 {
		return emptySVG(cfg, "No valuation data")
	}

	if cfg.Width == 0 {
		cfg = DefaultChartConfig()
	}
	if cfg.Title == "" {
		cfg.Title = fmt.Sprintf("Historical %s Band", bandName)
	}

	// Extract values for percentile calculation
	values := make([]float64, len(data))
	prices := make([]float64, len(data))
	labels := make([]string, len(data))
	for i, d := range data {
		values[i] = d.Value
		prices[i] = d.Price
		labels[i] = d.Date.Format("Jan 06")
	}

	return LineChart([]LineChartSeries{
		{Name: "Price", Values: prices, Color: "#2196f3"},
		{Name: bandName, Values: values, Color: "#ff9800"},
	}, labels, cfg)
}

// ════════════════════════════════════════════════════════════════════
// Option Payoff Diagram
// ════════════════════════════════════════════════════════════════════

// OptionPayoffChart generates an SVG chart of option strategy P&L vs underlying price.
func OptionPayoffChart(payoff []models.OptionPayoff, strategyName string, cfg ChartConfig) string {
	if len(payoff) == 0 {
		return emptySVG(cfg, "No payoff data")
	}

	if cfg.Width == 0 {
		cfg = DefaultChartConfig()
	}
	if cfg.Title == "" {
		cfg.Title = fmt.Sprintf("Payoff: %s", strategyName)
	}

	prices := make([]float64, len(payoff))
	pnls := make([]float64, len(payoff))
	labels := make([]string, len(payoff))
	for i, p := range payoff {
		prices[i] = p.UnderlyingPrice
		pnls[i] = p.PnL
		if i%(len(payoff)/6+1) == 0 {
			labels = append(labels[:i], append([]string{fmt.Sprintf("%.0f", p.UnderlyingPrice)}, labels[i+1:]...)...)
		}
	}

	// Use custom rendering for payoff (shows zero line)
	px, py, pw, ph := cfg.plotArea()

	minPnL, maxPnL := pnls[0], pnls[0]
	for _, v := range pnls {
		if v < minPnL {
			minPnL = v
		}
		if v > maxPnL {
			maxPnL = v
		}
	}
	vRange := maxPnL - minPnL
	if vRange < 0.001 {
		vRange = 1
	}
	minPnL -= vRange * 0.1
	maxPnL += vRange * 0.1
	vRange = maxPnL - minPnL

	minPrice, maxPrice := prices[0], prices[len(prices)-1]
	pRange := maxPrice - minPrice
	if pRange < 0.001 {
		pRange = 1
	}

	var sb strings.Builder
	sb.WriteString(svgHeader(cfg))
	sb.WriteString(fmt.Sprintf(`<rect x="0" y="0" width="%d" height="%d" fill="%s"/>`,
		cfg.Width, cfg.Height, cfg.BgColor))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="20" font-size="14" font-weight="bold" fill="%s" text-anchor="middle">%s</text>`,
		cfg.Width/2, cfg.TextColor, escapeXML(cfg.Title)))

	// Zero line
	if minPnL < 0 && maxPnL > 0 {
		zeroY := float64(py+ph) - (-minPnL/vRange)*float64(ph)
		sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%.1f" x2="%d" y2="%.1f" stroke="#999" stroke-width="1" stroke-dasharray="4,4"/>`,
			px, zeroY, px+pw, zeroY))
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%.1f" font-size="%d" fill="#999" text-anchor="end">0</text>`,
			px-5, zeroY+4, cfg.FontSize))
	}

	// Draw payoff line with profit (green) / loss (red) fill
	var pathParts []string
	for i, p := range payoff {
		cx := float64(px) + ((p.UnderlyingPrice-minPrice)/pRange)*float64(pw)
		cy := float64(py+ph) - ((p.PnL-minPnL)/vRange)*float64(ph)
		cmd := "L"
		if i == 0 {
			cmd = "M"
		}
		pathParts = append(pathParts, fmt.Sprintf("%s%.1f,%.1f", cmd, cx, cy))
	}
	sb.WriteString(fmt.Sprintf(`<path d="%s" fill="none" stroke="#2196f3" stroke-width="2.5"/>`,
		strings.Join(pathParts, " ")))

	// Y-axis labels
	for i := 0; i <= 5; i++ {
		val := minPnL + vRange*float64(i)/5
		y := py + ph - int(float64(ph)*float64(i)/5)
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="%d" fill="%s" text-anchor="end">%s</text>`,
			px-5, y+4, cfg.FontSize, cfg.TextColor, utils.FormatINR(val)))
	}

	// X-axis labels
	for i := 0; i <= 5; i++ {
		val := minPrice + pRange*float64(i)/5
		x := px + int(float64(pw)*float64(i)/5)
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="%d" fill="%s" text-anchor="middle">%.0f</text>`,
			x, py+ph+18, cfg.FontSize, cfg.TextColor, val))
	}

	sb.WriteString("</svg>")
	return sb.String()
}

// ════════════════════════════════════════════════════════════════════
// Gauge / Dial Chart (for signal strength)
// ════════════════════════════════════════════════════════════════════

// GaugeChart generates an SVG semicircular gauge for displaying a value
// like RSI, confidence score, or recommendation strength.
// value should be 0-100, label is the display label.
func GaugeChart(value float64, label string, width int) string {
	if width == 0 {
		width = 200
	}
	height := width/2 + 30

	cx := float64(width) / 2
	cy := float64(width)/2 - 10
	radius := float64(width)/2 - 20

	// Clamp value
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}

	// Angle: 180° (left) to 0° (right), value maps 0→180°, 100→0°
	angle := math.Pi - (value/100)*math.Pi
	needleX := cx + radius*0.85*math.Cos(angle)
	needleY := cy - radius*0.85*math.Sin(angle)

	// Color zones
	var color string
	switch {
	case value < 30:
		color = "#ef5350" // red
	case value < 50:
		color = "#ff9800" // orange
	case value < 70:
		color = "#ffc107" // yellow
	default:
		color = "#4caf50" // green
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height))
	sb.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="white"/>`, width, height))

	// Background arc
	sb.WriteString(fmt.Sprintf(`<path d="M%.1f,%.1f A%.1f,%.1f 0 0,1 %.1f,%.1f" fill="none" stroke="#e0e0e0" stroke-width="12" stroke-linecap="round"/>`,
		cx-radius, cy, radius, radius, cx+radius, cy))

	// Colored arc (proportional to value)
	endAngle := math.Pi - (value/100)*math.Pi
	endX := cx + radius*math.Cos(endAngle)
	endY := cy - radius*math.Sin(endAngle)
	largeArc := 0
	if value > 50 {
		largeArc = 1
	}
	sb.WriteString(fmt.Sprintf(`<path d="M%.1f,%.1f A%.1f,%.1f 0 %d,1 %.1f,%.1f" fill="none" stroke="%s" stroke-width="12" stroke-linecap="round"/>`,
		cx-radius, cy, radius, radius, largeArc, endX, endY, color))

	// Needle
	sb.WriteString(fmt.Sprintf(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#333" stroke-width="2"/>`,
		cx, cy, needleX, needleY))
	sb.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="5" fill="#333"/>`, cx, cy))

	// Value text
	sb.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" font-size="22" font-weight="bold" fill="%s" text-anchor="middle">%.0f</text>`,
		cx, cy+25, color, value))

	// Label
	sb.WriteString(fmt.Sprintf(`<text x="%.1f" y="%d" font-size="11" fill="#666" text-anchor="middle">%s</text>`,
		cx, height-5, escapeXML(label)))

	sb.WriteString("</svg>")
	return sb.String()
}

// ════════════════════════════════════════════════════════════════════
// SVG Helpers
// ════════════════════════════════════════════════════════════════════

func svgHeader(cfg ChartConfig) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" font-family="sans-serif">`,
		cfg.Width, cfg.Height, cfg.Width, cfg.Height)
}

func emptySVG(cfg ChartConfig, msg string) string {
	if cfg.Width == 0 {
		cfg.Width = 400
	}
	if cfg.Height == 0 {
		cfg.Height = 200
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d"><rect width="%d" height="%d" fill="#f5f5f5"/><text x="%d" y="%d" text-anchor="middle" fill="#999" font-size="14">%s</text></svg>`,
		cfg.Width, cfg.Height, cfg.Width, cfg.Height, cfg.Width/2, cfg.Height/2, escapeXML(msg))
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
