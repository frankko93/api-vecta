package reports

// PBRDetailReport represents detailed PBR report
type PBRDetailReport struct {
	CompanyID   int64            `json:"company_id"`
	CompanyName string           `json:"company_name"`
	Year        int              `json:"year"`
	Config      *CompanyConfig   `json:"config,omitempty"`
	Months      []PBRMonthlyData `json:"months"`
}

// PBRMonthlyData represents PBR data for a single month
type PBRMonthlyData struct {
	Month    string       `json:"month"` // "2025-01"
	Actual   *PBRDetail   `json:"actual"`
	Budget   *PBRDetail   `json:"budget"`
	Variance *PBRVariance `json:"variance,omitempty"`
}

// PBRDetail contains detailed PBR metrics
type PBRDetail struct {
	// Mining - Ore breakdown by mine type
	OpenPitOreT     float64 `json:"open_pit_ore_t"`
	UndergroundOreT float64 `json:"underground_ore_t"`
	OreMinedT       float64 `json:"ore_mined_t"` // Total

	// Mining - Waste and ratios
	WasteMinedT    float64 `json:"waste_mined_t"`
	StrippingRatio float64 `json:"stripping_ratio"` // Waste / OpenPit Ore
	WasteOreRatio  float64 `json:"waste_ore_ratio"` // Waste / Total Ore (calculated)
	TotalMoved     float64 `json:"total_moved"`     // Ore + Waste (calculated)

	// Mining - Grades by mine type
	MiningGradeSilverGpt      float64 `json:"mining_grade_silver_gpt"`
	MiningGradeGoldGpt        float64 `json:"mining_grade_gold_gpt"`
	OpenPitGradeSilverGpt     float64 `json:"open_pit_grade_silver_gpt"`
	UndergroundGradeSilverGpt float64 `json:"underground_grade_silver_gpt"`
	OpenPitGradeGoldGpt       float64 `json:"open_pit_grade_gold_gpt"`
	UndergroundGradeGoldGpt   float64 `json:"underground_grade_gold_gpt"`

	// Developments breakdown
	PrimaryDevelopmentM       float64 `json:"primary_development_m"`
	SecondaryDevelopmentOpexM float64 `json:"secondary_development_opex_m"`
	ExpansionaryDevelopmentM  float64 `json:"expansionary_development_m"`
	DevelopmentsM             float64 `json:"developments_m"` // Total

	// Processing
	TotalTonnesProcessed  float64 `json:"total_tonnes_processed"`
	FeedGradeSilverGpt    float64 `json:"feed_grade_silver_gpt"`
	FeedGradeGoldGpt      float64 `json:"feed_grade_gold_gpt"`
	RecoveryRateSilverPct float64 `json:"recovery_rate_silver_pct"`
	RecoveryRateGoldPct   float64 `json:"recovery_rate_gold_pct"`

	// Production (calculated)
	TotalProductionSilverOz float64 `json:"total_production_silver_oz"`
	TotalProductionGoldOz   float64 `json:"total_production_gold_oz"`

	// Headcount
	FullTimeEmployees int `json:"full_time_employees"`
	Contractors       int `json:"contractors"`
	TotalHeadcount    int `json:"total_headcount"`

	HasData bool `json:"has_data"`
}

// PBRVariance contains variance for PBR metrics
type PBRVariance struct {
	// Mining - Ore breakdown
	OpenPitOreT     VarianceMetric `json:"open_pit_ore_t"`
	UndergroundOreT VarianceMetric `json:"underground_ore_t"`
	OreMinedT       VarianceMetric `json:"ore_mined_t"`

	// Mining - Waste and ratios
	WasteMinedT    VarianceMetric `json:"waste_mined_t"`
	StrippingRatio VarianceMetric `json:"stripping_ratio"`
	WasteOreRatio  VarianceMetric `json:"waste_ore_ratio"`
	TotalMoved     VarianceMetric `json:"total_moved"`

	// Mining - Grades
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

	// Processing
	TotalTonnesProcessed  VarianceMetric `json:"total_tonnes_processed"`
	FeedGradeSilverGpt    VarianceMetric `json:"feed_grade_silver_gpt"`
	FeedGradeGoldGpt      VarianceMetric `json:"feed_grade_gold_gpt"`
	RecoveryRateSilverPct VarianceMetric `json:"recovery_rate_silver_pct"`
	RecoveryRateGoldPct   VarianceMetric `json:"recovery_rate_gold_pct"`

	// Production
	TotalProductionSilverOz VarianceMetric `json:"total_production_silver_oz"`
	TotalProductionGoldOz   VarianceMetric `json:"total_production_gold_oz"`

	// Headcount
	FullTimeEmployees VarianceMetric `json:"full_time_employees"`
	Contractors       VarianceMetric `json:"contractors"`
	TotalHeadcount    VarianceMetric `json:"total_headcount"`
}

// DoreDetailReport represents detailed Dore report
type DoreDetailReport struct {
	CompanyID   int64             `json:"company_id"`
	CompanyName string            `json:"company_name"`
	Year        int               `json:"year"`
	Config      *CompanyConfig    `json:"config,omitempty"`
	Months      []DoreMonthlyData `json:"months"`
}

// DoreMonthlyData represents Dore data for a single month
type DoreMonthlyData struct {
	Month    string        `json:"month"` // "2025-01"
	Actual   *DoreDetail   `json:"actual"`
	Budget   *DoreDetail   `json:"budget"`
	Variance *DoreVariance `json:"variance,omitempty"`
}

// DoreDetail contains detailed Dore metrics
type DoreDetail struct {
	// Production
	DoreProducedOz float64 `json:"dore_produced_oz"`
	SilverGradePct float64 `json:"silver_grade_pct"`
	GoldGradePct   float64 `json:"gold_grade_pct"`

	// Metal in Dore (before adjustments)
	MetalInDoreSilverOz float64 `json:"metal_in_dore_silver_oz"`
	MetalInDoreGoldOz   float64 `json:"metal_in_dore_gold_oz"`

	// Adjustments
	SilverAdjustmentOz float64 `json:"silver_adjustment_oz"`
	GoldAdjustmentOz   float64 `json:"gold_adjustment_oz"`

	// Metal Adjusted (after adjustments)
	MetalAdjustedSilverOz float64 `json:"metal_adjusted_silver_oz"`
	MetalAdjustedGoldOz   float64 `json:"metal_adjusted_gold_oz"`

	// Deductions
	AgDeductionsPct    float64 `json:"ag_deductions_pct"`
	AuDeductionsPct    float64 `json:"au_deductions_pct"`
	DeductionsSilverOz float64 `json:"deductions_silver_oz"`
	DeductionsGoldOz   float64 `json:"deductions_gold_oz"`

	// Payable Metal (after deductions)
	PayableSilverOz float64 `json:"payable_silver_oz"`
	PayableGoldOz   float64 `json:"payable_gold_oz"`

	// Prices
	PBRPriceSilver      float64 `json:"pbr_price_silver"`
	PBRPriceGold        float64 `json:"pbr_price_gold"`
	RealizedPriceSilver float64 `json:"realized_price_silver"`
	RealizedPriceGold   float64 `json:"realized_price_gold"`

	// Revenue
	GrossRevenueSilver float64 `json:"gross_revenue_silver"`
	GrossRevenueGold   float64 `json:"gross_revenue_gold"`
	GrossRevenueTotal  float64 `json:"gross_revenue_total"`

	// Charges
	TreatmentCharge      float64 `json:"treatment_charge"`
	RefiningDeductionsAu float64 `json:"refining_deductions_au"`
	TotalCharges         float64 `json:"total_charges"`

	// NSR
	NSRDore float64 `json:"nsr_dore"`

	HasData bool `json:"has_data"`
}

// DoreVariance contains variance for Dore metrics
type DoreVariance struct {
	DoreProducedOz        VarianceMetric `json:"dore_produced_oz"`
	SilverGradePct        VarianceMetric `json:"silver_grade_pct"`
	GoldGradePct          VarianceMetric `json:"gold_grade_pct"`
	MetalInDoreSilverOz   VarianceMetric `json:"metal_in_dore_silver_oz"`
	MetalInDoreGoldOz     VarianceMetric `json:"metal_in_dore_gold_oz"`
	SilverAdjustmentOz    VarianceMetric `json:"silver_adjustment_oz"`
	GoldAdjustmentOz      VarianceMetric `json:"gold_adjustment_oz"`
	MetalAdjustedSilverOz VarianceMetric `json:"metal_adjusted_silver_oz"`
	MetalAdjustedGoldOz   VarianceMetric `json:"metal_adjusted_gold_oz"`
	DeductionsSilverOz    VarianceMetric `json:"deductions_silver_oz"`
	DeductionsGoldOz      VarianceMetric `json:"deductions_gold_oz"`
	PayableSilverOz       VarianceMetric `json:"payable_silver_oz"`
	PayableGoldOz         VarianceMetric `json:"payable_gold_oz"`
	GrossRevenueSilver    VarianceMetric `json:"gross_revenue_silver"`
	GrossRevenueGold      VarianceMetric `json:"gross_revenue_gold"`
	GrossRevenueTotal     VarianceMetric `json:"gross_revenue_total"`
	TreatmentCharge       VarianceMetric `json:"treatment_charge"`
	RefiningDeductionsAu  VarianceMetric `json:"refining_deductions_au"`
	TotalCharges          VarianceMetric `json:"total_charges"`
	NSRDore               VarianceMetric `json:"nsr_dore"`
}

// OPEXDetailReport represents detailed OPEX report
type OPEXDetailReport struct {
	CompanyID     int64                          `json:"company_id"`
	CompanyName   string                         `json:"company_name"`
	Year          int                            `json:"year"`
	Config        *CompanyConfig                 `json:"config,omitempty"`
	Months        []OPEXMonthlyData              `json:"months"`
	ByCostCenter  map[string]OPEXCostCenterData  `json:"by_cost_center"`
	BySubcategory map[string]OPEXSubcategoryData `json:"by_subcategory"`
	ByExpenseType map[string]OPEXExpenseTypeData `json:"by_expense_type"`
}

// OPEXMonthlyData represents OPEX data for a single month
type OPEXMonthlyData struct {
	Month    string        `json:"month"` // "2025-01"
	Actual   *OPEXDetail   `json:"actual"`
	Budget   *OPEXDetail   `json:"budget"`
	Variance *OPEXVariance `json:"variance,omitempty"`
}

// OPEXDetail contains detailed OPEX metrics
type OPEXDetail struct {
	// By Cost Center
	Mine              float64 `json:"mine"`
	Processing        float64 `json:"processing"`
	GA                float64 `json:"ga"`
	TransportShipping float64 `json:"transport_shipping"`

	// Inventory
	InventoryVariations float64 `json:"inventory_variations"`

	// Total
	Total float64 `json:"total"`

	// Breakdown by subcategory
	BySubcategory map[string]float64 `json:"by_subcategory,omitempty"`

	// Breakdown by expense type (Labour, Materials, Third Party, Other)
	ByExpenseType map[string]float64 `json:"by_expense_type,omitempty"`

	HasData bool `json:"has_data"`
}

// OPEXVariance contains variance for OPEX metrics
type OPEXVariance struct {
	Mine                VarianceMetric `json:"mine"`
	Processing          VarianceMetric `json:"processing"`
	GA                  VarianceMetric `json:"ga"`
	TransportShipping   VarianceMetric `json:"transport_shipping"`
	InventoryVariations VarianceMetric `json:"inventory_variations"`
	Total               VarianceMetric `json:"total"`
}

// OPEXCostCenterData represents OPEX aggregated by cost center
type OPEXCostCenterData struct {
	CostCenter string         `json:"cost_center"`
	Actual     float64        `json:"actual"`
	Budget     float64        `json:"budget"`
	Variance   VarianceMetric `json:"variance"`
}

// OPEXSubcategoryData represents OPEX aggregated by subcategory
type OPEXSubcategoryData struct {
	Subcategory string         `json:"subcategory"`
	CostCenter  string         `json:"cost_center"` // Which cost center this subcategory belongs to
	Actual      float64        `json:"actual"`
	Budget      float64        `json:"budget"`
	Variance    VarianceMetric `json:"variance"`
}

// OPEXExpenseTypeData represents OPEX aggregated by expense type
type OPEXExpenseTypeData struct {
	ExpenseType string         `json:"expense_type"`
	Actual      float64        `json:"actual"`
	Budget      float64        `json:"budget"`
	Variance    VarianceMetric `json:"variance"`
}

// CAPEXDetailReport represents detailed CAPEX report
type CAPEXDetailReport struct {
	CompanyID   int64                        `json:"company_id"`
	CompanyName string                       `json:"company_name"`
	Year        int                          `json:"year"`
	Config      *CompanyConfig               `json:"config,omitempty"`
	Months      []CAPEXMonthlyData           `json:"months"`
	ByType      map[string]CAPEXTypeData     `json:"by_type"`
	ByCategory  map[string]CAPEXCategoryData `json:"by_category"`
}

// CAPEXMonthlyData represents CAPEX data for a single month
type CAPEXMonthlyData struct {
	Month    string               `json:"month"` // "2025-01"
	Actual   *CAPEXDetail         `json:"actual"`
	Budget   *CAPEXDetail         `json:"budget"`
	Variance *CAPEXVarianceDetail `json:"variance,omitempty"`
}

// CAPEXDetail contains detailed CAPEX metrics
type CAPEXDetail struct {
	Sustaining                      float64 `json:"sustaining"`
	Project                         float64 `json:"project"`
	Leasing                         float64 `json:"leasing"`
	AccretionOfMineClosureLiability float64 `json:"accretion_of_mine_closure_liability"`
	Total                           float64 `json:"total"`

	// Breakdown by category (e.g., "Mine Equipment", "Plant Upgrades", "Exploration/Mine Geology")
	ByCategory map[string]float64 `json:"by_category,omitempty"`

	// Breakdown by project (e.g., "C487EY21001 - CAPEX EXPLORACIONES")
	ByProject map[string]float64 `json:"by_project,omitempty"`

	HasData bool `json:"has_data"`
}

// CAPEXVarianceDetail contains variance for CAPEX metrics in detail reports
type CAPEXVarianceDetail struct {
	Sustaining                      VarianceMetric `json:"sustaining"`
	Project                         VarianceMetric `json:"project"`
	Leasing                         VarianceMetric `json:"leasing"`
	AccretionOfMineClosureLiability VarianceMetric `json:"accretion_of_mine_closure_liability"`
	Total                           VarianceMetric `json:"total"`
}

// CAPEXTypeData represents CAPEX aggregated by type
type CAPEXTypeData struct {
	Type     string         `json:"type"`
	Actual   float64        `json:"actual"`
	Budget   float64        `json:"budget"`
	Variance VarianceMetric `json:"variance"`
}

// CAPEXCategoryData represents CAPEX aggregated by category
type CAPEXCategoryData struct {
	Category string         `json:"category"`
	Actual   float64        `json:"actual"`
	Budget   float64        `json:"budget"`
	Variance VarianceMetric `json:"variance"`
}

// FinancialDetailReport represents detailed Financial report
type FinancialDetailReport struct {
	CompanyID   int64                  `json:"company_id"`
	CompanyName string                 `json:"company_name"`
	Year        int                    `json:"year"`
	Months      []FinancialMonthlyData `json:"months"`
}

// FinancialMonthlyData represents Financial data for a single month
type FinancialMonthlyData struct {
	Month    string             `json:"month"` // "2025-01"
	Actual   *FinancialDetail   `json:"actual"`
	Budget   *FinancialDetail   `json:"budget"`
	Variance *FinancialVariance `json:"variance,omitempty"`
}

// FinancialDetail contains detailed Financial metrics
type FinancialDetail struct {
	ShippingSelling     float64 `json:"shipping_selling"`
	SalesTaxesRoyalties float64 `json:"sales_taxes_royalties"`
	OtherAdjustments    float64 `json:"other_adjustments"`
	Total               float64 `json:"total"`

	HasData bool `json:"has_data"`
}

// FinancialVariance contains variance for Financial metrics
type FinancialVariance struct {
	ShippingSelling     VarianceMetric `json:"shipping_selling"`
	SalesTaxesRoyalties VarianceMetric `json:"sales_taxes_royalties"`
	OtherAdjustments    VarianceMetric `json:"other_adjustments"`
	Total               VarianceMetric `json:"total"`
}

// ProductionDetailReport represents detailed Production report
type ProductionDetailReport struct {
	CompanyID   int64                            `json:"company_id"`
	CompanyName string                           `json:"company_name"`
	Year        int                              `json:"year"`
	Months      []ProductionMonthlyData          `json:"months"`
	ByMineral   map[string]ProductionMineralData `json:"by_mineral"`
}

// ProductionMonthlyData represents Production data for a single month
type ProductionMonthlyData struct {
	Month    string                    `json:"month"` // "2025-01"
	Actual   *ProductionDetail         `json:"actual"`
	Budget   *ProductionDetail         `json:"budget"`
	Variance *ProductionVarianceDetail `json:"variance,omitempty"`
}

// ProductionDetail contains detailed Production metrics
type ProductionDetail struct {
	// Silver and Gold (from PBR)
	TotalProductionSilverOz float64 `json:"total_production_silver_oz"`
	TotalProductionGoldOz   float64 `json:"total_production_gold_oz"`

	// Other minerals (from ProductionData)
	ByMineral map[string]float64 `json:"by_mineral,omitempty"` // mineral_code -> quantity

	HasData bool `json:"has_data"`
}

// ProductionVarianceDetail contains variance for Production metrics in detail reports
type ProductionVarianceDetail struct {
	TotalProductionSilverOz VarianceMetric `json:"total_production_silver_oz"`
	TotalProductionGoldOz   VarianceMetric `json:"total_production_gold_oz"`
}

// ProductionMineralData represents Production aggregated by mineral
type ProductionMineralData struct {
	MineralCode string         `json:"mineral_code"`
	MineralName string         `json:"mineral_name"`
	Unit        string         `json:"unit"`
	Actual      float64        `json:"actual"`
	Budget      float64        `json:"budget"`
	Variance    VarianceMetric `json:"variance"`
}

// RevenueDetailReport represents detailed Revenue report
type RevenueDetailReport struct {
	CompanyID   int64                         `json:"company_id"`
	CompanyName string                        `json:"company_name"`
	Year        int                           `json:"year"`
	Months      []RevenueMonthlyData          `json:"months"`
	ByMineral   map[string]RevenueMineralData `json:"by_mineral"`
}

// RevenueMonthlyData represents Revenue data for a single month
type RevenueMonthlyData struct {
	Month    string           `json:"month"` // "2025-01"
	Actual   *RevenueDetail   `json:"actual"`
	Budget   *RevenueDetail   `json:"budget"`
	Variance *RevenueVariance `json:"variance,omitempty"`
}

// RevenueDetail contains detailed Revenue metrics
type RevenueDetail struct {
	// By mineral
	ByMineral map[string]RevenueMineralDetail `json:"by_mineral,omitempty"` // mineral_code -> detail

	// Totals
	TotalRevenue      float64 `json:"total_revenue"`
	TotalQuantitySold float64 `json:"total_quantity_sold"`
	AverageUnitPrice  float64 `json:"average_unit_price"`

	HasData bool `json:"has_data"`
}

// RevenueMineralDetail contains revenue detail for a specific mineral
type RevenueMineralDetail struct {
	MineralCode  string  `json:"mineral_code"`
	MineralName  string  `json:"mineral_name"`
	QuantitySold float64 `json:"quantity_sold"`
	UnitPrice    float64 `json:"unit_price"`
	Revenue      float64 `json:"revenue"`
	Currency     string  `json:"currency"`
}

// RevenueVariance contains variance for Revenue metrics
type RevenueVariance struct {
	TotalRevenue      VarianceMetric `json:"total_revenue"`
	TotalQuantitySold VarianceMetric `json:"total_quantity_sold"`
	AverageUnitPrice  VarianceMetric `json:"average_unit_price"`
}

// RevenueMineralData represents Revenue aggregated by mineral
type RevenueMineralData struct {
	MineralCode string         `json:"mineral_code"`
	MineralName string         `json:"mineral_name"`
	Currency    string         `json:"currency"`
	Actual      float64        `json:"actual"`
	Budget      float64        `json:"budget"`
	Variance    VarianceMetric `json:"variance"`
}
