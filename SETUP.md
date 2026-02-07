# Setup Local - Vecta API

## 🚀 Inicio Rápido

```bash
# 1. Setup database (solo primera vez)
./scripts/db-setup.sh

# 2. Start API
./scripts/api-restart.sh

# API en: http://localhost:3080
```

## 📋 Comandos Disponibles

```bash
# Setup completo de database (crea tablas, seeds)
./scripts/db-setup.sh

# Iniciar API (foreground - ver logs)
./scripts/api-start.sh

# Restart API (rebuild + start en background)
./scripts/api-restart.sh

# Detener API
./scripts/api-stop.sh

# Reset todo (limpia DB y API)
./scripts/reset-all.sh
```

## 🧪 Verificar que Funciona

```bash
# Health check
curl http://localhost:3080/version
curl http://localhost:3080/api/health/readiness

# Login
curl -X POST http://localhost:3080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"dni":"99999999","password":"admin123"}'
```

## 👤 Datos de Prueba

**Usuario:**
- DNI: `99999999`
- Password: `admin123`
- Permisos: `admin`

**Empresas:**
- Cerro Moro (Argentina) - Minerales: Au, Ag
- Minera Los Andes (Chile) - Minerales: Cu, Au, Ag

## 📝 Logs

```bash
# Ver logs de la API
tail -f api.log

# Ver logs de PostgreSQL
docker logs -f vecta_postgres_local
```

## 🔧 Troubleshooting

**API no conecta a BD:**
```bash
# Verificar que PostgreSQL esté corriendo
docker ps | grep postgres

# Verificar que la BD exista
docker exec vecta_postgres_local psql -U postgres -l
```

**Puerto 3080 ocupado:**
```bash
lsof -ti:3080 | xargs kill -9
```

**Reset completo:**
```bash
./scripts/reset-all.sh
./scripts/db-setup.sh
./scripts/api-restart.sh
```

