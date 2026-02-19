package datasource

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/seenimoa/openseai/pkg/models"
	"github.com/seenimoa/openseai/pkg/utils"
)

const screenerBaseURL = "https://www.screener.in"

// Screener implements the DataSource interface by scraping Screener.in.
type Screener struct {
	cache   *Cache
	limiter *RateLimiter
}

// NewScreener creates a new Screener.in data source.
func NewScreener() *Screener {
	return &Screener{
		cache:   NewCache(30 * time.Minute),
		limiter: NewRateLimiter(1, time.Second), // conservative: 1 req/s
	}
}

// Name returns the data source name.
func (s *Screener) Name() string { return "Screener.in" }

// --- Public methods ---

// GetFinancials returns financial statements scraped from Screener.in.
func (s *Screener) GetFinancials(ctx context.Context, ticker string) (*models.FinancialData, error) {
	symbol := utils.NormalizeTicker(ticker)

	cacheKey := "scr:fin:" + symbol
	if cached, ok := s.cache.Get(cacheKey); ok {
		return cached.(*models.FinancialData), nil
	}

	doc, err := s.fetchPage(ctx, symbol)
	if err != nil {
		return nil, err
	}

	fd := &models.FinancialData{
		Ticker: symbol,
	}

	// Parse quarterly results table.
	fd.QuarterlyIncome = s.parseIncomeTable(doc, "#quarters")

	// Parse annual profit & loss.
	fd.AnnualIncome = s.parseIncomeTable(doc, "#profit-loss")

	// Parse balance sheet.
	fd.AnnualBalanceSheet = s.parseBalanceSheet(doc, "#balance-sheet")

	// Parse cash flow.
	fd.AnnualCashFlow = s.parseCashFlow(doc, "#cash-flow")

	s.cache.SetWithTTL(cacheKey, fd, 1*time.Hour)
	return fd, nil
}

// GetFinancialRatios returns key ratios scraped from Screener.in.
func (s *Screener) GetFinancialRatios(ctx context.Context, ticker string) (*models.FinancialRatios, error) {
	symbol := utils.NormalizeTicker(ticker)

	cacheKey := "scr:ratios:" + symbol
	if cached, ok := s.cache.Get(cacheKey); ok {
		return cached.(*models.FinancialRatios), nil
	}

	doc, err := s.fetchPage(ctx, symbol)
	if err != nil {
		return nil, err
	}

	ratios := &models.FinancialRatios{}

	// Screener.in shows ratios in a top-level "ratios" list.
	doc.Find("#top-ratios li").Each(func(_ int, sel *goquery.Selection) {
		name := strings.TrimSpace(sel.Find(".name").Text())
		valStr := strings.TrimSpace(sel.Find(".number").Text())
		val := parseScreenerNumber(valStr)

		switch {
		case strings.Contains(name, "Stock P/E"):
			ratios.PE = val
		case strings.Contains(name, "Book Value"):
			ratios.BookValue = val
		case strings.Contains(name, "Price to book"):
			ratios.PB = val
		case strings.Contains(name, "Dividend Yield"):
			ratios.DividendYield = val
		case strings.Contains(name, "ROCE"):
			ratios.ROCE = val
		case strings.Contains(name, "ROE"):
			ratios.ROE = val
		case strings.Contains(name, "Face Value"):
			// skip, not in ratios struct
		case strings.Contains(name, "Debt to equity"):
			ratios.DebtEquity = val
		case strings.Contains(name, "Interest Coverage"):
			ratios.InterestCoverage = val
		case strings.Contains(name, "Current ratio"):
			ratios.CurrentRatio = val
		case strings.Contains(name, "EPS"):
			ratios.EPS = val
		case strings.Contains(name, "PEG"):
			ratios.PEGRatio = val
		}
	})

	s.cache.SetWithTTL(cacheKey, ratios, 1*time.Hour)
	return ratios, nil
}

// GetPeerComparison returns peer company comparison from Screener.in.
func (s *Screener) GetPeerComparison(ctx context.Context, ticker string) ([]map[string]string, error) {
	symbol := utils.NormalizeTicker(ticker)

	cacheKey := "scr:peers:" + symbol
	if cached, ok := s.cache.Get(cacheKey); ok {
		return cached.([]map[string]string), nil
	}

	doc, err := s.fetchPage(ctx, symbol)
	if err != nil {
		return nil, err
	}

	var peers []map[string]string
	var headers []string

	doc.Find("#peers table thead th").Each(func(_ int, sel *goquery.Selection) {
		headers = append(headers, strings.TrimSpace(sel.Text()))
	})

	doc.Find("#peers table tbody tr").Each(func(_ int, row *goquery.Selection) {
		peer := make(map[string]string)
		row.Find("td").Each(func(i int, cell *goquery.Selection) {
			if i < len(headers) {
				peer[headers[i]] = strings.TrimSpace(cell.Text())
			}
		})
		if len(peer) > 0 {
			peers = append(peers, peer)
		}
	})

	s.cache.SetWithTTL(cacheKey, peers, 1*time.Hour)
	return peers, nil
}

// --- DataSource interface ---

// GetQuote is not supported by Screener.in.
func (s *Screener) GetQuote(_ context.Context, _ string) (*models.Quote, error) {
	return nil, ErrNotSupported
}

// GetHistoricalData is not supported by Screener.in.
func (s *Screener) GetHistoricalData(_ context.Context, _ string, _, _ time.Time, _ models.Timeframe) ([]models.OHLCV, error) {
	return nil, ErrNotSupported
}

// GetOptionChain is not supported by Screener.in.
func (s *Screener) GetOptionChain(_ context.Context, _ string, _ string) (*models.OptionChain, error) {
	return nil, ErrNotSupported
}

// GetStockProfile returns a profile with financial data from Screener.in.
func (s *Screener) GetStockProfile(ctx context.Context, ticker string) (*models.StockProfile, error) {
	fd, err := s.GetFinancials(ctx, ticker)
	if err != nil {
		return nil, err
	}

	ratios, _ := s.GetFinancialRatios(ctx, ticker)

	return &models.StockProfile{
		Stock: models.Stock{
			Ticker:   utils.NormalizeTicker(ticker),
			Exchange: "NSE",
		},
		Financials: fd,
		Ratios:     ratios,
		FetchedAt:  time.Now(),
	}, nil
}

// --- Internal helpers ---

// fetchPage downloads and parses the Screener.in company page.
func (s *Screener) fetchPage(ctx context.Context, symbol string) (*goquery.Document, error) {
	if err := s.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/company/%s/consolidated/", screenerBaseURL, symbol)
	body, _, err := doGet(ctx, url, map[string]string{
		"Accept": "text/html",
	})
	if err != nil {
		// Try standalone if consolidated not found.
		url = fmt.Sprintf("%s/company/%s/", screenerBaseURL, symbol)
		body, _, err = doGet(ctx, url, map[string]string{
			"Accept": "text/html",
		})
		if err != nil {
			return nil, fmt.Errorf("screener.in %s: %w", symbol, err)
		}
	}
	defer body.Close()

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("parse screener HTML: %w", err)
	}

	return doc, nil
}

// parseIncomeTable parses an income statement table from Screener.in.
func (s *Screener) parseIncomeTable(doc *goquery.Document, sectionID string) []models.IncomeStatement {
	var periods []string
	var statements []models.IncomeStatement

	section := doc.Find(sectionID)
	if section.Length() == 0 {
		return nil
	}

	// Parse header row for period names.
	section.Find("table thead th").Each(func(i int, th *goquery.Selection) {
		if i > 0 { // skip row label column
			periods = append(periods, strings.TrimSpace(th.Text()))
		}
	})

	// Initialize statement slices.
	statements = make([]models.IncomeStatement, len(periods))
	for i, p := range periods {
		statements[i].Period = p
		if strings.Contains(sectionID, "quarter") {
			statements[i].PeriodType = "quarterly"
		} else {
			statements[i].PeriodType = "annual"
		}
	}

	// Parse data rows.
	section.Find("table tbody tr").Each(func(_ int, row *goquery.Selection) {
		label := strings.TrimSpace(row.Find("td:first-child").Text())
		row.Find("td").Each(func(i int, cell *goquery.Selection) {
			if i == 0 || i-1 >= len(statements) {
				return
			}
			val := parseScreenerNumber(strings.TrimSpace(cell.Text()))
			idx := i - 1

			switch {
			case strings.Contains(label, "Sales") || strings.Contains(label, "Revenue"):
				statements[idx].Revenue = val
			case strings.Contains(label, "Other Income"):
				statements[idx].OtherIncome = val
			case strings.Contains(label, "Total Income"):
				statements[idx].TotalIncome = val
			case strings.Contains(label, "Raw Material"):
				statements[idx].RawMaterials = val
			case strings.Contains(label, "Employee"):
				statements[idx].EmployeeCost = val
			case strings.Contains(label, "Total Expenses") || strings.Contains(label, "Expenses"):
				statements[idx].TotalExpenses = val
			case strings.Contains(label, "OPM"):
				statements[idx].OPMPct = val
			case strings.Contains(label, "EBITDA"):
				statements[idx].EBITDA = val
			case strings.Contains(label, "Depreciation"):
				statements[idx].Depreciation = val
			case strings.Contains(label, "Interest"):
				statements[idx].InterestExpense = val
			case strings.Contains(label, "Profit before tax"):
				statements[idx].PBT = val
			case strings.Contains(label, "Tax"):
				statements[idx].Tax = val
			case strings.Contains(label, "Net Profit"):
				statements[idx].PAT = val
			case strings.Contains(label, "EPS"):
				statements[idx].EPS = val
			}
		})
	})

	return statements
}

// parseBalanceSheet parses the balance sheet table.
func (s *Screener) parseBalanceSheet(doc *goquery.Document, sectionID string) []models.BalanceSheet {
	var periods []string
	section := doc.Find(sectionID)
	if section.Length() == 0 {
		return nil
	}

	section.Find("table thead th").Each(func(i int, th *goquery.Selection) {
		if i > 0 {
			periods = append(periods, strings.TrimSpace(th.Text()))
		}
	})

	sheets := make([]models.BalanceSheet, len(periods))
	for i, p := range periods {
		sheets[i].Period = p
		sheets[i].PeriodType = "annual"
	}

	section.Find("table tbody tr").Each(func(_ int, row *goquery.Selection) {
		label := strings.TrimSpace(row.Find("td:first-child").Text())
		row.Find("td").Each(func(i int, cell *goquery.Selection) {
			if i == 0 || i-1 >= len(sheets) {
				return
			}
			val := parseScreenerNumber(strings.TrimSpace(cell.Text()))
			idx := i - 1

			switch {
			case strings.Contains(label, "Share Capital"):
				sheets[idx].ShareCapital = val
			case strings.Contains(label, "Reserves"):
				sheets[idx].Reserves = val
			case strings.Contains(label, "Total Equity") || strings.Contains(label, "Equity"):
				sheets[idx].TotalEquity = val
			case strings.Contains(label, "Borrowings") && strings.Contains(label, "Long"):
				sheets[idx].LongTermBorrowings = val
			case strings.Contains(label, "Borrowings") && strings.Contains(label, "Short"):
				sheets[idx].ShortTermBorrowings = val
			case strings.Contains(label, "Total Debt") || strings.Contains(label, "Borrowings"):
				sheets[idx].TotalDebt = val
			case strings.Contains(label, "Fixed Assets"):
				sheets[idx].FixedAssets = val
			case strings.Contains(label, "CWIP"):
				sheets[idx].CWIP = val
			case strings.Contains(label, "Investments"):
				sheets[idx].Investments = val
			case strings.Contains(label, "Cash"):
				sheets[idx].CashEquivalents = val
			case strings.Contains(label, "Total Assets"):
				sheets[idx].TotalAssets = val
			case strings.Contains(label, "Current Assets"):
				sheets[idx].CurrentAssets = val
			case strings.Contains(label, "Current Liabilities"):
				sheets[idx].CurrentLiabilities = val
			}
		})
	})

	return sheets
}

// parseCashFlow parses the cash flow table.
func (s *Screener) parseCashFlow(doc *goquery.Document, sectionID string) []models.CashFlow {
	var periods []string
	section := doc.Find(sectionID)
	if section.Length() == 0 {
		return nil
	}

	section.Find("table thead th").Each(func(i int, th *goquery.Selection) {
		if i > 0 {
			periods = append(periods, strings.TrimSpace(th.Text()))
		}
	})

	flows := make([]models.CashFlow, len(periods))
	for i, p := range periods {
		flows[i].Period = p
		flows[i].PeriodType = "annual"
	}

	section.Find("table tbody tr").Each(func(_ int, row *goquery.Selection) {
		label := strings.TrimSpace(row.Find("td:first-child").Text())
		row.Find("td").Each(func(i int, cell *goquery.Selection) {
			if i == 0 || i-1 >= len(flows) {
				return
			}
			val := parseScreenerNumber(strings.TrimSpace(cell.Text()))
			idx := i - 1

			switch {
			case strings.Contains(label, "Operating") && strings.Contains(label, "Cash"):
				flows[idx].OperatingCashFlow = val
			case strings.Contains(label, "Investing") && strings.Contains(label, "Cash"):
				flows[idx].InvestingCashFlow = val
			case strings.Contains(label, "Financing") && strings.Contains(label, "Cash"):
				flows[idx].FinancingCashFlow = val
			case strings.Contains(label, "Net Cash") || strings.Contains(label, "Net cash"):
				flows[idx].NetCashFlow = val
			case strings.Contains(label, "Free Cash") || strings.Contains(label, "Free cash"):
				flows[idx].FreeCashFlow = val
			}
		})
	})

	return flows
}

// parseScreenerNumber parses a number from Screener.in format.
// Handles commas, percentages, and Cr/Lakh suffixes.
func parseScreenerNumber(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, ",", "", -1)
	s = strings.Replace(s, "%", "", -1)
	s = strings.Replace(s, "₹", "", -1)
	s = strings.TrimSpace(s)

	multiplier := 1.0
	if strings.HasSuffix(s, "Cr") || strings.HasSuffix(s, "Cr.") {
		s = strings.TrimSuffix(s, "Cr.")
		s = strings.TrimSuffix(s, "Cr")
		s = strings.TrimSpace(s)
		multiplier = 1e7 // 1 Crore = 10 million
	} else if strings.HasSuffix(s, "L") || strings.HasSuffix(s, "Lakh") {
		s = strings.TrimSuffix(s, "Lakh")
		s = strings.TrimSuffix(s, "L")
		s = strings.TrimSpace(s)
		multiplier = 1e5
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val * multiplier
}
