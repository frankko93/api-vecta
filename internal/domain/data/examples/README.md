# CSVs de Ejemplo - Estructura Organizada

## 📁 Organización:

### Budget (Anuales - 12 meses):
- `budget_2025_pbr.csv` - Plan Beneficio Regional (12 meses)
- `budget_2025_dore.csv` - Doré (12 meses)
- `budget_2025_opex.csv` - Costos operativos (varios registros)
- `budget_2025_capex.csv` - Inversiones (12 meses)
- `budget_2025_financial.csv` - Ajustes financieros (12 meses)

### Actual (Mensuales - 1 mes cada uno):
- `actual_2025_ene_{tipo}.csv` - Datos reales Enero
- `actual_2025_feb_{tipo}.csv` - Datos reales Febrero
- `actual_2025_mar_{tipo}.csv` - Datos reales Marzo

Donde {tipo} = pbr, dore, opex, capex, financial

## 🎯 Orden de Importación:

1. **Primero:** Budget anual (5 archivos)
2. **Después:** Actual mes a mes (5 archivos por mes)

## 📊 Total:
- Budget: 5 archivos
- Actual: 15 archivos (5 por cada mes)
- **Total: 20 archivos**
