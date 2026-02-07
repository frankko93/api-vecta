# Vecta API - Especificación para Frontend

Sistema para administración y visualización de datos mineros con reportes financieros.

## 🎯 Visión del Sistema

Sistema que permite:
1. **Gestionar empresas mineras** y sus configuraciones
2. **Importar datos** desde CSVs (producción, costos, ingresos, etc)
3. **Generar reportes automáticos** con cálculos financieros
4. **Comparar datos reales vs presupuestos** (múltiples versiones)
5. **Tomar decisiones** basadas en métricas calculadas

---

## 🔐 Autenticación

Todos los endpoints requieren autenticación con **Bearer Token** en header:
```
Authorization: Bearer {token}
```

### Login
```
POST /api/v1/auth/login

Request:
{
  "dni": "99999999",
  "password": "admin123"
}

Response:
{
  "token": "abc123...",  // 64 caracteres - guardar para requests
  "user": {
    "id": 1,
    "first_name": "Admin",
    "last_name": "User",
    "dni": "99999999",
    "birth_date": "1990-01-01",
    "work_area": "IT",
    "active": true,
    "permissions": ["admin"],  // admin, editor, viewer
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

### Usuario Actual
```
GET /api/v1/auth/me

Response: (mismo que user en login)
```

### Logout
```
POST /api/v1/auth/logout

Response:
{
  "message": "logged out successfully"
}
```

### Listar Usuarios (con paginación)
```
GET /api/v1/users?page=1&size=10

Response:
{
  "data": [
    {
      "id": 1,
      "first_name": "Admin",
      "last_name": "User",
      "dni": "99999999",
      "birth_date": "1990-01-01",
      "work_area": "IT",
      "active": true,
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "size": 10,
    "total": 1
  }
}
```

---

## 🏢 Configuración

### Listar Empresas Mineras
```
GET /api/v1/config/companies

Response:
[
  {
    "id": 1,
    "name": "Cerro Moro",
    "legal_name": "Cerro Moro S.A.",
    "tax_id": "20-12345678-9",
    "address": "Av. Principal 123",
    "contact_email": "contact@cerromoro.com",
    "contact_phone": "+54 261 123-4567",
    "active": true,
    "created_at": "2025-01-01T00:00:00Z"
  }
]
```

### Ver Empresa con Detalles
```
GET /api/v1/config/companies/{id}

Response:
{
  "id": 1,
  "name": "Cerro Moro",
  "legal_name": "Cerro Moro S.A.",
  ...
  "minerals": [
    { "id": 1, "name": "Oro", "code": "AU" },
    { "id": 2, "name": "Plata", "code": "AG" }
  ],
  "settings": {
    "company_id": 1,
    "is_open_pit": true,
    "country": "Argentina",
    "royalty_percentage": 3.5,
    "notes": ""
  }
}
```

### Crear Empresa (Admin)
```
POST /api/v1/config/companies

Request:
{
  "name": "Nueva Minera",
  "legal_name": "Nueva Minera S.A.",
  "tax_id": "20-87654321-9",
  "address": "Dirección",
  "contact_email": "email@minera.com",
  "contact_phone": "+54 261 999-9999",
  "is_open_pit": true,
  "country": "Argentina",
  "royalty_percentage": 3.5
}
```

### Listar Minerales
```
GET /api/v1/config/minerals

Response:
[
  { "id": 1, "name": "Oro", "code": "AU", "active": true },
  { "id": 2, "name": "Plata", "code": "AG", "active": true },
  { "id": 3, "name": "Cobre", "code": "CU", "active": true }
]

Nota: 7 minerales pre-cargados (Au, Ag, Cu, Zn, Pb, Li, Fe)
```

### Unidades de Medida Disponibles
```
GET /api/v1/config/units

Response:
{
  "data": [
    { "value": "tonnes", "label": "Toneladas" },
    { "value": "kilograms", "label": "Kilogramos" },
    { "value": "grams", "label": "Gramos" },
    { "value": "troy_ounces", "label": "Onzas Troy" }
  ]
}
```

---

## 📊 Importación de Datos

### Importar CSV
```
POST /api/v1/data/import (multipart/form-data)

Params:
- file: archivo.csv
- type: "pbr" | "dore" | "opex" | "capex" | "financial" | "production" | "revenue"
- data_type: "actual" | "budget"
- company_id: number
- version: number (opcional, default: 1)
- description: string (opcional, ej: "Budget ajustado Junio")

Response Exitosa:
{
  "success": true,
  "type": "pbr",
  "rows_total": 12,
  "rows_inserted": 12,
  "rows_failed": 0,
  "errors": []
}

Response con Errores:
{
  "success": false,
  "type": "pbr",
  "rows_total": 12,
  "rows_inserted": 0,
  "rows_failed": 2,
  "errors": [
    { "row": 5, "column": "ore_mined_t", "error": "invalid number: abc" },
    { "row": 8, "column": "date", "error": "invalid date format" }
  ]
}

Nota: Si HAY errores, NADA se inserta (transacción atómica)
```

### Listar Datos Importados
```
GET /api/v1/data/{type}/list
  ?company_id=1
  &year=2025
  &data_type=budget
  &version=2

Params:
- type: pbr, dore, opex, capex, financial
- company_id: required
- year: required
- data_type: required (actual o budget)
- version: opcional (default: 1)

Response: Array de datos según tipo
[
  {
    "id": 123,
    "company_id": 1,
    "date": "2025-01-15",
    "ore_mined_t": 24859,
    ... (campos específicos del tipo),
    "version": 2,
    "description": "Budget ajustado",
    "created_at": "2025-01-01T00:00:00Z"
  }
]
```

### Eliminar Datos (Soft Delete)
```
DELETE /api/v1/data/{type}/{id}

Example: DELETE /api/v1/data/pbr/123

Response:
{
  "message": "data deleted successfully"
}
```

---

## 📈 Reportes

### Summary Report (Principal)
```
GET /api/v1/reports/summary
  ?company_id=1
  &year=2025
  &months=1,2,3  (opcional - si no se pasa, devuelve los 12 meses)

Response:
{
  "company_id": 1,
  "company_name": "Cerro Moro",
  "year": 2025,
  "months": [
    {
      "month": "2025-01",
      "actual": {
        "mining": {
          "ore_mined_t": 24859,
          "waste_mined_t": 262591,
          "developments_m": 598,
          "has_data": true
        },
        "processing": {
          "total_tonnes_processed": 35951,
          "feed_grade_silver_gpt": 209.79,
          "feed_grade_gold_gpt": 7.35,
          "recovery_rate_silver_pct": 94.01,
          "recovery_rate_gold_pct": 95.36,
          "has_data": true
        },
        "production": {
          "total_production_silver_oz": 227957,  // CALCULADO
          "total_production_gold_oz": 8106,      // CALCULADO
          "payable_silver_oz": 227957,
          "payable_gold_oz": 8106,
          "has_data": true
        },
        "costs": {
          "mine": 8537997,
          "processing": 3613678,
          "ga": 5471220,
          "transport_shipping": 0,
          "inventory_variations": 1740162,
          "production_based_costs": 19363057,    // CALCULADO (suma)
          "production_based_margin": 0,
          "has_data": true
        },
        "nsr": {
          "nsr_dore": 27919108,                  // CALCULADO
          "shipping_selling": -202,
          "sales_taxes_royalties": 465867,
          "net_smelter_return": 27453443,        // CALCULADO
          "nsr_per_tonne": 763.6,                // CALCULADO
          "total_cost_per_tonne": 538.6,         // CALCULADO
          "margin_per_tonne": 225.0,             // CALCULADO
          "has_data": true
        },
        "capex": {
          "sustaining": 711052,
          "project": 0,
          "leasing": 0,
          "total": 711052,
          "production_based_margin": 8090385,    // CALCULADO (NSR - Costs)
          "pbr_net_cash_flow": 7379333,          // CALCULADO (Margin - CAPEX)
          "has_data": true
        },
        "cash_cost": {
          "cash_cost_per_oz_silver": -9.38,      // CALCULADO
          "aisc_per_oz_silver": -5.15,           // CALCULADO (incluye CAPEX)
          "gold_credit": 16090410,               // CALCULADO (by-product)
          "has_data": true
        }
      },
      "budget": {
        // MISMA ESTRUCTURA que actual
        "mining": { "has_data": true, ... },
        "nsr": { "has_data": false }  // ← Alerta si falta budget
      }
    },
    // ... resto de meses
  ],
  "ytd": null  // Por ahora null, futuro: acumulado del año
}
```

---

## 🎨 Guía de UI Sugerida

### **Secciones Principales:**

#### 1. **Dashboard / Home**
- Vista rápida de KPIs principales
- Última empresa seleccionada
- Alertas de datos faltantes

#### 2. **Empresas** (`/companies`)
- Lista de empresas mineras
- CRUD (admin only)
- Ver/editar configuración por empresa

#### 3. **Importación de Datos** (`/import`)
- Selector de empresa
- Selector de tipo (PBR, Dore, OPEX, etc)
- Selector: Actual vs Budget
- Input version (ej: Budget v2)
- Input description (ej: "Budget ajustado Junio")
- Upload CSV
- Mostrar errores si falló validación
- Historial de imports (con botón eliminar)

#### 4. **Reportes** (`/reports`)
- Selector de empresa
- Selector de año
- Selector de meses (checkboxes 1-12)
- Selector de versión de budget (dropdown: v1, v2, v3...)
- **Vista Summary** (tabla como Excel con filas):
  - Mining (Ore, Waste, Developments)
  - Processing (Tonnes, Grades, Recovery)
  - Production (Silver oz, Gold oz)
  - Costs breakdown
  - NSR metrics
  - CAPEX
  - Cash Cost & AISC
- Columnas: Actual | Budget | Fav(Unf) | % Variance
- **Gráficos sugeridos:**
  - Production trend (line chart)
  - Costs breakdown (pie chart)
  - Margin evolution (bar chart)
  - Actual vs Budget comparison (dual-axis chart)

---

## 🚨 Campos Importantes

### **Flags `has_data`**
```javascript
// Usar para mostrar alertas
if (!month.actual.nsr.has_data) {
  // Mostrar: "⚠️ Falta cargar datos de Dore para calcular NSR"
}

if (!month.budget.mining.has_data) {
  // Mostrar: "⚠️ No hay budget cargado para este mes"
}
```

### **Versiones de Budget**
```javascript
// Permitir seleccionar qué versión de budget comparar
<select>
  <option value="1">Budget Original (v1)</option>
  <option value="2">Budget Ajustado Junio (v2)</option>
  <option value="3">Budget Actualizado Sept (v3)</option>
</select>

// Luego filtrar en el GET:
GET /api/v1/reports/summary?company_id=1&year=2025&budget_version=2
```

### **Gestión de Imports**
```javascript
// Listar imports para poder corregir errores
GET /api/v1/data/pbr/list?company_id=1&year=2025&data_type=actual

// Mostrar tabla con botón eliminar por si se equivocaron
[
  { id: 123, date: "2025-01-15", version: 1, description: "...", [DELETE] }
]

// Eliminar y permitir re-importar
DELETE /api/v1/data/pbr/123
```

---

## 📋 Formatos CSV Esperados

Todos los CSVs **deben tener headers en la primera fila**.

### PBR (Plan Beneficio Regional)
```csv
date,ore_mined_t,waste_mined_t,developments_m,total_tonnes_processed,feed_grade_silver_gpt,feed_grade_gold_gpt,recovery_rate_silver_pct,recovery_rate_gold_pct
2025-01-15,24859,262591,598,35951,209.79,7.35,94.01,95.36
```

### Dore
```csv
date,dore_produced_oz,silver_grade_pct,gold_grade_pct,pbr_price_silver,pbr_price_gold,realized_price_silver,realized_price_gold,silver_adjustment_oz,gold_adjustment_oz,ag_deductions_pct,au_deductions_pct,treatment_charge,refining_deductions_au
2025-01-15,236064,85.5,14.5,24.50,2000,24.30,1985,10,5,2.5,1.5,5000,1200
```

### OPEX
```csv
date,cost_center,subcategory,expense_type,amount,currency
2025-01-15,Mine,Drilling,Labour,50000,USD
```
Cost centers: `Mine`, `Processing`, `G&A`, `Transport & Shipping`

### CAPEX
```csv
date,category,car_number,project_name,type,amount,currency
2025-01-15,Mine Equipment,C487MY25001,Equipment Purchase,sustaining,500000,USD
```
Types: `sustaining`, `project`, `leasing`

### Financial
```csv
date,shipping_selling,sales_taxes_royalties,other_adjustments
2025-01-15,-202,465867,0
```

Ejemplos completos en: `internal/domain/data/examples/`

---

## 🔢 Cálculos Automáticos (Backend)

El frontend **NO necesita calcular** esto, el backend ya lo devuelve:

```
✓ Total Production (desde PBR + recovery rates)
✓ Payable Metal (desde Dore con deductions)
✓ NSR Dore → Net Smelter Return
✓ Production Based Costs (suma de OPEX)
✓ Production Based Margin (NSR - Costs)
✓ PBR Net Cash Flow (Margin - CAPEX sustaining)
✓ NSR per Tonne
✓ Cost per Tonne
✓ Margin per Tonne
✓ Cash Cost per Oz Silver
✓ AISC per Oz Silver
✓ Gold Credit (by-product)
```

Frontend solo debe:
- ✅ Mostrar los valores
- ✅ Calcular Fav(Unf) = Actual - Budget
- ✅ Calcular % Variance = (Actual - Budget) / Budget * 100
- ✅ Generar gráficos

---

## 🎨 Flujo de Usuario Típico

1. **Login** → Obtener token
2. **Seleccionar empresa** → Ver lista y seleccionar
3. **Importar Budget anual** (12 meses):
   - Tipo: PBR, data_type: budget, version: 1
   - Tipo: OPEX, data_type: budget, version: 1
   - Tipo: CAPEX, data_type: budget, version: 1
   - Tipo: Financial, data_type: budget, version: 1
4. **Importar Actual mensual** (a medida que cierra cada mes):
   - Enero: PBR, Dore, OPEX, CAPEX, Financial (actual)
   - Febrero: PBR, Dore, OPEX, CAPEX, Financial (actual)
5. **Ver Reporte**:
   - Seleccionar año: 2025
   - Seleccionar meses: 1, 2 (o todos)
   - Ver tabla comparativa Actual vs Budget
   - Ver gráficos
6. **Si budget cambia** (ej: en Junio):
   - Importar nuevo budget completo (12 meses)
   - Version: 2
   - Description: "Budget ajustado desde Junio"
   - En reportes: seleccionar qué versión comparar

---

## 💡 Sugerencias de UX

### Validaciones en Upload:
- Mostrar preview del CSV antes de importar
- Validar formato de fecha client-side (YYYY-MM-DD)
- Mostrar errores fila por fila si falla

### Alertas Inteligentes:
```
⚠️ Falta cargar Dore de Enero (no se puede calcular NSR)
⚠️ No hay budget para comparar
⚠️ Production muy baja vs budget (-35%)
✓ Todos los datos de Enero están completos
```

### Navegación:
- Breadcrumbs: Home > Empresas > Cerro Moro > Reportes 2025
- Tabs: Summary | Detalles | Comparaciones | Exportar

### Gráficos Clave:
1. **Production Trend** (line): Silver oz & Gold oz por mes
2. **Cost Breakdown** (pie): Mine, Processing, G&A
3. **Actual vs Budget** (bar chart comparativo)
4. **Margin Evolution** (area chart)
5. **AISC Trend** (line con threshold)

---

## 🔄 Estados y Permisos

### Permisos:
- **viewer**: Solo VER empresas, minerales, reportes
- **editor**: VER + IMPORTAR datos
- **admin**: TODO (crear empresas, eliminar datos, etc)

### Estados de UI:
```javascript
// Loading states
isLoadingCompanies, isLoadingReport, isImporting

// Error states
importError, reportError

// Success states
importSuccess: "✓ 12 filas importadas correctamente"
```

---

## 📡 Manejo de Errores

```javascript
// Errores comunes y cómo manejarlos

401 Unauthorized:
→ Redirigir a /login
→ Token expiró (6 horas)

400 Bad Request con errors[]:
→ Mostrar tabla de errores del CSV
→ "Fila 5, columna 'ore_mined_t': valor inválido"

404 Not Found:
→ Empresa no existe
→ Datos no encontrados

500 Internal Server Error:
→ "Error del servidor, contactar soporte"
```

---

## 🚀 Quick Start para Desarrollo

```bash
# 1. Setup backend
docker-compose -f docker-compose-infra.yml up -d postgres
psql -h localhost -U postgres -d vecta_db -f database/schema.sql
psql -h localhost -U postgres -d vecta_db -f database/config_schema.sql
psql -h localhost -U postgres -d vecta_db -f database/data_schema.sql
go run cmd/go8/main.go

# Backend corre en: http://localhost:3080

# 2. Login de prueba
DNI: 99999999
Password: admin123
Permisos: admin

# 3. Empresa de prueba (crear con POST)
# 4. CSVs de ejemplo en: internal/domain/data/examples/
```

---

## 📞 Soporte

**Backend URL:** `http://localhost:3080`  
**Swagger (si está habilitado):** `http://localhost:3080/swagger/`  
**Health Check:** `http://localhost:3080/api/health/readiness`

**Tests:** `go test ./internal/domain/... -v`  
**E2E:** `docker-compose -f e2e/docker-compose.yml up --build`

