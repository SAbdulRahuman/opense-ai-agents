# OECD & IMF Provider Research — API Patterns for Go Implementation

> Research from OpenBB source code: `oss/OpenBB/openbb_platform/providers/oecd/` and `providers/imf/`

---

## Table of Contents
- [OECD Provider (9 Endpoints)](#oecd-provider-9-endpoints)
  - [Architecture & Helpers](#oecd-architecture--helpers)
  - [1. CompositeLeadingIndicator (CLI)](#1-compositeleadingindicator-cli)
  - [2. ConsumerPriceIndex (CPI)](#2-consumerpriceindex-cpi)
  - [3. CountryInterestRates](#3-countryinterestrates)
  - [4. GdpNominal](#4-gdpnominal)
  - [5. GdpReal](#5-gdpreal)
  - [6. GdpForecast](#6-gdpforecast)
  - [7. HousePriceIndex](#7-housepriceindex)
  - [8. SharePriceIndex](#8-sharepriceindex)
  - [9. Unemployment](#9-unemployment)
- [IMF Provider (8 Endpoints)](#imf-provider-8-endpoints)
  - [Architecture & Helpers](#imf-architecture--helpers)
  - [1. AvailableIndicators](#1-availableindicators)
  - [2. ConsumerPriceIndex (CPI)](#2-consumerpriceindex-cpi-1)
  - [3. DirectionOfTrade](#3-directionoftrade)
  - [4. EconomicIndicators](#4-economicindicators)
  - [5. MaritimeChokePointInfo](#5-maritimechokepointinfo)
  - [6. MaritimeChokePointVolume](#6-maritimechokepointvolume)
  - [7. PortInfo](#7-portinfo)
  - [8. PortVolume](#8-portvolume)
- [Summary & Go Implementation Notes](#summary--go-implementation-notes)

---

# OECD Provider (9 Endpoints)

## OECD Architecture & Helpers

### Base URL
All OECD endpoints use the **SDMX REST API** via:
```
https://sdmx.oecd.org/public/rest/data/{AGENCY_ID},{DSD_ID},{VERSION}/{KEY}?{PARAMS}
```

### Response Formats
- **Primary**: CSV (`format=csvfile` or `Accept: application/vnd.sdmx.data+csv`)
- **Secondary**: SDMX XML (parsed with `defusedxml.ElementTree`)
- Columns extracted from CSV: `REF_AREA`, `TIME_PERIOD`, `OBS_VALUE` (+ others per endpoint)

### Common Request Headers
```
Accept: application/vnd.sdmx.data+csv; charset=utf-8
```

### Common Query Parameters (all endpoints)
| Parameter | Description | Format |
|---|---|---|
| `startPeriod` | Start date | `YYYY-MM` or `YYYY-MM-DD` |
| `endPeriod` | End date | `YYYY-MM` or `YYYY-MM-DD` |
| `dimensionAtObservation` | Always `TIME_PERIOD` | String |
| `detail` | Always `dataonly` | String |
| `format` | `csvfile` (some endpoints) | String |

### SSL Workaround
OECD requires legacy SSL support. Python uses a `CustomHttpAdapter` with `OP_LEGACY_SERVER_CONNECT`. In Go, configure `tls.Config` with `MinVersion: tls.VersionTLS10`.

### Caching
Results cached as CSV/Parquet files in `~/.openbb_platform/user_data/oecd/`. Cache valid for 1 day (date-based timestamp file).

### Date Parsing (`oecd_date_to_python_date`)
- `"2023-Q1"` → parse as quarter start date
- `"2023"` → `2023-01-01`
- `"2023-05"` → parse as month start date

### Country Code Mapping
Each endpoint has its own country-to-ISO3 map. Key pattern: `snake_case_country_name` → `ISO3_CODE`. Multiple countries joined with `+` in URL key segment. Empty string = all countries.

### Third-Party Dependencies (Python)
- `requests`, `urllib3` (HTTP with custom SSL)
- `defusedxml` (XML parsing)
- `pandas` (CSV parsing, data manipulation)
- `pyarrow` (Parquet cache)

### Go Equivalents Needed
- `encoding/csv` or `gocsv` for CSV parsing
- `encoding/xml` for SDMX XML
- `crypto/tls` for SSL configuration
- `net/http` for requests
- Simple file-based cache

---

## 1. CompositeLeadingIndicator (CLI)

### API URL
```
https://sdmx.oecd.org/public/rest/data/OECD.SDD.STES,DSD_STES@DF_CLI,4.1/{COUNTRY}.M.LI...{ADJUSTMENT}.{GROWTH_RATE}..H?startPeriod={START}&endPeriod={END}&dimensionAtObservation=TIME_PERIOD&detail=dataonly&format=csvfile
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `country` | No | string | `"g20"` | 22 choices + `"all"` (multi-value with `,`) |
| `start_date` | No | date | `1947-01-01` (or `2020-01-01` if all) | Date |
| `end_date` | No | date | current year Dec 31 | Date |
| `adjustment` | No | enum | `"amplitude"` | `"amplitude"` → `AA`, `"normalized"` → `NOR` |
| `growth_rate` | No | bool | `false` | `true` → `GY`, `false` → `IX` (if GY, adjustment="") |

### URL Key Pattern
`{COUNTRIES}.M.LI...{ADJUSTMENT}.{GROWTH_RATE}..H`

### Response: CSV
Columns: `REF_AREA`, `TIME_PERIOD`, `OBS_VALUE` (+ others ignored)

### Data Processing
1. Parse CSV, extract 3 columns
2. Map `REF_AREA` codes back to country names (title case)
3. Convert `TIME_PERIOD` to date
4. If `growth_rate=true`, divide value by 100
5. Filter nulls, sort by date+country

### Country Codes (22 countries)
```
g20→G20, g7→G7, asia5→A5M, north_america→NAFTA, europe4→G4E,
australia→AUS, brazil→BRA, canada→CAN, china→CHN, france→FRA,
germany→DEU, india→IND, indonesia→IDN, italy→ITA, japan→JPN,
mexico→MEX, spain→ESP, south_africa→ZAF, south_korea→KOR,
turkey→TUR, united_states→USA, united_kingdom→GBR
```

---

## 2. ConsumerPriceIndex (CPI)

### API URL
```
https://sdmx.oecd.org/public/rest/data/OECD.SDD.TPS,DSD_PRICES@DF_PRICES_ALL,1.0/{COUNTRY}.{FREQ}.{METHODOLOGY}.CPI.{UNITS}.{EXPENDITURE}.N.
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `country` | No | string | `"united_states"` | ~45 countries + `"all"` |
| `start_date` | No | date | `1950-01-01` | Date |
| `end_date` | No | date | current year Dec 31 | Date |
| `frequency` | No | enum | (from standard) | `"monthly"→M`, `"quarter"→Q`, `"annual"→A` |
| `harmonized` | No | bool | `false` | `true` → `HICP`, `false` → `N` |
| `transform` | No | enum | `"index"` | `"index"→IX`, `"yoy"→PA`, `"period"→PC (mom)` |
| `expenditure` | No | enum | `"total"` | 30 choices + `"all"` (COICOP codes) |

### Data Processing (uses caching via `get_possibly_cached_data`)
1. Fetch SDMX XML, parse to DataFrame
2. Filter by query params using DataFrame query syntax
3. Extract: `REF_AREA`, `TIME_PERIOD`, `VALUE`, `EXPENDITURE`
4. Map country codes, expenditure codes to human-readable names
5. Parse dates, filter date range
6. If transform is yoy/period: divide value by 100

### Expenditure Code Map
```
_T→total, CP01→food_non_alcoholic_beverages, CP02→alcoholic_beverages_tobacco_narcotics,
CP03→clothing_footwear, CP04→housing_water_electricity_gas, CP05→furniture_household_equipment,
CP06→health, CP07→transport, CP08→communication, CP09→recreation_culture,
CP10→education, CP11→restaurants_hotels, CP12→miscellaneous_goods_services,
CP045_0722→energy, GD→goods, SERV→services, ...
```

---

## 3. CountryInterestRates

### API URL
```
https://sdmx.oecd.org/public/rest/data/OECD.SDD.STES,DSD_KEI@DF_KEI,4.0/{COUNTRY}.{FREQ}.{DURATION}....?startPeriod={START}&endPeriod={END}&dimensionAtObservation=TIME_PERIOD&detail=dataonly
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `country` | No | string | `"united_states"` | ~50 countries + `"all"` |
| `start_date` | No | date | `1954-01-01` | Date |
| `end_date` | No | date | current year Dec 31 | Date |
| `frequency` | No | enum | `"monthly"` | `"monthly"→M`, `"quarter"→Q`, `"annual"→A` |
| `duration` | No | enum | `"short"` | `"immediate"→IRSTCI`, `"short"→IR3TIB`, `"long"→IRLT` |

### Response: CSV
Accept header requests CSV format.

### Data Processing
1. Parse CSV, extract `REF_AREA`, `TIME_PERIOD`, `OBS_VALUE`
2. Map country codes, parse dates
3. **Divide value by 100** (rates expressed as percentage)
4. Filter nulls, sort by date+country

---

## 4. GdpNominal

### API URL
```
https://sdmx.oecd.org/public/rest/data/OECD.SDD.NAD,DSD_NAMAIN1@DF_QNA_EXPENDITURE_{UNIT},1.1/{FREQ}..{COUNTRY}.S1..B1GQ.....{PRICE_BASE}..?&startPeriod={START}&endPeriod={END}&dimensionAtObservation=TIME_PERIOD&detail=dataonly&format=csvfile
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `country` | No | string | `"united_states"` | ~60 countries + `"all"` |
| `start_date` | No | date | `1947-01-01` | Date |
| `end_date` | No | date | current year Dec 31 | Date |
| `frequency` | No | enum | `"quarter"` | `"quarter"→Q`, `"annual"→A` |
| `units` | No | enum | `"level"` | `"level"→USD`, `"index"→INDICES`, `"capita"→CAPITA` |
| `price_base` | No | enum | `"current_prices"` | `"current_prices"→V`, `"volume"→LR` |

### Special Logic
- If `units=index` AND `price_base=current_prices` → change `price_base` to `DR`
- If `units=capita` → replace `B1GQ` with `B1GQ_POP` in URL
- If `units=level` → multiply value by 1,000,000 and cast to int64

### DSD Pattern
`OECD.SDD.NAD,DSD_NAMAIN1@DF_QNA_EXPENDITURE_{UNIT},1.1`

Where `{UNIT}` = `USD` | `INDICES` | `CAPITA`

---

## 5. GdpReal

### API URL
```
https://sdmx.oecd.org/public/rest/data/OECD.SDD.NAD,DSD_NAMAIN1@DF_QNA,1.1/{FREQ}..{COUNTRY}.S1..B1GQ._Z...USD_PPP.LR.LA.T0102?&startPeriod={START}&endPeriod={END}&dimensionAtObservation=TIME_PERIOD&detail=dataonly&format=csvfile
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `country` | No | string | `"united_states"` | ~60 countries + `"all"` |
| `start_date` | No | date | `1947-01-01` | Date |
| `end_date` | No | date | current year Dec 31 | Date |
| `frequency` | No | enum | `"quarter"` | `"quarter"→Q`, `"annual"→A` |

### Data Processing
- Same as GdpNominal but fixed unit: USD PPP, chain-linked volume
- Value multiplied by 1,000,000, cast to int64

---

## 6. GdpForecast

### API URL
```
https://sdmx.oecd.org/public/rest/data/OECD.ECO.MAD,DSD_EO@DF_EO,1.1/{COUNTRY}.{MEASURE}.{FREQ}?startPeriod={START}&endPeriod={END}&dimensionAtObservation=TIME_PERIOD&detail=dataonly&format=csvfile
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `country` | No | string | `"all"` | ~55 countries + groups + `"all"` |
| `start_date` | No | date | current year Jan 1 | Date |
| `end_date` | No | date | current year + 2 Dec 31 | Date |
| `frequency` | No | enum | `"annual"` | `"annual"→A`, `"quarter"→Q` |
| `units` | No | enum | `"volume"` | See below |

### Units/Measure Map
```
current_prices → GDP_USD
volume         → GDPV_USD
capita         → GDPVD_CAP (annual only)
growth         → GDPV_ANNPCT
deflator       → PGDP
```

### Data Processing
- If `units != growth`: cast value to int64, filter value > 0
- If `units == growth`: divide value by 100 (percent), filter value > 0

### Special Country Codes
Includes special groups: `asia→DAE`, `world→W`, `rest_of_the_world→WXD`, `other_major_oil_producers→OIL_O`

---

## 7. HousePriceIndex

### API URL
```
https://sdmx.oecd.org/public/rest/data/OECD.SDD.TPS,DSD_RHPI_TARGET@DF_RHPI_TARGET,1.0/COU.{COUNTRY}.{FREQ}.RHPI.{TRANSFORM}....?startPeriod={START}&endPeriod={END}&dimensionAtObservation=TIME_PERIOD&detail=dataonly
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `country` | No | string | `"united_states"` | ~52 countries + `"all"` |
| `start_date` | No | date | `1969-01-01` | Date |
| `end_date` | No | date | current year Dec 31 | Date |
| `frequency` | No | enum | (from standard) | `"monthly"→M`, `"quarter"→Q`, `"annual"→A` |
| `transform` | No | enum | (from standard) | `"yoy"→PA`, `"period"→PC`, `"index"→IX` |

### Special Logic
- If monthly data returns 404, auto-fallback to quarterly

---

## 8. SharePriceIndex

### API URL
```
https://sdmx.oecd.org/public/rest/data/OECD.SDD.STES,DSD_STES@DF_FINMARK,4.0/{COUNTRY}.{FREQ}.SHARE......?startPeriod={START}&endPeriod={END}&dimensionAtObservation=TIME_PERIOD&detail=dataonly
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `country` | No | string | `"united_states"` | ~52 countries + `"all"` |
| `start_date` | No | date | `1958-01-01` | Date |
| `end_date` | No | date | current year Dec 31 | Date |
| `frequency` | No | enum | (from standard) | `"monthly"→M`, `"quarter"→Q`, `"annual"→A` |

### Data Processing
- Standard CSV parse → rename → map country codes → parse dates → sort

---

## 9. Unemployment

### API URL
```
https://sdmx.oecd.org/public/rest/data/OECD.SDD.TPS,DSD_LFS@DF_IALFS_UNE_M,1.0/{COUNTRY}..._Z.{SEASONAL_ADJ}.{SEX}.{AGE}..{FREQ}?startPeriod={START}&endPeriod={END}&dimensionAtObservation=TIME_PERIOD&detail=dataonly
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `country` | No | string | `"united_states"` | ~45 countries + `"all"` |
| `start_date` | No | date | `1950-01-01` | Date |
| `end_date` | No | date | current year Dec 31 | Date |
| `frequency` | No | enum | (from standard) | `"monthly"→M`, `"quarter"→Q`, `"annual"→A` |
| `sex` | No | enum | `"total"` | `"total"→_T`, `"male"→M`, `"female"→F` |
| `age` | No | enum | `"total"` | `"total"→Y_GE15`, `"15-24"→Y15T24`, `"25+"→Y_GE25` |
| `seasonal_adjustment` | No | bool | `false` | `true→Y`, `false→N` |

### Data Processing
- **Divide value by 100** (unemployment rate as decimal)
- Map country codes, parse dates, filter nulls
- Replace NaN country values with "all"

---

# IMF Provider (8 Endpoints)

## IMF Architecture & Helpers

### Base URLs — Two Distinct API Families

**1. SDMX REST API v3.0** (for economic data: CPI, EconomicIndicators, DirectionOfTrade, AvailableIndicators)
```
https://api.imf.org/external/sdmx/3.0/data/dataflow/{AGENCY_ID}/{DATAFLOW}/+/{KEY}?{PARAMS}
```

**2. ArcGIS Feature Services** (for maritime/port data: PortInfo, PortVolume, MaritimeChokePointInfo, MaritimeChokePointVolume)
```
https://services9.arcgis.com/weJ1QsnbMYJlCHdG/arcgis/rest/services/{SERVICE}/FeatureServer/0/query?{PARAMS}
```

### SDMX API Pattern
The IMF SDMX API uses a **dimension-based key** where each position in a dot-separated key corresponds to a dimension in the Data Structure Definition (DSD). Dimensions are ordered by position from the DSD.

Example key: `USA.CPI._T.IX.M` (country.index_type.coicop.transformation.frequency)

Query params:
| Parameter | Description |
|---|---|
| `c[TIME_PERIOD]` | Date filter: `ge:2020-01-01+le:2025-12-31` |
| `dimensionAtObservation` | Always `TIME_PERIOD` |
| `detail` | `full` |
| `includeHistory` | `false` |
| `lastNObservations` | Limit records per series |

### ArcGIS API Pattern
Standard ArcGIS REST query:
```
?where={SQL_WHERE}&outFields=*&returnGeometry=false&f=json
```
or for GeoJSON:
```
?outFields=*&where=1%3D1&f=geojson
```

Pagination via `resultOffset` when `exceededTransferLimit` is true.

### Response Formats
- **SDMX endpoints**: Custom SDMX JSON parsed by `ImfQueryBuilder.fetch_data()` → returns `{data: [...], metadata: {...}}`
- **ArcGIS endpoints**: JSON with `features[].attributes` or `features[].properties` (GeoJSON)

### Metadata System (`ImfMetadata` singleton)
- Pre-cached metadata loaded from `assets/imf_cache.pkl.gz` (gzip-pickled)
- Contains: dataflows, data structures, concept schemes, codelists, hierarchies, constraints
- Provides: `search_indicators()`, `get_dataflow_parameters()`, `get_available_constraints()`, `get_dataflow_hierarchies()`
- Thread-safe singleton pattern

### Go Implementation Note for Metadata
The compressed pickle cache won't work in Go. Options:
1. Convert to JSON at build time and embed with `embed.go`
2. Fetch metadata on first use from IMF API and cache locally
3. Pre-generate Go maps from the pickle data

### Third-Party Dependencies (Python)
- `async_lru` (caching async calls)
- `gzip`, `pickle` (metadata cache)
- Standard `aiohttp` session via OpenBB helpers
- No special SSL requirements (unlike OECD)

---

## 1. AvailableIndicators

### API
No direct HTTP call — queries the local `ImfMetadata` singleton cache.

### Query Parameters
| Parameter | Required | Type | Default | Description |
|---|---|---|---|---|
| `query` | No | string | None | Search text with AND (+), OR (\|), quoted phrases; semicolons separate phrases |
| `dataflows` | No | string/list | None | Filter by IMF dataflow IDs (semicolon-separated) |
| `keywords` | No | string/list | None | Single-word filters; prefix with "not" to exclude |

### Response Fields
| Field | Type | Description |
|---|---|---|
| `symbol` | string | `{dataflow_id}::{indicator_code}` |
| `description` | string | Label of the indicator |
| `agency_id` | string | Agency responsible |
| `dataflow_id` | string | IMF dataflow ID |
| `dataflow_name` | string | Dataflow name |
| `structure_id` | string | DSD ID |
| `dimension_id` | string | Dimension ID in the DSD |
| `long_description` | string | Detailed description |
| `member_of` | list[str] | Table symbols this indicator belongs to |

### Data Processing
- Calls `metadata.search_indicators(query, dataflows, keywords)`
- Constructs `symbol` as `dataflow_id::indicator_code` for each result

---

## 2. ConsumerPriceIndex (CPI)

### API URL (via ImfQueryBuilder)
```
https://api.imf.org/external/sdmx/3.0/data/dataflow/{AGENCY}/CPI/+/{KEY}?c[TIME_PERIOD]=ge:{START}+le:{END}&dimensionAtObservation=TIME_PERIOD&detail=full&includeHistory=false
```

### Key Dimensions (ordered by DSD position)
```
{COUNTRY}.{INDEX_TYPE}.{COICOP_1999}.{TYPE_OF_TRANSFORMATION}.{FREQUENCY}
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `country` | Yes | string | `"united_states"` | ISO3 codes or snake_case names; ~196 countries + `"*"` |
| `start_date` | No | date | None | Date |
| `end_date` | No | date | None | Date |
| `frequency` | No | enum | (from standard) | `"monthly"→M`, `"quarter"→Q`, `"annual"→A` |
| `harmonized` | No | bool | `false` | `true→HICP`, `false→CPI` |
| `transform` | No | enum | `"index"` | See transform map below |
| `expenditure` | No | enum | `"total"` | 14 COICOP categories + `"all"` |
| `limit` | No | int | None | Max records per series |

### Transform Map (IMF CPI)
```
index          → IX
period         → POP_PCH_PA_PT
yoy            → YOY_PCH_PA_PT
ref_index      → SRP_IX
ref_period     → SRP_POP_PCH_PA_PT
ref_yoy        → SRP_YOY_PCH_PA_PT
weight         → WGT
weight_percent → WGT_PT
```

### Response Processing
- Data comes from `ImfQueryBuilder.fetch_data("CPI", ...)`
- Returns `{data: [...], metadata: {...}}`
- Each record has: `TIME_PERIOD`, `OBS_VALUE`, `COUNTRY`, `COICOP_1999`, `TYPE_OF_TRANSFORMATION`, etc.
- If unit contains "percent", divide value by 100, set multiplier=100
- Build title: `"{freq} {index_type} - {expenditure} - {transformation}"`
- Sort by date → country → expenditure order

---

## 3. DirectionOfTrade

### API URL (via `imts_query` helper)
```
https://api.imf.org/external/sdmx/3.0/data/dataflow/{AGENCY}/IMTS/+/{KEY}?{PARAMS}
```

### Key Dimensions (IMTS dataflow)
```
{COUNTRY}.{INDICATOR}.{FREQUENCY}.{COUNTERPART_COUNTRY}
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `country` | Yes | string | — | ISO3 codes, snake_case names, or `"*"` |
| `counterpart` | Yes | string | — | Same as country + `"all"` for wildcard |
| `direction` | No | enum | — | `"exports"→XG_FOB_USD`, `"imports"→MG_CIF_USD`, `"balance"→TBG_USD`, `"all"→*` |
| `frequency` | No | enum | `"annual"` | `"annual"→A`, `"quarter"→Q`, `"month"→M` |
| `start_date` | No | date | None | Date |
| `end_date` | No | date | None | Date |
| `limit` | No | int | None | lastNObservations |

### Country Resolution
- Uses `ImfMetadata.get_dataflow_parameters("IMTS")["COUNTRY"]` for valid codes
- Supports common aliases: `"world"→G001`, `"euro_area"→G163`, `"eu"→G998`
- Dynamically loaded from metadata, not hardcoded

### Response Fields
```
date, country, country_code, counterpart, counterpart_code, symbol, title, value, scale, unit, unit_multiplier
```

---

## 4. EconomicIndicators

### API URL (via ImfQueryBuilder or ImfTableBuilder)
Same SDMX pattern but supports **any** IMF dataflow.

### Symbol Format
```
{DATAFLOW}::{IDENTIFIER}
```
- Identifier starting with `H_` = hierarchical table request
- Otherwise = individual indicator request

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `symbol` | Yes | string | — | `"BOP::H_BOP_BOP_AGG_STANDARD_PRESENTATION"`, `"WEO::NGDP_RPCH"`, etc. |
| `country` | Yes | string | — | ISO3 codes, wildcard `"*"` |
| `start_date` | No | date | None | Date |
| `end_date` | No | date | None | Date |
| `frequency` | No | string | None | `"annual"→A`, `"quarter"→Q`, `"month"→M`, `"all"→*` |
| `transform` | No | string | None | Dataflow-specific (resolved via `detect_transform_dimension()`) |
| `dimension_values` | No | list[str] | None | `"DIM_ID:DIM_VALUE"` format for extra filters |
| `limit` | No | int | None | Max records |
| `pivot` | No | bool | false | Pivot to presentation view |

### Two Modes
1. **Table mode**: Uses `ImfTableBuilder.get_table()` — supports hierarchical views with parent/child, ordering, indentation
2. **Indicator mode**: Uses `ImfQueryBuilder.fetch_data()` — flat list of observations

### Key Construction
Dynamic per dataflow. Dimension keys are built from the DSD:
```python
dimensions = sorted(dsd["dimensions"], key=lambda x: x["position"])
key = ".".join([value_for_dim or "*" for dim in dimensions])
```

### Known Table IDs (30+ pre-configured)
```
bop_standard, bop_analytic, dip, iip_aggregated, eer, irfcl_reserve_assets,
fsi_core_and_additional, mfs_monetary_aggs, gfs_balance, gdp_annual_expenditure,
cpi, fas_indicator_by_country, isora_indicators_by_topic, ...
```

### Response Processing
- Table mode: returns hierarchy with `order`, `level`, `parent_id`, `is_category_header`
- Indicator mode: returns flat records with dimensions translated to human-readable labels
- Each record: `TIME_PERIOD`, `OBS_VALUE`, `COUNTRY`/`JURISDICTION`, `series_id`, `title`, etc.

---

## 5. MaritimeChokePointInfo

### API URL
```
https://services9.arcgis.com/weJ1QsnbMYJlCHdG/arcgis/rest/services/PortWatch_chokepoints_database/FeatureServer/0/query?outFields=*&where=1%3D1&f=geojson
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `theme` | No | enum | None | `"dark"`, `"light"` (charting only) |

No filtering — returns all 24 global chokepoints.

### Response: GeoJSON
```json
{
  "features": [
    {
      "properties": {
        "portid": "chokepoint1",
        "portname": "Suez Canal",
        "lat": 30.45,
        "lon": 32.35,
        "vessel_count_total": 18000,
        "vessel_count_tanker": 4500,
        "vessel_count_container": 5200,
        "vessel_count_general_cargo": 2100,
        "vessel_count_dry_bulk": 3400,
        "vessel_count_RoRo": 2800,
        "industry_top1": "Petroleum",
        "industry_top2": "Chemicals",
        "industry_top3": "Electronics"
      }
    }
  ]
}
```

---

## 6. MaritimeChokePointVolume

### API URL
```
https://services9.arcgis.com/weJ1QsnbMYJlCHdG/arcgis/rest/services/Daily_Chokepoints_Data/FeatureServer/0/query?where=portid%20%3D%20%27{CHOKEPOINT_ID}%27AND%20date%20>%3D%20TIMESTAMP%20%27{START}%2000%3A00%3A00%27%20AND%20date%20<%3D%20TIMESTAMP%20%27{END}%2000%3A00%3A00%27&outFields=*&orderByFields=date&returnZ=true&resultOffset={OFFSET}&resultRecordCount=1000&maxRecordCountFactor=5&outSR=&f=json
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `chokepoint` | No | string | None (all) | 24 chokepoint names or IDs; multi-value with `,` |
| `start_date` | No | date | None | Date |
| `end_date` | No | date | None | Date |

### Chokepoint IDs (24 total)
```
chokepoint1→Suez Canal, chokepoint2→Panama Canal, chokepoint3→Bosporus Strait,
chokepoint4→Bab el-Mandeb Strait, chokepoint5→Malacca Strait, chokepoint6→Strait of Hormuz,
chokepoint7→Cape of Good Hope, chokepoint8→Gibraltar Strait, ...
chokepoint24→Mona Passage
```

### Response: JSON (ArcGIS Feature)
```json
{
  "features": [
    {
      "attributes": {
        "portid": "CHOKEPOINT1",
        "portname": "Suez Canal",
        "year": 2024, "month": 1, "day": 15,
        "n_total": 55, "n_cargo": 35, "n_tanker": 20,
        "n_container": 15, "n_general_cargo": 8, "n_dry_bulk": 10, "n_roro": 2,
        "capacity": 1850000.0,
        "capacity_cargo": 1200000.0, "capacity_tanker": 650000.0, ...
      }
    }
  ]
}
```

### Pagination
- 1000 records per page
- Check `exceededTransferLimit` flag
- Increment `resultOffset`

### Data Processing
- Combine year/month/day into date string `YYYY-MM-DD`
- Remove raw year/month/day/date/ObjectId fields

---

## 7. PortInfo

### API URL
```
https://services9.arcgis.com/weJ1QsnbMYJlCHdG/arcgis/rest/services/PortWatch_ports_database/FeatureServer/0/query?where=1%3D1&outFields=*&returnGeometry=false&outSR=&f=json
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `continent` | No | enum | None | `"north_america"`, `"europe"`, `"asia_pacific"`, `"south_america"`, `"africa"` |
| `country` | No | string | None | ISO3 country code (supersedes continent) |
| `limit` | No | int | None | Limit results (by vessel_count_total ranking) |

### Response: JSON
```json
{
  "features": [
    {
      "attributes": {
        "portid": "port1114",
        "portname": "Shanghai",
        "fullname": "Shanghai Port, China",
        "ISO3": "CHN",
        "countrynoaccents": "China",
        "continent": "Asia & Pacific",
        "lat": 31.23, "lon": 121.47,
        "vessel_count_total": 85000,
        "vessel_count_tanker": 12000,
        "vessel_count_container": 35000,
        "vessel_count_general_cargo": 15000,
        "vessel_count_dry_bulk": 18000,
        "vessel_count_RoRo": 5000,
        "industry_top1": "Electronics",
        "share_country_maritime_import": 25.5,
        "share_country_maritime_export": 30.2
      }
    }
  ]
}
```

### Pagination
Same pattern — offset-based when `exceededTransferLimit` is true.

### Data Processing
- Sort by `vessel_count_total` descending
- Filter by country ISO3 or continent label
- Normalize `share_country_maritime_*` by dividing by 100

---

## 8. PortVolume

### API URL
```
https://services9.arcgis.com/weJ1QsnbMYJlCHdG/arcgis/rest/services/Daily_Trade_Data/FeatureServer/0/query?where=portid%20%3D%20%27{PORT_ID}%27AND%20date%20>%3D%20TIMESTAMP%20%27{START}%27%20AND%20date%20<%3D%20TIMESTAMP%20%27{END}%27&outFields=*&orderByFields=date&resultOffset={OFFSET}&resultRecordCount=2000&f=json
```

### Query Parameters
| Parameter | Required | Type | Default | Values |
|---|---|---|---|---|
| `port_code` | No* | string | `"port1114"` | Port ID(s), comma-separated |
| `country` | No* | string | None | ISO3 code (resolves to all ports in country) |
| `start_date` | No | date | None | Min: 2019-01-01 |
| `end_date` | No | date | None | Date |

*At least one of `port_code` or `country` required. `country` is resolved to port IDs.

### Response Fields (per daily record)
```
date, portid, portname, ISO3,
portcalls, portcalls_tanker, portcalls_container, portcalls_general_cargo, portcalls_dry_bulk, portcalls_roro,
import (total), import_cargo, import_tanker, import_container, import_general_cargo, import_dry_bulk, import_roro,
export (total), export_cargo, export_tanker, export_container, export_general_cargo, export_dry_bulk, export_roro
```

All import/export values in metric tons.

---

# Summary & Go Implementation Notes

## API Base URLs

| Provider | API Type | Base URL |
|---|---|---|
| **OECD** | SDMX REST | `https://sdmx.oecd.org/public/rest/data/{AGENCY},{DSD},{VER}/{KEY}?{PARAMS}` |
| **IMF SDMX** | SDMX REST v3 | `https://api.imf.org/external/sdmx/3.0/data/dataflow/{AGENCY}/{DATAFLOW}/+/{KEY}?{PARAMS}` |
| **IMF PortWatch** | ArcGIS REST | `https://services9.arcgis.com/weJ1QsnbMYJlCHdG/arcgis/rest/services/{SERVICE}/FeatureServer/0/query?{PARAMS}` |

## Response Format Summary

| Provider | Endpoint | Format | Key Fields |
|---|---|---|---|
| OECD (all 9) | SDMX | CSV (or XML) | `REF_AREA`, `TIME_PERIOD`, `OBS_VALUE` |
| IMF CPI/EconIndicators/DOT | SDMX | JSON | `COUNTRY`, `TIME_PERIOD`, `OBS_VALUE`, dimensions |
| IMF AvailableIndicators | Local cache | Dict | Metadata fields |
| IMF ChokePointInfo | ArcGIS | GeoJSON | `features[].properties` |
| IMF ChokePointVolume | ArcGIS | JSON | `features[].attributes` |
| IMF PortInfo | ArcGIS | JSON | `features[].attributes` |
| IMF PortVolume | ArcGIS | JSON | `features[].attributes` |

## Go-Specific Considerations

### 1. OECD SSL
```go
transport := &http.Transport{
    TLSClientConfig: &tls.Config{
        MinVersion: tls.VersionTLS10,
    },
}
client := &http.Client{Transport: transport}
```

### 2. CSV Parsing (OECD)
```go
// Use encoding/csv to parse SDMX CSV responses
reader := csv.NewReader(strings.NewReader(responseBody))
records, _ := reader.ReadAll()
// Find column indices for REF_AREA, TIME_PERIOD, OBS_VALUE
```

### 3. ArcGIS Pagination (IMF)
```go
type ArcGISResponse struct {
    Features             []Feature `json:"features"`
    ExceededTransferLimit bool      `json:"exceededTransferLimit"`
}
// Loop with resultOffset until ExceededTransferLimit == false
```

### 4. SDMX Key Building (IMF)
```go
// Dimensions ordered by position from DSD
// Each segment: specific value, "+" for multi-value, "*" for wildcard
key := strings.Join(dimensionValues, ".")
url := fmt.Sprintf("https://api.imf.org/external/sdmx/3.0/data/dataflow/%s/%s/+/%s", agency, dataflow, key)
```

### 5. Country Code Maps
Embed as Go maps. Example:
```go
var OECDCountryToCodeGDP = map[string]string{
    "united_states": "USA",
    "united_kingdom": "GBR",
    // ...
}
```

### 6. Date Parsing
```go
func ParseOECDDate(input string) time.Time {
    if strings.Contains(input, "Q") {
        // Parse quarterly: "2023-Q1" → 2023-01-01
    } else if len(input) == 4 {
        // Annual: "2023" → 2023-01-01
    } else {
        // Monthly: "2023-05" → 2023-05-01
    }
}
```

### 7. No External Dependencies Required
- All OECD endpoints return CSV — Go `encoding/csv` is built-in
- IMF SDMX returns JSON — Go `encoding/json` is built-in
- ArcGIS returns JSON — same
- XML parsing via `encoding/xml` if needed for OECD fallback

### 8. Concurrency
- IMF PortVolume/ChokePointVolume fetch multiple ports/chokepoints concurrently (Python uses `asyncio.gather`)
- Go equivalent: goroutines + `sync.WaitGroup` or `errgroup`

## Priority for Implementation (by complexity)

1. **Simple** (static URL, CSV response): CLI, SharePriceIndex, CountryInterestRates, HousePriceIndex, Unemployment
2. **Medium** (URL construction logic): GdpNominal, GdpReal, GdpForecast, OECD CPI
3. **Medium** (ArcGIS JSON, pagination): MaritimeChokePointInfo, PortInfo, MaritimeChokePointVolume, PortVolume
4. **Complex** (dynamic metadata, DSD parsing): IMF CPI, DirectionOfTrade, EconomicIndicators, AvailableIndicators
