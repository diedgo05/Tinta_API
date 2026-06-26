# Tinta · Backend (Go monorepo)

API SOA de **Tinta**, plataforma de club de lectura digital con tutor IA on-device.

## 📦 Servicios

| Servicio        | Puerto | Esquema BD       | Función                                                         |
|-----------------|--------|------------------|-----------------------------------------------------------------|
| identity        | 8001   | identity         | Usuarios, autenticación JWT, registro                           |
| community       | 8002   | community        | Clubes de lectura                                               |
| recommendations | 8003   | recommendations  | Recomendaciones del motor ML                                    |
| **catalog**     | 8004   | catalog          | Libros, géneros, capítulos                                      |
| **reading**     | 8005   | reading          | Progreso de lectura + anotaciones                               |
| **knowledge**   | 8006   | knowledge        | Temas académicos y fragmentos del RAG (base de conocimientos)   |
| **notifications** | 8007 | notifications    | Notificaciones in-app                                           |

## 🚀 Setup local

```bash
# 1. Generar llaves JWT
make keys

# 2. Levantar todo (Postgres + Redis + 7 servicios)
make up

# 3. Aplicar migraciones (Esto solo funciona con `migrate` instalado nativo)
make migrate-all

# 4. Insertar admins iniciales
make seed
```

Credenciales precargadas tras `make seed`:
- `adrian@tinta.app` / `admin123`
- `diego@tinta.app` / `admin123`
- `gael@tinta.app` / `admin123`
- `system@tinta.app` / `admin123`

## 📂 Estructura

```
.
├── go.work                          # Workspace Go (8 módulos)
├── docker-compose.yml               # Local: Postgres + Redis + 7 servicios
├── Makefile                         # Comandos make
├── scripts/
│   ├── init-db.sql                  # Crea 7 esquemas separados (SOA)
│   └── seed/main.go                 # Inserta admins iniciales
├── shared/                          # Código compartido entre servicios
│   ├── httpx/                       # Helpers de respuesta HTTP
│   ├── jwtauth/                     # Firmador/verificador JWT RS256
│   ├── logger/                      # Wrapper de zerolog
│   └── middleware/                  # RequireAuth middleware
└── services/
    ├── identity/                    # 8001
    ├── community/                   # 8002
    ├── recommendations/             # 8003
    ├── catalog/                     # 8004 ← NUEVO
    ├── reading/                     # 8005 ← NUEVO
    ├── knowledge/                   # 8006 ← NUEVO
    └── notifications/               # 8007 ← NUEVO
```

Cada servicio sigue **Clean Architecture + Screaming Architecture**:

```
services/<svc>/
├── cmd/api/main.go                          # Composition root
├── internal/
│   ├── <bounded-context>/                   # Carpetas que "gritan" el dominio
│   │   ├── domain/                          # Entidades + errores + validación
│   │   ├── ports/                           # Interfaces (puertos)
│   │   ├── application/                     # Casos de uso
│   │   └── infrastructure/
│   │       ├── postgres/                    # Repo concreto (pgx)
│   │       └── http/                        # Handler Fiber + DTOs
│   └── platform/                            # config, database pool, server
├── migrations/                              # SQL up/down
├── sqlc/                                    # queries.sql (sqlc opcional)
├── sqlc.yaml
├── Dockerfile                               # Multi-stage + COPY keys
└── go.mod
```

## 🛣️ Endpoints

### Identity (8001)
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `POST /api/v1/users` (registro público)
- `GET  /api/v1/users/me`
- `PATCH /api/v1/users/me`
- `DELETE /api/v1/users/me`

### Community (8002)
- `POST /api/v1/clubs`
- `GET /api/v1/clubs`
- `GET /api/v1/clubs/:id`
- `PATCH /api/v1/clubs/:id` (solo creador)
- `DELETE /api/v1/clubs/:id` (solo creador)

### Recommendations (8003)
- `GET /api/v1/recommendations`
- `POST /api/v1/recommendations/:id/feedback`
- `DELETE /api/v1/recommendations/:id`
- `POST /api/v1/recommendations/regenerate`

### Catalog (8004) — NUEVO

**Libros** (lectura pública, escritura requiere auth)
- `GET /api/v1/books` · `?page=1&page_size=20&genre_id=&search=`
- `GET /api/v1/books/:id`
- `POST /api/v1/books`
- `PATCH /api/v1/books/:id`
- `DELETE /api/v1/books/:id`

**Géneros**
- `GET /api/v1/genres`
- `GET /api/v1/genres/:id`
- `POST /api/v1/genres`
- `PATCH /api/v1/genres/:id`
- `DELETE /api/v1/genres/:id`

### Reading (8005) — NUEVO

**Progreso de lectura** (todos requieren auth)
- `POST /api/v1/reading` · iniciar/actualizar progreso (upsert)
- `GET /api/v1/reading` · listar mi progreso · `?status=reading|paused|finished|abandoned`
- `GET /api/v1/reading/:book_id` · progreso de un libro
- `PATCH /api/v1/reading/:book_id`
- `DELETE /api/v1/reading/:book_id`

**Anotaciones**
- `POST /api/v1/annotations`
- `GET /api/v1/annotations` · `?book_id=` o `?personal_doc_id=`
- `PATCH /api/v1/annotations/:id` (solo dueño)
- `DELETE /api/v1/annotations/:id` (solo dueño)

### Knowledge (8006) — NUEVO

**Catálogo de temas (público)**
- `GET /api/v1/topics`
- `GET /api/v1/topics/:id`

**Selección del usuario (auth)**
- `PUT /api/v1/topics/me` · reemplaza selección · body: `{"topic_ids":[uuid,...]}` (2-5)
- `GET /api/v1/topics/me`
- `POST /api/v1/topics/me/:topic_id/downloaded`
- `DELETE /api/v1/topics/me/:topic_id`

**Fragmentos del RAG (auth) — el celular descarga la base de conocimientos**
- `GET /api/v1/topics/:topic_id/fragments` · `?page=1&page_size=100`
- `GET /api/v1/topics/:topic_id/documents`
- `POST /api/v1/documents` (admin)
- `POST /api/v1/fragments` (admin)

### Notifications (8007) — NUEVO

Todos requieren auth.
- `GET /api/v1/notifications` · `?page=1&page_size=20&unread=true`
- `POST /api/v1/notifications` (creación interna)
- `POST /api/v1/notifications/:id/read`
- `POST /api/v1/notifications/read-all`
- `DELETE /api/v1/notifications/:id`

## 🚢 Despliegue en Railway

**Dockerfiles ya están configurados** con `COPY keys /app/keys` para que las llaves JWT estén dentro de la imagen. Para desplegar:

1. **Crea un servicio nuevo en Railway** por cada microservicio (catalog, reading, knowledge, notifications).
2. **Apunta cada servicio al mismo repo** y configura el path del Dockerfile (`services/<svc>/Dockerfile`).
3. **Variables de entorno** que debes configurar en cada servicio:
   - `HTTP_PORT` = 8004 / 8005 / 8006 / 8007 (Railway las puede sobrescribir con su `PORT`)
   - `DATABASE_URL` = el `DATABASE_URL` del Postgres de Railway con `?search_path=<schema>` añadido
   - `JWT_PUBLIC_KEY_PATH` = `/app/keys/jwt_public.pem`
   - `LOG_LEVEL` = `info`
4. **Antes del primer arranque**, ejecuta las migraciones del nuevo esquema:
   ```sql
   CREATE SCHEMA IF NOT EXISTS catalog;
   CREATE SCHEMA IF NOT EXISTS reading;
   CREATE SCHEMA IF NOT EXISTS knowledge;
   CREATE SCHEMA IF NOT EXISTS notifications;
   ```
   Y aplica las migraciones SQL con `psql` o desde Railway.

⚠️ **IMPORTANTE**: Lo que ya está desplegado en Railway (identity, community, recommendations) **NO se ve afectado** por estos cambios. Los nuevos servicios son independientes.

## 🧪 Probar con Postman

Importa los archivos en `Tinta_Postman.zip`:
- `Tinta_API_v1.postman_collection.json`
- `Tinta_Local.postman_environment.json`
- `Tinta_Railway.postman_environment.json` (URLs de tus servicios desplegados)

## 📋 Próximos pasos (Turno 2)

- Email verification + password recovery (Identity)
- Members + discussions (Community)
- Pipeline ML real con Asynq + Redis (Recommendations)
- OpenAPI/Swagger por servicio
- Kong API Gateway
- Tests unitarios
