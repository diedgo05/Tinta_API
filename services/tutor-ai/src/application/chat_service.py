"""
Servicio de aplicación del chat con RAG.

Orquesta:
    1. Si hay document_id: recuperar chunks relevantes (RAG)
    2. Armar el system prompt (con o sin contexto RAG)
    3. Construir el prompt de Gemma con historial
    4. Ejecutar generación en streaming
    5. Al final, incluir las fuentes usadas
"""
from __future__ import annotations

from typing import AsyncIterator, Optional
from uuid import UUID

import structlog

from src.config import Settings
from src.domain.entities import ChatMessage, ChunkWithScore, Source
from src.domain.prompts import build_gemma_chat_prompt, build_system_prompt
from src.infrastructure.embeddings.minilm_model import EmbeddingsModel
from src.infrastructure.llm.llama_runner import LlamaRunner
from src.infrastructure.vector_db.pgvector_repo import PgVectorRepo

log = structlog.get_logger()


class ChatService:
    def __init__(
        self,
        settings: Settings,
        llm: LlamaRunner,
        embeddings: EmbeddingsModel,
        repo: PgVectorRepo,
    ):
        self._settings = settings
        self._llm = llm
        self._embeddings = embeddings
        self._repo = repo

    async def stream_response(
        self,
        question: str,
        history: list[ChatMessage],
        document_id: Optional[UUID],
    ) -> AsyncIterator[dict]:
        """
        Genera la respuesta del tutor como stream de eventos.

        Cada evento es un dict que la capa API convertirá a SSE. Formatos:
            {"token": "palabra "}
            {"done": true, "sources": [...]}
            {"error": "mensaje"}
        """
        # 1. RAG: recuperar chunks si hay documento
        chunks: list[ChunkWithScore] = []
        if document_id is not None:
            try:
                query_vec = await self._embeddings.encode(question)
                chunks = await self._repo.search_similar_chunks(
                    query_embedding=query_vec,
                    document_id=document_id,
                    top_k=self._settings.top_k_chunks,
                    min_similarity=self._settings.min_similarity,
                )
                log.info(
                    "chat.rag_retrieved",
                    doc_id=str(document_id),
                    chunks_found=len(chunks),
                )
            except Exception:
                log.exception("chat.rag_failed", doc_id=str(document_id))
                # Continuar sin RAG en vez de fallar la respuesta

        # 2. Armar prompts
        system_prompt = build_system_prompt(chunks if chunks else None)
        prompt = build_gemma_chat_prompt(system_prompt, history, question)

        # 3. Streaming tokens
        try:
            async for token in self._llm.generate(prompt):
                yield {"token": token}
        except Exception as e:
            log.exception("chat.generation_failed")
            yield {"error": f"Error al generar respuesta: {e}"}
            return

        # 4. Incluir fuentes al final
        sources = self._chunks_to_sources(chunks)
        yield {"done": True, "sources": [s.model_dump(mode="json") for s in sources]}

    def _chunks_to_sources(
        self,
        chunks: list[ChunkWithScore],
    ) -> list[Source]:
        """Convierte los chunks usados en objetos Source para el cliente."""
        sources: list[Source] = []
        for item in chunks:
            excerpt = item.chunk.chunk_text[:280]
            if len(item.chunk.chunk_text) > 280:
                excerpt += "..."
            sources.append(Source(
                chunk_id=item.chunk.id,
                page_number=item.chunk.page_number,
                excerpt=excerpt,
                similarity=item.similarity,
            ))
        return sources
