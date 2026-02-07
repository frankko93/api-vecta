package reports

import (
	"github.com/gmhafiz/go8/internal/domain/data"
)

// Calculator calculates all derived metrics from raw data
type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

// CalculateDataSet calculates all metrics for a dataset
func (c *Calculator) CalculateDataSet(
	pbr *data.PBRData,
	dore *data.DoreData,
	financial *data.FinancialData,
	opexList []*data.OPEXData,
	capexList []*data.CAPEXData,
) *DataSet {
	ds := &DataSet{}

	// Mining & Processing from PBR
	if pbr != nil {
		ds.Mining = MiningMetrics{
			OreMinedT:     pbr.OreMinedT,
			WasteMinedT:   pbr.WasteMinedT,
			DevelopmentsM: pbr.DevelopmentsM,
			HasData:       true,
		}

		ds.Processing = ProcessingMetrics{
			TotalTonnesProcessed:  pbr.TotalTonnesProcessed,
			FeedGradeSilverGpt:    pbr.FeedGradeSilverGpt,
			FeedGradeGoldGpt:      pbr.FeedGradeGoldGpt,
			RecoveryRateSilverPct: pbr.RecoveryRateSilverPct,
			RecoveryRateGoldPct:   pbr.RecoveryRateGoldPct,
			HasData:               true,
		}

		// Calculate production from PBR
		ds.Production = c.calculateProduction(pbr)
	}

	// Calculate costs from OPEX
	if len(opexList) > 0 {
		ds.Costs = c.calculateCosts(opexList)
	}

	// Calculate NSR from Dore + Financial
	if dore != nil {
		ds.NSR = c.calculateNSR(dore, financial, pbr, ds.Costs)
		// Update ProductionBasedMargin in Costs after NSR is calculated
		ds.Costs.ProductionBasedMargin = ds.NSR.NetSmelterReturn - ds.Costs.ProductionBasedCosts
	}

	// Calculate CAPEX
	if len(capexList) > 0 {
		ds.CAPEX = c.calculateCAPEX(capexList, ds.NSR, ds.Costs)
	}

	// Calculate Cash Cost & AISC
	if ds.Production.HasData && ds.Costs.HasData {
		ds.CashCost = c.calculateCashCost(ds.Costs, ds.CAPEX, ds.Production, dore)
	}

	return ds
}

// calculateProduction calculates production from PBR data
func (c *Calculator) calculateProduction(pbr *data.PBRData) ProductionMetrics {
	// Formula: Feed Grade (g/t) * Tonnes Processed * Recovery Rate / 31.1035 (grams per oz)
	silverOz := pbr.FeedGradeSilverGpt * pbr.TotalTonnesProcessed * (pbr.RecoveryRateSilverPct / 100) / 31.1035
	goldOz := pbr.FeedGradeGoldGpt * pbr.TotalTonnesProcessed * (pbr.RecoveryRateGoldPct / 100) / 31.1035
	doreProductionOz := silverOz + goldOz

	return ProductionMetrics{
		TotalProductionSilverOz: silverOz,
		TotalProductionGoldOz:   goldOz,
		PayableSilverOz:         silverOz, // From dore would be adjusted, using same for now
		PayableGoldOz:           goldOz,
		DoreProductionOz:        doreProductionOz,
		HasData:                 true,
	}
}

// calculateCosts calculates cost breakdown from OPEX
func (c *Calculator) calculateCosts(opexList []*data.OPEXData) CostMetrics {
	var mine, processing, ga, transport, inventory float64

	for _, opex := range opexList {
		// Inventory variations handling
		if opex.Subcategory == "Inventory Variation" || opex.Subcategory == "Stockpile/WIP" || opex.Subcategory == "Inventory Variations" {
			inventory += opex.Amount
			continue
		}

		switch opex.CostCenter {
		case "Mine":
			mine += opex.Amount
		case "Processing":
			processing += opex.Amount
		case "G&A":
			ga += opex.Amount
		case "Transport & Shipping":
			transport += opex.Amount
		}
	}

	productionBasedCosts := mine + processing + ga + transport + inventory

	return CostMetrics{
		Mine:                  mine,
		Processing:            processing,
		GA:                    ga,
		TransportShipping:     transport,
		InventoryVariations:   inventory,
		ProductionBasedCosts:  productionBasedCosts,
		ProductionBasedMargin: 0, // Calculated later with NSR
		HasData:               true,
	}
}

// calculateNSR calculates Net Smelter Return from Dore data + Financial adjustments
func (c *Calculator) calculateNSR(dore *data.DoreData, financial *data.FinancialData, pbr *data.PBRData, costs CostMetrics) NSRMetrics {
	// Calculate metal in dore
	metalSilverOz := dore.DoreProducedOz * (dore.SilverGradePct / 100)
	metalGoldOz := dore.DoreProducedOz * (dore.GoldGradePct / 100)

	// Apply adjustments
	metalSilverAdjusted := metalSilverOz + dore.SilverAdjustmentOz
	metalGoldAdjusted := metalGoldOz + dore.GoldAdjustmentOz

	// Calculate deductions
	agDeductionsOz := metalSilverAdjusted * (dore.AgDeductionsPct / 100)
	auDeductionsOz := metalGoldAdjusted * (dore.AuDeductionsPct / 100)

	// Payable metal
	payableSilverOz := metalSilverAdjusted - agDeductionsOz
	payableGoldOz := metalGoldAdjusted - auDeductionsOz

	// Gross revenue
	grossRevenueSilver := payableSilverOz * dore.RealizedPriceSilver
	grossRevenueGold := payableGoldOz * dore.RealizedPriceGold
	doreRevenue := grossRevenueSilver + grossRevenueGold

	// Total charges (Smelting & Refining)
	smeltingRefiningCharges := dore.TreatmentCharge + dore.RefiningDeductionsAu

	// NSR Dore
	nsrDore := doreRevenue - smeltingRefiningCharges

	// Apply financial adjustments
	var shippingSelling, salesTaxesRoyalties float64
	if financial != nil {
		shippingSelling = financial.ShippingSelling
		salesTaxesRoyalties = financial.SalesTaxesRoyalties
	}

	// Net Smelter Return = NSR Dore + Shipping/Selling + Sales Taxes
	netSmelterReturn := nsrDore + shippingSelling + salesTaxesRoyalties

	var nsrPerTonne, costPerTonne, marginPerTonne float64
	if pbr != nil && pbr.TotalTonnesProcessed > 0 {
		nsrPerTonne = netSmelterReturn / pbr.TotalTonnesProcessed
		costPerTonne = costs.ProductionBasedCosts / pbr.TotalTonnesProcessed
		marginPerTonne = nsrPerTonne - costPerTonne
	}

	return NSRMetrics{
		NSRDore:                nsrDore,
		ShippingSelling:        shippingSelling,
		SalesTaxesRoyalties:    salesTaxesRoyalties,
		SmeltingRefiningCharges: smeltingRefiningCharges,
		NetSmelterReturn:       netSmelterReturn,
		NSRPerTonne:            nsrPerTonne,
		TotalCostPerTonne:      costPerTonne,
		MarginPerTonne:         marginPerTonne,
		HasData:                true,
	}
}

// calculateCAPEX calculates CAPEX breakdown
func (c *Calculator) calculateCAPEX(capexList []*data.CAPEXData, nsr NSRMetrics, costs CostMetrics) CAPEXMetrics {
	var sustaining, project, leasing, accretion float64

	for _, capex := range capexList {
		switch capex.Type {
		case "sustaining":
			sustaining += capex.Amount
		case "project":
			project += capex.Amount
		case "leasing":
			leasing += capex.Amount
		}
		// Accretion of Mine Closure Liability - check if it's in category or subcategory
		// For now, we'll need to add this as a separate field in CAPEXData or check category
		// TODO: Add AccretionOfMineClosureLiability field to CAPEXData model
	}

	total := sustaining + project + leasing + accretion

	// Production Based Margin = Net Smelter Return - Production Based Costs
	productionBasedMargin := nsr.NetSmelterReturn - costs.ProductionBasedCosts

	// PBR Net Cash Flow = Production Based Margin - AISC Sustaining Capital
	pbrNetCashFlow := productionBasedMargin - sustaining

	return CAPEXMetrics{
		Sustaining:                     sustaining,
		Project:                        project,
		Leasing:                        leasing,
		AccretionOfMineClosureLiability: accretion,
		Total:                          total,
		ProductionBasedMargin:          productionBasedMargin,
		PBRNetCashFlow:                 pbrNetCashFlow,
		HasData:                        len(capexList) > 0,
	}
}

// calculateCashCost calculates cash cost and AISC per ounce
func (c *Calculator) calculateCashCost(costs CostMetrics, capex CAPEXMetrics, production ProductionMetrics, dore *data.DoreData) CashCostMetrics {
	// Gold credit (by-product credit)
	var goldCredit float64
	if dore != nil && production.PayableGoldOz > 0 {
		goldCredit = production.PayableGoldOz * dore.RealizedPriceGold
	}

	// Cash costs for silver (production costs - gold credit)
	cashCostsSilver := costs.ProductionBasedCosts - goldCredit

	// Cash cost per payable ounce of silver
	var cashCostPerOzSilver, aiscPerOzSilver float64
	if production.PayableSilverOz > 0 {
		cashCostPerOzSilver = cashCostsSilver / production.PayableSilverOz

		// AISC = Cash costs + Sustaining CAPEX + Accretion of Mine Closure Liability
		aiscSilver := cashCostsSilver + capex.Sustaining + capex.AccretionOfMineClosureLiability
		aiscPerOzSilver = aiscSilver / production.PayableSilverOz
	}

	return CashCostMetrics{
		CashCostPerOzSilver: cashCostPerOzSilver,
		AISCPerOzSilver:     aiscPerOzSilver,
		CashCostsSilver:     cashCostsSilver,
		AISCSilver:          cashCostsSilver + capex.Sustaining + capex.AccretionOfMineClosureLiability,
		GoldCredit:          goldCredit,
		HasData:             production.HasData && costs.HasData,
	}
}

// Helper function to calculate variance percentage
func calculateVariancePct(actual, budget float64) float64 {
	if budget == 0 {
		return 0
	}
	return ((actual - budget) / budget) * 100
}

// CalculateVarianceData calculates variance for a monthly comparison
func (c *Calculator) CalculateVarianceData(actual, budget *DataSet) *VarianceData {
	if actual == nil || budget == nil {
		return nil
	}

	return &VarianceData{
		Mining: MiningVariance{
			OreMinedT:     VarianceMetric{Actual: actual.Mining.OreMinedT, Budget: budget.Mining.OreMinedT, Variance: actual.Mining.OreMinedT - budget.Mining.OreMinedT, VariancePct: calculateVariancePct(actual.Mining.OreMinedT, budget.Mining.OreMinedT)},
			WasteMinedT:   VarianceMetric{Actual: actual.Mining.WasteMinedT, Budget: budget.Mining.WasteMinedT, Variance: actual.Mining.WasteMinedT - budget.Mining.WasteMinedT, VariancePct: calculateVariancePct(actual.Mining.WasteMinedT, budget.Mining.WasteMinedT)},
			DevelopmentsM: VarianceMetric{Actual: actual.Mining.DevelopmentsM, Budget: budget.Mining.DevelopmentsM, Variance: actual.Mining.DevelopmentsM - budget.Mining.DevelopmentsM, VariancePct: calculateVariancePct(actual.Mining.DevelopmentsM, budget.Mining.DevelopmentsM)},
		},
		Processing: ProcessingVariance{
			TotalTonnesProcessed:  VarianceMetric{Actual: actual.Processing.TotalTonnesProcessed, Budget: budget.Processing.TotalTonnesProcessed, Variance: actual.Processing.TotalTonnesProcessed - budget.Processing.TotalTonnesProcessed, VariancePct: calculateVariancePct(actual.Processing.TotalTonnesProcessed, budget.Processing.TotalTonnesProcessed)},
			FeedGradeSilverGpt:    VarianceMetric{Actual: actual.Processing.FeedGradeSilverGpt, Budget: budget.Processing.FeedGradeSilverGpt, Variance: actual.Processing.FeedGradeSilverGpt - budget.Processing.FeedGradeSilverGpt, VariancePct: calculateVariancePct(actual.Processing.FeedGradeSilverGpt, budget.Processing.FeedGradeSilverGpt)},
			FeedGradeGoldGpt:      VarianceMetric{Actual: actual.Processing.FeedGradeGoldGpt, Budget: budget.Processing.FeedGradeGoldGpt, Variance: actual.Processing.FeedGradeGoldGpt - budget.Processing.FeedGradeGoldGpt, VariancePct: calculateVariancePct(actual.Processing.FeedGradeGoldGpt, budget.Processing.FeedGradeGoldGpt)},
			RecoveryRateSilverPct: VarianceMetric{Actual: actual.Processing.RecoveryRateSilverPct, Budget: budget.Processing.RecoveryRateSilverPct, Variance: actual.Processing.RecoveryRateSilverPct - budget.Processing.RecoveryRateSilverPct, VariancePct: calculateVariancePct(actual.Processing.RecoveryRateSilverPct, budget.Processing.RecoveryRateSilverPct)},
			RecoveryRateGoldPct:   VarianceMetric{Actual: actual.Processing.RecoveryRateGoldPct, Budget: budget.Processing.RecoveryRateGoldPct, Variance: actual.Processing.RecoveryRateGoldPct - budget.Processing.RecoveryRateGoldPct, VariancePct: calculateVariancePct(actual.Processing.RecoveryRateGoldPct, budget.Processing.RecoveryRateGoldPct)},
		},
		Production: ProductionVariance{
			TotalProductionSilverOz: VarianceMetric{Actual: actual.Production.TotalProductionSilverOz, Budget: budget.Production.TotalProductionSilverOz, Variance: actual.Production.TotalProductionSilverOz - budget.Production.TotalProductionSilverOz, VariancePct: calculateVariancePct(actual.Production.TotalProductionSilverOz, budget.Production.TotalProductionSilverOz)},
			TotalProductionGoldOz:   VarianceMetric{Actual: actual.Production.TotalProductionGoldOz, Budget: budget.Production.TotalProductionGoldOz, Variance: actual.Production.TotalProductionGoldOz - budget.Production.TotalProductionGoldOz, VariancePct: calculateVariancePct(actual.Production.TotalProductionGoldOz, budget.Production.TotalProductionGoldOz)},
			PayableSilverOz:         VarianceMetric{Actual: actual.Production.PayableSilverOz, Budget: budget.Production.PayableSilverOz, Variance: actual.Production.PayableSilverOz - budget.Production.PayableSilverOz, VariancePct: calculateVariancePct(actual.Production.PayableSilverOz, budget.Production.PayableSilverOz)},
			PayableGoldOz:           VarianceMetric{Actual: actual.Production.PayableGoldOz, Budget: budget.Production.PayableGoldOz, Variance: actual.Production.PayableGoldOz - budget.Production.PayableGoldOz, VariancePct: calculateVariancePct(actual.Production.PayableGoldOz, budget.Production.PayableGoldOz)},
			DoreProductionOz:        VarianceMetric{Actual: actual.Production.DoreProductionOz, Budget: budget.Production.DoreProductionOz, Variance: actual.Production.DoreProductionOz - budget.Production.DoreProductionOz, VariancePct: calculateVariancePct(actual.Production.DoreProductionOz, budget.Production.DoreProductionOz)},
		},
		Costs: CostVariance{
			Mine:                  VarianceMetric{Actual: actual.Costs.Mine, Budget: budget.Costs.Mine, Variance: actual.Costs.Mine - budget.Costs.Mine, VariancePct: calculateVariancePct(actual.Costs.Mine, budget.Costs.Mine)},
			Processing:            VarianceMetric{Actual: actual.Costs.Processing, Budget: budget.Costs.Processing, Variance: actual.Costs.Processing - budget.Costs.Processing, VariancePct: calculateVariancePct(actual.Costs.Processing, budget.Costs.Processing)},
			GA:                    VarianceMetric{Actual: actual.Costs.GA, Budget: budget.Costs.GA, Variance: actual.Costs.GA - budget.Costs.GA, VariancePct: calculateVariancePct(actual.Costs.GA, budget.Costs.GA)},
			TransportShipping:     VarianceMetric{Actual: actual.Costs.TransportShipping, Budget: budget.Costs.TransportShipping, Variance: actual.Costs.TransportShipping - budget.Costs.TransportShipping, VariancePct: calculateVariancePct(actual.Costs.TransportShipping, budget.Costs.TransportShipping)},
			InventoryVariations:   VarianceMetric{Actual: actual.Costs.InventoryVariations, Budget: budget.Costs.InventoryVariations, Variance: actual.Costs.InventoryVariations - budget.Costs.InventoryVariations, VariancePct: calculateVariancePct(actual.Costs.InventoryVariations, budget.Costs.InventoryVariations)},
			ProductionBasedCosts:  VarianceMetric{Actual: actual.Costs.ProductionBasedCosts, Budget: budget.Costs.ProductionBasedCosts, Variance: actual.Costs.ProductionBasedCosts - budget.Costs.ProductionBasedCosts, VariancePct: calculateVariancePct(actual.Costs.ProductionBasedCosts, budget.Costs.ProductionBasedCosts)},
			ProductionBasedMargin: VarianceMetric{Actual: actual.Costs.ProductionBasedMargin, Budget: budget.Costs.ProductionBasedMargin, Variance: actual.Costs.ProductionBasedMargin - budget.Costs.ProductionBasedMargin, VariancePct: calculateVariancePct(actual.Costs.ProductionBasedMargin, budget.Costs.ProductionBasedMargin)},
		},
		NSR: NSRVariance{
			NSRDore:                VarianceMetric{Actual: actual.NSR.NSRDore, Budget: budget.NSR.NSRDore, Variance: actual.NSR.NSRDore - budget.NSR.NSRDore, VariancePct: calculateVariancePct(actual.NSR.NSRDore, budget.NSR.NSRDore)},
			ShippingSelling:        VarianceMetric{Actual: actual.NSR.ShippingSelling, Budget: budget.NSR.ShippingSelling, Variance: actual.NSR.ShippingSelling - budget.NSR.ShippingSelling, VariancePct: calculateVariancePct(actual.NSR.ShippingSelling, budget.NSR.ShippingSelling)},
			SalesTaxesRoyalties:    VarianceMetric{Actual: actual.NSR.SalesTaxesRoyalties, Budget: budget.NSR.SalesTaxesRoyalties, Variance: actual.NSR.SalesTaxesRoyalties - budget.NSR.SalesTaxesRoyalties, VariancePct: calculateVariancePct(actual.NSR.SalesTaxesRoyalties, budget.NSR.SalesTaxesRoyalties)},
			SmeltingRefiningCharges: VarianceMetric{Actual: actual.NSR.SmeltingRefiningCharges, Budget: budget.NSR.SmeltingRefiningCharges, Variance: actual.NSR.SmeltingRefiningCharges - budget.NSR.SmeltingRefiningCharges, VariancePct: calculateVariancePct(actual.NSR.SmeltingRefiningCharges, budget.NSR.SmeltingRefiningCharges)},
			NetSmelterReturn:       VarianceMetric{Actual: actual.NSR.NetSmelterReturn, Budget: budget.NSR.NetSmelterReturn, Variance: actual.NSR.NetSmelterReturn - budget.NSR.NetSmelterReturn, VariancePct: calculateVariancePct(actual.NSR.NetSmelterReturn, budget.NSR.NetSmelterReturn)},
			NSRPerTonne:            VarianceMetric{Actual: actual.NSR.NSRPerTonne, Budget: budget.NSR.NSRPerTonne, Variance: actual.NSR.NSRPerTonne - budget.NSR.NSRPerTonne, VariancePct: calculateVariancePct(actual.NSR.NSRPerTonne, budget.NSR.NSRPerTonne)},
			TotalCostPerTonne:      VarianceMetric{Actual: actual.NSR.TotalCostPerTonne, Budget: budget.NSR.TotalCostPerTonne, Variance: actual.NSR.TotalCostPerTonne - budget.NSR.TotalCostPerTonne, VariancePct: calculateVariancePct(actual.NSR.TotalCostPerTonne, budget.NSR.TotalCostPerTonne)},
			MarginPerTonne:         VarianceMetric{Actual: actual.NSR.MarginPerTonne, Budget: budget.NSR.MarginPerTonne, Variance: actual.NSR.MarginPerTonne - budget.NSR.MarginPerTonne, VariancePct: calculateVariancePct(actual.NSR.MarginPerTonne, budget.NSR.MarginPerTonne)},
		},
		CAPEX: CAPEXVariance{
			Sustaining:                     VarianceMetric{Actual: actual.CAPEX.Sustaining, Budget: budget.CAPEX.Sustaining, Variance: actual.CAPEX.Sustaining - budget.CAPEX.Sustaining, VariancePct: calculateVariancePct(actual.CAPEX.Sustaining, budget.CAPEX.Sustaining)},
			Project:                        VarianceMetric{Actual: actual.CAPEX.Project, Budget: budget.CAPEX.Project, Variance: actual.CAPEX.Project - budget.CAPEX.Project, VariancePct: calculateVariancePct(actual.CAPEX.Project, budget.CAPEX.Project)},
			Leasing:                        VarianceMetric{Actual: actual.CAPEX.Leasing, Budget: budget.CAPEX.Leasing, Variance: actual.CAPEX.Leasing - budget.CAPEX.Leasing, VariancePct: calculateVariancePct(actual.CAPEX.Leasing, budget.CAPEX.Leasing)},
			AccretionOfMineClosureLiability: VarianceMetric{Actual: actual.CAPEX.AccretionOfMineClosureLiability, Budget: budget.CAPEX.AccretionOfMineClosureLiability, Variance: actual.CAPEX.AccretionOfMineClosureLiability - budget.CAPEX.AccretionOfMineClosureLiability, VariancePct: calculateVariancePct(actual.CAPEX.AccretionOfMineClosureLiability, budget.CAPEX.AccretionOfMineClosureLiability)},
			Total:                          VarianceMetric{Actual: actual.CAPEX.Total, Budget: budget.CAPEX.Total, Variance: actual.CAPEX.Total - budget.CAPEX.Total, VariancePct: calculateVariancePct(actual.CAPEX.Total, budget.CAPEX.Total)},
			ProductionBasedMargin:          VarianceMetric{Actual: actual.CAPEX.ProductionBasedMargin, Budget: budget.CAPEX.ProductionBasedMargin, Variance: actual.CAPEX.ProductionBasedMargin - budget.CAPEX.ProductionBasedMargin, VariancePct: calculateVariancePct(actual.CAPEX.ProductionBasedMargin, budget.CAPEX.ProductionBasedMargin)},
			PBRNetCashFlow:                 VarianceMetric{Actual: actual.CAPEX.PBRNetCashFlow, Budget: budget.CAPEX.PBRNetCashFlow, Variance: actual.CAPEX.PBRNetCashFlow - budget.CAPEX.PBRNetCashFlow, VariancePct: calculateVariancePct(actual.CAPEX.PBRNetCashFlow, budget.CAPEX.PBRNetCashFlow)},
		},
		CashCost: CashCostVariance{
			CashCostPerOzSilver: VarianceMetric{Actual: actual.CashCost.CashCostPerOzSilver, Budget: budget.CashCost.CashCostPerOzSilver, Variance: actual.CashCost.CashCostPerOzSilver - budget.CashCost.CashCostPerOzSilver, VariancePct: calculateVariancePct(actual.CashCost.CashCostPerOzSilver, budget.CashCost.CashCostPerOzSilver)},
			AISCPerOzSilver:     VarianceMetric{Actual: actual.CashCost.AISCPerOzSilver, Budget: budget.CashCost.AISCPerOzSilver, Variance: actual.CashCost.AISCPerOzSilver - budget.CashCost.AISCPerOzSilver, VariancePct: calculateVariancePct(actual.CashCost.AISCPerOzSilver, budget.CashCost.AISCPerOzSilver)},
			CashCostsSilver:     VarianceMetric{Actual: actual.CashCost.CashCostsSilver, Budget: budget.CashCost.CashCostsSilver, Variance: actual.CashCost.CashCostsSilver - budget.CashCost.CashCostsSilver, VariancePct: calculateVariancePct(actual.CashCost.CashCostsSilver, budget.CashCost.CashCostsSilver)},
			AISCSilver:          VarianceMetric{Actual: actual.CashCost.AISCSilver, Budget: budget.CashCost.AISCSilver, Variance: actual.CashCost.AISCSilver - budget.CashCost.AISCSilver, VariancePct: calculateVariancePct(actual.CashCost.AISCSilver, budget.CashCost.AISCSilver)},
			GoldCredit:          VarianceMetric{Actual: actual.CashCost.GoldCredit, Budget: budget.CashCost.GoldCredit, Variance: actual.CashCost.GoldCredit - budget.CashCost.GoldCredit, VariancePct: calculateVariancePct(actual.CashCost.GoldCredit, budget.CashCost.GoldCredit)},
		},
	}
}
