"""
Entrypoint del servicio tutor-ai.

Inicializa FastAPI, configura el lifespan (carga de modelos y conexión a
DB al arrancar; cierre limpio al detenerse), y registra los routers.
"""
from __future__ import annotations

import logging
from contextlib import asynccontextmanager

import structlog
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from src.api import chat, documents, health
from src.application.chat_service import ChatService
from src.application.document_service import DocumentService
from src.config import get_settings
from src.infrastructure.embeddings.minilm_model import EmbeddingsModel
from src.infrastructure.llm.llama_runner import LlamaRunner
from src.infrastructure.pdf.extractor import PdfExtractor
from src.infrastructure.vector_db.pgvector_repo import PgVectorRepo


# ── Logging estructurado ─────────────────────────────────────────
def _configure_logging(level: str) -> None:
    logging.basicConfig(
        format="%(message)s",
        level=level,
    )
    structlog.configure(
        processors=[
            structlog.contextvars.merge_contextvars,
            structlog.processors.add_log_level,
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.JSONRenderer(),
        ],
    )


@asynccontextmanager
async def lifespan(app: FastAPI):
    """
    Ciclo de vida de la app:
        - Startup: carga LLM, embeddings, conecta a DB, inicializa servicios.
        - Shutdown: cierra conexiones, libera memoria.

    Railway espera health check en menos de 5 min. La carga del LLM
    tarda ~10-15 seg, así que hay tiempo de sobra.
    """
    settings = get_settings()
    _configure_logging(settings.log_level)
    log = structlog.get_logger()

    log.info("startup.begin", service=settings.service_name, env=settings.environment)

    # 1. Infrastructure
    llm = LlamaRunner(settings)
    embeddings = EmbeddingsModel(settings)
    repo = PgVectorRepo(settings)
    pdf_extractor = PdfExtractor(
        chunk_size=settings.chunk_size,
        chunk_overlap=settings.chunk_overlap,
    )

    # 2. Cargar en paralelo lo que sí es paralelizable
    #    (DB primero, luego modelos secuenciales para evitar OOM)
    await repo.connect()
    await embeddings.load()
    await llm.load()

    # 3. Application services
    chat_service = ChatService(settings, llm, embeddings, repo)
    document_service = DocumentService(pdf_extractor, embeddings, repo)

    # 4. Poner todo en app.state para las dependencies
    app.state.llm = llm
    app.state.embeddings = embeddings
    app.state.repo = repo
    app.state.chat_service = chat_service
    app.state.document_service = document_service

    log.info("startup.ready")

    yield  # ← aquí corre la app

    # Shutdown
    log.info("shutdown.begin")
    await llm.unload()
    await repo.close()
    log.info("shutdown.done")


# ── App instance ────────────────────────────────────────────────
app = FastAPI(
    title="Tinta Tutor AI",
    version="0.1.0",
    description="Microservicio de tutoría con LLM y RAG.",
    lifespan=lifespan,
)

# CORS: en producción restringir a los orígenes reales de la app.
# Para el MVP con Flutter mobile no aplica (mobile no manda Origin),
# pero por si se prueba desde web.
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# ── Routers ─────────────────────────────────────────────────────
API_PREFIX = "/api/v1/tutor"

app.include_router(health.router, prefix=API_PREFIX)
app.include_router(chat.router, prefix=API_PREFIX)
app.include_router(documents.router, prefix=API_PREFIX)


# ── Root ────────────────────────────────────────────────────────
@app.get("/")
async def root():
    return {"service": "tutor-ai", "version": "0.1.0"}
