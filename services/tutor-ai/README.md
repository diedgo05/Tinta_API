# tutor-ai — Microservicio de tutoría con LLM + RAG

Servicio Python + FastAPI que expone el chat del Tutor IA con recuperación
aumentada (RAG) sobre PDFs subidos por los usuarios.

## Arquitectura

```
Cliente Flutter
    │
    │ HTTPS + JWT (identity)
    ▼
tutor-ai (este servicio)
    │
    ├─→ Gemma 3 4B GGUF Q4_K_M   (llama-cpp-python)
    ├─→ MiniLM multilingüe        (sentence-transformers)
    └─→ Postgres + pgvector       (compartido con otros servicios)
```

## Estructura

Sigue Clean Architecture como el resto del monorepo:

```
services/tutor-ai/
├── src/
│   ├── main.py               FastAPI entrypoint + lifespan
│   ├── config.py             Settings tipadas (Pydantic)
│   ├── api/                  Endpoints REST + auth
│   ├── domain/               Entidades y prompts
│   ├── application/          Servicios de aplicación
│   └── infrastructure/       LLM, embeddings, DB, PDF
├── migrations/               SQL para el esquema `knowledge`
├── scripts/                  Descarga de modelos + init DB
├── Dockerfile                Multi-stage con modelos baked-in
├── railway.toml              Config de Railway
└── requirements.txt
```

## Endpoints

Todos protegidos con JWT del servicio `identity` (compartido vía clave pública).

### `POST /api/v1/tutor/documents`
Sube un PDF. Devuelve 202 con `document_id` en estado `pending`.
Procesamiento en background: extract → chunk → embed → save.

### `GET /api/v1/tutor/documents/{document_id}`
Consulta el estado del procesamiento.
Estados: `pending`, `processing`, `ready`, `failed`.

### `POST /api/v1/tutor/chat`
Chat con streaming SSE. Cuerpo:
```json
{
  "question": "¿Qué es la Guerra de Reforma?",
  "document_id": "550e8400-...",   // opcional; activa RAG
  "history": [...]
}
```
Respuesta: stream `text/event-stream` con `{"token": "..."}` por cada token
y al final `{"done": true, "sources": [...]}`.

### `GET /api/v1/tutor/health`
Health check. Devuelve estado de LLM, embeddings y DB.

## Deploy en Railway

### 1. Crear un servicio nuevo

En tu proyecto de Railway:

- **New** → **GitHub Repo** → selecciona `Tinta_API`
- **Root Directory**: `services/tutor-ai`
- El builder detectará `railway.toml` y usará Dockerfile.

### 2. Configurar variables de entorno

Copia el contenido de `.env.example` en Settings → Variables. Los valores
críticos:

- `DATABASE_URL`: usa el mismo Postgres compartido de tus otros servicios.
  Añade `?search_path=knowledge,public` al final.
- `JWT_PUBLIC_KEY`: pega la clave RSA pública del servicio `identity`
  con saltos de línea reales.

### 3. Aplicar migraciones

Una sola vez, después del primer deploy:

```bash
railway run --service tutor-ai python -m scripts.init_db
```

Esto crea el esquema `knowledge` con sus 3 tablas + índices HNSW.

### 4. Deploy y healthcheck

El primer build tarda 15-25 minutos (descarga de Gemma 3 4B).
Los siguientes solo cambian código, cachean la layer de modelos.

Verificar:
```bash
curl https://tutor-ai-production.up.railway.app/api/v1/tutor/health
```

Respuesta esperada:
```json
{
  "status": "ok",
  "llm_loaded": true,
  "embeddings_loaded": true,
  "db_connected": true
}
```

## Probar el flujo completo con curl

Antes de tocar Flutter, se puede probar todo con curl. Necesitas:
- Un JWT válido emitido por `identity`
- Un PDF chico (2-10 páginas para probar rápido)

### 1. Subir un PDF

```bash
TOKEN="eyJhbGc..."

curl -X POST https://tutor-ai-production.up.railway.app/api/v1/tutor/documents \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@ejemplo.pdf"
```

Respuesta:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "ejemplo.pdf",
  "status": "pending",
  ...
}
```

### 2. Esperar a que esté listo

```bash
DOC_ID="550e8400-..."

curl https://tutor-ai-production.up.railway.app/api/v1/tutor/documents/$DOC_ID \
  -H "Authorization: Bearer $TOKEN"
```

Repite hasta ver `"status": "ready"`.

### 3. Preguntar al tutor

```bash
curl -N -X POST https://tutor-ai-production.up.railway.app/api/v1/tutor/chat \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "¿De qué trata este documento?",
    "document_id": "'$DOC_ID'",
    "history": []
  }'
```

Verás los tokens llegando en streaming SSE:
```
data: {"token": "Este "}
data: {"token": "documento "}
data: {"token": "trata "}
...
data: {"done": true, "sources": [{"page_number": 1, "excerpt": "...", ...}]}
```

## Desarrollo local

```bash
cd services/tutor-ai
cp .env.example .env
# editar .env con valores locales

# Descargar modelos (unos GB, tarda)
python scripts/download_models.py

# Postgres local con pgvector
docker run -d --name tinta-pg -e POSTGRES_PASSWORD=dev \
  -p 5432:5432 pgvector/pgvector:pg16

# Aplicar migraciones
DATABASE_URL="postgresql://postgres:dev@localhost:5432/postgres" \
  python scripts/init_db.py

# Correr el servicio
pip install -r requirements.txt
uvicorn src.main:app --reload
```

## Recursos y costos estimados

Corriendo en Railway Hobby con 1 usuario haciendo consultas moderadas:

| Recurso | Valor |
|---|---|
| RAM en steady state | ~3 GB |
| RAM peak durante inferencia | ~4 GB |
| CPU en inferencia | 100% × ~10 seg por respuesta |
| Imagen Docker | ~3.5 GB |
| Costo mensual estimado | $10-20 USD |

## Limitaciones conocidas

- **Cold start**: cargar el modelo tarda 10-15 seg. Railway apaga containers
  inactivos en plan Hobby, así que la primera consulta después de un rato
  será lenta.
- **Sin persistencia del chat**: cada request es independiente. El historial
  lo mantiene el cliente y lo envía en cada request.
- **RAG limitado a un documento por consulta**: la búsqueda solo cubre el
  `document_id` enviado. La base académica se añadirá en el futuro.
- **Modelo pequeño (4B)**: para respuestas más ricas se puede escalar a
  Gemma 3 12B cambiando el modelo en el Dockerfile y subiendo el plan a Pro.
