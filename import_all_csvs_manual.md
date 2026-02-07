# Comandos cURL para Importación Manual

Esta guía contiene todos los comandos cURL necesarios para importar los CSVs de ejemplo manualmente.

## 🔐 Paso 0: Login y Obtener Token

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:3080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "dni": "99999999",
    "password": "admin123"
  }' | grep -o '"token":"[^"]*' | cut -d'"' -f4)

echo "Token: $TOKEN"
```

O manualmente:
```bash
curl -X POST http://localhost:3080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "dni": "99999999",
    "password": "admin123"
  }'
```

Copia el `token` de la respuesta y úsalo en los siguientes comandos reemplazando `{TOKEN}`.

---

## 📦 PASO 1: IMPORTAR BUDGET (12 meses)

### 1.1 PBR Budget (PRIMERO - Requerido para Dore)

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=pbr" \
  -F "data_type=budget" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/budget_2025_pbr.csv"
```

### 1.2 Dore Budget (Requiere PBR ya importado)

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=dore" \
  -F "data_type=budget" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/budget_2025_dore.csv"
```

### 1.3 OPEX Budget

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=opex" \
  -F "data_type=budget" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/budget_2025_opex.csv"
```

### 1.4 CAPEX Budget

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=capex" \
  -F "data_type=budget" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/budget_2025_capex.csv"
```

### 1.5 Financial Budget

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=financial" \
  -F "data_type=budget" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/budget_2025_financial.csv"
```

---

## 📦 PASO 2: IMPORTAR ACTUAL - ENERO

### 2.1 PBR Actual Enero (PRIMERO - Requerido para Dore)

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=pbr" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_ene_pbr.csv"
```

### 2.2 Dore Actual Enero (Requiere PBR ya importado)

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=dore" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_ene_dore.csv"
```

### 2.3 OPEX Actual Enero

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=opex" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_ene_opex.csv"
```

### 2.4 CAPEX Actual Enero

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=capex" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_ene_capex.csv"
```

### 2.5 Financial Actual Enero

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=financial" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_ene_financial.csv"
```

---

## 📦 PASO 3: IMPORTAR ACTUAL - FEBRERO

### 3.1 PBR Actual Febrero

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=pbr" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_feb_pbr.csv"
```

### 3.2 Dore Actual Febrero

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=dore" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_feb_dore.csv"
```

### 3.3 OPEX Actual Febrero

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=opex" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_feb_opex.csv"
```

### 3.4 CAPEX Actual Febrero

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=capex" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_feb_capex.csv"
```

### 3.5 Financial Actual Febrero

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=financial" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_feb_financial.csv"
```

---

## 📦 PASO 4: IMPORTAR ACTUAL - MARZO

### 4.1 PBR Actual Marzo

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=pbr" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_mar_pbr.csv"
```

### 4.2 Dore Actual Marzo

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=dore" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_mar_dore.csv"
```

### 4.3 OPEX Actual Marzo

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=opex" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_mar_opex.csv"
```

### 4.4 CAPEX Actual Marzo

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=capex" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_mar_capex.csv"
```

### 4.5 Financial Actual Marzo

```bash
curl -X POST http://localhost:3080/api/v1/data/import \
  -H "Authorization: Bearer {TOKEN}" \
  -F "type=financial" \
  -F "data_type=actual" \
  -F "company_id=1" \
  -F "file=@internal/domain/data/examples/actual_2025_mar_financial.csv"
```

---

## 🔍 Verificación Post-Importación

### Verificar datos importados (PBR Budget)

```bash
curl -H "Authorization: Bearer {TOKEN}" \
  "http://localhost:3080/api/v1/data/list?type=pbr&company_id=1&year=2025&type_filter=budget" | jq
```

### Generar Summary Report (Valida automáticamente)

```bash
curl -H "Authorization: Bearer {TOKEN}" \
  "http://localhost:3080/api/v1/reports/summary?company_id=1&year=2025" | jq
```

Si hay errores de validación, se mostrarán claramente en la respuesta.

---

## 📝 Script Automatizado

Para importar todo automáticamente, usa el script:

```bash
chmod +x import_all_csvs.sh
./import_all_csvs.sh
```

O con URL y token personalizados:

```bash
./import_all_csvs.sh http://localhost:3080 {TOKEN}
```

---

## ⚠️ Notas Importantes

1. **Orden crítico:** PBR debe importarse antes de Dore
2. **Token válido:** Reemplaza `{TOKEN}` con el token real obtenido del login
3. **Paths relativos:** Los paths de archivos son relativos a la raíz del proyecto
4. **Validaciones:** El Summary validará automáticamente alineación y dependencias
