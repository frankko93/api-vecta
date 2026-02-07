package data

import "time"

// ProductionData represents production data
type ProductionData struct {
	ID          int64      `db:"id" json:"id"`
	CompanyID   int64      `db:"company_id" json:"company_id"`
	Date        time.Time  `db:"date" json:"date"`
	MineralID   int        `db:"mineral_id" json:"mineral_id"`
	Quantity    float64    `db:"quantity" json:"quantity"`
	Unit        string     `db:"unit" json:"unit"`
	DataType    string     `db:"data_type" json:"data_type"`
	Version     int        `db:"version" json:"version"`
	Description string     `db:"description" json:"description,omitempty"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedBy   int64      `db:"created_by" json:"created_by"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}

// DoreData represents dore production data
type DoreData struct {
	ID                   int64      `db:"id" json:"id"`
	CompanyID            int64      `db:"company_id" json:"company_id"`
	Date                 time.Time  `db:"date" json:"date"`
	DoreProducedOz       float64    `db:"dore_produced_oz" json:"dore_produced_oz"`
	SilverGradePct       float64    `db:"silver_grade_pct" json:"silver_grade_pct"`
	GoldGradePct         float64    `db:"gold_grade_pct" json:"gold_grade_pct"`
	PBRPriceSilver       float64    `db:"pbr_price_silver" json:"pbr_price_silver"`
	PBRPriceGold         float64    `db:"pbr_price_gold" json:"pbr_price_gold"`
	RealizedPriceSilver  float64    `db:"realized_price_silver" json:"realized_price_silver"`
	RealizedPriceGold    float64    `db:"realized_price_gold" json:"realized_price_gold"`
	SilverAdjustmentOz   float64    `db:"silver_adjustment_oz" json:"silver_adjustment_oz"`
	GoldAdjustmentOz     float64    `db:"gold_adjustment_oz" json:"gold_adjustment_oz"`
	AgDeductionsPct      float64    `db:"ag_deductions_pct" json:"ag_deductions_pct"`
	AuDeductionsPct      float64    `db:"au_deductions_pct" json:"au_deductions_pct"`
	TreatmentCharge      float64    `db:"treatment_charge" json:"treatment_charge"`
	RefiningDeductionsAu float64    `db:"refining_deductions_au" json:"refining_deductions_au"`
	Streaming            float64    `db:"streaming" json:"streaming"` // Streaming agreement value (usually negative)
	DataType             string     `db:"data_type" json:"data_type"`
	Version              int        `db:"version" json:"version"`
	Description          string     `db:"description" json:"description,omitempty"`
	DeletedAt            *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedBy            int64      `db:"created_by" json:"created_by"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
}

// PBRData represents Plan Beneficio Regional data
type PBRData struct {
	ID        int64     `db:"id" json:"id"`
	CompanyID int64     `db:"company_id" json:"company_id"`
	Date      time.Time `db:"date" json:"date"`

	// Mining - Ore breakdown by mine type
	OpenPitOreT     float64 `db:"open_pit_ore_t" json:"open_pit_ore_t"`
	UndergroundOreT float64 `db:"underground_ore_t" json:"underground_ore_t"`
	OreMinedT       float64 `db:"ore_mined_t" json:"ore_mined_t"` // Total = OpenPit + Underground

	// Mining - Waste and ratios
	WasteMinedT   float64 `db:"waste_mined_t" json:"waste_mined_t"`
	StrippingRatio float64 `db:"stripping_ratio" json:"stripping_ratio"` // Waste / OpenPit Ore

	// Mining - Grades by mine type
	MiningGradeSilverGpt      float64 `db:"mining_grade_silver_gpt" json:"mining_grade_silver_gpt"`
	MiningGradeGoldGpt        float64 `db:"mining_grade_gold_gpt" json:"mining_grade_gold_gpt"`
	OpenPitGradeSilverGpt     float64 `db:"open_pit_grade_silver_gpt" json:"open_pit_grade_silver_gpt"`
	UndergroundGradeSilverGpt float64 `db:"underground_grade_silver_gpt" json:"underground_grade_silver_gpt"`
	OpenPitGradeGoldGpt       float64 `db:"open_pit_grade_gold_gpt" json:"open_pit_grade_gold_gpt"`
	UndergroundGradeGoldGpt   float64 `db:"underground_grade_gold_gpt" json:"underground_grade_gold_gpt"`

	// Developments breakdown
	PrimaryDevelopmentM       float64 `db:"primary_development_m" json:"primary_development_m"`
	SecondaryDevelopmentOpexM float64 `db:"secondary_development_opex_m" json:"secondary_development_opex_m"`
	ExpansionaryDevelopmentM  float64 `db:"expansionary_development_m" json:"expansionary_development_m"`
	DevelopmentsM             float64 `db:"developments_m" json:"developments_m"` // Total

	// Processing
	TotalTonnesProcessed  float64 `db:"total_tonnes_processed" json:"total_tonnes_processed"`
	FeedGradeSilverGpt    float64 `db:"feed_grade_silver_gpt" json:"feed_grade_silver_gpt"`
	FeedGradeGoldGpt      float64 `db:"feed_grade_gold_gpt" json:"feed_grade_gold_gpt"`
	RecoveryRateSilverPct float64 `db:"recovery_rate_silver_pct" json:"recovery_rate_silver_pct"`
	RecoveryRateGoldPct   float64 `db:"recovery_rate_gold_pct" json:"recovery_rate_gold_pct"`

	// Headcount
	FullTimeEmployees int `db:"full_time_employees" json:"full_time_employees"`
	Contractors       int `db:"contractors" json:"contractors"`
	TotalHeadcount    int `db:"total_headcount" json:"total_headcount"`

	// Metadata
	DataType    string     `db:"data_type" json:"data_type"`
	Version     int        `db:"version" json:"version"`
	Description string     `db:"description" json:"description,omitempty"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedBy   int64      `db:"created_by" json:"created_by"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}

// OPEXData represents operational expenditure data
type OPEXData struct {
	ID          int64      `db:"id" json:"id"`
	CompanyID   int64      `db:"company_id" json:"company_id"`
	Date        time.Time  `db:"date" json:"date"`
	CostCenter  string     `db:"cost_center" json:"cost_center"`
	Subcategory string     `db:"subcategory" json:"subcategory"`
	ExpenseType string     `db:"expense_type" json:"expense_type"`
	Amount      float64    `db:"amount" json:"amount"`
	Currency    string     `db:"currency" json:"currency"`
	DataType    string     `db:"data_type" json:"data_type"`
	Version     int        `db:"version" json:"version"`
	Description string     `db:"description" json:"description,omitempty"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedBy   int64      `db:"created_by" json:"created_by"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}

// CAPEXData represents capital expenditure data
type CAPEXData struct {
	ID                              int64      `db:"id" json:"id"`
	CompanyID                       int64      `db:"company_id" json:"company_id"`
	Date                            time.Time  `db:"date" json:"date"`
	Category                        string     `db:"category" json:"category"`
	CARNumber                       string     `db:"car_number" json:"car_number"`
	ProjectName                     string     `db:"project_name" json:"project_name"`
	Type                            string     `db:"type" json:"type"`
	Amount                          float64    `db:"amount" json:"amount"`
	AccretionOfMineClosureLiability float64    `db:"accretion_of_mine_closure_liability" json:"accretion_of_mine_closure_liability"`
	Currency                        string     `db:"currency" json:"currency"`
	DataType                        string     `db:"data_type" json:"data_type"`
	Version                         int        `db:"version" json:"version"`
	Description                     string     `db:"description" json:"description,omitempty"`
	DeletedAt                       *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedBy                       int64      `db:"created_by" json:"created_by"`
	CreatedAt                       time.Time  `db:"created_at" json:"created_at"`
}

// RevenueData represents revenue data
type RevenueData struct {
	ID           int64      `db:"id" json:"id"`
	CompanyID    int64      `db:"company_id" json:"company_id"`
	Date         time.Time  `db:"date" json:"date"`
	MineralID    int        `db:"mineral_id" json:"mineral_id"`
	QuantitySold float64    `db:"quantity_sold" json:"quantity_sold"`
	UnitPrice    float64    `db:"unit_price" json:"unit_price"`
	Currency     string     `db:"currency" json:"currency"`
	DataType     string     `db:"data_type" json:"data_type"`
	Version      int        `db:"version" json:"version"`
	Description  string     `db:"description" json:"description,omitempty"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedBy    int64      `db:"created_by" json:"created_by"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

// FinancialData represents financial adjustments (shipping, taxes, royalties)
type FinancialData struct {
	ID                  int64      `db:"id" json:"id"`
	CompanyID           int64      `db:"company_id" json:"company_id"`
	Date                time.Time  `db:"date" json:"date"`
	ShippingSelling     float64    `db:"shipping_selling" json:"shipping_selling"`
	SalesTaxesRoyalties float64    `db:"sales_taxes_royalties" json:"sales_taxes_royalties"`
	OtherAdjustments    float64    `db:"other_adjustments" json:"other_adjustments"`
	Currency            string     `db:"currency" json:"currency"`
	DataType            string     `db:"data_type" json:"data_type"`
	Version             int        `db:"version" json:"version"`
	Description         string     `db:"description" json:"description,omitempty"`
	DeletedAt           *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedBy           int64      `db:"created_by" json:"created_by"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
}
