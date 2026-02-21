package federalreserve

// ---------------------------------------------------------------------------
// NY Fed Markets API response types.
// ---------------------------------------------------------------------------

// nyfedRatesResponse wraps the ref rates (EFFR, SOFR, OBFR).
type nyfedRatesResponse struct {
	RefRates []nyfedRefRate `json:"refRates"`
}

// nyfedRefRate is a single ref rate entry from the NY Fed.
type nyfedRefRate struct {
	EffectiveDate      string  `json:"effectiveDate"`
	Type               string  `json:"type,omitempty"`
	PercentRate        float64 `json:"percentRate"`
	TargetRateTo       float64 `json:"targetRateTo,omitempty"`
	TargetRateFrom     float64 `json:"targetRateFrom,omitempty"`
	PercentPercentile1  float64 `json:"percentPercentile1,omitempty"`
	PercentPercentile25 float64 `json:"percentPercentile25,omitempty"`
	PercentPercentile75 float64 `json:"percentPercentile75,omitempty"`
	PercentPercentile99 float64 `json:"percentPercentile99,omitempty"`
	VolumeInBillions   float64 `json:"volumeInBillions,omitempty"`
	IntraDayLow        float64 `json:"intraDayLow,omitempty"`
	IntraDayHigh       float64 `json:"intraDayHigh,omitempty"`
	StdDeviation       float64 `json:"stdDeviation,omitempty"`
	RevisionIndicator  string  `json:"revisionIndicator,omitempty"`
}

// ---------------------------------------------------------------------------
// NY Fed SOMA (System Open Market Account) response types.
// ---------------------------------------------------------------------------

// nyfedSomaResponse wraps SOMA holdings data.
type nyfedSomaResponse struct {
	Soma nyfedSomaPayload `json:"soma"`
}

// nyfedSomaPayload is the inner SOMA payload.
type nyfedSomaPayload struct {
	AsOfDates []string          `json:"asOfDates,omitempty"`
	Summary   []nyfedSomaEntry  `json:"summary,omitempty"`
	Holdings  []nyfedSomaEntry  `json:"holdings,omitempty"`
}

// nyfedSomaEntry represents a single SOMA holding record.
type nyfedSomaEntry struct {
	AsOfDate              string  `json:"asOfDate"`
	SecurityDescription   string  `json:"securityDescription,omitempty"`
	SecurityTypes         any     `json:"securityTypes,omitempty"` // string or []string
	CUSIP                 string  `json:"cusip,omitempty"`
	Issuer                string  `json:"issuer,omitempty"`
	MaturityDate          string  `json:"maturityDate,omitempty"`
	Term                  string  `json:"term,omitempty"`
	Coupon                float64 `json:"coupon,omitempty"`
	Spread                float64 `json:"spread,omitempty"`
	CurrentFaceValue      float64 `json:"currentFaceValue,omitempty"`
	ParValue              float64 `json:"parValue,omitempty"`
	PercentOutstanding    float64 `json:"percentOutstanding,omitempty"`
	// Summary-specific fields:
	Bills                 float64 `json:"bills,omitempty"`
	FRN                   float64 `json:"frn,omitempty"`
	NotesBonds            float64 `json:"notesbonds,omitempty"`
	TIPS                  float64 `json:"tips,omitempty"`
	MBS                   float64 `json:"mbs,omitempty"`
	CMBS                  float64 `json:"cmbs,omitempty"`
	Agencies              float64 `json:"agencies,omitempty"`
	Total                 float64 `json:"total,omitempty"`
	InflationCompensation float64 `json:"inflationCompensation,omitempty"`
	ChangeFromPriorWeek   float64 `json:"changeFromPriorWeek,omitempty"`
	ChangeFromPriorYear   float64 `json:"changeFromPriorYear,omitempty"`
}

// ---------------------------------------------------------------------------
// NY Fed Primary Dealer response types.
// ---------------------------------------------------------------------------

// nyfedPDResponse wraps primary dealer data.
type nyfedPDResponse struct {
	PD nyfedPDPayload `json:"pd"`
}

// nyfedPDPayload is the inner PD payload.
type nyfedPDPayload struct {
	Timeseries []nyfedPDEntry `json:"timeseries"`
}

// nyfedPDEntry is a single primary dealer time series entry.
type nyfedPDEntry struct {
	KeyID    string `json:"keyid"`
	AsOfDate string `json:"asofdate"`
	Value    string `json:"value"`
}
