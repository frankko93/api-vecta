package reports

// SummaryReport represents the complete summary report for a company
// Always returns full year - frontend will filter/aggregate as needed
type SummaryReport struct {
	CompanyID   int64         `json:"company_id"`
	CompanyName string        `json:"company_name"`
	Year        int           `json:"year"`
	Months      []MonthlyData `json:"months"` // Always 12 months (or empty if no data)
}

// MonthlyData represents data for a single month
type MonthlyData struct {
	Month    string        `json:"month"` // "2025-01"
	Actual   *DataSet      `json:"actual"`
	Budget   *DataSet      `json:"budget"`
	Variance *VarianceData `json:"variance,omitempty"` // Variance calculations (Actual - Budget)
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
	OreMinedT     float64 `json:"ore_mined_t"`
	WasteMinedT   float64 `json:"waste_mined_t"`
	DevelopmentsM float64 `json:"developments_m"`
	HasData       bool    `json:"has_data"`
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
	ShippingSelling         float64 `json:"shipping_selling"`
	SalesTaxesRoyalties     float64 `json:"sales_taxes_royalties"`
	SmeltingRefiningCharges float64 `json:"smelting_refining_charges"` // Treatment + Refining charges separated
	NetSmelterReturn        float64 `json:"net_smelter_return"`
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
	CashCostPerOzSilver float64 `json:"cash_cost_per_oz_silver"`
	AISCPerOzSilver     float64 `json:"aisc_per_oz_silver"`
	CashCostsSilver     float64 `json:"cash_costs_silver"` // Total cash costs before dividing by ounces
	AISCSilver          float64 `json:"aisc_silver"`       // Total AISC before dividing by ounces
	GoldCredit          float64 `json:"gold_credit"`
	HasData             bool    `json:"has_data"`
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
	OreMinedT     VarianceMetric `json:"ore_mined_t"`
	WasteMinedT   VarianceMetric `json:"waste_mined_t"`
	DevelopmentsM VarianceMetric `json:"developments_m"`
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
	ShippingSelling         VarianceMetric `json:"shipping_selling"`
	SalesTaxesRoyalties     VarianceMetric `json:"sales_taxes_royalties"`
	SmeltingRefiningCharges VarianceMetric `json:"smelting_refining_charges"`
	NetSmelterReturn        VarianceMetric `json:"net_smelter_return"`
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
	CashCostPerOzSilver VarianceMetric `json:"cash_cost_per_oz_silver"`
	AISCPerOzSilver     VarianceMetric `json:"aisc_per_oz_silver"`
	CashCostsSilver     VarianceMetric `json:"cash_costs_silver"`
	AISCSilver          VarianceMetric `json:"aisc_silver"`
	GoldCredit          VarianceMetric `json:"gold_credit"`
}

// VarianceMetric represents variance calculation for a single metric
type VarianceMetric struct {
	Actual      float64 `json:"actual"`
	Budget      float64 `json:"budget"`
	Variance    float64 `json:"variance"`     // Actual - Budget (Fav/Unf)
	VariancePct float64 `json:"variance_pct"` // ((Actual - Budget) / Budget) * 100
}
