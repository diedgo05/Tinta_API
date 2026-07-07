"""
Entidades del dominio del tutor IA.

Estas clases representan los conceptos centrales sin acoplamiento a
frameworks, bases de datos o librerías externas. Son la lengua franca
entre capas.
"""
from datetime import datetime
from enum import Enum
from typing import Optional
from uuid import UUID

from pydantic import BaseModel, Field


# ══════════════════════════════════════════════════════════════════════════
# DOCUMENTOS
# ══════════════════════════════════════════════════════════════════════════

class DocumentStatus(str, Enum):
    """Estados posibles del pipeline de ingesta de un documento."""
    PENDING = "pending"
    PROCESSING = "processing"
    READY = "ready"
    FAILED = "failed"


class Document(BaseModel):
    """
    Un documento PDF registrado en el sistema.

    Si `user_id` es None, es contenido curado por el equipo (base académica).
    Si `user_id` está presente, es un PDF subido por ese usuario.
    """
    id: UUID
    user_id: Optional[UUID] = None
    filename: str
    total_pages: Optional[int] = None
    status: DocumentStatus
    error_message: Optional[str] = None
    uploaded_at: datetime
    processed_at: Optional[datetime] = None
    chunks_count: int = 0


# ══════════════════════════════════════════════════════════════════════════
# CHUNKS
# ══════════════════════════════════════════════════════════════════════════

class Chunk(BaseModel):
    """Fragmento de texto de ~500 palabras extraído de un documento."""
    id: UUID
    document_id: UUID
    chunk_text: str
    chunk_position: int
    page_number: Optional[int] = None


class ChunkWithScore(BaseModel):
    """Chunk recuperado por búsqueda semántica, con su score de similitud."""
    chunk: Chunk
    similarity: float = Field(..., ge=0.0, le=1.0)


# ══════════════════════════════════════════════════════════════════════════
# CHAT
# ══════════════════════════════════════════════════════════════════════════

class ChatRole(str, Enum):
    USER = "user"
    ASSISTANT = "assistant"
    SYSTEM = "system"


class ChatMessage(BaseModel):
    """Un mensaje del historial de chat."""
    role: ChatRole
    content: str


class ChatRequest(BaseModel):
    """Payload del endpoint POST /chat."""
    question: str = Field(..., min_length=1, max_length=2000)
    document_id: Optional[UUID] = None
    history: list[ChatMessage] = Field(default_factory=list)


class Source(BaseModel):
    """
    Referencia a un chunk usado por el LLM para responder.
    Se devuelve al cliente al final del stream para que muestre las fuentes.
    """
    chunk_id: UUID
    page_number: Optional[int]
    excerpt: str = Field(..., max_length=300)
    similarity: float
