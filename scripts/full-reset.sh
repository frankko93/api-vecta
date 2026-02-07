#!/bin/bash
# Full reset y setup de prueba completo

cd "$(dirname "$0")/.."

echo "🔄 RESET COMPLETO DEL SISTEMA"
echo "=============================="
echo ""

# 1. Stop API
echo "1️⃣ Deteniendo API..."
./scripts/api-stop.sh

# 2. Reset PostgreSQL
echo ""
echo "2️⃣ Reseteando PostgreSQL..."
docker-compose down -v
docker-compose up -d
sleep 5

# 3. Ejecutar schemas
echo ""
echo "3️⃣ Ejecutando schemas..."
docker exec -i vecta_postgres_local psql -U postgres -d vecta_db < database/schema.sql > /dev/null 2>&1
docker exec -i vecta_postgres_local psql -U postgres -d vecta_db < database/config_schema.sql > /dev/null 2>&1
docker exec -i vecta_postgres_local psql -U postgres -d vecta_db < database/data_schema.sql > /dev/null 2>&1
docker exec -i vecta_postgres_local psql -U postgres -d vecta_db < database/reports_schema.sql > /dev/null 2>&1

# 4. Crear usuarios de prueba
echo ""
echo "4️⃣ Creando usuarios de prueba..."
docker exec -i vecta_postgres_local psql -U postgres -d vecta_db << 'SQL'
-- Usuario Editor
INSERT INTO users (first_name, last_name, dni, birth_date, work_area, password_hash, active)
VALUES ('Editor', 'Test', '88888888', '1992-05-15', 'Operations', 
        '$argon2id$v=19$m=65536,t=1,p=11$26wRAe/3D66n2EZzzR0QNw$FLiJupf5T0vQCFLryzB2gWdrR4jLMX8sFVAfq2UbnwE',
        true);

-- Usuario Viewer
INSERT INTO users (first_name, last_name, dni, birth_date, work_area, password_hash, active)
VALUES ('Viewer', 'Test', '77777777', '1993-08-20', 'Finance',
        '$argon2id$v=19$m=65536,t=1,p=11$26wRAe/3D66n2EZzzR0QNw$FLiJupf5T0vQCFLryzB2gWdrR4jLMX8sFVAfq2UbnwE',
        true);

-- Asignar permisos
INSERT INTO user_permissions (user_id, permission_id)
SELECT u.id, p.id FROM users u, permissions p 
WHERE u.dni = '88888888' AND p.name = 'editor';

INSERT INTO user_permissions (user_id, permission_id)
SELECT u.id, p.id FROM users u, permissions p 
WHERE u.dni = '77777777' AND p.name = 'viewer';

SELECT 
    u.first_name || ' ' || u.last_name as nombre,
    u.dni,
    array_agg(p.name) as permisos
FROM users u
LEFT JOIN user_permissions up ON u.id = up.user_id
LEFT JOIN permissions p ON up.permission_id = p.id
GROUP BY u.id, u.first_name, u.last_name, u.dni
ORDER BY u.id;
SQL

# 5. Verificar minerales cargados
echo ""
echo "5️⃣ Verificando minerales disponibles..."
docker exec vecta_postgres_local psql -U postgres -d vecta_db -c "SELECT id, code, name FROM minerals WHERE active = true ORDER BY id"

echo ""
echo "✅ Setup completo!"
echo ""
echo "👥 Usuarios creados:"
echo "   - DNI: 99999999 / Password: admin123 (admin) - puede crear empresas"
echo "   - DNI: 88888888 / Password: admin123 (editor) - puede importar datos"
echo "   - DNI: 77777777 / Password: admin123 (viewer) - solo ver"
echo ""
echo "💎 Minerales disponibles: Au, Ag, Cu, Zn, Pb, Li, Fe"
echo ""
echo "📋 Próximos pasos:"
echo "   1. ./scripts/api-restart.sh"
echo "   2. Login desde frontend como admin"
echo "   3. Crear empresa (POST /api/v1/config/companies)"
echo "   4. Asignar minerales (PUT /api/v1/config/companies/{id}/minerals)"
echo "   5. Importar datos"
