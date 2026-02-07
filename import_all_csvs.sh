#!/bin/bash
# Script para importar todos los CSVs de ejemplo
# Uso: ./import_all_csvs.sh [API_URL] [TOKEN]
# Si no se proporciona TOKEN, se hará login automáticamente

API_URL="${1:-http://localhost:3080}"
EXAMPLES_DIR="internal/domain/data/examples"
COMPANY_ID=1

echo "🚀 Importando todos los CSVs de ejemplo..."
echo "API URL: $API_URL"
echo "Company ID: $COMPANY_ID"
echo ""

# Función para hacer login y obtener token
login() {
    echo "🔐 Haciendo login..." >&2
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "dni": "99999999",
            "password": "admin123"
        }')
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" -ne 200 ]; then
        echo "❌ Error: Login falló con código HTTP $HTTP_CODE" >&2
        echo "Respuesta: $BODY" >&2
        return 1
    fi
    
    # Intentar extraer token con jq si está disponible, sino con grep
    if command -v jq &> /dev/null; then
        TOKEN=$(echo "$BODY" | jq -r '.token // empty')
    else
        TOKEN=$(echo "$BODY" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    fi
    
    if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
        echo "❌ Error: No se pudo obtener token de la respuesta" >&2
        echo "Respuesta completa: $BODY" >&2
        return 1
    fi
    
    echo "✅ Login exitoso" >&2
    echo "" >&2
    # Solo imprimir el token en stdout (sin >&2)
    echo "$TOKEN"
}

# Función para importar un CSV
import_csv() {
    local TYPE=$1
    local DATA_TYPE=$2
    local FILE=$3
    local TOKEN=$4
    
    if [ ! -f "$FILE" ]; then
        echo "⚠️  Archivo no encontrado: $FILE"
        return 1
    fi
    
    echo "📤 Importando: $(basename $FILE) (type=$TYPE, data_type=$DATA_TYPE)"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/api/v1/data/import" \
        -H "Authorization: Bearer $TOKEN" \
        -F "type=$TYPE" \
        -F "data_type=$DATA_TYPE" \
        -F "company_id=$COMPANY_ID" \
        -F "file=@$FILE")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" -eq 200 ]; then
        # Intentar usar jq si está disponible
        if command -v jq &> /dev/null; then
            SUCCESS=$(echo "$BODY" | jq -r '.success // false')
            ROWS_INSERTED=$(echo "$BODY" | jq -r '.rows_inserted // 0')
        else
            SUCCESS=$(echo "$BODY" | grep -o '"success":[^,}]*' | cut -d':' -f2 | tr -d ' ')
            ROWS_INSERTED=$(echo "$BODY" | grep -o '"rows_inserted":[^,}]*' | cut -d':' -f2 | tr -d ' ')
        fi
        
        if [ "$SUCCESS" = "true" ]; then
            echo "  ✅ Éxito: $ROWS_INSERTED filas insertadas"
            return 0
        else
            echo "  ❌ Error en importación:"
            if command -v jq &> /dev/null; then
                echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
            else
                echo "$BODY"
            fi
            return 1
        fi
    else
        echo "  ❌ Error HTTP $HTTP_CODE:"
        if command -v jq &> /dev/null; then
            echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
        else
            echo "$BODY"
        fi
        return 1
    fi
}

# Obtener token
if [ -z "$2" ]; then
    TOKEN=$(login)
    if [ $? -ne 0 ] || [ -z "$TOKEN" ]; then
        echo "❌ No se pudo obtener token. Abortando."
        exit 1
    fi
else
    TOKEN="$2"
    echo "✅ Usando token proporcionado"
    echo ""
fi

# Contador de errores
ERRORS=0

echo "═══════════════════════════════════════════════════════════"
echo "PASO 1: IMPORTAR BUDGET (12 meses)"
echo "═══════════════════════════════════════════════════════════"
echo ""

# Budget: PBR primero (requerido para Dore)
import_csv "pbr" "budget" "$EXAMPLES_DIR/budget_2025_pbr.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Budget: Dore (requiere PBR)
import_csv "dore" "budget" "$EXAMPLES_DIR/budget_2025_dore.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Budget: OPEX
import_csv "opex" "budget" "$EXAMPLES_DIR/budget_2025_opex.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Budget: CAPEX
import_csv "capex" "budget" "$EXAMPLES_DIR/budget_2025_capex.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Budget: Financial
import_csv "financial" "budget" "$EXAMPLES_DIR/budget_2025_financial.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Budget: Revenue
import_csv "revenue" "budget" "$EXAMPLES_DIR/budget_2025_revenue.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

echo "═══════════════════════════════════════════════════════════"
echo "PASO 2: IMPORTAR ACTUAL - ENERO"
echo "═══════════════════════════════════════════════════════════"
echo ""

# Actual Enero: PBR primero
import_csv "pbr" "actual" "$EXAMPLES_DIR/actual_2025_ene_pbr.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Enero: Dore (requiere PBR)
import_csv "dore" "actual" "$EXAMPLES_DIR/actual_2025_ene_dore.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Enero: OPEX
import_csv "opex" "actual" "$EXAMPLES_DIR/actual_2025_ene_opex.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Enero: CAPEX
import_csv "capex" "actual" "$EXAMPLES_DIR/actual_2025_ene_capex.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Enero: Financial
import_csv "financial" "actual" "$EXAMPLES_DIR/actual_2025_ene_financial.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Enero: Revenue
import_csv "revenue" "actual" "$EXAMPLES_DIR/actual_2025_ene_revenue.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

echo "═══════════════════════════════════════════════════════════"
echo "PASO 3: IMPORTAR ACTUAL - FEBRERO"
echo "═══════════════════════════════════════════════════════════"
echo ""

# Actual Febrero: PBR primero
import_csv "pbr" "actual" "$EXAMPLES_DIR/actual_2025_feb_pbr.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Febrero: Dore (requiere PBR)
import_csv "dore" "actual" "$EXAMPLES_DIR/actual_2025_feb_dore.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Febrero: OPEX
import_csv "opex" "actual" "$EXAMPLES_DIR/actual_2025_feb_opex.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Febrero: CAPEX
import_csv "capex" "actual" "$EXAMPLES_DIR/actual_2025_feb_capex.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Febrero: Financial
import_csv "financial" "actual" "$EXAMPLES_DIR/actual_2025_feb_financial.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Febrero: Revenue
import_csv "revenue" "actual" "$EXAMPLES_DIR/actual_2025_feb_revenue.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

echo "═══════════════════════════════════════════════════════════"
echo "PASO 4: IMPORTAR ACTUAL - MARZO"
echo "═══════════════════════════════════════════════════════════"
echo ""

# Actual Marzo: PBR primero
import_csv "pbr" "actual" "$EXAMPLES_DIR/actual_2025_mar_pbr.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Marzo: Dore (requiere PBR)
import_csv "dore" "actual" "$EXAMPLES_DIR/actual_2025_mar_dore.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Marzo: OPEX
import_csv "opex" "actual" "$EXAMPLES_DIR/actual_2025_mar_opex.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Marzo: CAPEX
import_csv "capex" "actual" "$EXAMPLES_DIR/actual_2025_mar_capex.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Marzo: Financial
import_csv "financial" "actual" "$EXAMPLES_DIR/actual_2025_mar_financial.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

# Actual Marzo: Revenue
import_csv "revenue" "actual" "$EXAMPLES_DIR/actual_2025_mar_revenue.csv" "$TOKEN" || ERRORS=$((ERRORS+1))
echo ""

echo "═══════════════════════════════════════════════════════════"
echo "RESUMEN"
echo "═══════════════════════════════════════════════════════════"
echo ""

if [ $ERRORS -eq 0 ]; then
    echo "✅ ¡Todos los imports fueron exitosos!"
    echo ""
    echo "🔍 Verificar datos:"
    echo "   curl -H \"Authorization: Bearer $TOKEN\" \"$API_URL/api/v1/data/list?type=pbr&company_id=$COMPANY_ID&year=2025&type_filter=budget\""
    echo ""
    echo "📊 Generar Summary (validará automáticamente):"
    echo "   curl -H \"Authorization: Bearer $TOKEN\" \"$API_URL/api/v1/reports/summary?company_id=$COMPANY_ID&year=2025\" | jq"
else
    echo "❌ Se encontraron $ERRORS errores durante la importación"
    echo "   Revisa los mensajes de error arriba"
    exit 1
fi
