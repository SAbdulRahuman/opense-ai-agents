package models

import "time"

// CommodityPrice represents a commodity spot/futures price.
type CommodityPrice struct {
	Date      time.Time `json:"date"`
	Commodity string    `json:"commodity"` // e.g., "Crude Oil WTI", "Gold", "Silver"
	Price     float64   `json:"price"`
	Change    float64   `json:"change,omitempty"`
	ChangePct float64   `json:"change_pct,omitempty"`
	Unit      string    `json:"unit,omitempty"` // e.g., "USD/barrel", "USD/oz"
	Source    string    `json:"source,omitempty"`
}

// PetroleumStatusData represents EIA petroleum status report data.
type PetroleumStatusData struct {
	Date                time.Time `json:"date"`
	CrudeOilInventory   float64   `json:"crude_oil_inventory,omitempty"`
	GasolineInventory   float64   `json:"gasoline_inventory,omitempty"`
	DistillateInventory float64   `json:"distillate_inventory,omitempty"`
	CrudeOilProduction  float64   `json:"crude_oil_production,omitempty"`
	GasolineProduction  float64   `json:"gasoline_production,omitempty"`
	CrudeOilImports     float64   `json:"crude_oil_imports,omitempty"`
	Unit                string    `json:"unit,omitempty"` // e.g., "thousand barrels"
}

// EnergyOutlookData represents short-term energy outlook from EIA.
type EnergyOutlookData struct {
	Date           time.Time `json:"date"`
	Category       string    `json:"category"`
	Metric         string    `json:"metric"`
	Value          float64   `json:"value"`
	Unit           string    `json:"unit,omitempty"`
	ForecastPeriod string    `json:"forecast_period,omitempty"`
}

// CommodityPSDData represents USDA PSD (Production, Supply, Distribution) data.
type CommodityPSDData struct {
	Country       string  `json:"country"`
	Commodity     string  `json:"commodity"`
	MarketYear    string  `json:"market_year"`
	Production    float64 `json:"production,omitempty"`
	Imports       float64 `json:"imports,omitempty"`
	Exports       float64 `json:"exports,omitempty"`
	DomesticUse   float64 `json:"domestic_use,omitempty"`
	EndingStocks  float64 `json:"ending_stocks,omitempty"`
	Unit          string  `json:"unit,omitempty"`
}

// WeatherBulletinData represents weather bulletin data relevant to commodities.
type WeatherBulletinData struct {
	Date     time.Time `json:"date"`
	Region   string    `json:"region"`
	Title    string    `json:"title"`
	Summary  string    `json:"summary,omitempty"`
	URL      string    `json:"url,omitempty"`
	FileType string    `json:"file_type,omitempty"` // "pdf", "html"
}
