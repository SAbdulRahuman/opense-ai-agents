package provider

// ModelType represents a standard data model type, matching OpenBB's standard_models.
// Each ModelType maps to a specific data structure in pkg/models/.
type ModelType string

// --- Equity / Price ---
const (
	ModelEquityHistorical     ModelType = "EquityHistorical"
	ModelEquityQuote          ModelType = "EquityQuote"
	ModelEquityInfo           ModelType = "EquityInfo"
	ModelEquitySearch         ModelType = "EquitySearch"
	ModelEquityScreener       ModelType = "EquityScreener"
	ModelEquityNBBO           ModelType = "EquityNBBO"
	ModelEquityPeers          ModelType = "EquityPeers"
	ModelMarketSnapshots      ModelType = "MarketSnapshots"
	ModelHistoricalMarketCap  ModelType = "HistoricalMarketCap"
	ModelPricePerformance     ModelType = "PricePerformance"
)

// --- Equity / Fundamentals ---
const (
	ModelBalanceSheet              ModelType = "BalanceSheet"
	ModelBalanceSheetGrowth        ModelType = "BalanceSheetGrowth"
	ModelIncomeStatement           ModelType = "IncomeStatement"
	ModelIncomeStatementGrowth     ModelType = "IncomeStatementGrowth"
	ModelCashFlowStatement         ModelType = "CashFlowStatement"
	ModelCashFlowStatementGrowth   ModelType = "CashFlowStatementGrowth"
	ModelFinancialRatios           ModelType = "FinancialRatios"
	ModelKeyMetrics                ModelType = "KeyMetrics"
	ModelKeyExecutives             ModelType = "KeyExecutives"
	ModelExecutiveCompensation     ModelType = "ExecutiveCompensation"
	ModelRevenueGeographic         ModelType = "RevenueGeographic"
	ModelRevenueBusinessLine       ModelType = "RevenueBusinessLine"
	ModelReportedFinancials        ModelType = "ReportedFinancials"
	ModelHistoricalDividends       ModelType = "HistoricalDividends"
	ModelHistoricalSplits          ModelType = "HistoricalSplits"
	ModelHistoricalEps             ModelType = "HistoricalEps"
	ModelHistoricalEmployees       ModelType = "HistoricalEmployees"
	ModelEarningsCallTranscript    ModelType = "EarningsCallTranscript"
	ModelTrailingDividendYield     ModelType = "TrailingDividendYield"
	ModelManagementDiscussionAnalysis ModelType = "ManagementDiscussionAnalysis"
	ModelEsgScore                  ModelType = "EsgScore"
	ModelShareStatistics           ModelType = "ShareStatistics"
)

// --- Equity / Estimates ---
const (
	ModelPriceTarget          ModelType = "PriceTarget"
	ModelPriceTargetConsensus ModelType = "PriceTargetConsensus"
	ModelAnalystEstimates     ModelType = "AnalystEstimates"
	ModelAnalystSearch        ModelType = "AnalystSearch"
	ModelForwardEpsEstimates    ModelType = "ForwardEpsEstimates"
	ModelForwardEbitdaEstimates ModelType = "ForwardEbitdaEstimates"
	ModelForwardPeEstimates     ModelType = "ForwardPeEstimates"
	ModelForwardSalesEstimates  ModelType = "ForwardSalesEstimates"
)

// --- Equity / Calendar ---
const (
	ModelCalendarEarnings ModelType = "CalendarEarnings"
	ModelCalendarDividend ModelType = "CalendarDividend"
	ModelCalendarIpo      ModelType = "CalendarIpo"
	ModelCalendarSplits   ModelType = "CalendarSplits"
	ModelCalendarEvents   ModelType = "CalendarEvents"
)

// --- Equity / Discovery ---
const (
	ModelEquityGainers             ModelType = "EquityGainers"
	ModelEquityLosers              ModelType = "EquityLosers"
	ModelEquityActive              ModelType = "EquityActive"
	ModelEquityUndervaluedLargeCaps ModelType = "EquityUndervaluedLargeCaps"
	ModelEquityUndervaluedGrowth   ModelType = "EquityUndervaluedGrowth"
	ModelEquityAggressiveSmallCaps ModelType = "EquityAggressiveSmallCaps"
	ModelGrowthTechEquities        ModelType = "GrowthTechEquities"
	ModelTopRetail                 ModelType = "TopRetail"
	ModelDiscoveryFilings          ModelType = "DiscoveryFilings"
	ModelLatestFinancialReports    ModelType = "LatestFinancialReports"
)

// --- Equity / Ownership ---
const (
	ModelEquityOwnership       ModelType = "EquityOwnership"
	ModelInstitutionalOwnership ModelType = "InstitutionalOwnership"
	ModelInsiderTrading        ModelType = "InsiderTrading"
	ModelGovernmentTrades      ModelType = "GovernmentTrades"
	ModelForm13FHR             ModelType = "Form13FHR"
)

// --- Equity / Shorts ---
const (
	ModelEquityFTD          ModelType = "EquityFTD"
	ModelShortVolume        ModelType = "ShortVolume"
	ModelEquityShortInterest ModelType = "EquityShortInterest"
)

// --- Equity / Compare ---
const (
	ModelCompareGroups       ModelType = "CompareGroups"
	ModelCompareCompanyFacts ModelType = "CompareCompanyFacts"
	ModelOTCAggregate        ModelType = "OTCAggregate"
)

// --- Derivatives / Options ---
const (
	ModelOptionsChains    ModelType = "OptionsChains"
	ModelOptionsUnusual   ModelType = "OptionsUnusual"
	ModelOptionsSnapshots ModelType = "OptionsSnapshots"
)

// --- Derivatives / Futures ---
const (
	ModelFuturesHistorical  ModelType = "FuturesHistorical"
	ModelFuturesCurve       ModelType = "FuturesCurve"
	ModelFuturesInstruments ModelType = "FuturesInstruments"
	ModelFuturesInfo        ModelType = "FuturesInfo"
)

// --- ETF ---
const (
	ModelEtfSearch           ModelType = "EtfSearch"
	ModelEtfHistorical       ModelType = "EtfHistorical"
	ModelEtfInfo             ModelType = "EtfInfo"
	ModelEtfHoldings         ModelType = "EtfHoldings"
	ModelEtfSectors          ModelType = "EtfSectors"
	ModelEtfCountries        ModelType = "EtfCountries"
	ModelEtfPricePerformance ModelType = "EtfPricePerformance"
	ModelEtfEquityExposure   ModelType = "EtfEquityExposure"
	ModelNportDisclosure     ModelType = "NportDisclosure"
	ModelETFGainers          ModelType = "ETFGainers"
	ModelETFLosers           ModelType = "ETFLosers"
	ModelETFActive           ModelType = "ETFActive"
)

// --- Index ---
const (
	ModelIndexHistorical  ModelType = "IndexHistorical"
	ModelIndexConstituents ModelType = "IndexConstituents"
	ModelIndexSnapshots   ModelType = "IndexSnapshots"
	ModelIndexSectors     ModelType = "IndexSectors"
	ModelAvailableIndices ModelType = "AvailableIndices"
	ModelIndexSearch      ModelType = "IndexSearch"
	ModelSP500Multiples   ModelType = "SP500Multiples"
)

// --- Crypto ---
const (
	ModelCryptoHistorical ModelType = "CryptoHistorical"
	ModelCryptoSearch     ModelType = "CryptoSearch"
)

// --- Currency / Forex ---
const (
	ModelCurrencyHistorical     ModelType = "CurrencyHistorical"
	ModelCurrencyPairs          ModelType = "CurrencyPairs"
	ModelCurrencyReferenceRates ModelType = "CurrencyReferenceRates"
	ModelCurrencySnapshots      ModelType = "CurrencySnapshots"
)

// --- News ---
const (
	ModelCompanyNews ModelType = "CompanyNews"
	ModelWorldNews   ModelType = "WorldNews"
)

// --- Economy ---
const (
	ModelEconomicCalendar         ModelType = "EconomicCalendar"
	ModelConsumerPriceIndex        ModelType = "ConsumerPriceIndex"
	ModelUnemployment              ModelType = "Unemployment"
	ModelGdpReal                   ModelType = "GdpReal"
	ModelGdpNominal                ModelType = "GdpNominal"
	ModelGdpForecast               ModelType = "GdpForecast"
	ModelCompositeLeadingIndicator ModelType = "CompositeLeadingIndicator"
	ModelCountryProfile            ModelType = "CountryProfile"
	ModelAvailableIndicators       ModelType = "AvailableIndicators"
	ModelEconomicIndicators        ModelType = "EconomicIndicators"
	ModelBalanceOfPayments         ModelType = "BalanceOfPayments"
	ModelDirectionOfTrade          ModelType = "DirectionOfTrade"
	ModelExportDestinations        ModelType = "ExportDestinations"
	ModelMoneyMeasures             ModelType = "MoneyMeasures"
	ModelCentralBankHoldings       ModelType = "CentralBankHoldings"
	ModelRiskPremium               ModelType = "RiskPremium"
	ModelSharePriceIndex           ModelType = "SharePriceIndex"
	ModelHousePriceIndex           ModelType = "HousePriceIndex"
	ModelCountryInterestRates      ModelType = "CountryInterestRates"
	ModelRetailPrices              ModelType = "RetailPrices"
	ModelPersonalConsumptionExpenditures ModelType = "PersonalConsumptionExpenditures"
	ModelNonFarmPayrolls           ModelType = "NonFarmPayrolls"
)

// --- Economy / FRED ---
const (
	ModelFredSearch       ModelType = "FredSearch"
	ModelFredSeries       ModelType = "FredSeries"
	ModelFredReleaseTable ModelType = "FredReleaseTable"
	ModelFredRegional     ModelType = "FredRegional"
)

// --- Economy / Surveys ---
const (
	ModelUniversityOfMichigan              ModelType = "UniversityOfMichigan"
	ModelSeniorLoanOfficerSurvey           ModelType = "SeniorLoanOfficerSurvey"
	ModelManufacturingOutlookNY            ModelType = "ManufacturingOutlookNY"
	ModelManufacturingOutlookTexas         ModelType = "ManufacturingOutlookTexas"
	ModelSurveyOfEconomicConditionsChicago ModelType = "SurveyOfEconomicConditionsChicago"
	ModelBlsSearch                         ModelType = "BlsSearch"
	ModelBlsSeries                         ModelType = "BlsSeries"
)

// --- Economy / FOMC & Fed ---
const (
	ModelFomcDocuments              ModelType = "FomcDocuments"
	ModelPrimaryDealerPositioning   ModelType = "PrimaryDealerPositioning"
	ModelPrimaryDealerFails         ModelType = "PrimaryDealerFails"
	ModelTotalFactorProductivity    ModelType = "TotalFactorProductivity"
	ModelInflationExpectations      ModelType = "InflationExpectations"
)

// --- Economy / Shipping ---
const (
	ModelPortInfo                 ModelType = "PortInfo"
	ModelPortVolume               ModelType = "PortVolume"
	ModelMaritimeChokePointInfo   ModelType = "MaritimeChokePointInfo"
	ModelMaritimeChokePointVolume ModelType = "MaritimeChokePointVolume"
)

// --- Fixed Income / Rates ---
const (
	ModelSOFR                              ModelType = "SOFR"
	ModelSONIA                             ModelType = "SONIA"
	ModelAmeribor                          ModelType = "Ameribor"
	ModelFederalFundsRate                  ModelType = "FederalFundsRate"
	ModelProjections                       ModelType = "Projections"
	ModelEuroShortTermRate                 ModelType = "EuroShortTermRate"
	ModelEuropeanCentralBankInterestRates  ModelType = "EuropeanCentralBankInterestRates"
	ModelIORB                              ModelType = "IORB"
	ModelDiscountWindowPrimaryCreditRate   ModelType = "DiscountWindowPrimaryCreditRate"
	ModelOvernightBankFundingRate          ModelType = "OvernightBankFundingRate"
)

// --- Fixed Income / Government ---
const (
	ModelYieldCurve         ModelType = "YieldCurve"
	ModelTreasuryRates      ModelType = "TreasuryRates"
	ModelTreasuryAuctions   ModelType = "TreasuryAuctions"
	ModelTreasuryPrices     ModelType = "TreasuryPrices"
	ModelTipsYields         ModelType = "TipsYields"
	ModelSvenssonYieldCurve ModelType = "SvenssonYieldCurve"
)

// --- Fixed Income / Spreads ---
const (
	ModelTreasuryConstantMaturity         ModelType = "TreasuryConstantMaturity"
	ModelSelectedTreasuryConstantMaturity ModelType = "SelectedTreasuryConstantMaturity"
	ModelSelectedTreasuryBill             ModelType = "SelectedTreasuryBill"
)

// --- Fixed Income / Corporate ---
const (
	ModelHighQualityMarketCorporateBond ModelType = "HighQualityMarketCorporateBond"
	ModelSpotRate                       ModelType = "SpotRate"
	ModelCommercialPaper                ModelType = "CommercialPaper"
	ModelBondPrices                     ModelType = "BondPrices"
	ModelBondIndices                    ModelType = "BondIndices"
	ModelMortgageIndices                ModelType = "MortgageIndices"
)

// --- Commodity ---
const (
	ModelCommoditySpotPrices       ModelType = "CommoditySpotPrices"
	ModelPetroleumStatusReport     ModelType = "PetroleumStatusReport"
	ModelShortTermEnergyOutlook    ModelType = "ShortTermEnergyOutlook"
	ModelCommodityPsdData          ModelType = "CommodityPsdData"
	ModelCommodityPsdReport        ModelType = "CommodityPsdReport"
	ModelWeatherBulletin           ModelType = "WeatherBulletin"
	ModelWeatherBulletinDownload   ModelType = "WeatherBulletinDownload"
)

// --- Regulators / SEC ---
const (
	ModelCompanyFilings      ModelType = "CompanyFilings"
	ModelCikMap              ModelType = "CikMap"
	ModelSymbolMap           ModelType = "SymbolMap"
	ModelSicSearch           ModelType = "SicSearch"
	ModelInstitutionsSearch  ModelType = "InstitutionsSearch"
	ModelSchemaFiles         ModelType = "SchemaFiles"
	ModelRssLitigation       ModelType = "RssLitigation"
	ModelSecFiling           ModelType = "SecFiling"
	ModelSecHtmFile          ModelType = "SecHtmFile"
)

// --- Regulators / CFTC ---
const (
	ModelCOT       ModelType = "COT"
	ModelCOTSearch ModelType = "COTSearch"
)

// --- Factor Models ---
const (
	ModelFamaFrenchFactors                    ModelType = "FamaFrenchFactors"
	ModelFamaFrenchBreakpoints                ModelType = "FamaFrenchBreakpoints"
	ModelFamaFrenchUSPortfolioReturns         ModelType = "FamaFrenchUSPortfolioReturns"
	ModelFamaFrenchCountryPortfolioReturns    ModelType = "FamaFrenchCountryPortfolioReturns"
	ModelFamaFrenchRegionalPortfolioReturns   ModelType = "FamaFrenchRegionalPortfolioReturns"
	ModelFamaFrenchInternationalIndexReturns  ModelType = "FamaFrenchInternationalIndexReturns"
)

// --- Congress ---
const (
	ModelCongressBills    ModelType = "CongressBills"
	ModelCongressBillInfo ModelType = "CongressBillInfo"
	ModelCongressBillText ModelType = "CongressBillText"
)

// AllModels returns all defined model types. Useful for iteration and validation.
func AllModels() []ModelType {
	return []ModelType{
		// Equity / Price
		ModelEquityHistorical, ModelEquityQuote, ModelEquityInfo,
		ModelEquitySearch, ModelEquityScreener, ModelEquityNBBO,
		ModelEquityPeers, ModelMarketSnapshots, ModelHistoricalMarketCap,
		ModelPricePerformance,
		// Equity / Fundamentals
		ModelBalanceSheet, ModelBalanceSheetGrowth,
		ModelIncomeStatement, ModelIncomeStatementGrowth,
		ModelCashFlowStatement, ModelCashFlowStatementGrowth,
		ModelFinancialRatios, ModelKeyMetrics,
		ModelKeyExecutives, ModelExecutiveCompensation,
		ModelRevenueGeographic, ModelRevenueBusinessLine,
		ModelReportedFinancials,
		ModelHistoricalDividends, ModelHistoricalSplits,
		ModelHistoricalEps, ModelHistoricalEmployees,
		ModelEarningsCallTranscript, ModelTrailingDividendYield,
		ModelManagementDiscussionAnalysis, ModelEsgScore,
		ModelShareStatistics,
		// Equity / Estimates
		ModelPriceTarget, ModelPriceTargetConsensus,
		ModelAnalystEstimates, ModelAnalystSearch,
		ModelForwardEpsEstimates, ModelForwardEbitdaEstimates,
		ModelForwardPeEstimates, ModelForwardSalesEstimates,
		// Equity / Calendar
		ModelCalendarEarnings, ModelCalendarDividend,
		ModelCalendarIpo, ModelCalendarSplits, ModelCalendarEvents,
		// Equity / Discovery
		ModelEquityGainers, ModelEquityLosers, ModelEquityActive,
		ModelEquityUndervaluedLargeCaps, ModelEquityUndervaluedGrowth,
		ModelEquityAggressiveSmallCaps, ModelGrowthTechEquities,
		ModelTopRetail, ModelDiscoveryFilings, ModelLatestFinancialReports,
		// Equity / Ownership
		ModelEquityOwnership, ModelInstitutionalOwnership,
		ModelInsiderTrading, ModelGovernmentTrades, ModelForm13FHR,
		// Equity / Shorts
		ModelEquityFTD, ModelShortVolume, ModelEquityShortInterest,
		// Equity / Compare
		ModelCompareGroups, ModelCompareCompanyFacts, ModelOTCAggregate,
		// Derivatives
		ModelOptionsChains, ModelOptionsUnusual, ModelOptionsSnapshots,
		ModelFuturesHistorical, ModelFuturesCurve,
		ModelFuturesInstruments, ModelFuturesInfo,
		// ETF
		ModelEtfSearch, ModelEtfHistorical, ModelEtfInfo,
		ModelEtfHoldings, ModelEtfSectors, ModelEtfCountries,
		ModelEtfPricePerformance, ModelEtfEquityExposure,
		ModelNportDisclosure, ModelETFGainers, ModelETFLosers, ModelETFActive,
		// Index
		ModelIndexHistorical, ModelIndexConstituents,
		ModelIndexSnapshots, ModelIndexSectors,
		ModelAvailableIndices, ModelIndexSearch, ModelSP500Multiples,
		// Crypto
		ModelCryptoHistorical, ModelCryptoSearch,
		// Currency
		ModelCurrencyHistorical, ModelCurrencyPairs,
		ModelCurrencyReferenceRates, ModelCurrencySnapshots,
		// News
		ModelCompanyNews, ModelWorldNews,
		// Economy
		ModelEconomicCalendar, ModelConsumerPriceIndex,
		ModelUnemployment, ModelGdpReal, ModelGdpNominal, ModelGdpForecast,
		ModelCompositeLeadingIndicator,
		ModelCountryProfile, ModelAvailableIndicators, ModelEconomicIndicators,
		ModelBalanceOfPayments, ModelDirectionOfTrade, ModelExportDestinations,
		ModelMoneyMeasures, ModelCentralBankHoldings,
		ModelRiskPremium, ModelSharePriceIndex, ModelHousePriceIndex,
		ModelCountryInterestRates, ModelRetailPrices,
		ModelPersonalConsumptionExpenditures, ModelNonFarmPayrolls,
		// Economy / FRED
		ModelFredSearch, ModelFredSeries, ModelFredReleaseTable, ModelFredRegional,
		// Economy / Surveys
		ModelUniversityOfMichigan, ModelSeniorLoanOfficerSurvey,
		ModelManufacturingOutlookNY, ModelManufacturingOutlookTexas,
		ModelSurveyOfEconomicConditionsChicago,
		ModelBlsSearch, ModelBlsSeries,
		// Economy / FOMC
		ModelFomcDocuments, ModelPrimaryDealerPositioning,
		ModelPrimaryDealerFails, ModelTotalFactorProductivity,
		ModelInflationExpectations,
		// Economy / Shipping
		ModelPortInfo, ModelPortVolume,
		ModelMaritimeChokePointInfo, ModelMaritimeChokePointVolume,
		// Fixed Income / Rates
		ModelSOFR, ModelSONIA, ModelAmeribor, ModelFederalFundsRate,
		ModelProjections, ModelEuroShortTermRate,
		ModelEuropeanCentralBankInterestRates, ModelIORB,
		ModelDiscountWindowPrimaryCreditRate, ModelOvernightBankFundingRate,
		// Fixed Income / Government
		ModelYieldCurve, ModelTreasuryRates, ModelTreasuryAuctions,
		ModelTreasuryPrices, ModelTipsYields, ModelSvenssonYieldCurve,
		// Fixed Income / Spreads
		ModelTreasuryConstantMaturity, ModelSelectedTreasuryConstantMaturity,
		ModelSelectedTreasuryBill,
		// Fixed Income / Corporate
		ModelHighQualityMarketCorporateBond, ModelSpotRate,
		ModelCommercialPaper, ModelBondPrices, ModelBondIndices, ModelMortgageIndices,
		// Commodity
		ModelCommoditySpotPrices, ModelPetroleumStatusReport,
		ModelShortTermEnergyOutlook,
		ModelCommodityPsdData, ModelCommodityPsdReport,
		ModelWeatherBulletin, ModelWeatherBulletinDownload,
		// Regulators / SEC
		ModelCompanyFilings, ModelCikMap, ModelSymbolMap,
		ModelSicSearch, ModelInstitutionsSearch, ModelSchemaFiles,
		ModelRssLitigation, ModelSecFiling, ModelSecHtmFile,
		// Regulators / CFTC
		ModelCOT, ModelCOTSearch,
		// Factor Models
		ModelFamaFrenchFactors, ModelFamaFrenchBreakpoints,
		ModelFamaFrenchUSPortfolioReturns,
		ModelFamaFrenchCountryPortfolioReturns,
		ModelFamaFrenchRegionalPortfolioReturns,
		ModelFamaFrenchInternationalIndexReturns,
		// Congress
		ModelCongressBills, ModelCongressBillInfo, ModelCongressBillText,
	}
}

// ModelCategory maps model types to their category for grouping.
func ModelCategory(m ModelType) string {
	switch m {
	case ModelEquityHistorical, ModelEquityQuote, ModelEquityInfo,
		ModelEquitySearch, ModelEquityScreener, ModelEquityNBBO,
		ModelEquityPeers, ModelMarketSnapshots, ModelHistoricalMarketCap,
		ModelPricePerformance:
		return "Equity / Price"
	case ModelBalanceSheet, ModelBalanceSheetGrowth,
		ModelIncomeStatement, ModelIncomeStatementGrowth,
		ModelCashFlowStatement, ModelCashFlowStatementGrowth,
		ModelFinancialRatios, ModelKeyMetrics,
		ModelKeyExecutives, ModelExecutiveCompensation,
		ModelRevenueGeographic, ModelRevenueBusinessLine,
		ModelReportedFinancials,
		ModelHistoricalDividends, ModelHistoricalSplits,
		ModelHistoricalEps, ModelHistoricalEmployees,
		ModelEarningsCallTranscript, ModelTrailingDividendYield,
		ModelManagementDiscussionAnalysis, ModelEsgScore,
		ModelShareStatistics:
		return "Equity / Fundamentals"
	case ModelPriceTarget, ModelPriceTargetConsensus,
		ModelAnalystEstimates, ModelAnalystSearch,
		ModelForwardEpsEstimates, ModelForwardEbitdaEstimates,
		ModelForwardPeEstimates, ModelForwardSalesEstimates:
		return "Equity / Estimates"
	case ModelCalendarEarnings, ModelCalendarDividend,
		ModelCalendarIpo, ModelCalendarSplits, ModelCalendarEvents:
		return "Equity / Calendar"
	case ModelEquityGainers, ModelEquityLosers, ModelEquityActive,
		ModelEquityUndervaluedLargeCaps, ModelEquityUndervaluedGrowth,
		ModelEquityAggressiveSmallCaps, ModelGrowthTechEquities,
		ModelTopRetail, ModelDiscoveryFilings, ModelLatestFinancialReports:
		return "Equity / Discovery"
	case ModelEquityOwnership, ModelInstitutionalOwnership,
		ModelInsiderTrading, ModelGovernmentTrades, ModelForm13FHR:
		return "Equity / Ownership"
	case ModelEquityFTD, ModelShortVolume, ModelEquityShortInterest:
		return "Equity / Shorts"
	case ModelCompareGroups, ModelCompareCompanyFacts, ModelOTCAggregate:
		return "Equity / Compare"
	case ModelOptionsChains, ModelOptionsUnusual, ModelOptionsSnapshots:
		return "Derivatives / Options"
	case ModelFuturesHistorical, ModelFuturesCurve,
		ModelFuturesInstruments, ModelFuturesInfo:
		return "Derivatives / Futures"
	case ModelEtfSearch, ModelEtfHistorical, ModelEtfInfo,
		ModelEtfHoldings, ModelEtfSectors, ModelEtfCountries,
		ModelEtfPricePerformance, ModelEtfEquityExposure,
		ModelNportDisclosure, ModelETFGainers, ModelETFLosers, ModelETFActive:
		return "ETF"
	case ModelIndexHistorical, ModelIndexConstituents,
		ModelIndexSnapshots, ModelIndexSectors,
		ModelAvailableIndices, ModelIndexSearch, ModelSP500Multiples:
		return "Index"
	case ModelCryptoHistorical, ModelCryptoSearch:
		return "Crypto"
	case ModelCurrencyHistorical, ModelCurrencyPairs,
		ModelCurrencyReferenceRates, ModelCurrencySnapshots:
		return "Currency"
	case ModelCompanyNews, ModelWorldNews:
		return "News"
	case ModelCommoditySpotPrices, ModelPetroleumStatusReport,
		ModelShortTermEnergyOutlook,
		ModelCommodityPsdData, ModelCommodityPsdReport,
		ModelWeatherBulletin, ModelWeatherBulletinDownload:
		return "Commodity"
	case ModelCompanyFilings, ModelCikMap, ModelSymbolMap,
		ModelSicSearch, ModelInstitutionsSearch, ModelSchemaFiles,
		ModelRssLitigation, ModelSecFiling, ModelSecHtmFile:
		return "Regulators / SEC"
	case ModelCOT, ModelCOTSearch:
		return "Regulators / CFTC"
	default:
		// Check remaining categories.
		switch {
		case m == ModelSOFR || m == ModelSONIA || m == ModelAmeribor ||
			m == ModelFederalFundsRate || m == ModelProjections ||
			m == ModelEuroShortTermRate || m == ModelEuropeanCentralBankInterestRates ||
			m == ModelIORB || m == ModelDiscountWindowPrimaryCreditRate ||
			m == ModelOvernightBankFundingRate:
			return "Fixed Income / Rates"
		case m == ModelYieldCurve || m == ModelTreasuryRates ||
			m == ModelTreasuryAuctions || m == ModelTreasuryPrices ||
			m == ModelTipsYields || m == ModelSvenssonYieldCurve:
			return "Fixed Income / Government"
		case m == ModelTreasuryConstantMaturity ||
			m == ModelSelectedTreasuryConstantMaturity ||
			m == ModelSelectedTreasuryBill:
			return "Fixed Income / Spreads"
		case m == ModelHighQualityMarketCorporateBond || m == ModelSpotRate ||
			m == ModelCommercialPaper || m == ModelBondPrices ||
			m == ModelBondIndices || m == ModelMortgageIndices:
			return "Fixed Income / Corporate"
		default:
			return "Economy"
		}
	}
}
