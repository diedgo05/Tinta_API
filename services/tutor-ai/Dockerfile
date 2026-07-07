# ═══════════════════════════════════════════════════════════════════════════
# Dockerfile multi-stage para tutor-ai
#
# Stage 1 (models): descarga Gemma 3 4B y MiniLM. Se cachea entre builds
#                    porque cambia rara vez.
# Stage 2 (runtime): imagen final con Python + dependencies + modelos + código.
#
# Tamaño final aproximado: ~3.5 GB (mayoría es Gemma + MiniLM + torch).
# ═══════════════════════════════════════════════════════════════════════════

# ── STAGE 1: Descarga de modelos ─────────────────────────────────────────
FROM python:3.11-slim AS models

# Solo necesitamos sentence-transformers para la parte de embeddings
RUN pip install --no-cache-dir sentence-transformers==3.2.1

WORKDIR /app
COPY scripts/download_models.py ./scripts/
RUN python scripts/download_models.py

# En este punto tenemos:
#   /models/gemma-3-4b-it-q4_k_m.gguf
#   /root/.cache/huggingface/hub/... (embeddings model)


# ── STAGE 2: Runtime ─────────────────────────────────────────────────────
FROM python:3.11-slim AS runtime

# Dependencias del sistema para llama-cpp-python
RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential \
        cmake \
        libpq-dev \
        curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copiar requirements primero para cachear la layer de pip install
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Copiar modelos desde el stage anterior
COPY --from=models /models /models
COPY --from=models /root/.cache/huggingface /root/.cache/huggingface

# Copiar código fuente y migraciones
COPY src ./src
COPY migrations ./migrations
COPY scripts ./scripts

# Directorio temporal para uploads
RUN mkdir -p /tmp/tinta-uploads

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PORT=8000

EXPOSE 8000

# Railway inyecta $PORT; el shell expande la variable en runtime.
CMD uvicorn src.main:app --host 0.0.0.0 --port ${PORT:-8000}
