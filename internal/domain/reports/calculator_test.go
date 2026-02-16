package reports

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gmhafiz/go8/internal/domain/data"
)

func TestCalculateProduction(t *testing.T) {
	calc := NewCalculator()
	pbr := newTestPBRData()

	production := calc.calculateProduction(pbr)

	// Expected silver oz = 209.79 * 35951 * 0.9401 / 31.1035 ≈ 227,957
	assert.InDelta(t, expectedTotalProductionSilverOz, production.TotalProductionSilverOz, 100)

	// Expected gold oz = 7.35 * 35951 * 0.9536 / 31.1035 ≈ 8,106
	assert.InDelta(t, expectedTotalProductionGoldOz, production.TotalProductionGoldOz, 50)

	// Dore Production = Silver + Gold
	expectedDoreProduction := expectedTotalProductionSilverOz + expectedTotalProductionGoldOz
	assert.InDelta(t, expectedDoreProduction, production.DoreProductionOz, 150)

	assert.True(t, production.HasData)
}

func TestCalculateCosts(t *testing.T) {
	calc := NewCalculator()
	opexList := newTestOPEXList()

	costs := calc.calculateCosts(opexList)

	// Verify individual cost centers match our test data
	assert.Equal(t, 8537997.0, costs.Mine)
	assert.Equal(t, 3613678.0, costs.Processing)
	assert.Equal(t, 5471220.0, costs.GA)
	assert.Equal(t, 1740162.0, costs.InventoryVariations)

	// Total should match expected
	assert.Equal(t, expectedProductionBasedCosts, costs.ProductionBasedCosts)
	assert.True(t, costs.HasData)
}

func TestCalculateNSR(t *testing.T) {
	calc := NewCalculator()
	dore := newTestDoreData()
	financial := newTestFinancialData()
	pbr := newTestPBRData()

	costs := CostMetrics{
		ProductionBasedCosts: expectedProductionBasedCosts,
	}

	nsr := calc.calculateNSR(dore, financial, pbr, costs)

	// Verify NSR components are calculated
	assert.Greater(t, nsr.NSRDore, 0.0)
	assert.Equal(t, financial.ShippingSelling, nsr.ShippingSelling)
	assert.Equal(t, financial.SalesTaxes, nsr.SalesTaxes)
	assert.Equal(t, financial.Royalties, nsr.Royalties)
	assert.Equal(t, financial.SalesTaxes+financial.Royalties, nsr.SalesTaxesRoyalties)
	assert.Equal(t, financial.OtherSalesDeductions, nsr.OtherSalesDeductions)
	assert.Greater(t, nsr.NetSmelterReturn, 0.0)

	// Smelting & Refining Charges = Treatment Charge + Refining Deductions
	expectedSmeltingRefiningCharges := dore.TreatmentCharge + dore.RefiningDeductionsAu
	assert.Equal(t, expectedSmeltingRefiningCharges, nsr.SmeltingRefiningCharges)

	// Streaming should be passed through from dore data
	assert.Equal(t, dore.Streaming, nsr.Streaming)

	// PBR Revenue = NSR Dore + Streaming
	expectedPBRRevenue := nsr.NSRDore + nsr.Streaming
	assert.Equal(t, expectedPBRRevenue, nsr.PBRRevenue)

	// Gold Credit should be negative (by-product credit)
	assert.Less(t, nsr.GoldCredit, 0.0)

	// NSR = NSR Dore + Shipping + SalesTaxes + Royalties + OtherSalesDeductions
	expectedNSR := nsr.NSRDore + nsr.ShippingSelling + nsr.SalesTaxes + nsr.Royalties + nsr.OtherSalesDeductions
	assert.Equal(t, expectedNSR, nsr.NetSmelterReturn)

	// Metal prices should be propagated from dore data
	assert.Equal(t, dore.RealizedPriceSilver, nsr.SilverPricePerOz)
	assert.Equal(t, dore.RealizedPriceGold, nsr.GoldPricePerOz)

	// Per tonne calculations
	assert.Greater(t, nsr.NSRPerTonne, 0.0)
	assert.Greater(t, nsr.TotalCostPerTonne, 0.0)
	assert.NotZero(t, nsr.MarginPerTonne)

	assert.True(t, nsr.HasData)
}

func TestCalculateCAPEX(t *testing.T) {
	calc := NewCalculator()
	capexList := newTestCAPEXList()

	nsr := NSRMetrics{
		NetSmelterReturn: expectedNetSmelterReturn,
	}

	costs := CostMetrics{
		ProductionBasedCosts: expectedProductionBasedCosts,
	}

	capex := calc.calculateCAPEX(capexList, nsr, costs)

	assert.Equal(t, expectedSustainingCAPEX, capex.Sustaining)
	assert.Equal(t, 350000.0, capex.Project)
	assert.Equal(t, 100000.0, capex.Leasing)

	// AccretionOfMineClosureLiability should be 0 (not in test data)
	assert.Equal(t, 0.0, capex.AccretionOfMineClosureLiability)

	// Total = Sustaining + Project + Leasing + Accretion
	expectedTotal := expectedSustainingCAPEX + 350000.0 + 100000.0 + 0.0
	assert.Equal(t, expectedTotal, capex.Total)

	// Production Based Margin = NSR - Costs
	expectedMargin := expectedNetSmelterReturn - expectedProductionBasedCosts
	assert.Equal(t, expectedMargin, capex.ProductionBasedMargin)

	// Net Cash Flow = Margin - Sustaining CAPEX
	expectedCashFlow := expectedMargin - expectedSustainingCAPEX
	assert.Equal(t, expectedCashFlow, capex.PBRNetCashFlow)

	assert.True(t, capex.HasData)
}

func TestCalculateCashCost(t *testing.T) {
	calc := NewCalculator()

	dore := newTestDoreData()
	financial := newTestFinancialData()

	costs := CostMetrics{
		ProductionBasedCosts: expectedProductionBasedCosts,
		HasData:              true,
	}

	capex := CAPEXMetrics{
		Sustaining:                      expectedSustainingCAPEX,
		AccretionOfMineClosureLiability: 0.0, // Not in test data
	}

	production := ProductionMetrics{
		PayableSilverOz: expectedTotalProductionSilverOz,
		PayableGoldOz:   expectedTotalProductionGoldOz,
		HasData:         true,
	}

	// NSR metrics needed for smelting_refining_charges
	nsr := NSRMetrics{
		SmeltingRefiningCharges: dore.TreatmentCharge + dore.RefiningDeductionsAu,
	}

	cashCost := calc.calculateCashCost(costs, capex, production, dore, nsr, financial)

	// Gold Credit = Payable Gold * Price
	expectedGoldCredit := expectedTotalProductionGoldOz * dore.RealizedPriceGold
	assert.Equal(t, expectedGoldCredit, cashCost.GoldCredit)

	// CORRECTED Cash Costs Silver (total) = ProdCosts + Shipping + Smelting + SalesTaxes + Royalties + OtherDeductions - GoldCredit
	cashCostsSilver := expectedProductionBasedCosts +
		financial.ShippingSelling +
		nsr.SmeltingRefiningCharges +
		financial.SalesTaxes +
		financial.Royalties +
		financial.OtherSalesDeductions -
		expectedGoldCredit
	assert.Equal(t, cashCostsSilver, cashCost.CashCostsSilver)

	// Cash Cost Silver per Oz = Cash Costs Silver / Payable Silver Oz
	expectedCashCostPerOz := cashCostsSilver / expectedTotalProductionSilverOz
	assert.InDelta(t, expectedCashCostPerOz, cashCost.CashCostPerOzSilver, 0.1)

	// AISC Silver (total) = Cash Costs Silver + Sustaining CAPEX + Accretion
	aiscSilverTotal := cashCostsSilver + expectedSustainingCAPEX + capex.AccretionOfMineClosureLiability
	assert.Equal(t, aiscSilverTotal, cashCost.AISCSilver)

	// AISC per Oz = AISC Silver Total / Payable Silver Oz
	expectedAISC := aiscSilverTotal / expectedTotalProductionSilverOz
	assert.InDelta(t, expectedAISC, cashCost.AISCPerOzSilver, 0.1)

	// Sustaining Capital per Oz = Sustaining CAPEX / Payable Silver Oz
	expectedSustCapPerOz := expectedSustainingCAPEX / expectedTotalProductionSilverOz
	assert.InDelta(t, expectedSustCapPerOz, cashCost.SustainingCapitalPerOz, 0.1)

	assert.True(t, cashCost.HasData)
}

func TestCostCenterValidation(t *testing.T) {
	validCenters := []string{"Mine", "Processing", "G&A", "Transport & Shipping"}

	for _, center := range validCenters {
		cc := data.CostCenter(center)
		assert.True(t, cc.IsValid(), "Cost center %s should be valid", center)
	}

	invalidCenters := []string{"Invalid", "Mining", "Proc", ""}
	for _, center := range invalidCenters {
		cc := data.CostCenter(center)
		assert.False(t, cc.IsValid(), "Cost center %s should be invalid", center)
	}
}

func TestExpenseTypeValidation(t *testing.T) {
	validTypes := []string{"Labour", "Materials", "Third Party", "Other"}

	for _, expType := range validTypes {
		et := data.ExpenseType(expType)
		assert.True(t, et.IsValid(), "Expense type %s should be valid", expType)
	}

	invalidTypes := []string{"Invalid", "Labor", "Material", ""}
	for _, expType := range invalidTypes {
		et := data.ExpenseType(expType)
		assert.False(t, et.IsValid(), "Expense type %s should be invalid", expType)
	}
}
