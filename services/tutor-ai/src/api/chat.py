"""
Endpoint de chat con streaming Server-Sent Events (SSE).

Cliente envía POST con la pregunta e (opcionalmente) el document_id.
Respuesta es un stream text/event-stream con tokens del LLM.

Formato de eventos (cada uno como línea `data: {...}\n\n`):
    {"token": "palabra "}      → un token generado
    {"done": true, "sources": [...]}  → fin, con fuentes RAG
    {"error": "mensaje"}       → algo falló
"""
from __future__ import annotations

import json
from uuid import UUID

import structlog
from fastapi import APIRouter, Depends, HTTPException, status
from sse_starlette.sse import EventSourceResponse

from src.api.auth import get_current_user_id
from src.api.deps import get_chat_service, get_repo
from src.application.chat_service import ChatService
from src.domain.entities import ChatRequest
from src.infrastructure.vector_db.pgvector_repo import PgVectorRepo

log = structlog.get_logger()

router = APIRouter(tags=["chat"])


@router.post("/chat")
async def chat(
    payload: ChatRequest,
    user_id: UUID = Depends(get_current_user_id),
    chat_service: ChatService = Depends(get_chat_service),
    repo: PgVectorRepo = Depends(get_repo),
):
    """
    Envía una pregunta al tutor y devuelve la respuesta como stream SSE.
    """
    # Validar acceso al documento si se especificó
    if payload.document_id is not None:
        doc = await repo.get_document(payload.document_id)
        if doc is None:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Documento no encontrado",
            )
        if doc.user_id is not None and doc.user_id != user_id:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="No tienes acceso a este documento",
            )
        if doc.status.value != "ready":
            raise HTTPException(
                status_code=status.HTTP_409_CONFLICT,
                detail=f"El documento aún no está listo (estado: {doc.status.value})",
            )

    log.info(
        "chat.request",
        user_id=str(user_id),
        doc_id=str(payload.document_id) if payload.document_id else None,
        history_len=len(payload.history),
    )

    async def event_generator():
        """
        Genera los eventos SSE. sse-starlette los formatea con
        `data: {...}\n\n` automáticamente.
        """
        async for event in chat_service.stream_response(
            question=payload.question,
            history=payload.history,
            document_id=payload.document_id,
        ):
            yield {"data": json.dumps(event)}

    return EventSourceResponse(event_generator())
