from __future__ import annotations

import re
from typing import AsyncIterator, Optional
from uuid import UUID

import structlog

from src.config import Settings
from src.domain.entities import ChatMessage, ChunkWithScore, Source
from src.domain.prompts import build_gemma_chat_prompt, build_system_prompt, OUT_OF_SCOPE_MESSAGE
from src.infrastructure.embeddings.minilm_model import EmbeddingsModel
from src.infrastructure.llm.llama_runner import LlamaRunner
from src.infrastructure.vector_db.pgvector_repo import PgVectorRepo

log = structlog.get_logger()


# ── Detección de preguntas "amplias" sobre todo el documento ──────
_BROAD_QUESTION_PATTERNS = [
    r"resum(e|en|ir)",
    r"ideas?\s+principal",
    r"tema\s+central",
    r"de\s+qu[ée]\s+trata",
    r"sobre\s+qu[ée]\s+trata",
    r"ejemplos?\s+del?\s+concepto",
    r"conceptos?\s+clave",
    r"qu[ée]\s+dice\s+el\s+documento",
    r"contenido\s+del\s+documento",
    r"explica(me)?\s+el\s+documento",
    r"de\s+qu[ée]\s+se\s+trata",
]
_broad_question_regex = re.compile(
    "|".join(_BROAD_QUESTION_PATTERNS), re.IGNORECASE
)


def _is_broad_question(question: str) -> bool:
    """Heurística simple, no un clasificador. Cubre los casos comunes
    de preguntas meta sobre el documento (resumen, temas, ejemplos).
    No es perfecta: una pregunta amplia pero genuinamente ajena al
    documento ("resume las noticias de hoy") pasaría el filtro igual,
    pero es un caso raro y el costo de ese falso positivo es bajo."""
    return bool(_broad_question_regex.search(question))


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
        ...
        """
        chunks: list[ChunkWithScore] = []
        document_requested = document_id is not None
        is_broad = _is_broad_question(question)

        if document_requested:
            try:
                query_vec = await self._embeddings.encode(question)

                # Preguntas amplias ("resume", "tema central", etc.)
                effective_min_similarity = (
                    0.0 if is_broad else self._settings.min_similarity
                )
                effective_top_k = (
                    self._settings.top_k_chunks * 2
                    if is_broad
                    else self._settings.top_k_chunks
                )

                chunks = await self._repo.search_similar_chunks(
                    query_embedding=query_vec,
                    document_id=document_id,
                    top_k=effective_top_k,
                    min_similarity=effective_min_similarity,
                )
                log.info(
                    "chat.rag_retrieved",
                    doc_id=str(document_id),
                    chunks_found=len(chunks),
                    is_broad_question=is_broad,
                )
            except Exception:
                log.exception("chat.rag_failed", doc_id=str(document_id))

        # Gate de relevancia: solo aplica a preguntas puntuales.
        if document_requested and not chunks:
            log.info(
                "chat.rag_out_of_scope",
                doc_id=str(document_id),
                question=question,
            )
            yield {"token": OUT_OF_SCOPE_MESSAGE}
            yield {"done": True, "sources": []}
            return

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