#!/bin/bash
# Limpia empresas y datos importados pero MANTIENE usuarios

cd "$(dirname "$0")/.."

echo "🧹 LIMPIANDO DATOS (manteniendo usuarios)"
echo "========================================="
echo ""

docker exec vecta_postgres_local psql -U postgres -d vecta_db << 'SQL'
-- Limpiar datos importados
DELETE FROM production_data;
DELETE FROM dore_data;
DELETE FROM pbr_data;
DELETE FROM opex_data;
DELETE FROM capex_data;
DELETE FROM revenue_data;
DELETE FROM financial_data;

-- Limpiar reportes guardados
DELETE FROM saved_reports;

-- Limpiar empresas (esto limpia settings y minerals por CASCADE)
DELETE FROM mining_companies;

SELECT 
    'Users: ' || COUNT(*) as estado FROM users
UNION ALL
SELECT 'Companies: ' || COUNT(*) FROM mining_companies
UNION ALL
SELECT 'Data tables: ' || COUNT(*) FROM pbr_data;
SQL

echo ""
echo "✅ Datos limpiados!"
echo ""
echo "Mantenido:"
echo "  ✅ Usuarios (3)"
echo "  ✅ Minerales (7)"
echo ""
echo "Limpiado:"
echo "  ❌ Empresas"
echo "  ❌ Datos importados"
echo "  ❌ Reportes guardados"
echo ""
echo "Ahora el frontend puede crear empresas desde cero"
