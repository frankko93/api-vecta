# Análisis de Atributos y Cálculos: Lo que Tenemos vs Lo que Necesitamos

## 📊 Estructura del Reporte Completo

### Pestañas del Reporte (según CSV completo):
1. **Summary** - Resumen ejecutivo
2. **PBR** - Detalle de Plan Beneficio Regional
3. **Dore** - Detalle de producción de dore
4. **OPEX** - Detalle de costos operativos
5. **CAPEX** - Detalle de gastos de capital
6. **Financial** - Detalle financiero
7. **Production** - Detalle de producción
8. **Revenue** - Detalle de ingresos

---

## 1. SUMMARY - Comparación de Atributos

### ✅ Lo que TENEMOS

```go
SummaryReport {
  CompanyID, CompanyName, Year
  Months[] {
    Month: "2025-01"
    Actual: DataSet
    Budget: DataSet
  }
  YTD: ComparisonData (pero retorna nil)
}
```

**DataSet incluye:**
- Mining: OreMinedT, WasteMinedT, DevelopmentsM
- Processing: TotalTonnesProcessed, FeedGradeSilverGpt, FeedGradeGoldGpt, RecoveryRateSilverPct, RecoveryRateGoldPct
- Production: TotalProductionSilverOz, TotalProductionGoldOz, PayableSilverOz, PayableGoldOz
- Costs: Mine, Processing, GA, TransportShipping, InventoryVariations, ProductionBasedCosts, ProductionBasedMargin
- NSR: NSRDore, ShippingSelling, SalesTaxesRoyalties, NetSmelterReturn, NSRPerTonne, TotalCostPerTonne, MarginPerTonne
- CAPEX: Sustaining, Project, Leasing, Total, ProductionBasedMargin, PBRNetCashFlow
- CashCost: CashCostPerOzSilver, AISCPerOzSilver, GoldCredit
```

### 🔴 Lo que FALTA en Summary

#### A. Agregaciones Temporales
- ❌ **Quarters (Q1, Q2, Q3, Q4)** - Agregación trimestral
- ❌ **Semesters (H1, H2)** - Agregación semestral
- ❌ **YTD (Year to Date)** - Implementado pero retorna `nil`
- ❌ **YearTotal** - Total anual

**Necesitamos:**
```go
SummaryReport {
  Months: []MonthlyData
  Quarters: []QuarterlyData    // NUEVO
  Semesters: []SemesterData    // NUEVO
  YTD: ComparisonData         // Implementar cálculo real
  YearTotal: ComparisonData   // NUEVO
}
```

#### B. Comparaciones y Variances
- ❌ **Variance (Fav/Unf)** - Diferencia entre Actual y Budget
- ❌ **% Variance** - Porcentaje de variación

**Necesitamos agregar a cada métrica:**
```go
type MetricWithVariance struct {
  Actual    float64
  Budget    float64
  Variance  float64  // Actual - Budget
  VariancePct float64 // ((Actual - Budget) / Budget) * 100
}
```

#### C. Campos Faltantes en DataSet

**En ProductionMetrics:**
- ✅ TotalProductionSilverOz
- ✅ TotalProductionGoldOz
- ✅ PayableSilverOz
- ✅ PayableGoldOz
- ❌ **DoreProductionOz** - Total de dore producido (Silver + Gold)

**En NSRMetrics:**
- ✅ NSRDore
- ✅ ShippingSelling
- ✅ SalesTaxesRoyalties
- ✅ NetSmelterReturn
- ❌ **SmeltingRefiningCharges** - Separado de NSR Dore
- ❌ **OtherSalesDeductions** - Deducciones adicionales

**En CostMetrics:**
- ✅ Todos los campos básicos
- ✅ ProductionBasedMargin (pero se calcula en CAPEX, debería estar aquí)

**En CAPEXMetrics:**
- ✅ Sustaining
- ✅ Project
- ✅ Leasing
- ✅ Total
- ❌ **AccretionOfMineClosureLiability** - No está en ningún modelo

**En CashCostMetrics:**
- ✅ CashCostPerOzSilver
- ✅ AISCPerOzSilver
- ✅ GoldCredit
- ❌ **CashCostsSilver** - Valor total antes de dividir por onzas
- ❌ **AISCSilver** - Valor total antes de dividir por onzas

---

## 2. PBR (Plan Beneficio Regional) - Pestaña Detallada

### ✅ Lo que TENEMOS
- Datos básicos de PBR en el modelo
- Cálculo de producción desde PBR

### 🔴 Lo que FALTA
- ❌ **Pestaña completa de PBR** - Solo tenemos datos en Summary
- ❌ **Desglose mensual detallado de PBR**
- ❌ **Comparaciones Actual vs Budget en PBR**
- ❌ **Ratios calculados:**
  - ❌ Waste/Ore Ratio
  - ❌ Total Moved (Ore + Waste)
  - ❌ Processing Efficiency

**Necesitamos endpoint:**
```
GET /api/v1/reports/pbr?company_id=1&year=2025&months=1,2,3
```

---

## 3. DORE - Pestaña Detallada

### ✅ Lo que TENEMOS
- Modelo DoreData completo
- Cálculo de producción desde PBR
- Cálculo de NSR Dore

### 🔴 Lo que FALTA
- ❌ **Pestaña completa de Dore** - Solo tenemos datos en Summary
- ❌ **Desglose detallado:**
  - ❌ Metal in Dore (antes de ajustes)
  - ❌ Metal Adjusted (después de ajustes)
  - ❌ Deductions (Ag, Au)
  - ❌ Payable Metal (después de deducciones)
  - ❌ Gross Revenue (Silver + Gold)
  - ❌ Treatment Charges (separado)
  - ❌ Refining Deductions (separado)
  - ❌ NSR Dore (después de cargos)

**Necesitamos endpoint:**
```
GET /api/v1/reports/dore?company_id=1&year=2025&months=1,2,3
```

**Campos adicionales necesarios en respuesta:**
```go
type DoreDetailMetrics struct {
  MetalInDoreSilverOz    float64
  MetalInDoreGoldOz       float64
  MetalAdjustedSilverOz   float64
  MetalAdjustedGoldOz     float64
  DeductionsSilverOz      float64
  DeductionsGoldOz        float64
  PayableSilverOz         float64
  PayableGoldOz           float64
  GrossRevenueSilver      float64
  GrossRevenueGold        float64
  TreatmentCharges        float64
  RefiningDeductions      float64
  NSRDore                 float64
}
```

---

## 4. OPEX - Pestaña Detallada

### ✅ Lo que TENEMOS
- Modelo OPEXData completo
- Agrupación por Cost Center
- Cálculo de costos totales

### 🔴 Lo que FALTA
- ❌ **Pestaña completa de OPEX** - Solo tenemos agregados en Summary
- ❌ **Desglose por:**
  - ❌ Subcategory (detalle de cada subcategoría)
  - ❌ ExpenseType (tipo de gasto)
  - ❌ Por mes con comparaciones
  - ❌ Por Cost Center con comparaciones
- ❌ **Agregaciones:**
  - ❌ Por trimestre
  - ❌ Por semestre
  - ❌ YTD

**Necesitamos endpoint:**
```
GET /api/v1/reports/opex?company_id=1&year=2025&group_by=month|quarter|cost_center|subcategory
```

**Estructura de respuesta necesaria:**
```go
type OPEXDetailReport struct {
  ByMonth: []OPEXMonthlyData
  ByQuarter: []OPEXQuarterlyData
  ByCostCenter: map[string]OPEXCostCenterData
  BySubcategory: map[string]OPEXSubcategoryData
  Total: OPEXTotalData
}
```

---

## 5. CAPEX - Pestaña Detallada

### ✅ Lo que TENEMOS
- Modelo CAPEXData completo
- Agrupación por Type (sustaining, project, leasing)
- Cálculo de PBR Net Cash Flow

### 🔴 Lo que FALTA
- ❌ **Pestaña completa de CAPEX** - Solo tenemos agregados en Summary
- ❌ **Desglose por:**
  - ❌ Category (Mine Equipment, etc.)
  - ❌ Project Name
  - ❌ CAR Number
  - ❌ Por mes con comparaciones
- ❌ **Campo faltante:**
  - ❌ **AccretionOfMineClosureLiability** - No está en el modelo
- ❌ **Agregaciones:**
  - ❌ Por trimestre
  - ❌ Por semestre
  - ❌ YTD

**Necesitamos agregar al modelo CAPEX:**
```go
type CAPEXData struct {
  // ... campos existentes
  AccretionOfMineClosureLiability float64 // NUEVO
}
```

**Necesitamos endpoint:**
```
GET /api/v1/reports/capex?company_id=1&year=2025&group_by=month|quarter|type|category
```

---

## 6. FINANCIAL - Pestaña Detallada

### ✅ Lo que TENEMOS
- Modelo FinancialData básico
- ShippingSelling, SalesTaxesRoyalties, OtherAdjustments

### 🔴 Lo que FALTA
- ❌ **Pestaña completa de Financial** - Solo tenemos datos en Summary
- ❌ **Campo faltante:**
  - ❌ **OtherSalesDeductions** - Mencionado en CSV pero no en modelo
- ❌ **Desglose detallado:**
  - ❌ Por mes
  - ❌ Comparaciones Actual vs Budget
  - ❌ Impacto en NSR

**Necesitamos verificar modelo:**
```go
type FinancialData struct {
  ShippingSelling     float64
  SalesTaxesRoyalties float64
  OtherAdjustments    float64
  OtherSalesDeductions float64 // ¿Es lo mismo que OtherAdjustments?
}
```

---

## 7. PRODUCTION - Pestaña Detallada

### ✅ Lo que TENEMOS
- Modelo ProductionData (para otros minerales)
- Cálculo de producción desde PBR (Silver y Gold)

### 🔴 Lo que FALTA
- ❌ **Pestaña completa de Production**
- ❌ **Desglose por mineral:**
  - ❌ Producción por mineral (no solo Silver/Gold)
  - ❌ Comparaciones Actual vs Budget
  - ❌ Por mes, trimestre, semestre, YTD

**Necesitamos endpoint:**
```
GET /api/v1/reports/production?company_id=1&year=2025&mineral_id=1,2,3
```

---

## 8. REVENUE - Pestaña Detallada

### ✅ Lo que TENEMOS
- Modelo RevenueData básico

### 🔴 Lo que FALTA
- ❌ **Pestaña completa de Revenue**
- ❌ **Desglose detallado:**
  - ❌ Revenue por mineral
  - ❌ Revenue por mes
  - ❌ Comparaciones Actual vs Budget
  - ❌ Unit Price trends

**Necesitamos endpoint:**
```
GET /api/v1/reports/revenue?company_id=1&year=2025&mineral_id=1,2,3
```

---

## 📋 Resumen de Atributos Faltantes por Prioridad

### 🔴 CRÍTICOS (Deben agregarse)

1. **Agregaciones Temporales**
   - Quarters (Q1-Q4)
   - Semesters (H1-H2)
   - YTD (implementar cálculo real)
   - YearTotal

2. **Variances**
   - Variance (Fav/Unf) para cada métrica
   - % Variance para cada métrica

3. **Campos faltantes en modelos:**
   - `AccretionOfMineClosureLiability` en CAPEX
   - `OtherSalesDeductions` en Financial (o clarificar si es OtherAdjustments)
   - `DoreProductionOz` en ProductionMetrics
   - `SmeltingRefiningCharges` separado en NSRMetrics
   - `CashCostsSilver` y `AISCSilver` (valores totales) en CashCostMetrics

### ⚠️ IMPORTANTES (Deben implementarse)

4. **Pestañas detalladas:**
   - Endpoint `/reports/pbr` con desglose completo
   - Endpoint `/reports/dore` con desglose completo
   - Endpoint `/reports/opex` con desglose completo
   - Endpoint `/reports/capex` con desglose completo
   - Endpoint `/reports/financial` con desglose completo
   - Endpoint `/reports/production` con desglose completo
   - Endpoint `/reports/revenue` con desglose completo

5. **Desgloses adicionales:**
   - Por subcategoría en OPEX
   - Por proyecto en CAPEX
   - Por mineral en Production y Revenue
   - Ratios calculados (Waste/Ore, etc.)

### 📝 MEJORAS (Opcionales)

6. **Métricas derivadas:**
   - Waste/Ore Ratio
   - Total Moved
   - Processing Efficiency
   - Recovery Efficiency trends

---

## 🎯 Plan de Implementación Sugerido

### Fase 1: Expandir Summary
1. Agregar campos faltantes a modelos existentes
2. Implementar cálculo de Variances
3. Implementar agregaciones temporales (Quarters, Semesters, YTD, YearTotal)

### Fase 2: Crear Pestañas Detalladas
4. Implementar endpoint `/reports/pbr`
5. Implementar endpoint `/reports/dore`
6. Implementar endpoint `/reports/opex`
7. Implementar endpoint `/reports/capex`
8. Implementar endpoint `/reports/financial`
9. Implementar endpoint `/reports/production`
10. Implementar endpoint `/reports/revenue`

### Fase 3: Agregar Desgloses
11. Agregar desgloses por subcategoría, proyecto, mineral
12. Agregar ratios y métricas derivadas
