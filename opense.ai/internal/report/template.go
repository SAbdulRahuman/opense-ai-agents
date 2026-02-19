package report

// ReportTemplate is the HTML template for the research report.
// It is embedded as a Go constant — no external file dependencies.
const ReportTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.Title}}</title>
<style>
  :root {
    --bg: #ffffff;
    --text: #1a1a2e;
    --muted: #6b7280;
    --border: #e5e7eb;
    --accent: #2563eb;
    --green: #16a34a;
    --red: #dc2626;
    --orange: #ea580c;
    --section-bg: #f8fafc;
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    color: var(--text);
    background: var(--bg);
    line-height: 1.6;
    max-width: 900px;
    margin: 0 auto;
    padding: 20px;
  }
  h1, h2, h3, h4 { font-weight: 600; }
  h1 { font-size: 1.5rem; margin-bottom: 4px; }
  h2 { font-size: 1.2rem; margin: 24px 0 12px; padding-bottom: 6px; border-bottom: 2px solid var(--accent); }
  h3 { font-size: 1rem; margin: 16px 0 8px; }
  p { margin: 6px 0; }
  .muted { color: var(--muted); font-size: 0.85rem; }

  /* Header */
  .header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    border-bottom: 3px solid var(--accent);
    padding-bottom: 12px;
    margin-bottom: 16px;
  }
  .header-left h1 { color: var(--accent); }
  .header-right { text-align: right; }
  .ticker-badge {
    display: inline-block;
    background: var(--accent);
    color: white;
    padding: 2px 12px;
    border-radius: 4px;
    font-weight: 700;
    font-size: 1.1rem;
    margin-right: 8px;
  }

  /* Quote bar */
  .quote-bar {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 8px;
    background: var(--section-bg);
    padding: 12px;
    border-radius: 8px;
    margin-bottom: 16px;
  }
  .quote-item { text-align: center; }
  .quote-item .label { font-size: 0.75rem; color: var(--muted); text-transform: uppercase; }
  .quote-item .value { font-size: 1rem; font-weight: 600; }
  .positive { color: var(--green); }
  .negative { color: var(--red); }

  /* Recommendation badge */
  .rec-box {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px;
    border-radius: 8px;
    margin: 12px 0;
  }
  .rec-box.strong-buy { background: #dcfce7; border-left: 5px solid var(--green); }
  .rec-box.buy { background: #ecfdf5; border-left: 5px solid #22c55e; }
  .rec-box.hold { background: #fefce8; border-left: 5px solid #eab308; }
  .rec-box.sell { background: #fef2f2; border-left: 5px solid #f97316; }
  .rec-box.strong-sell { background: #fef2f2; border-left: 5px solid var(--red); }
  .rec-label { font-size: 1.4rem; font-weight: 700; }
  .rec-box.strong-buy .rec-label { color: var(--green); }
  .rec-box.buy .rec-label { color: #22c55e; }
  .rec-box.hold .rec-label { color: #eab308; }
  .rec-box.sell .rec-label { color: #f97316; }
  .rec-box.strong-sell .rec-label { color: var(--red); }

  /* Trade box */
  .trade-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 10px;
    margin: 12px 0;
  }
  .trade-item {
    background: var(--section-bg);
    padding: 10px;
    border-radius: 6px;
    text-align: center;
  }
  .trade-item .label { font-size: 0.75rem; color: var(--muted); text-transform: uppercase; }
  .trade-item .value { font-size: 1.05rem; font-weight: 600; }

  /* Signal table */
  table { width: 100%; border-collapse: collapse; margin: 8px 0 16px; font-size: 0.9rem; }
  th { background: var(--section-bg); text-align: left; padding: 8px; font-weight: 600; }
  td { padding: 8px; border-bottom: 1px solid var(--border); }
  .signal-badge {
    display: inline-block;
    padding: 1px 8px;
    border-radius: 3px;
    font-size: 0.8rem;
    font-weight: 600;
  }
  .signal-badge.buy { background: #dcfce7; color: var(--green); }
  .signal-badge.sell { background: #fef2f2; color: var(--red); }
  .signal-badge.neutral { background: #f3f4f6; color: var(--muted); }

  /* Ratio grid */
  .ratio-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 8px;
    margin: 10px 0 16px;
  }
  .ratio-card {
    background: var(--section-bg);
    padding: 8px 12px;
    border-radius: 6px;
    display: flex;
    justify-content: space-between;
  }
  .ratio-card .label { color: var(--muted); font-size: 0.85rem; }
  .ratio-card .value { font-weight: 600; }

  /* Chart container */
  .chart-container {
    margin: 12px 0;
    overflow-x: auto;
  }
  .chart-container svg { max-width: 100%; height: auto; }

  /* Section */
  .section { margin: 20px 0; }
  .section-summary {
    background: var(--section-bg);
    padding: 12px;
    border-radius: 6px;
    margin: 8px 0;
    font-size: 0.95rem;
    line-height: 1.7;
  }

  /* Footer */
  .footer {
    margin-top: 30px;
    padding-top: 12px;
    border-top: 2px solid var(--border);
    font-size: 0.8rem;
    color: var(--muted);
    text-align: center;
  }

  /* Gauge inline */
  .gauge-inline { display: flex; align-items: center; gap: 12px; }
  .gauge-inline svg { flex-shrink: 0; }

  @media print {
    body { max-width: 100%; padding: 10px; }
    .section { page-break-inside: avoid; }
  }
</style>
</head>
<body>

<!-- ═══════ HEADER ═══════ -->
<div class="header">
  <div class="header-left">
    <h1><span class="ticker-badge">{{.Ticker}}</span> {{.CompanyName}}</h1>
    <p class="muted">{{.Exchange}} · {{.Sector}} · {{.Industry}}</p>
  </div>
  <div class="header-right">
    <p class="muted">{{.GeneratedAt}}</p>
    <p class="muted">{{.Author}}</p>
  </div>
</div>

<!-- ═══════ QUOTE BAR ═══════ -->
{{if .LastPrice}}
<div class="quote-bar">
  <div class="quote-item">
    <div class="label">Last Price</div>
    <div class="value">{{.LastPrice}}</div>
  </div>
  <div class="quote-item">
    <div class="label">Change</div>
    <div class="value">{{.Change}} ({{.ChangePct}})</div>
  </div>
  <div class="quote-item">
    <div class="label">Day Range</div>
    <div class="value">{{.DayLow}} — {{.DayHigh}}</div>
  </div>
  <div class="quote-item">
    <div class="label">52W Range</div>
    <div class="value">{{.WeekLow52}} — {{.WeekHigh52}}</div>
  </div>
  <div class="quote-item">
    <div class="label">Volume</div>
    <div class="value">{{.Volume}}</div>
  </div>
  <div class="quote-item">
    <div class="label">Market Cap</div>
    <div class="value">{{.MarketCap}}</div>
  </div>
  <div class="quote-item">
    <div class="label">P/E</div>
    <div class="value">{{.PE}}</div>
  </div>
  <div class="quote-item">
    <div class="label">Div Yield</div>
    <div class="value">{{.DividendYield}}</div>
  </div>
</div>
{{end}}

<!-- ═══════ RECOMMENDATION ═══════ -->
{{if .ShowRecommend}}
<div class="section">
  <h2>Recommendation</h2>
  <div class="rec-box {{.RecommendationClass}}">
    <div>
      <div class="rec-label">{{.Recommendation}}</div>
      <div class="muted">Confidence: {{.Confidence}} · Timeframe: {{.Timeframe}}</div>
    </div>
    <div class="gauge-inline">{{.GaugeChart}}</div>
  </div>

  {{if .EntryPrice}}
  <div class="trade-grid">
    <div class="trade-item"><div class="label">Entry</div><div class="value">{{.EntryPrice}}</div></div>
    <div class="trade-item"><div class="label">Target</div><div class="value positive">{{.TargetPrice}}</div></div>
    <div class="trade-item"><div class="label">Stop Loss</div><div class="value negative">{{.StopLoss}}</div></div>
    <div class="trade-item"><div class="label">Risk/Reward</div><div class="value">{{.RiskReward}}</div></div>
  </div>
  {{end}}

  <div class="section-summary">{{.Summary}}</div>
</div>
{{end}}

<!-- ═══════ PRICE CHART ═══════ -->
{{if .PriceChart}}
<div class="section">
  <h2>Price Chart</h2>
  <div class="chart-container">{{.PriceChart}}</div>
</div>
{{end}}

<!-- ═══════ FUNDAMENTAL ═══════ -->
{{if .ShowFundamental}}
<div class="section">
  <h2>Fundamental Analysis</h2>
  <div class="section-summary">{{.FundamentalSummary}}</div>

  {{if .FundamentalSignals}}
  <table>
    <thead><tr><th>Source</th><th>Signal</th><th>Confidence</th><th>Reason</th></tr></thead>
    <tbody>
    {{range .FundamentalSignals}}
    <tr>
      <td>{{.Source}}</td>
      <td><span class="signal-badge {{.TypeClass}}">{{.Type}}</span></td>
      <td>{{.Confidence}}</td>
      <td>{{.Reason}}</td>
    </tr>
    {{end}}
    </tbody>
  </table>
  {{end}}

  {{if .FinancialRatios}}
  <h3>Key Ratios</h3>
  <div class="ratio-grid">
    {{range .FinancialRatios}}
    <div class="ratio-card">
      <span class="label">{{.Label}}</span>
      <span class="value">{{.Value}}</span>
    </div>
    {{end}}
  </div>
  {{end}}
</div>
{{end}}

<!-- ═══════ TECHNICAL ═══════ -->
{{if .ShowTechnical}}
<div class="section">
  <h2>Technical Analysis</h2>
  <div class="section-summary">{{.TechnicalSummary}}</div>

  {{if .TechnicalSignals}}
  <table>
    <thead><tr><th>Indicator</th><th>Signal</th><th>Confidence</th><th>Reason</th></tr></thead>
    <tbody>
    {{range .TechnicalSignals}}
    <tr>
      <td>{{.Source}}</td>
      <td><span class="signal-badge {{.TypeClass}}">{{.Type}}</span></td>
      <td>{{.Confidence}}</td>
      <td>{{.Reason}}</td>
    </tr>
    {{end}}
    </tbody>
  </table>
  {{end}}
</div>
{{end}}

<!-- ═══════ DERIVATIVES ═══════ -->
{{if .ShowDerivatives}}
<div class="section">
  <h2>Derivatives &amp; F&amp;O View</h2>
  <div class="section-summary">{{.DerivativesSummary}}</div>

  {{if .DerivativesSignals}}
  <table>
    <thead><tr><th>Source</th><th>Signal</th><th>Confidence</th><th>Reason</th></tr></thead>
    <tbody>
    {{range .DerivativesSignals}}
    <tr>
      <td>{{.Source}}</td>
      <td><span class="signal-badge {{.TypeClass}}">{{.Type}}</span></td>
      <td>{{.Confidence}}</td>
      <td>{{.Reason}}</td>
    </tr>
    {{end}}
    </tbody>
  </table>
  {{end}}

  {{if .OptionStrategy}}
  <h3>Option Strategy: {{.OptionStrategy}}</h3>
  <div class="trade-grid">
    <div class="trade-item"><div class="label">Max Profit</div><div class="value positive">{{.MaxProfit}}</div></div>
    <div class="trade-item"><div class="label">Max Loss</div><div class="value negative">{{.MaxLoss}}</div></div>
    <div class="trade-item"><div class="label">Breakevens</div><div class="value">{{.Breakevens}}</div></div>
  </div>
  {{end}}

  {{if .PayoffChart}}
  <div class="chart-container">{{.PayoffChart}}</div>
  {{end}}
</div>
{{end}}

<!-- ═══════ SENTIMENT ═══════ -->
{{if .ShowSentiment}}
<div class="section">
  <h2>Sentiment Analysis</h2>
  <div class="section-summary">{{.SentimentSummary}}</div>

  {{if .SentimentSignals}}
  <table>
    <thead><tr><th>Source</th><th>Signal</th><th>Confidence</th><th>Reason</th></tr></thead>
    <tbody>
    {{range .SentimentSignals}}
    <tr>
      <td>{{.Source}}</td>
      <td><span class="signal-badge {{.TypeClass}}">{{.Type}}</span></td>
      <td>{{.Confidence}}</td>
      <td>{{.Reason}}</td>
    </tr>
    {{end}}
    </tbody>
  </table>
  {{end}}
</div>
{{end}}

<!-- ═══════ RISK ═══════ -->
{{if .ShowRisk}}
<div class="section">
  <h2>Risk Assessment</h2>
  <div class="section-summary">{{.RiskSummary}}</div>

  {{if .RiskSignals}}
  <table>
    <thead><tr><th>Factor</th><th>Signal</th><th>Confidence</th><th>Detail</th></tr></thead>
    <tbody>
    {{range .RiskSignals}}
    <tr>
      <td>{{.Source}}</td>
      <td><span class="signal-badge {{.TypeClass}}">{{.Type}}</span></td>
      <td>{{.Confidence}}</td>
      <td>{{.Reason}}</td>
    </tr>
    {{end}}
    </tbody>
  </table>
  {{end}}
</div>
{{end}}

<!-- ═══════ FOOTER ═══════ -->
<div class="footer">
  <p><strong>Disclaimer:</strong> This report is AI-generated by OpeNSE.ai for educational and informational purposes only.
  It does not constitute financial advice. Always consult a SEBI-registered investment advisor before making investment decisions.</p>
  <p>© {{.GeneratedAt}} OpeNSE.ai · Generated on {{.GeneratedAt}}</p>
</div>

</body>
</html>`
