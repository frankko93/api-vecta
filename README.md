# Vecta API

API para administración y visualización de datos mineros.

## 🚀 Inicio Rápido

```bash
# 1. Setup (solo primera vez)
./scripts/db-setup.sh

# 2. Start API
./scripts/api-restart.sh

# API: http://localhost:3080
```

## 📋 Comandos

```bash
./scripts/db-setup.sh       # Setup PostgreSQL + schemas
./scripts/api-restart.sh    # Rebuild + restart API
./scripts/api-stop.sh       # Detener API
./scripts/reset-all.sh      # Limpiar todo
```

## 🧪 Tests

```bash
go test ./internal/domain/... -v     # 20 tests unitarios
docker-compose -f e2e/docker-compose.yml up --build  # E2E tests
```

## 👤 Datos de Prueba

- **Usuario:** DNI `99999999` / Password `admin123`
- **Empresas:** Cerro Moro (Argentina), Minera Los Andes (Chile)

## 📄 Documentación

- **FRONTEND_API_SPEC.md** - Especificación completa para frontend
- **SETUP.md** - Guía detallada de setup
- **IMPORT_GUIDE.md** - Guía de importación de CSVs desde frontend
- **DATA_SETUP_CHECKLIST.md** - Checklist para setup de datos de prueba
- **INPUT_CONTRACT_AND_CALCULATION_RULEBOOK.md** - Contrato de entrada y reglas de cálculo

## 🎯 Endpoints Principales

```
POST /api/v1/auth/login
GET  /api/v1/users
GET  /api/v1/config/companies
GET  /api/v1/config/minerals
POST /api/v1/data/import
GET  /api/v1/reports/summary
```

Ver **FRONTEND_API_SPEC.md** para detalles completos.
