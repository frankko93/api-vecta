package reports

// CompanyConfig contains company configuration metadata for dynamic UI rendering
type CompanyConfig struct {
	MiningType string   `json:"mining_type"` // "open_pit", "underground", "both"
	Minerals   []string `json:"minerals"`    // List of mineral codes: ["AU", "AG", "CU", etc.]
}

// SummaryReport represents the complete summary report for a company
// Always returns full year - frontend will filter/aggregate as needed
type SummaryReport struct {
	CompanyID   int64          `json:"company_id"`
	CompanyName string         `json:"company_name"`
	Year        int            `json:"year"`
	Config      *CompanyConfig `json:"config,omitempty"` // Company configuration for dynamic UI
	Months      []MonthlyData  `json:"months"`           // Always 12 months (or empty if no data)
	Coverage    *DataCoverage  `json:"coverage,omitempty"`
}

// DataCoverage indicates which months have data loaded
type DataCoverage struct {
	ActualMonths      []int `json:"actual_months"`
	BudgetMonths      []int `json:"budget_months"`
	ActualLastMonth   int   `json:"actual_last_month"`
	BudgetLastMonth   int   `json:"budget_last_month"`
	ActualIsPartial   bool  `json:"actual_is_partial"`
	BudgetIsPartial   bool  `json:"budget_is_partial"`
	HasAnyActual      bool  `json:"has_any_actual"`
	HasAnyBudget      bool  `json:"has_any_budget"`
	HasCompleteActual bool  `json:"has_complete_actual"`
	HasCompleteBudget bool  `json:"has_complete_budget"`
}

// MonthlyData represents data for a single month
type MonthlyData struct {
	Month    string        `json:"month"` // "2025-01"
	Actual   *DataSet      `json:"actual"`
	Budget   *DataSet      `json:"budget"`
	Variance *VarianceData `json:"variance,omitempty"` // Variance calculations (Actual - Budget)
	YTD      *YTDData      `json:"ytd,omitempty"`      // Year-to-date calculations
}

// YTDData represents year-to-date aggregated data
type YTDData struct {
	Actual   *DataSet      `json:"actual"`
	Budget   *DataSet      `json:"budget"`
	Variance *VarianceData `json:"variance,omitempty"`
}

// DataSet contains all metrics for actual or budget
type DataSet struct {
	Mining     MiningMetrics     `json:"mining"`
	Processing ProcessingMetrics `json:"processing"`
	Production ProductionMetrics `json:"production"`
	Costs      CostMetrics       `json:"costs"`
	NSR        NSRMetrics        `json:"nsr"`
	CAPEX      CAPEXMetrics      `json:"capex"`
	CashCost   CashCostMetrics   `json:"cash_cost"`
}

// MiningMetrics represents mining data
type MiningMetrics struct {
	// Ore breakdown by mine type
	OpenPitOreT     float64 `json:"open_pit_ore_t"`
	UndergroundOreT float64 `json:"underground_ore_t"`
	OreMinedT       float64 `json:"ore_mined_t"` // Total = OpenPit + Underground

	// Waste and ratios
	WasteMinedT    float64 `json:"waste_mined_t"`
	StrippingRatio float64 `json:"stripping_ratio"` // Waste / OpenPit Ore

	// Mining grades by mine type (g/t)
	MiningGradeSilverGpt      float64 `json:"mining_grade_silver_gpt"`
	MiningGradeGoldGpt        float64 `json:"mining_grade_gold_gpt"`
	OpenPitGradeSilverGpt     float64 `json:"open_pit_grade_silver_gpt"`
	UndergroundGradeSilverGpt float64 `json:"underground_grade_silver_gpt"`
	OpenPitGradeGoldGpt       float64 `json:"open_pit_grade_gold_gpt"`
	UndergroundGradeGoldGpt   float64 `json:"underground_grade_gold_gpt"`

	// Developments breakdown (meters)
	PrimaryDevelopmentM       float64 `json:"primary_development_m"`
	SecondaryDevelopmentOpexM float64 `json:"secondary_development_opex_m"`
	ExpansionaryDevelopmentM  float64 `json:"expansionary_development_m"`
	DevelopmentsM             float64 `json:"developments_m"` // Total

	// Headcount
	FullTimeEmployees int `json:"full_time_employees"`
	Contractors       int `json:"contractors"`
	TotalHeadcount    int `json:"total_headcount"`

	HasData bool `json:"has_data"`
}

// ProcessingMetrics represents processing data
type ProcessingMetrics struct {
	TotalTonnesProcessed  float64 `json:"total_tonnes_processed"`
	FeedGradeSilverGpt    float64 `json:"feed_grade_silver_gpt"`
	FeedGradeGoldGpt      float64 `json:"feed_grade_gold_gpt"`
	RecoveryRateSilverPct float64 `json:"recovery_rate_silver_pct"`
	RecoveryRateGoldPct   float64 `json:"recovery_rate_gold_pct"`
	HasData               bool    `json:"has_data"`
}

// ProductionMetrics represents calculated production
type ProductionMetrics struct {
	TotalProductionSilverOz float64 `json:"total_production_silver_oz"`
	TotalProductionGoldOz   float64 `json:"total_production_gold_oz"`
	PayableSilverOz         float64 `json:"payable_silver_oz"`
	PayableGoldOz           float64 `json:"payable_gold_oz"`
	DoreProductionOz        float64 `json:"dore_production_oz"` // Total dore (Silver + Gold)
	HasData                 bool    `json:"has_data"`
}

// CostMetrics represents cost breakdown
type CostMetrics struct {
	Mine                  float64 `json:"mine"`
	Processing            float64 `json:"processing"`
	GA                    float64 `json:"ga"`
	TransportShipping     float64 `json:"transport_shipping"`
	InventoryVariations   float64 `json:"inventory_variations"`
	ProductionBasedCosts  float64 `json:"production_based_costs"`
	ProductionBasedMargin float64 `json:"production_based_margin"`
	HasData               bool    `json:"has_data"`
}

// NSRMetrics represents Net Smelter Return metrics
type NSRMetrics struct {
	NSRDore                 float64 `json:"nsr_dore"`
	Streaming               float64 `json:"streaming"`                  // Streaming agreement value (usually negative)
	PBRRevenue              float64 `json:"pbr_revenue"`                // NSR Dore + Streaming
	ShippingSelling         float64 `json:"shipping_selling"`
	SalesTaxes              float64 `json:"sales_taxes"`                // Sales taxes (split from combined)
	Royalties               float64 `json:"royalties"`                  // Royalties (split from combined)
	SalesTaxesRoyalties     float64 `json:"sales_taxes_royalties"`      // Calculated: SalesTaxes + Royalties
	OtherSalesDeductions    float64 `json:"other_sales_deductions"`     // Other sales deductions
	SmeltingRefiningCharges float64 `json:"smelting_refining_charges"`  // Treatment + Refining charges
	NetSmelterReturn        float64 `json:"net_smelter_return"`
	GoldCredit              float64 `json:"gold_credit"`                // Gold by-product credit (negative)
	SilverPricePerOz        float64 `json:"silver_price_per_oz"`        // Realized silver price $/oz
	GoldPricePerOz          float64 `json:"gold_price_per_oz"`          // Realized gold price $/oz
	NSRPerTonne             float64 `json:"nsr_per_tonne"`
	TotalCostPerTonne       float64 `json:"total_cost_per_tonne"`
	MarginPerTonne          float64 `json:"margin_per_tonne"`
	HasData                 bool    `json:"has_data"`
}

// CAPEXMetrics represents capital expenditure metrics
type CAPEXMetrics struct {
	Sustaining                      float64 `json:"sustaining"`
	Project                         float64 `json:"project"`
	Leasing                         float64 `json:"leasing"`
	AccretionOfMineClosureLiability float64 `json:"accretion_of_mine_closure_liability"` // New field
	Total                           float64 `json:"total"`
	ProductionBasedMargin           float64 `json:"production_based_margin"`
	PBRNetCashFlow                  float64 `json:"pbr_net_cash_flow"`
	HasData                         bool    `json:"has_data"`
}

// CashCostMetrics represents cash cost and AISC metrics
type CashCostMetrics struct {
	CashCostPerOzSilver    float64 `json:"cash_cost_per_oz_silver"`
	AISCPerOzSilver        float64 `json:"aisc_per_oz_silver"`
	CashCostsSilver        float64 `json:"cash_costs_silver"`         // Total cash costs before dividing by ounces
	AISCSilver             float64 `json:"aisc_silver"`               // Total AISC before dividing by ounces
	GoldCredit             float64 `json:"gold_credit"`
	SustainingCapitalPerOz float64 `json:"sustaining_capital_per_oz"` // Sustaining CAPEX / payable silver oz
	HasData                bool    `json:"has_data"`
}

// ComparisonData for YTD comparisons
type ComparisonData struct {
	Actual   *DataSet      `json:"actual"`
	Budget   *DataSet      `json:"budget"`
	Variance *VarianceData `json:"variance,omitempty"` // Variance calculations
}

// VarianceData contains variance calculations for all metrics
type VarianceData struct {
	Mining     MiningVariance     `json:"mining"`
	Processing ProcessingVariance `json:"processing"`
	Production ProductionVariance `json:"production"`
	Costs      CostVariance       `json:"costs"`
	NSR        NSRVariance        `json:"nsr"`
	CAPEX      CAPEXVariance      `json:"capex"`
	CashCost   CashCostVariance   `json:"cash_cost"`
}

// Variance helper structs for each metric type
type MiningVariance struct {
	// Ore breakdown
	OpenPitOreT     VarianceMetric `json:"open_pit_ore_t"`
	UndergroundOreT VarianceMetric `json:"underground_ore_t"`
	OreMinedT       VarianceMetric `json:"ore_mined_t"`

	// Waste and ratios
	WasteMinedT    VarianceMetric `json:"waste_mined_t"`
	StrippingRatio VarianceMetric `json:"stripping_ratio"`

	// Mining grades
	MiningGradeSilverGpt      VarianceMetric `json:"mining_grade_silver_gpt"`
	MiningGradeGoldGpt        VarianceMetric `json:"mining_grade_gold_gpt"`
	OpenPitGradeSilverGpt     VarianceMetric `json:"open_pit_grade_silver_gpt"`
	UndergroundGradeSilverGpt VarianceMetric `json:"underground_grade_silver_gpt"`
	OpenPitGradeGoldGpt       VarianceMetric `json:"open_pit_grade_gold_gpt"`
	UndergroundGradeGoldGpt   VarianceMetric `json:"underground_grade_gold_gpt"`

	// Developments breakdown
	PrimaryDevelopmentM       VarianceMetric `json:"primary_development_m"`
	SecondaryDevelopmentOpexM VarianceMetric `json:"secondary_development_opex_m"`
	ExpansionaryDevelopmentM  VarianceMetric `json:"expansionary_development_m"`
	DevelopmentsM             VarianceMetric `json:"developments_m"`

	// Headcount
	FullTimeEmployees VarianceMetric `json:"full_time_employees"`
	Contractors       VarianceMetric `json:"contractors"`
	TotalHeadcount    VarianceMetric `json:"total_headcount"`
}

type ProcessingVariance struct {
	TotalTonnesProcessed  VarianceMetric `json:"total_tonnes_processed"`
	FeedGradeSilverGpt    VarianceMetric `json:"feed_grade_silver_gpt"`
	FeedGradeGoldGpt      VarianceMetric `json:"feed_grade_gold_gpt"`
	RecoveryRateSilverPct VarianceMetric `json:"recovery_rate_silver_pct"`
	RecoveryRateGoldPct   VarianceMetric `json:"recovery_rate_gold_pct"`
}

type ProductionVariance struct {
	TotalProductionSilverOz VarianceMetric `json:"total_production_silver_oz"`
	TotalProductionGoldOz   VarianceMetric `json:"total_production_gold_oz"`
	PayableSilverOz         VarianceMetric `json:"payable_silver_oz"`
	PayableGoldOz           VarianceMetric `json:"payable_gold_oz"`
	DoreProductionOz        VarianceMetric `json:"dore_production_oz"`
}

type CostVariance struct {
	Mine                  VarianceMetric `json:"mine"`
	Processing            VarianceMetric `json:"processing"`
	GA                    VarianceMetric `json:"ga"`
	TransportShipping     VarianceMetric `json:"transport_shipping"`
	InventoryVariations   VarianceMetric `json:"inventory_variations"`
	ProductionBasedCosts  VarianceMetric `json:"production_based_costs"`
	ProductionBasedMargin VarianceMetric `json:"production_based_margin"`
}

type NSRVariance struct {
	NSRDore                 VarianceMetric `json:"nsr_dore"`
	Streaming               VarianceMetric `json:"streaming"`
	PBRRevenue              VarianceMetric `json:"pbr_revenue"`
	ShippingSelling         VarianceMetric `json:"shipping_selling"`
	SalesTaxes              VarianceMetric `json:"sales_taxes"`
	Royalties               VarianceMetric `json:"royalties"`
	SalesTaxesRoyalties     VarianceMetric `json:"sales_taxes_royalties"`
	OtherSalesDeductions    VarianceMetric `json:"other_sales_deductions"`
	SmeltingRefiningCharges VarianceMetric `json:"smelting_refining_charges"`
	NetSmelterReturn        VarianceMetric `json:"net_smelter_return"`
	GoldCredit              VarianceMetric `json:"gold_credit"`
	SilverPricePerOz        VarianceMetric `json:"silver_price_per_oz"`
	GoldPricePerOz          VarianceMetric `json:"gold_price_per_oz"`
	NSRPerTonne             VarianceMetric `json:"nsr_per_tonne"`
	TotalCostPerTonne       VarianceMetric `json:"total_cost_per_tonne"`
	MarginPerTonne          VarianceMetric `json:"margin_per_tonne"`
}

type CAPEXVariance struct {
	Sustaining                      VarianceMetric `json:"sustaining"`
	Project                         VarianceMetric `json:"project"`
	Leasing                         VarianceMetric `json:"leasing"`
	AccretionOfMineClosureLiability VarianceMetric `json:"accretion_of_mine_closure_liability"`
	Total                           VarianceMetric `json:"total"`
	ProductionBasedMargin           VarianceMetric `json:"production_based_margin"`
	PBRNetCashFlow                  VarianceMetric `json:"pbr_net_cash_flow"`
}

type CashCostVariance struct {
	CashCostPerOzSilver    VarianceMetric `json:"cash_cost_per_oz_silver"`
	AISCPerOzSilver        VarianceMetric `json:"aisc_per_oz_silver"`
	CashCostsSilver        VarianceMetric `json:"cash_costs_silver"`
	AISCSilver             VarianceMetric `json:"aisc_silver"`
	GoldCredit             VarianceMetric `json:"gold_credit"`
	SustainingCapitalPerOz VarianceMetric `json:"sustaining_capital_per_oz"`
}

// VarianceMetric represents variance calculation for a single metric
type VarianceMetric struct {
	Actual      float64 `json:"actual"`
	Budget      float64 `json:"budget"`
	Variance    float64 `json:"variance"`     // Actual - Budget (Fav/Unf)
	VariancePct float64 `json:"variance_pct"` // ((Actual - Budget) / Budget) * 100
}
