# Tinta Backend

Backend del proyecto **Tinta** — tutor académico privado con IA local.

Esta API expone los servicios de negocio que viven en el servidor (usuarios, autenticación, clubes de lectura, recomendaciones generadas por ML). **No** ejecuta el LLM ni el RAG: eso vive dentro de la app móvil del usuario.

## Arquitectura

Monorepo con tres microservicios independientes en Go, siguiendo **Clean Architecture** y **Screaming Architecture**:

```
tinta-backend/
├── services/
│   ├── identity/         # Usuarios, login, JWT  → puerto 8001
│   ├── community/        # Clubes de lectura     → puerto 8002
│   └── recommendations/  # Resultados de ML      → puerto 8003
├── shared/               # Código compartido (JWT auth, middlewares, logger)
├── scripts/              # Scripts de inicialización y seed
└── docs/                 # OpenAPI por servicio
```

Cada microservicio:
- Tiene su propio binario, puerto y migraciones.
- Usa **su propio esquema** dentro de la base PostgreSQL `tinta` (`identity`, `community`, `recommendations`).
- Verifica JWT con la clave pública del servicio Identity (sólo Identity firma; los demás verifican).
- Sigue capas: `domain → application → ports → infrastructure`.

## Stack

| Categoría | Tecnología |
|---|---|
| Lenguaje | Go 1.22 |
| Web framework | Fiber v2 |
| Acceso a datos | sqlc + pgx |
| Migraciones | golang-migrate |
| BD relacional | PostgreSQL 16 |
| Caché / sesiones | Redis 7 |
| Autenticación | JWT RS256 + Argon2id |
| Contenedores | Docker + Docker Compose |

## Prerrequisitos

- **Docker** y **Docker Compose**
- **Go 1.22+**
- **make**
- **openssl** (para generar llaves JWT)
- **sqlc** ([instalación](https://docs.sqlc.dev/en/latest/overview/install.html))
- **golang-migrate** ([instalación](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate))

Para instalar las herramientas de Go:

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Levantar el proyecto (paso a paso)

```bash
# 1. Generar llaves JWT (sólo la primera vez)
make keys

# 2. Levantar infraestructura (Postgres + Redis + los 3 servicios)
make up

# 3. Aplicar migraciones a los tres esquemas
make migrate-all

# 4. Insertar los 3 admins iniciales
make seed
```

## Comandos útiles

Lista completa: `make help`

```bash
make up                  # Levantar todo
make down                # Detener todo
make logs SERVICE=identity   # Ver logs de un servicio
make migrate-identity    # Migraciones sólo de identity
make sqlc-all            # Regenerar código de sqlc en los 3 servicios
make test                # Tests de todos los servicios
```

## Versionado de la API

Todas las rutas tienen prefijo `/api/v1`. Cuando rompamos contratos, subimos a `/api/v2` y mantenemos `v1` hasta que la app móvil migre.

## Endpoints disponibles (V1 inicial)

### Identity (puerto 8001)
- `POST /api/v1/users` — Registrar usuario
- `GET  /api/v1/users/me` — Mi perfil (requiere JWT)
- `GET  /api/v1/users/{id}` — Perfil público
- `PATCH /api/v1/users/me` — Actualizar mi perfil
- `DELETE /api/v1/users/me` — Eliminar mi cuenta (ARCO)
- `POST /api/v1/auth/login` — Login (email + password)
- `POST /api/v1/auth/refresh` — Renovar tokens
- `POST /api/v1/auth/logout` — Cerrar sesión

### Community (puerto 8002)
- `POST   /api/v1/clubs` — Crear club
- `GET    /api/v1/clubs` — Listar clubes (paginado)
- `GET    /api/v1/clubs/{id}` — Detalle
- `PATCH  /api/v1/clubs/{id}` — Actualizar (solo creador)
- `DELETE /api/v1/clubs/{id}` — Eliminar (solo creador)

### Recommendations (puerto 8003)
- `GET    /api/v1/recommendations` — Mis recomendaciones
- `POST   /api/v1/recommendations/{id}/feedback` — Feedback (👍 / 👎)
- `DELETE /api/v1/recommendations/{id}` — Descartar
- `POST   /api/v1/recommendations/regenerate` — Forzar regeneración

## Credenciales iniciales (solo desarrollo)

Tras correr `make seed`:

| Email | Password | Rol |
|---|---|---|
| adrian@tinta.app | admin123 | admin |
| diego@tinta.app | admin123 | admin |
| gael@tinta.app | admin123 | admin |
| system@tinta.app | admin123 | system |

> ⚠️ Cambiar estas contraseñas en producción.

## Probar la API rápido (curl)

```bash
# 1. Registrar un usuario
curl -X POST http://localhost:8001/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@tinta.app","password":"test1234","name":"Test User"}'

# 2. Login (devuelve access_token y refresh_token)
TOKEN=$(curl -s -X POST http://localhost:8001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"adrian@tinta.app","password":"admin123"}' \
  | jq -r '.access_token')

# 3. Obtener mi perfil (requiere JWT)
curl http://localhost:8001/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"

# 4. Crear un club
curl -X POST http://localhost:8002/api/v1/clubs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Cien años de soledad","description":"Lectura colectiva","is_private":false}'

# 5. Listar clubes
curl "http://localhost:8002/api/v1/clubs?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"

# 6. Ver mis recomendaciones (vacío hasta que el pipeline ML genere)
curl http://localhost:8003/api/v1/recommendations \
  -H "Authorization: Bearer $TOKEN"
```

## Equipo

| Matrícula | Nombre | Grupo |
|---|---|---|
| 233405 | Ángel Adrian Sánchez García | C |
| 233358 | Diego Jiménez Pérez | C |
| — | Gael Andre Hueytlelt Villalobos | C |
