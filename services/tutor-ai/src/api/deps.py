"""
Providers de dependencias para inyectar los servicios en los endpoints.

Estos objetos son singletons vivos en `app.state`, inicializados en el
lifespan de FastAPI (main.py).
"""
from __future__ import annotations

from fastapi import Depends, Request

from src.application.chat_service import ChatService
from src.application.document_service import DocumentService
from src.infrastructure.embeddings.minilm_model import EmbeddingsModel
from src.infrastructure.llm.llama_runner import LlamaRunner
from src.infrastructure.vector_db.pgvector_repo import PgVectorRepo


def get_llm(request: Request) -> LlamaRunner:
    return request.app.state.llm


def get_embeddings(request: Request) -> EmbeddingsModel:
    return request.app.state.embeddings


def get_repo(request: Request) -> PgVectorRepo:
    return request.app.state.repo


def get_chat_service(request: Request) -> ChatService:
    return request.app.state.chat_service


def get_document_service(request: Request) -> DocumentService:
    return request.app.state.document_service
