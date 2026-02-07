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

## 🎯 Orden de Importación (CRÍTICO):

**IMPORTANTE:** El orden es crítico. Dore depende de PBR, por lo que PBR debe importarse primero.

### Paso 1: Budget (5 archivos en este orden)
1. `budget_2025_pbr.csv` - **PRIMERO** (requerido para Dore)
2. `budget_2025_dore.csv` - Requiere PBR ya importado
3. `budget_2025_opex.csv`
4. `budget_2025_capex.csv`
5. `budget_2025_financial.csv`

### Paso 2: Actual - Por mes (5 archivos por mes en este orden)
Para cada mes (Enero, Febrero, Marzo):
1. `actual_2025_{mes}_pbr.csv` - **PRIMERO** (requerido para Dore)
2. `actual_2025_{mes}_dore.csv` - Requiere PBR ya importado
3. `actual_2025_{mes}_opex.csv`
4. `actual_2025_{mes}_capex.csv`
5. `actual_2025_{mes}_financial.csv`

## 📊 Total:
- Budget: 5 archivos
- Actual: 15 archivos (5 por cada mes: Enero, Febrero, Marzo)
- **Total: 20 archivos**

## ⚠️ Validaciones Automáticas

Después de importar, cuando generes el Summary report, el sistema validará:
- ✅ Alineación de meses entre todos los archivos
- ✅ Dependencias Dore → PBR (cada fecha en Dore debe tener PBR)
- ✅ Consistencia de años

Si alguna validación falla, el Summary retornará error descriptivo.

## 📖 Ver Guía Completa

Ver `IMPORT_GUIDE.md` en la raíz del proyecto para instrucciones detalladas de importación desde el frontend.
