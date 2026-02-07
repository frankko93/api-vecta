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
			// Ore breakdown
			OpenPitOreT:     pbr.OpenPitOreT,
			UndergroundOreT: pbr.UndergroundOreT,
			OreMinedT:       pbr.OreMinedT,

			// Waste and ratios
			WasteMinedT:    pbr.WasteMinedT,
			StrippingRatio: pbr.StrippingRatio,

			// Mining grades
			MiningGradeSilverGpt:      pbr.MiningGradeSilverGpt,
			MiningGradeGoldGpt:        pbr.MiningGradeGoldGpt,
			OpenPitGradeSilverGpt:     pbr.OpenPitGradeSilverGpt,
			UndergroundGradeSilverGpt: pbr.UndergroundGradeSilverGpt,
			OpenPitGradeGoldGpt:       pbr.OpenPitGradeGoldGpt,
			UndergroundGradeGoldGpt:   pbr.UndergroundGradeGoldGpt,

			// Developments breakdown
			PrimaryDevelopmentM:       pbr.PrimaryDevelopmentM,
			SecondaryDevelopmentOpexM: pbr.SecondaryDevelopmentOpexM,
			ExpansionaryDevelopmentM:  pbr.ExpansionaryDevelopmentM,
			DevelopmentsM:             pbr.DevelopmentsM,

			// Headcount
			FullTimeEmployees: pbr.FullTimeEmployees,
			Contractors:       pbr.Contractors,
			TotalHeadcount:    pbr.TotalHeadcount,

			HasData: true,
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

	// Streaming (from dore data, usually negative)
	streaming := dore.Streaming

	// PBR Revenue = NSR Dore + Streaming (streaming is typically negative)
	pbrRevenue := nsrDore + streaming

	// Apply financial adjustments
	var shippingSelling, salesTaxesRoyalties float64
	if financial != nil {
		shippingSelling = financial.ShippingSelling
		salesTaxesRoyalties = financial.SalesTaxesRoyalties
	}

	// Net Smelter Return = NSR Dore + Shipping/Selling + Sales Taxes
	netSmelterReturn := nsrDore + shippingSelling + salesTaxesRoyalties

	// Gold credit (by-product credit) - negative value
	goldCredit := -(payableGoldOz * dore.RealizedPriceGold)

	var nsrPerTonne, costPerTonne, marginPerTonne float64
	if pbr != nil && pbr.TotalTonnesProcessed > 0 {
		nsrPerTonne = netSmelterReturn / pbr.TotalTonnesProcessed
		costPerTonne = costs.ProductionBasedCosts / pbr.TotalTonnesProcessed
		marginPerTonne = nsrPerTonne - costPerTonne
	}

	return NSRMetrics{
		NSRDore:                 nsrDore,
		Streaming:               streaming,
		PBRRevenue:              pbrRevenue,
		ShippingSelling:         shippingSelling,
		SalesTaxesRoyalties:     salesTaxesRoyalties,
		SmeltingRefiningCharges: smeltingRefiningCharges,
		NetSmelterReturn:        netSmelterReturn,
		GoldCredit:              goldCredit,
		NSRPerTonne:             nsrPerTonne,
		TotalCostPerTonne:       costPerTonne,
		MarginPerTonne:          marginPerTonne,
		HasData:                 true,
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
		// Accretion of Mine Closure Liability - now comes from the field
		accretion += capex.AccretionOfMineClosureLiability
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
			// Ore breakdown
			OpenPitOreT:     VarianceMetric{Actual: actual.Mining.OpenPitOreT, Budget: budget.Mining.OpenPitOreT, Variance: actual.Mining.OpenPitOreT - budget.Mining.OpenPitOreT, VariancePct: calculateVariancePct(actual.Mining.OpenPitOreT, budget.Mining.OpenPitOreT)},
			UndergroundOreT: VarianceMetric{Actual: actual.Mining.UndergroundOreT, Budget: budget.Mining.UndergroundOreT, Variance: actual.Mining.UndergroundOreT - budget.Mining.UndergroundOreT, VariancePct: calculateVariancePct(actual.Mining.UndergroundOreT, budget.Mining.UndergroundOreT)},
			OreMinedT:       VarianceMetric{Actual: actual.Mining.OreMinedT, Budget: budget.Mining.OreMinedT, Variance: actual.Mining.OreMinedT - budget.Mining.OreMinedT, VariancePct: calculateVariancePct(actual.Mining.OreMinedT, budget.Mining.OreMinedT)},
			// Waste and ratios
			WasteMinedT:    VarianceMetric{Actual: actual.Mining.WasteMinedT, Budget: budget.Mining.WasteMinedT, Variance: actual.Mining.WasteMinedT - budget.Mining.WasteMinedT, VariancePct: calculateVariancePct(actual.Mining.WasteMinedT, budget.Mining.WasteMinedT)},
			StrippingRatio: VarianceMetric{Actual: actual.Mining.StrippingRatio, Budget: budget.Mining.StrippingRatio, Variance: actual.Mining.StrippingRatio - budget.Mining.StrippingRatio, VariancePct: calculateVariancePct(actual.Mining.StrippingRatio, budget.Mining.StrippingRatio)},
			// Mining grades
			MiningGradeSilverGpt:      VarianceMetric{Actual: actual.Mining.MiningGradeSilverGpt, Budget: budget.Mining.MiningGradeSilverGpt, Variance: actual.Mining.MiningGradeSilverGpt - budget.Mining.MiningGradeSilverGpt, VariancePct: calculateVariancePct(actual.Mining.MiningGradeSilverGpt, budget.Mining.MiningGradeSilverGpt)},
			MiningGradeGoldGpt:        VarianceMetric{Actual: actual.Mining.MiningGradeGoldGpt, Budget: budget.Mining.MiningGradeGoldGpt, Variance: actual.Mining.MiningGradeGoldGpt - budget.Mining.MiningGradeGoldGpt, VariancePct: calculateVariancePct(actual.Mining.MiningGradeGoldGpt, budget.Mining.MiningGradeGoldGpt)},
			OpenPitGradeSilverGpt:     VarianceMetric{Actual: actual.Mining.OpenPitGradeSilverGpt, Budget: budget.Mining.OpenPitGradeSilverGpt, Variance: actual.Mining.OpenPitGradeSilverGpt - budget.Mining.OpenPitGradeSilverGpt, VariancePct: calculateVariancePct(actual.Mining.OpenPitGradeSilverGpt, budget.Mining.OpenPitGradeSilverGpt)},
			UndergroundGradeSilverGpt: VarianceMetric{Actual: actual.Mining.UndergroundGradeSilverGpt, Budget: budget.Mining.UndergroundGradeSilverGpt, Variance: actual.Mining.UndergroundGradeSilverGpt - budget.Mining.UndergroundGradeSilverGpt, VariancePct: calculateVariancePct(actual.Mining.UndergroundGradeSilverGpt, budget.Mining.UndergroundGradeSilverGpt)},
			OpenPitGradeGoldGpt:       VarianceMetric{Actual: actual.Mining.OpenPitGradeGoldGpt, Budget: budget.Mining.OpenPitGradeGoldGpt, Variance: actual.Mining.OpenPitGradeGoldGpt - budget.Mining.OpenPitGradeGoldGpt, VariancePct: calculateVariancePct(actual.Mining.OpenPitGradeGoldGpt, budget.Mining.OpenPitGradeGoldGpt)},
			UndergroundGradeGoldGpt:   VarianceMetric{Actual: actual.Mining.UndergroundGradeGoldGpt, Budget: budget.Mining.UndergroundGradeGoldGpt, Variance: actual.Mining.UndergroundGradeGoldGpt - budget.Mining.UndergroundGradeGoldGpt, VariancePct: calculateVariancePct(actual.Mining.UndergroundGradeGoldGpt, budget.Mining.UndergroundGradeGoldGpt)},
			// Developments
			PrimaryDevelopmentM:       VarianceMetric{Actual: actual.Mining.PrimaryDevelopmentM, Budget: budget.Mining.PrimaryDevelopmentM, Variance: actual.Mining.PrimaryDevelopmentM - budget.Mining.PrimaryDevelopmentM, VariancePct: calculateVariancePct(actual.Mining.PrimaryDevelopmentM, budget.Mining.PrimaryDevelopmentM)},
			SecondaryDevelopmentOpexM: VarianceMetric{Actual: actual.Mining.SecondaryDevelopmentOpexM, Budget: budget.Mining.SecondaryDevelopmentOpexM, Variance: actual.Mining.SecondaryDevelopmentOpexM - budget.Mining.SecondaryDevelopmentOpexM, VariancePct: calculateVariancePct(actual.Mining.SecondaryDevelopmentOpexM, budget.Mining.SecondaryDevelopmentOpexM)},
			ExpansionaryDevelopmentM:  VarianceMetric{Actual: actual.Mining.ExpansionaryDevelopmentM, Budget: budget.Mining.ExpansionaryDevelopmentM, Variance: actual.Mining.ExpansionaryDevelopmentM - budget.Mining.ExpansionaryDevelopmentM, VariancePct: calculateVariancePct(actual.Mining.ExpansionaryDevelopmentM, budget.Mining.ExpansionaryDevelopmentM)},
			DevelopmentsM:             VarianceMetric{Actual: actual.Mining.DevelopmentsM, Budget: budget.Mining.DevelopmentsM, Variance: actual.Mining.DevelopmentsM - budget.Mining.DevelopmentsM, VariancePct: calculateVariancePct(actual.Mining.DevelopmentsM, budget.Mining.DevelopmentsM)},
			// Headcount
			FullTimeEmployees: VarianceMetric{Actual: float64(actual.Mining.FullTimeEmployees), Budget: float64(budget.Mining.FullTimeEmployees), Variance: float64(actual.Mining.FullTimeEmployees - budget.Mining.FullTimeEmployees), VariancePct: calculateVariancePct(float64(actual.Mining.FullTimeEmployees), float64(budget.Mining.FullTimeEmployees))},
			Contractors:       VarianceMetric{Actual: float64(actual.Mining.Contractors), Budget: float64(budget.Mining.Contractors), Variance: float64(actual.Mining.Contractors - budget.Mining.Contractors), VariancePct: calculateVariancePct(float64(actual.Mining.Contractors), float64(budget.Mining.Contractors))},
			TotalHeadcount:    VarianceMetric{Actual: float64(actual.Mining.TotalHeadcount), Budget: float64(budget.Mining.TotalHeadcount), Variance: float64(actual.Mining.TotalHeadcount - budget.Mining.TotalHeadcount), VariancePct: calculateVariancePct(float64(actual.Mining.TotalHeadcount), float64(budget.Mining.TotalHeadcount))},
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
			NSRDore:                 VarianceMetric{Actual: actual.NSR.NSRDore, Budget: budget.NSR.NSRDore, Variance: actual.NSR.NSRDore - budget.NSR.NSRDore, VariancePct: calculateVariancePct(actual.NSR.NSRDore, budget.NSR.NSRDore)},
			Streaming:               VarianceMetric{Actual: actual.NSR.Streaming, Budget: budget.NSR.Streaming, Variance: actual.NSR.Streaming - budget.NSR.Streaming, VariancePct: calculateVariancePct(actual.NSR.Streaming, budget.NSR.Streaming)},
			PBRRevenue:              VarianceMetric{Actual: actual.NSR.PBRRevenue, Budget: budget.NSR.PBRRevenue, Variance: actual.NSR.PBRRevenue - budget.NSR.PBRRevenue, VariancePct: calculateVariancePct(actual.NSR.PBRRevenue, budget.NSR.PBRRevenue)},
			ShippingSelling:         VarianceMetric{Actual: actual.NSR.ShippingSelling, Budget: budget.NSR.ShippingSelling, Variance: actual.NSR.ShippingSelling - budget.NSR.ShippingSelling, VariancePct: calculateVariancePct(actual.NSR.ShippingSelling, budget.NSR.ShippingSelling)},
			SalesTaxesRoyalties:     VarianceMetric{Actual: actual.NSR.SalesTaxesRoyalties, Budget: budget.NSR.SalesTaxesRoyalties, Variance: actual.NSR.SalesTaxesRoyalties - budget.NSR.SalesTaxesRoyalties, VariancePct: calculateVariancePct(actual.NSR.SalesTaxesRoyalties, budget.NSR.SalesTaxesRoyalties)},
			SmeltingRefiningCharges: VarianceMetric{Actual: actual.NSR.SmeltingRefiningCharges, Budget: budget.NSR.SmeltingRefiningCharges, Variance: actual.NSR.SmeltingRefiningCharges - budget.NSR.SmeltingRefiningCharges, VariancePct: calculateVariancePct(actual.NSR.SmeltingRefiningCharges, budget.NSR.SmeltingRefiningCharges)},
			NetSmelterReturn:        VarianceMetric{Actual: actual.NSR.NetSmelterReturn, Budget: budget.NSR.NetSmelterReturn, Variance: actual.NSR.NetSmelterReturn - budget.NSR.NetSmelterReturn, VariancePct: calculateVariancePct(actual.NSR.NetSmelterReturn, budget.NSR.NetSmelterReturn)},
			GoldCredit:              VarianceMetric{Actual: actual.NSR.GoldCredit, Budget: budget.NSR.GoldCredit, Variance: actual.NSR.GoldCredit - budget.NSR.GoldCredit, VariancePct: calculateVariancePct(actual.NSR.GoldCredit, budget.NSR.GoldCredit)},
			NSRPerTonne:             VarianceMetric{Actual: actual.NSR.NSRPerTonne, Budget: budget.NSR.NSRPerTonne, Variance: actual.NSR.NSRPerTonne - budget.NSR.NSRPerTonne, VariancePct: calculateVariancePct(actual.NSR.NSRPerTonne, budget.NSR.NSRPerTonne)},
			TotalCostPerTonne:       VarianceMetric{Actual: actual.NSR.TotalCostPerTonne, Budget: budget.NSR.TotalCostPerTonne, Variance: actual.NSR.TotalCostPerTonne - budget.NSR.TotalCostPerTonne, VariancePct: calculateVariancePct(actual.NSR.TotalCostPerTonne, budget.NSR.TotalCostPerTonne)},
			MarginPerTonne:          VarianceMetric{Actual: actual.NSR.MarginPerTonne, Budget: budget.NSR.MarginPerTonne, Variance: actual.NSR.MarginPerTonne - budget.NSR.MarginPerTonne, VariancePct: calculateVariancePct(actual.NSR.MarginPerTonne, budget.NSR.MarginPerTonne)},
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

// AccumulateYTD accumulates YTD values by summing current month with previous YTD
// monthDore is the Dore data for the current month (needed for gold credit calculation)
func (c *Calculator) AccumulateYTD(ytd *DataSet, month *DataSet, monthDore *data.DoreData) *DataSet {
	if month == nil {
		return ytd
	}

	if ytd == nil {
		// First month: YTD = month
		return c.copyDataSet(month)
	}

	// Accumulate: YTD = previous YTD + current month
	accumulated := &DataSet{}

	// Mining: sum (except grades which are weighted averages, and headcount which takes latest)
	totalOreYTD := ytd.Mining.OreMinedT + month.Mining.OreMinedT
	openPitOreYTD := ytd.Mining.OpenPitOreT + month.Mining.OpenPitOreT
	undergroundOreYTD := ytd.Mining.UndergroundOreT + month.Mining.UndergroundOreT

	// Calculate weighted average grades
	var miningGradeSilverYTD, miningGradeGoldYTD float64
	var opGradeSilverYTD, ugGradeSilverYTD, opGradeGoldYTD, ugGradeGoldYTD float64

	if totalOreYTD > 0 {
		miningGradeSilverYTD = (ytd.Mining.MiningGradeSilverGpt*ytd.Mining.OreMinedT +
			month.Mining.MiningGradeSilverGpt*month.Mining.OreMinedT) / totalOreYTD
		miningGradeGoldYTD = (ytd.Mining.MiningGradeGoldGpt*ytd.Mining.OreMinedT +
			month.Mining.MiningGradeGoldGpt*month.Mining.OreMinedT) / totalOreYTD
	}
	if openPitOreYTD > 0 {
		opGradeSilverYTD = (ytd.Mining.OpenPitGradeSilverGpt*ytd.Mining.OpenPitOreT +
			month.Mining.OpenPitGradeSilverGpt*month.Mining.OpenPitOreT) / openPitOreYTD
		opGradeGoldYTD = (ytd.Mining.OpenPitGradeGoldGpt*ytd.Mining.OpenPitOreT +
			month.Mining.OpenPitGradeGoldGpt*month.Mining.OpenPitOreT) / openPitOreYTD
	}
	if undergroundOreYTD > 0 {
		ugGradeSilverYTD = (ytd.Mining.UndergroundGradeSilverGpt*ytd.Mining.UndergroundOreT +
			month.Mining.UndergroundGradeSilverGpt*month.Mining.UndergroundOreT) / undergroundOreYTD
		ugGradeGoldYTD = (ytd.Mining.UndergroundGradeGoldGpt*ytd.Mining.UndergroundOreT +
			month.Mining.UndergroundGradeGoldGpt*month.Mining.UndergroundOreT) / undergroundOreYTD
	}

	// Calculate stripping ratio YTD
	var strippingRatioYTD float64
	wasteYTD := ytd.Mining.WasteMinedT + month.Mining.WasteMinedT
	if openPitOreYTD > 0 {
		strippingRatioYTD = wasteYTD / openPitOreYTD
	}

	accumulated.Mining = MiningMetrics{
		// Ore breakdown
		OpenPitOreT:     openPitOreYTD,
		UndergroundOreT: undergroundOreYTD,
		OreMinedT:       totalOreYTD,

		// Waste and ratios
		WasteMinedT:    wasteYTD,
		StrippingRatio: strippingRatioYTD,

		// Mining grades (weighted averages)
		MiningGradeSilverGpt:      miningGradeSilverYTD,
		MiningGradeGoldGpt:        miningGradeGoldYTD,
		OpenPitGradeSilverGpt:     opGradeSilverYTD,
		UndergroundGradeSilverGpt: ugGradeSilverYTD,
		OpenPitGradeGoldGpt:       opGradeGoldYTD,
		UndergroundGradeGoldGpt:   ugGradeGoldYTD,

		// Developments (sum)
		PrimaryDevelopmentM:       ytd.Mining.PrimaryDevelopmentM + month.Mining.PrimaryDevelopmentM,
		SecondaryDevelopmentOpexM: ytd.Mining.SecondaryDevelopmentOpexM + month.Mining.SecondaryDevelopmentOpexM,
		ExpansionaryDevelopmentM:  ytd.Mining.ExpansionaryDevelopmentM + month.Mining.ExpansionaryDevelopmentM,
		DevelopmentsM:             ytd.Mining.DevelopmentsM + month.Mining.DevelopmentsM,

		// Headcount (take current month - latest)
		FullTimeEmployees: month.Mining.FullTimeEmployees,
		Contractors:       month.Mining.Contractors,
		TotalHeadcount:    month.Mining.TotalHeadcount,

		HasData: ytd.Mining.HasData || month.Mining.HasData,
	}

	// Processing: sum for tonnes, weighted average for grades, ratio for recovery rates
	totalTonnesYTD := ytd.Processing.TotalTonnesProcessed + month.Processing.TotalTonnesProcessed
	if totalTonnesYTD > 0 {
		// Feed grade YTD: weighted by TotalTonnesProcessed
		feedGradeSilverYTD := (ytd.Processing.FeedGradeSilverGpt*ytd.Processing.TotalTonnesProcessed +
			month.Processing.FeedGradeSilverGpt*month.Processing.TotalTonnesProcessed) / totalTonnesYTD
		feedGradeGoldYTD := (ytd.Processing.FeedGradeGoldGpt*ytd.Processing.TotalTonnesProcessed +
			month.Processing.FeedGradeGoldGpt*month.Processing.TotalTonnesProcessed) / totalTonnesYTD
		
		// Recovery YTD: sum(recovered metal) / sum(contained metal) * 100
		// Contained metal = Feed Grade * Tonnes Processed / 31.1035 (grams per oz)
		containedSilverYTD := (ytd.Processing.FeedGradeSilverGpt*ytd.Processing.TotalTonnesProcessed +
			month.Processing.FeedGradeSilverGpt*month.Processing.TotalTonnesProcessed) / 31.1035
		containedGoldYTD := (ytd.Processing.FeedGradeGoldGpt*ytd.Processing.TotalTonnesProcessed +
			month.Processing.FeedGradeGoldGpt*month.Processing.TotalTonnesProcessed) / 31.1035
		
		// Recovered metal = Production (already accumulated)
		recoveredSilverYTD := accumulated.Production.TotalProductionSilverOz
		recoveredGoldYTD := accumulated.Production.TotalProductionGoldOz
		
		recoveryRateSilverYTD := 0.0
		recoveryRateGoldYTD := 0.0
		if containedSilverYTD > 0 {
			recoveryRateSilverYTD = (recoveredSilverYTD / containedSilverYTD) * 100
		}
		if containedGoldYTD > 0 {
			recoveryRateGoldYTD = (recoveredGoldYTD / containedGoldYTD) * 100
		}
		
		accumulated.Processing = ProcessingMetrics{
			TotalTonnesProcessed:  totalTonnesYTD,
			FeedGradeSilverGpt:    feedGradeSilverYTD,
			FeedGradeGoldGpt:       feedGradeGoldYTD,
			RecoveryRateSilverPct: recoveryRateSilverYTD,
			RecoveryRateGoldPct:    recoveryRateGoldYTD,
			HasData:               ytd.Processing.HasData || month.Processing.HasData,
		}
	} else {
		accumulated.Processing = ProcessingMetrics{
			TotalTonnesProcessed:  totalTonnesYTD,
			FeedGradeSilverGpt:    0,
			FeedGradeGoldGpt:      0,
			RecoveryRateSilverPct: 0,
			RecoveryRateGoldPct:   0,
			HasData:               ytd.Processing.HasData || month.Processing.HasData,
		}
	}

	// Production: sum
	accumulated.Production = ProductionMetrics{
		TotalProductionSilverOz: ytd.Production.TotalProductionSilverOz + month.Production.TotalProductionSilverOz,
		TotalProductionGoldOz:   ytd.Production.TotalProductionGoldOz + month.Production.TotalProductionGoldOz,
		PayableSilverOz:         ytd.Production.PayableSilverOz + month.Production.PayableSilverOz,
		PayableGoldOz:           ytd.Production.PayableGoldOz + month.Production.PayableGoldOz,
		DoreProductionOz:        ytd.Production.DoreProductionOz + month.Production.DoreProductionOz,
		HasData:                 ytd.Production.HasData || month.Production.HasData,
	}

	// Costs: sum
	accumulated.Costs = CostMetrics{
		Mine:                  ytd.Costs.Mine + month.Costs.Mine,
		Processing:            ytd.Costs.Processing + month.Costs.Processing,
		GA:                    ytd.Costs.GA + month.Costs.GA,
		TransportShipping:     ytd.Costs.TransportShipping + month.Costs.TransportShipping,
		InventoryVariations:   ytd.Costs.InventoryVariations + month.Costs.InventoryVariations,
		ProductionBasedCosts:  ytd.Costs.ProductionBasedCosts + month.Costs.ProductionBasedCosts,
		ProductionBasedMargin: ytd.Costs.ProductionBasedMargin + month.Costs.ProductionBasedMargin,
		HasData:               ytd.Costs.HasData || month.Costs.HasData,
	}

	// NSR: sum
	accumulated.NSR = NSRMetrics{
		NSRDore:                 ytd.NSR.NSRDore + month.NSR.NSRDore,
		Streaming:               ytd.NSR.Streaming + month.NSR.Streaming,
		PBRRevenue:              ytd.NSR.PBRRevenue + month.NSR.PBRRevenue,
		ShippingSelling:         ytd.NSR.ShippingSelling + month.NSR.ShippingSelling,
		SalesTaxesRoyalties:     ytd.NSR.SalesTaxesRoyalties + month.NSR.SalesTaxesRoyalties,
		SmeltingRefiningCharges: ytd.NSR.SmeltingRefiningCharges + month.NSR.SmeltingRefiningCharges,
		NetSmelterReturn:        ytd.NSR.NetSmelterReturn + month.NSR.NetSmelterReturn,
		GoldCredit:              ytd.NSR.GoldCredit + month.NSR.GoldCredit,
		// Per tonne metrics: recalculate from accumulated totals (handle division by zero)
		NSRPerTonne:       0,
		TotalCostPerTonne: 0,
		MarginPerTonne:    0,
		HasData:           ytd.NSR.HasData || month.NSR.HasData,
	}
	if accumulated.Processing.TotalTonnesProcessed > 0 {
		accumulated.NSR.NSRPerTonne = accumulated.NSR.NetSmelterReturn / accumulated.Processing.TotalTonnesProcessed
		accumulated.NSR.TotalCostPerTonne = accumulated.Costs.ProductionBasedCosts / accumulated.Processing.TotalTonnesProcessed
		accumulated.NSR.MarginPerTonne = accumulated.NSR.NSRPerTonne - accumulated.NSR.TotalCostPerTonne
	}

	// CAPEX: sum
	accumulated.CAPEX = CAPEXMetrics{
		Sustaining:                      ytd.CAPEX.Sustaining + month.CAPEX.Sustaining,
		Project:                         ytd.CAPEX.Project + month.CAPEX.Project,
		Leasing:                         ytd.CAPEX.Leasing + month.CAPEX.Leasing,
		AccretionOfMineClosureLiability: ytd.CAPEX.AccretionOfMineClosureLiability + month.CAPEX.AccretionOfMineClosureLiability,
		Total:                           ytd.CAPEX.Total + month.CAPEX.Total,
		ProductionBasedMargin:            accumulated.Costs.ProductionBasedMargin,
		PBRNetCashFlow:                   accumulated.Costs.ProductionBasedMargin - accumulated.CAPEX.Sustaining,
		HasData:                          ytd.CAPEX.HasData || month.CAPEX.HasData,
	}

	// Cash Cost: recalculate from accumulated totals
	// GoldCredit_YTD = sum(PayableGoldOz_m * RealizedPriceGold_m) for all months
	if accumulated.Production.HasData && accumulated.Costs.HasData {
		// Calculate gold credit for current month
		var monthGoldCredit float64
		if monthDore != nil && month.Production.PayableGoldOz > 0 {
			monthGoldCredit = month.Production.PayableGoldOz * monthDore.RealizedPriceGold
		}
		
		// Accumulate gold credit: YTD = previous YTD + current month
		var goldCreditYTD float64
		if ytd != nil && ytd.CashCost.HasData {
			goldCreditYTD = ytd.CashCost.GoldCredit + monthGoldCredit
		} else {
			goldCreditYTD = monthGoldCredit
		}
		
		// Cash costs for silver (production costs - gold credit)
		cashCostsSilverYTD := accumulated.Costs.ProductionBasedCosts - goldCreditYTD
		
		var cashCostPerOzSilver, aiscPerOzSilver float64
		if accumulated.Production.PayableSilverOz > 0 {
			cashCostPerOzSilver = cashCostsSilverYTD / accumulated.Production.PayableSilverOz
			aiscSilverYTD := cashCostsSilverYTD + accumulated.CAPEX.Sustaining + accumulated.CAPEX.AccretionOfMineClosureLiability
			aiscPerOzSilver = aiscSilverYTD / accumulated.Production.PayableSilverOz
			
			accumulated.CashCost = CashCostMetrics{
				CashCostPerOzSilver: cashCostPerOzSilver,
				AISCPerOzSilver:     aiscPerOzSilver,
				CashCostsSilver:     cashCostsSilverYTD,
				AISCSilver:          aiscSilverYTD,
				GoldCredit:          goldCreditYTD,
				HasData:             true,
			}
		} else {
			accumulated.CashCost = CashCostMetrics{
				CashCostPerOzSilver: 0,
				AISCPerOzSilver:     0,
				CashCostsSilver:      cashCostsSilverYTD,
				AISCSilver:           cashCostsSilverYTD + accumulated.CAPEX.Sustaining + accumulated.CAPEX.AccretionOfMineClosureLiability,
				GoldCredit:           goldCreditYTD,
				HasData:             true,
			}
		}
	}

	return accumulated
}

// copyDataSet creates a deep copy of a DataSet
func (c *Calculator) copyDataSet(ds *DataSet) *DataSet {
	if ds == nil {
		return nil
	}
	return &DataSet{
		Mining:     ds.Mining,
		Processing: ds.Processing,
		Production: ds.Production,
		Costs:      ds.Costs,
		NSR:        ds.NSR,
		CAPEX:      ds.CAPEX,
		CashCost:   ds.CashCost,
	}
}
