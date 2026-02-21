# OpenBB Feature Integration Plan for opense.ai

> **Goal**: Systematically integrate OpenBB's 170+ standard data models, 32 data providers, and 18 feature modules into the opense.ai Go platform — phase by phase, provider by provider.

---

## Current State (Already Implemented)

| Component | What Exists |
|---|---|
| **Data Sources** | YFinance, NSE India, NSE Derivatives, Screener.in, News (RSS), FII/DII |
| **DataSource Interface** | `GetQuote`, `GetHistoricalData`, `GetFinancials`, `GetOptionChain`, `GetStockProfile` |
| **Models** | Stock/OHLCV/Quote, FinancialData (Income/BS/CF), OptionChain, Orders, Analysis |
| **Analysis** | Technical (RSI/MACD/SMA/EMA/BB/SuperTrend/ATR/VWAP), Fundamental, Derivatives, Sentiment |
| **Brokers** | Paper, Zerodha, IBKR |
| **LLMs** | OpenAI, Ollama, Gemini, Anthropic |
| **Agents** | Orchestrator, Fundamental, Technical, Sentiment, F&O, Risk, Executor, Reporter |

---

## Architecture: Provider Pattern

Adopt OpenBB's provider pattern in Go:

```
internal/
  provider/
    provider.go          # Provider interface + registry
    standard_models.go   # Canonical data schemas (like OpenBB standard_models)
    fetcher.go           # Fetcher interface (QueryParams → Data)
  providers/
    yfinance/            # Each provider in its own package
      provider.go        # Register fetchers
      models/            # Provider-specific model mappings
      fetcher_*.go       # One fetcher per data type
    fmp/
    fred/
    ...
```

Each provider implements `Fetcher[QueryParams, Data]` and registers with a central `ProviderRegistry`. The aggregator routes requests to the appropriate provider based on user preference or auto-selection.

---

# PART 1: FEATURE PHASES

---

## Phase 1: Foundation — Provider Framework & Extended Interface

**Goal**: Build the provider abstraction layer and extend models to support all OpenBB data categories.

**Steps**:
1. Create `internal/provider/provider.go` — `Provider` interface with `Name()`, `Description()`, `Credentials()`, `Fetchers()` methods
2. Create `internal/provider/registry.go` — Global `ProviderRegistry` with `Register()`, `Get()`, `List()`, `GetFetchersFor(model)` methods
3. Create `internal/provider/fetcher.go` — Generic `Fetcher` interface: `Fetch(ctx, params) → (data, error)` with standard query/data types
4. Extend `pkg/models/` with new model files:
   - `equity_extended.go` — EquitySearch, EquityScreener, EquityInfo, MarketSnapshots, HistoricalMarketCap, ShareStatistics, EquityPeers
   - `estimates.go` — PriceTarget, PriceTargetConsensus, AnalystEstimates, ForwardEPS/EBITDA/PE/Sales
   - `calendar.go` — CalendarEarnings, CalendarDividend, CalendarIPO, CalendarSplits, CalendarEvents
   - `etf.go` — ETFSearch, ETFInfo, ETFHoldings, ETFSectors, ETFCountries, ETFPerformance
   - `index.go` — IndexHistorical, IndexConstituents, IndexSnapshots, AvailableIndices, SP500Multiples
   - `crypto.go` — CryptoHistorical, CryptoSearch
   - `currency.go` — CurrencyHistorical, CurrencyPairs, CurrencyReferenceRates, CurrencySnapshots
   - `economy.go` — GDP (Real/Nominal/Forecast), CPI, Unemployment, EconomicCalendar, EconomicIndicators
   - `fixedincome.go` — YieldCurve, TreasuryRates, SOFR, SONIA, BondIndices, BondPrices
   - `commodity.go` — CommoditySpotPrices, PetroleumStatusReport
   - `regulatory.go` — CompanyFilings, InsiderTrading, InstitutionalOwnership, GovernmentTrades, COT
   - `news_extended.go` — CompanyNews, WorldNews (structured, not just RSS)
   - `derivatives_extended.go` — FuturesCurve, FuturesHistorical, FuturesInfo, OptionsUnusual, OptionsSnapshots
5. Create `internal/provider/standard_models.go` — Enum/constants for all ~170 standard model names (matching OpenBB naming)
6. Update `internal/datasource/aggregator.go` to support provider-based routing alongside legacy sources

**Verification**: Unit tests for registry, fetcher dispatch, model serialization. `go build` passes.

---

## Phase 2: Equity — Full Coverage

**Goal**: Match OpenBB's equity extension with all sub-routers.

### Phase 2A: Equity Price & Quote
- EquityHistorical (OHLCV with provider selection)
- EquityQuote (real-time quotes with extended fields)
- PricePerformance (1D/1W/1M/3M/6M/1Y/3Y/5Y/10Y/YTD)
- EquityNBBO (National Best Bid & Offer)

### Phase 2B: Equity Fundamentals
- BalanceSheet + BalanceSheetGrowth
- IncomeStatement + IncomeStatementGrowth
- CashFlowStatement + CashFlowStatementGrowth
- FinancialRatios, KeyMetrics
- KeyExecutives, ExecutiveCompensation
- RevenueGeographic, RevenueBusinessLine
- ReportedFinancials
- HistoricalDividends, HistoricalSplits, HistoricalEps, HistoricalEmployees
- EarningsCallTranscript
- TrailingDividendYield
- ManagementDiscussionAnalysis
- EsgScore

### Phase 2C: Equity Estimates
- PriceTarget, PriceTargetConsensus
- AnalystEstimates (historical), AnalystSearch
- ForwardEpsEstimates, ForwardEbitdaEstimates
- ForwardPeEstimates, ForwardSalesEstimates

### Phase 2D: Equity Calendar
- CalendarEarnings, CalendarDividend, CalendarIPO
- CalendarSplits, CalendarEvents

### Phase 2E: Equity Discovery
- EquityGainers, EquityLosers, EquityActive
- UndervaluedLargeCaps, UndervaluedGrowth, AggressiveSmallCaps
- GrowthTechEquities, TopRetail
- DiscoveryFilings, LatestFinancialReports

### Phase 2F: Equity Ownership & Shorts
- EquityOwnership (major holders)
- InstitutionalOwnership (13-F)
- InsiderTrading, GovernmentTrades
- ShareStatistics, Form13FHR
- EquityFTD (fails to deliver)
- ShortVolume, ShortInterest

### Phase 2G: Equity Compare & Darkpool
- EquityPeers, CompareGroups, CompareCompanyFacts
- OTCAggregate (dark pool data)

### Phase 2H: Equity Search & Screener
- EquitySearch (symbol/name/CIK/LEI)
- EquityScreener (multi-criteria filtering)
- MarketSnapshots, HistoricalMarketCap

**Verification**: Integration tests hitting at least 2 providers per endpoint. Agent tool registration for fundamental/technical agents.

---

## Phase 3: Derivatives — Options & Futures

**Goal**: Full derivatives support matching OpenBB's derivatives extension.

### Phase 3A: Options
- OptionsChains (extend existing NSE implementation + add multi-provider)
- OptionsUnusual (unusual activity detection)
- OptionsSnapshots (market-wide options snapshot)
- Options volatility surface computation (POST-style, from chains data)

### Phase 3B: Futures
- FuturesHistorical (historical futures prices)
- FuturesCurve (term structure / forward curve)
- FuturesInstruments (available contracts)
- FuturesInfo (current trading statistics)

**Verification**: F&O agent updated with new tools. Test with CBOE, Deribit, YFinance, Tradier providers.

---

## Phase 4: ETF & Index

**Goal**: Full ETF and index data coverage.

### Phase 4A: ETF
- EtfSearch, EtfHistorical, EtfInfo
- EtfHoldings, EtfSectors, EtfCountries
- EtfPricePerformance, EtfEquityExposure
- NportDisclosure (SEC N-PORT filings)
- ETF Discovery: Gainers, Losers, Active

### Phase 4B: Index
- IndexHistorical, IndexConstituents
- IndexSnapshots, IndexSectors
- AvailableIndices, IndexSearch
- SP500Multiples (Shiller PE, earnings yield, etc.)

**Verification**: Tests with YFinance, FMP, CBOE, TMX, WSJ providers.

---

## Phase 5: Economy & Macroeconomics

**Goal**: Comprehensive macroeconomic data matching OpenBB's economy extension.

### Phase 5A: Core Economic Indicators
- EconomicCalendar (global events)
- ConsumerPriceIndex (CPI by country)
- Unemployment (global)
- GDP: Real, Nominal, Forecast
- CompositeLeadingIndicator (CLI)

### Phase 5B: FRED Integration
- FredSearch, FredSeries, FredReleaseTable
- FredRegional (geo-mapped data)
- PersonalConsumptionExpenditures (PCE)
- NonFarmPayrolls
- RetailPrices

### Phase 5C: Economic Surveys
- UniversityOfMichigan (consumer sentiment)
- SeniorLoanOfficerSurvey (SLOOS)
- ManufacturingOutlookNY (Empire State)
- ManufacturingOutlookTexas
- SurveyOfEconomicConditionsChicago
- BLS Search + Series

### Phase 5D: International Economics
- CountryProfile, AvailableIndicators, EconomicIndicators
- BalanceOfPayments, DirectionOfTrade
- ExportDestinations
- MoneyMeasures, CentralBankHoldings
- RiskPremium, SharePriceIndex, HousePriceIndex
- CountryInterestRates

### Phase 5E: FOMC & Federal Reserve
- FomcDocuments
- PrimaryDealerPositioning, PrimaryDealerFails
- TotalFactorProductivity, InflationExpectations

### Phase 5F: Shipping & Maritime
- PortInfo, PortVolume
- MaritimeChokePointInfo, MaritimeChokePointVolume

**Verification**: Tests with FRED, OECD, IMF, EconDB, BLS providers.

---

## Phase 6: Fixed Income & Rates

**Goal**: Full fixed income coverage matching OpenBB's fixedincome extension.

### Phase 6A: Interest Rates
- SOFR, SONIA, Ameribor
- FederalFundsRate + Projections
- EuroShortTermRate, ECB Interest Rates
- IORB, DiscountWindowPrimaryCreditRate
- OvernightBankFundingRate

### Phase 6B: Government Bonds
- YieldCurve (multi-country)
- TreasuryRates, TreasuryAuctions, TreasuryPrices
- TipsYields
- SvenssonYieldCurve

### Phase 6C: Spreads
- TreasuryConstantMaturity (10Y minus selected)
- SelectedTreasuryConstantMaturity (maturity minus EFFR)
- SelectedTreasuryBill (T-Bill minus EFFR)

### Phase 6D: Corporate Bonds
- HighQualityMarketCorporateBond (HQM)
- SpotRate (zero coupon)
- CommercialPaper
- BondPrices, BondIndices

**Verification**: Tests with FRED, Federal Reserve, ECB, TMX, Government US providers.

---

## Phase 7: Crypto, Currency & Commodity

### Phase 7A: Cryptocurrency
- CryptoSearch, CryptoHistorical
- (from YFinance, FMP, Tiingo providers)

### Phase 7B: Currency / Forex
- CurrencyPairs (search available pairs)
- CurrencyHistorical (OHLCV for FX)
- CurrencyReferenceRates (ECB official rates)
- CurrencySnapshots

### Phase 7C: Commodities
- CommoditySpotPrices (FRED)
- PetroleumStatusReport (EIA weekly)
- ShortTermEnergyOutlook (EIA STEO)
- CommodityPsdData, CommodityPsdReport (USDA)
- WeatherBulletin, WeatherBulletinDownload (USDA)

**Verification**: Tests with YFinance, FMP, Tiingo, ECB, EIA, Government US, FRED providers.

---

## Phase 8: News, Regulators & Advanced Analytics

### Phase 8A: Structured News
- CompanyNews (multi-provider: YFinance, FMP, Benzinga, Tiingo, Intrinio, TMX)
- WorldNews (multi-provider: FMP, Benzinga, Tiingo, Intrinio, Biztoc)
- Replace/augment current RSS-based news with structured API news

### Phase 8B: Regulators — SEC
- CompanyFilings, SecFiling (raw filing access)
- CikMap, SymbolMap, SicSearch
- InstitutionsSearch, SchemaFiles
- RssLitigation, SecHtmFile
- Form13FHR, NportDisclosure
- CompareCompanyFacts, LatestFinancialReports
- ManagementDiscussionAnalysis

### Phase 8C: Regulators — CFTC
- COT (Commitments of Traders data)
- COTSearch

### Phase 8D: Technical Analysis (OpenBB-style)
- Extend existing technical indicators with OpenBB's full set:
  - Add: Fibonacci, OBV, Fisher Transform, Aroon, Demark Sequential, Donchian Channels, Ichimoku Cloud, Clenow Momentum, ADX, CCI, Stochastic, Keltner Channels, Center of Gravity, Volatility Cones, AD/ADOSC, ZLMA, HMA, WMA, Relative Rotation
- Implement as POST-style analysis (operate on passed-in data)

### Phase 8E: Quantitative Analysis
- Summary statistics, Normality tests, Unit root tests (ADF/KPSS)
- CAPM
- Rolling: skew, variance, stdev, kurtosis, quantile, mean
- Performance: Omega ratio, Sharpe ratio, Sortino ratio

### Phase 8F: Econometrics & Factor Models
- Fama-French factors, breakpoints, portfolio returns
- Regional/country/international index returns
- US Congress: Bills, BillInfo, BillText

**Verification**: End-to-end tests, agent tool integration, API endpoint tests.

---

# PART 2: PROVIDER PHASES (Implementation Priority)

---

## Provider Tier 1: Free Core Providers (No API Key, Broad Coverage)

> **Priority**: Highest — these are the backbone, free, and cover the most endpoints.

### P1.1: YFinance (Yahoo Finance) — 29 endpoints ✅ PARTIALLY EXISTS
- **Status**: `GetQuote` and `GetHistoricalData` already implemented
- **Remaining**:
  - BalanceSheet, IncomeStatement, CashFlowStatement
  - CryptoHistorical, CurrencyHistorical
  - EquityActive, EquityGainers, EquityLosers
  - EquityInfo (profile), EquityScreener, EquityQuote (extended)
  - UndervaluedGrowth, UndervaluedLargeCaps, AggressiveSmallCaps, GrowthTechEquities
  - EtfHistorical, EtfInfo
  - FuturesCurve, FuturesHistorical
  - HistoricalDividends, KeyExecutives, KeyMetrics
  - IndexHistorical, AvailableIndices
  - OptionsChains (extend beyond NSE)
  - PriceTargetConsensus, ShareStatistics
  - CompanyNews

### P1.2: CBOE (Chicago Board Options Exchange) — 11 endpoints
- **No API key needed**
- **Endpoints**: AvailableIndices, EquityHistorical, EquityQuote, EquitySearch, EtfHistorical, FuturesCurve, IndexConstituents, IndexHistorical, IndexSearch, IndexSnapshots, OptionsChains
- **Value**: Authoritative options/derivatives data, index data

### P1.3: SEC (Securities & Exchange Commission) — 18 endpoints
- **No API key needed**
- **Endpoints**: CikMap, CompanyFilings, CompareCompanyFacts, EquityFTD, EquitySearch, Form13FHR, InsiderTrading, InstitutionsSearch, LatestFinancialReports, ManagementDiscussionAnalysis, NportDisclosure, RssLitigation, SchemaFiles, SecFiling, SecHtmFile, SicSearch, SymbolMap
- **Value**: Authoritative regulatory/filing data, cannot get elsewhere

### P1.4: TMX (Toronto Stock Exchange) — 24 endpoints
- **No API key needed**
- **Endpoints**: AvailableIndices, BondPrices, CalendarEarnings, CompanyFilings, CompanyNews, EquityHistorical/Info/Quote/Search, EtfSearch/Holdings/Sectors/Countries/Historical/Info, EquityGainers, HistoricalDividends, IndexConstituents/Sectors/Snapshots, InsiderTrading, OptionsChains, PriceTargetConsensus, TreasuryPrices
- **Value**: Canadian market coverage, 24 free endpoints

---

## Provider Tier 2: Free Specialized Providers (No API Key, Focused)

### P2.1: Federal Reserve — 13 endpoints
- **Endpoints**: CentralBankHoldings, FederalFundsRate, FomcDocuments, InflationExpectations, MoneyMeasures, OvernightBankFundingRate, PrimaryDealerFails/Positioning, SOFR, SvenssonYieldCurve, TFP, TreasuryRates, YieldCurve
- **Value**: Authoritative US monetary policy data

### P2.2: OECD — 9 endpoints
- **Endpoints**: CLI, CPI, CountryInterestRates, GdpNominal/Real/Forecast, HousePriceIndex, SharePriceIndex, Unemployment
- **Value**: International macro data for OECD countries

### P2.3: IMF — 8 endpoints
- **Endpoints**: AvailableIndicators, CPI, DirectionOfTrade, EconomicIndicators, MaritimeChokePointInfo/Volume, PortInfo/Volume
- **Value**: Global economic data, trade flows, shipping

### P2.4: ECB (European Central Bank) — 3 endpoints
- **Endpoints**: BalanceOfPayments, CurrencyReferenceRates, YieldCurve
- **Value**: Official EUR reference rates, European yield curves

### P2.5: Finviz — 7 endpoints
- **Endpoints**: CompareGroups, EquityInfo, EquityScreener, EtfPricePerformance, KeyMetrics, PricePerformance, PriceTarget
- **Value**: Popular free screener, visual market maps data

### P2.6: FINRA — 2 endpoints
- **Endpoints**: OTCAggregate, EquityShortInterest
- **Value**: Authoritative short interest data

### P2.7: Stockgrid — 1 endpoint
- **Endpoint**: ShortVolume
- **Value**: Dark pool / short volume data

### P2.8: Multpl — 1 endpoint
- **Endpoint**: SP500Multiples
- **Value**: Historical S&P 500 valuation multiples, Shiller PE

### P2.9: WSJ — 3 endpoints
- **Endpoints**: ETFGainers, ETFLosers, ETFActive
- **Value**: ETF market movers

### P2.10: Seeking Alpha — 3 endpoints
- **Endpoints**: CalendarEarnings, ForwardEpsEstimates, ForwardSalesEstimates
- **Value**: Earnings calendar, forward estimates

### P2.11: Deribit — 5 endpoints
- **Endpoints**: FuturesCurve, FuturesHistorical, FuturesInfo, FuturesInstruments, OptionsChains
- **Value**: Crypto derivatives (BTC/ETH options and futures)

### P2.12: Government US (Data.gov) — 6 endpoints
- **Endpoints**: CommodityPsdData/Report, TreasuryAuctions/Prices, WeatherBulletin/Download
- **Value**: Treasury auction data, USDA commodity data

---

## Provider Tier 3: Free with API Key (Easy Registration)

### P3.1: FMP (Financial Modeling Prep) — 69 endpoints ⭐ HIGHEST PRIORITY
- **API Key**: Free tier available
- **Categories**: Equity (prices/quotes/screener/discovery), Fundamentals (all statements + growth), ETF (full), Crypto, Forex, Calendar (earnings/dividends/IPO/splits/events), Estimates, Index, News, ESG, Government Trades, Treasury, Yield Curves
- **Value**: Single provider covers the most ground — 69 endpoints, free tier

### P3.2: FRED (Federal Reserve Economic Data) — 35 endpoints
- **API Key**: Free registration
- **Categories**: Interest rates (SOFR/SONIA/EFFR/Ameribor), CPI, Manufacturing surveys, Mortgage indices, Non-farm payrolls, PCE, Yield curves, Bond indices, Commodity spot prices, University of Michigan, Treasury data
- **Value**: 816K+ economic time series, gold standard for US economic data

### P3.3: Tiingo — 7 endpoints
- **API Key**: Free token
- **Categories**: Equity/ETF historical, Crypto, Forex, News, Dividends
- **Value**: Enterprise-grade market data, good free tier

### P3.4: Alpha Vantage — 3 endpoints
- **API Key**: Free
- **Categories**: Equity/ETF historical, Historical EPS
- **Value**: Well-known free market data API

### P3.5: Nasdaq — 9 endpoints
- **API Key**: Free
- **Categories**: Calendar (dividends/earnings/IPO), Equity search/screener, Economic calendar, Historical dividends, TopRetail
- **Value**: Authoritative calendar data

### P3.6: EconDB — 8 endpoints
- **API Key**: Optional (uses temp token if absent)
- **Categories**: Economic indicators, GDP, Yield curves, Trade, Port volume, Country profiles
- **Value**: Macro data processing layer

### P3.7: EIA (Energy Information Administration) — 2 endpoints
- **API Key**: Free
- **Categories**: Petroleum Status Report, Short-Term Energy Outlook
- **Value**: Authoritative US energy data

### P3.8: BLS (Bureau of Labor Statistics) — 2 endpoints
- **API Key**: Free
- **Categories**: Labor/employment series search + data
- **Value**: Authoritative US employment data

### P3.9: CFTC — 2 endpoints
- **API Key**: Optional (app_token)
- **Categories**: COT reports, COT search
- **Value**: Commitment of Traders data

### P3.10: Congress.gov — 3 endpoints
- **API Key**: Free
- **Categories**: Congressional bills, bill info, bill text
- **Value**: Legislative data for policy-sensitive analysis

---

## Provider Tier 4: Paid / Premium Providers

### P4.1: Intrinio — 38 endpoints
- **API Key**: Paid subscription
- **Categories**: Equity, ETF, Options (chains/snapshots/unusual), Fundamentals, Forward estimates (EPS/EBITDA/PE/Sales), News, Index, Insider trading, FRED series, Market snapshots
- **Value**: Enterprise-grade, excellent options data, unique attributes API

### P4.2: Benzinga — 4 endpoints
- **API Key**: Paid
- **Categories**: Company news, World news, Analyst search, Price targets
- **Value**: High-quality financial news, analyst research

### P4.3: Trading Economics — 1 endpoint
- **API Key**: Paid
- **Categories**: Economic calendar (196 countries, 20M+ indicators)
- **Value**: Most comprehensive international economic calendar

### P4.4: Biztoc — 1 endpoint
- **API Key**: Paid (RapidAPI)
- **Categories**: World news aggregation
- **Value**: Alternative news source

---

## Provider Tier 5: Regional / Niche Providers

### P5.1: Tradier — 5 endpoints
- **API Key**: Required (account_type)
- **Categories**: Equity historical/quote/search, ETF, Options chains
- **Value**: Brokerage-integrated data, good options chains

### P5.2: Famafrench — 6 endpoints
- **No API Key**
- **Categories**: Factor breakpoints, Country/Regional/US portfolio returns, International index returns, Factor data (SMB, HML, etc.)
- **Value**: Academic factor models for quantitative analysis

---

## Provider Tier 6: India-Specific (Already Partially Implemented)

### P6.1: NSE India ✅ EXISTS
- **Status**: GetQuote, GetHistoricalData, GetShareholding, GetStockProfile implemented
- **Enhance**: Add NSE-specific endpoints not in OpenBB (NSE is not an OpenBB provider)

### P6.2: NSE Derivatives ✅ EXISTS
- **Status**: GetOptionChain, GetIndiaVIX, GetFIIDIIData, GetFuturesData implemented
- **Enhance**: OI buildup analysis, multi-expiry chains

### P6.3: Screener.in ✅ EXISTS
- **Status**: GetFinancials, GetFinancialRatios, GetPeerComparison implemented

### P6.4: Indian News ✅ EXISTS
- **Status**: RSS-based news from Moneycontrol, ET, LiveMint, Business Standard

---

# PART 3: IMPLEMENTATION ROADMAP

---

## Sprint 1 (Weeks 1-2): Foundation
- [ ] Phase 1: Provider framework, registry, fetcher interface
- [ ] Extended models for all data categories
- [ ] Provider auto-selection logic in aggregator

## Sprint 2 (Weeks 3-4): Core Free Providers
- [ ] P1.1: YFinance (complete remaining 25+ endpoints)
- [ ] P3.1: FMP (start with equity price/quote/fundamental — ~30 endpoints)

## Sprint 3 (Weeks 5-6): Equity Complete
- [ ] Phase 2A-2D: Equity price, fundamentals, estimates, calendar
- [ ] P1.3: SEC provider (filings, insider trading)
- [ ] P3.2: FRED provider (start with key rates, CPI)

## Sprint 4 (Weeks 7-8): Derivatives & ETF
- [ ] Phase 3: Options & Futures (CBOE, Deribit providers)
- [ ] Phase 4: ETF & Index
- [ ] P1.2: CBOE provider

## Sprint 5 (Weeks 9-10): Economy & Fixed Income
- [ ] Phase 5A-5B: Core economic indicators + FRED
- [ ] Phase 6: Fixed income & rates
- [ ] P2.1: Federal Reserve, P2.2: OECD, P2.3: IMF

## Sprint 6 (Weeks 11-12): Remaining Features
- [ ] Phase 7: Crypto, Currency, Commodity
- [ ] Phase 8A-8C: News, Regulators
- [ ] P2.4-P2.12: Remaining free specialized providers

## Sprint 7 (Weeks 13-14): Analytics & Polish
- [ ] Phase 8D: Extended technical analysis
- [ ] Phase 8E: Quantitative analysis
- [ ] Phase 8F: Econometrics & factor models
- [ ] P3.3-P3.10: Remaining API-key providers

## Sprint 8 (Weeks 15-16): Premium & Agent Integration
- [ ] P4.1-P4.4: Premium providers (Intrinio, Benzinga, etc.)
- [ ] P5.1-P5.2: Niche providers
- [ ] Full agent tool registration for all new data
- [ ] API endpoints for all new features
- [ ] Documentation & integration tests

---

# PART 4: PROVIDER DETAIL — ENDPOINT REFERENCE

---

## YFinance (29 endpoints)
| # | Standard Model | Description |
|---|---|---|
| 1 | AvailableIndices | List all available indices |
| 2 | BalanceSheet | Balance sheet data |
| 3 | CashFlowStatement | Cash flow statement |
| 4 | CompanyNews | Company-specific news |
| 5 | CryptoHistorical | Crypto OHLCV data |
| 6 | CurrencyHistorical | Forex OHLCV data |
| 7 | EquityActive | Most active stocks |
| 8 | EquityAggressiveSmallCaps | Top small caps |
| 9 | EquityGainers | Top gainers |
| 10 | EquityHistorical | Stock OHLCV data |
| 11 | EquityInfo | Company profile |
| 12 | EquityLosers | Top losers |
| 13 | EquityQuote | Real-time quote |
| 14 | EquityScreener | Stock screener |
| 15 | EquityUndervaluedGrowth | Undervalued growth stocks |
| 16 | EquityUndervaluedLargeCaps | Undervalued large caps |
| 17 | EtfHistorical | ETF OHLCV |
| 18 | EtfInfo | ETF information |
| 19 | FuturesCurve | Futures term structure |
| 20 | FuturesHistorical | Futures OHLCV |
| 21 | GrowthTechEquities | Top tech growth |
| 22 | HistoricalDividends | Dividend history |
| 23 | IncomeStatement | Income statement |
| 24 | IndexHistorical | Index OHLCV |
| 25 | KeyExecutives | Management team |
| 26 | KeyMetrics | Fundamental metrics |
| 27 | OptionsChains | Full options chain |
| 28 | PriceTargetConsensus | Analyst consensus |
| 29 | ShareStatistics | Float & share stats |

## FMP (69 endpoints)
| # | Standard Model | Description |
|---|---|---|
| 1 | AnalystEstimates | Historical analyst estimates |
| 2 | AvailableIndices | Available indices |
| 3 | BalanceSheet | Balance sheet |
| 4 | BalanceSheetGrowth | Balance sheet growth |
| 5 | CalendarDividend | Dividend calendar |
| 6 | CalendarEarnings | Earnings calendar |
| 7 | CalendarEvents | Events calendar |
| 8 | CalendarIpo | IPO calendar |
| 9 | CalendarSplits | Splits calendar |
| 10 | CashFlowStatement | Cash flow statement |
| 11 | CashFlowStatementGrowth | Cash flow growth |
| 12 | CompanyFilings | SEC filings |
| 13 | CompanyNews | Company news |
| 14 | CryptoHistorical | Crypto OHLCV |
| 15 | CryptoSearch | Search crypto pairs |
| 16 | CurrencyHistorical | Forex OHLCV |
| 17 | CurrencyPairs | Available FX pairs |
| 18 | CurrencySnapshots | FX snapshots |
| 19 | DiscoveryFilings | New SEC filings |
| 20 | EarningsCallTranscript | Earnings transcripts |
| 21 | EconomicCalendar | Economic events |
| 22 | EquityActive | Most active |
| 23 | EquityHistorical | Stock OHLCV |
| 24 | EquityOwnership | Major holders |
| 25 | EquityPeers | Peer companies |
| 26 | EquityInfo | Company profile |
| 27 | EquityGainers | Top gainers |
| 28 | EquityLosers | Top losers |
| 29 | EquityQuote | Real-time quote |
| 30 | EquityScreener | Stock screener |
| 31 | EsgScore | ESG ratings |
| 32 | EtfCountries | ETF country weights |
| 33 | EtfEquityExposure | ETF stock exposure |
| 34 | EtfHoldings | ETF holdings |
| 35 | EtfInfo | ETF overview |
| 36 | EtfPricePerformance | ETF performance |
| 37 | EtfSearch | Search ETFs |
| 38 | EtfSectors | ETF sector weights |
| 39 | EtfHistorical | ETF OHLCV |
| 40 | ExecutiveCompensation | Executive comp |
| 41 | FinancialRatios | Financial ratios |
| 42 | ForwardEbitdaEstimates | Forward EBITDA |
| 43 | ForwardEpsEstimates | Forward EPS |
| 44 | GovernmentTrades | Congress trades |
| 45 | HistoricalDividends | Dividend history |
| 46 | HistoricalEmployees | Employee count |
| 47 | HistoricalEps | EPS history |
| 48 | HistoricalMarketCap | Market cap history |
| 49 | HistoricalSplits | Split history |
| 50 | IncomeStatement | Income statement |
| 51 | IncomeStatementGrowth | Income growth |
| 52 | IndexConstituents | Index members |
| 53 | IndexHistorical | Index OHLCV |
| 54 | InsiderTrading | Insider trades |
| 55 | InstitutionalOwnership | 13-F holdings |
| 56 | KeyExecutives | Management team |
| 57 | KeyMetrics | Key metrics |
| 58 | MarketSnapshots | Market snapshot |
| 59 | NportDisclosure | N-PORT filings |
| 60 | PricePerformance | Price returns |
| 61 | PriceTarget | Analyst targets |
| 62 | PriceTargetConsensus | Target consensus |
| 63 | RevenueBusinessLine | Rev by segment |
| 64 | RevenueGeographic | Rev by geography |
| 65 | RiskPremium | Market risk premium |
| 66 | ShareStatistics | Share stats |
| 67 | TreasuryRates | Treasury rates |
| 68 | WorldNews | Global news |
| 69 | YieldCurve | Yield curve |

## FRED (35 endpoints)
| # | Standard Model | Description |
|---|---|---|
| 1 | BalanceOfPayments | Balance of payments |
| 2 | BondIndices | Bond index data |
| 3 | CommoditySpotPrices | Commodity spot prices |
| 4 | ConsumerPriceIndex | CPI data |
| 5 | SOFR | Secured overnight rate |
| 6 | EuroShortTermRate | Euro short-term rate |
| 7 | SONIA | Sterling overnight rate |
| 8 | Ameribor | AMERIBOR rate |
| 9 | FederalFundsRate | Fed funds rate |
| 10 | PROJECTIONS | Fed rate projections |
| 11 | IORB | Interest on reserves |
| 12 | DiscountWindowPrimaryCreditRate | Discount window rate |
| 13 | EuropeanCentralBankInterestRates | ECB rates |
| 14 | ManufacturingOutlookNY | Empire State |
| 15 | ManufacturingOutlookTexas | Texas manufacturing |
| 16 | MortgageIndices | Mortgage rates |
| 17 | NonFarmPayrolls | NFP data |
| 18 | OvernightBankFundingRate | OBFR |
| 19 | PersonalConsumptionExpenditures | PCE |
| 20 | CommercialPaper | CP rates |
| 21 | FredReleaseTable | Release data |
| 22 | FredSearch | Series search |
| 23 | FredSeries | Series data |
| 24 | FredRegional | Regional data |
| 25 | RetailPrices | Retail prices |
| 26 | SeniorLoanOfficerSurvey | SLOOS |
| 27 | SpotRate | Spot rate |
| 28 | HighQualityMarketCorporateBond | HQM bond |
| 29 | TreasuryConstantMaturity | TCM |
| 30 | SelectedTreasuryConstantMaturity | Selected TCM |
| 31 | SelectedTreasuryBill | Selected T-Bill |
| 32 | SurveyOfEconomicConditionsChicago | Chicago survey |
| 33 | TipsYields | TIPS yields |
| 34 | UniversityOfMichigan | UMich sentiment |
| 35 | YieldCurve | Yield curve |

## Intrinio (38 endpoints)
| # | Standard Model | Description |
|---|---|---|
| 1 | BalanceSheet | Balance sheet |
| 2 | CalendarIpo | IPO calendar |
| 3 | CashFlowStatement | Cash flow |
| 4 | CompanyFilings | Company filings |
| 5 | CompanyNews | Company news |
| 6 | CurrencyPairs | FX pairs |
| 7 | EquityHistorical | Stock OHLCV |
| 8 | EquityInfo | Company info |
| 9 | EquityQuote | Quote |
| 10 | EquitySearch | Search |
| 11 | EtfHistorical | ETF OHLCV |
| 12 | EtfHoldings | ETF holdings |
| 13 | EtfInfo | ETF info |
| 14 | EtfPricePerformance | ETF perf |
| 15 | EtfSearch | ETF search |
| 16 | FinancialRatios | Ratios |
| 17 | ForwardEbitdaEstimates | Fwd EBITDA |
| 18 | ForwardEpsEstimates | Fwd EPS |
| 19 | ForwardPeEstimates | Fwd PE |
| 20 | ForwardSalesEstimates | Fwd Sales |
| 21 | FredSeries | FRED data |
| 22 | HistoricalAttributes | Hist attributes |
| 23 | HistoricalDividends | Dividends |
| 24 | HistoricalMarketCap | Market cap |
| 25 | IncomeStatement | Income stmt |
| 26 | IndexHistorical | Index OHLCV |
| 27 | InsiderTrading | Insider trades |
| 28 | KeyMetrics | Key metrics |
| 29 | LatestAttributes | Latest attrs |
| 30 | MarketSnapshots | Snapshot |
| 31 | OptionsChains | Options chain |
| 32 | OptionsSnapshots | Options mkt snapshot |
| 33 | OptionsUnusual | Unusual options |
| 34 | PriceTargetConsensus | Target consensus |
| 35 | ReportedFinancials | As-reported |
| 36 | SearchAttributes | Search data tags |
| 37 | ShareStatistics | Share stats |
| 38 | WorldNews | Global news |

## SEC (18 endpoints)
| # | Standard Model | Description |
|---|---|---|
| 1 | CikMap | Ticker → CIK mapping |
| 2 | CompanyFilings | Company filings |
| 3 | CompareCompanyFacts | Compare reported facts |
| 4 | EquityFTD | Fails to deliver |
| 5 | EquitySearch | Search companies |
| 6 | Filings | Filing access |
| 7 | Form13FHR | 13-F filing data |
| 8 | SecHtmFile | Raw HTML filing |
| 9 | InsiderTrading | Insider trades |
| 10 | InstitutionsSearch | Search institutions |
| 11 | LatestFinancialReports | Newest reports |
| 12 | ManagementDiscussionAnalysis | MD&A section |
| 13 | NportDisclosure | N-PORT data |
| 14 | RssLitigation | Litigation feed |
| 15 | SchemaFiles | XBRL schemas |
| 16 | SecFiling | Filing index |
| 17 | SicSearch | SIC code search |
| 18 | SymbolMap | CIK → Ticker |

## CBOE (11 endpoints)
| # | Standard Model | Description |
|---|---|---|
| 1 | AvailableIndices | Available indices |
| 2 | EquityHistorical | Stock OHLCV |
| 3 | EquityQuote | Stock quote |
| 4 | EquitySearch | Search stocks |
| 5 | EtfHistorical | ETF OHLCV |
| 6 | FuturesCurve | Futures curve |
| 7 | IndexConstituents | Index members |
| 8 | IndexHistorical | Index OHLCV |
| 9 | IndexSearch | Search indices |
| 10 | IndexSnapshots | Index snapshots |
| 11 | OptionsChains | Options chain |

## TMX (24 endpoints)
| # | Standard Model | Description |
|---|---|---|
| 1 | AvailableIndices | Canadian indices |
| 2 | BondPrices | Bond prices |
| 3 | CalendarEarnings | Earnings cal |
| 4 | CompanyFilings | Filings |
| 5 | CompanyNews | News |
| 6 | EquityHistorical | Stock OHLCV |
| 7 | EquityInfo | Profile |
| 8 | EquityQuote | Quote |
| 9 | EquitySearch | Search |
| 10 | EtfSearch | ETF search |
| 11 | EtfHoldings | ETF holdings |
| 12 | EtfSectors | ETF sectors |
| 13 | EtfCountries | ETF countries |
| 14 | EtfHistorical | ETF OHLCV |
| 15 | EtfInfo | ETF info |
| 16 | EquityGainers | Top gainers |
| 17 | HistoricalDividends | Dividends |
| 18 | IndexConstituents | Index members |
| 19 | IndexSectors | Index sectors |
| 20 | IndexSnapshots | Index levels |
| 21 | InsiderTrading | Insider trades |
| 22 | OptionsChains | Options chain |
| 23 | PriceTargetConsensus | Target consensus |
| 24 | TreasuryPrices | Treasury prices |

## Federal Reserve (13 endpoints)
| # | Standard Model | Description |
|---|---|---|
| 1 | CentralBankHoldings | Fed holdings |
| 2 | FederalFundsRate | Fed funds rate |
| 3 | FomcDocuments | FOMC docs |
| 4 | InflationExpectations | Inflation exp |
| 5 | MoneyMeasures | M1/M2 |
| 6 | OvernightBankFundingRate | OBFR |
| 7 | PrimaryDealerFails | Dealer fails |
| 8 | PrimaryDealerPositioning | Dealer positions |
| 9 | SOFR | SOFR rate |
| 10 | SvenssonYieldCurve | Svensson curve |
| 11 | TotalFactorProductivity | TFP |
| 12 | TreasuryRates | Treasury rates |
| 13 | YieldCurve | Yield curve |

## OECD (9 endpoints)
| # | Standard Model | Description |
|---|---|---|
| 1 | CompositeLeadingIndicator | CLI |
| 2 | ConsumerPriceIndex | CPI |
| 3 | CountryInterestRates | Interest rates |
| 4 | GdpNominal | Nominal GDP |
| 5 | GdpReal | Real GDP |
| 6 | GdpForecast | GDP forecast |
| 7 | HousePriceIndex | House prices |
| 8 | SharePriceIndex | Share prices |
| 9 | Unemployment | Unemployment |

## IMF (8 endpoints)
| # | Standard Model | Description |
|---|---|---|
| 1 | AvailableIndicators | Available indicators |
| 2 | ConsumerPriceIndex | CPI |
| 3 | DirectionOfTrade | Trade flows |
| 4 | EconomicIndicators | Economic data |
| 5 | MaritimeChokePointInfo | Chokepoint info |
| 6 | MaritimeChokePointVolume | Chokepoint volume |
| 7 | PortInfo | Port metadata |
| 8 | PortVolume | Port trade volume |

## Remaining Providers (Summary)
| Provider | Endpoints | Key Data |
|---|---|---|
| Tiingo | 7 | Equity/Crypto/FX historical, News, Dividends |
| Finviz | 7 | Screener, Key metrics, Price performance/targets |
| EconDB | 8 | GDP, Yield curves, Economic indicators, Trade |
| Nasdaq | 9 | Calendar data, Equity search/screener, TopRetail |
| FamaFrench | 6 | Factor models, Portfolio returns |
| Government US | 6 | Treasury auctions/prices, Commodity PSD, Weather |
| Tradier | 5 | Equity data, Options chains |
| Deribit | 5 | Crypto futures/options |
| Benzinga | 4 | News, Analyst search, Price targets |
| Alpha Vantage | 3 | Equity/ETF historical, EPS |
| ECB | 3 | FX reference rates, Yield curve, BoP |
| Seeking Alpha | 3 | Earnings calendar, Forward estimates |
| WSJ | 3 | ETF movers |
| Congress.gov | 3 | Legislative bills |
| BLS | 2 | Labor statistics |
| CFTC | 2 | COT reports |
| EIA | 2 | Petroleum, Energy outlook |
| FINRA | 2 | Short interest, OTC |
| Trading Economics | 1 | Economic calendar (196 countries) |
| Biztoc | 1 | News aggregation |
| Stockgrid | 1 | Short volume |
| Multpl | 1 | S&P 500 multiples |

---

## Summary Statistics

| Metric | Count |
|---|---|
| **Total Providers** | 32 |
| **Total Endpoints (all providers)** | ~350+ |
| **Feature Phases** | 8 (with 25+ sub-phases) |
| **Provider Tiers** | 6 |
| **Standard Models** | ~170 |
| **Extensions/Modules** | 18 |
| **Free Providers (no key)** | 16 |
| **Free with API Key** | 10 |
| **Paid Providers** | 4 |
| **Existing (India-specific)** | 4 (NSE, NSE Derivatives, Screener.in, News RSS) |
