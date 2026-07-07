"""
Endpoints de documentos:
    POST /documents          → subir un PDF (procesamiento en background)
    GET  /documents/{id}     → consultar estado del PDF
"""
from __future__ import annotations

from pathlib import Path
from uuid import UUID, uuid4

import structlog
from fastapi import (
    APIRouter,
    BackgroundTasks,
    Depends,
    File,
    HTTPException,
    UploadFile,
    status,
)

from src.api.auth import get_current_user_id
from src.api.deps import get_document_service, get_repo
from src.application.document_service import DocumentService
from src.config import Settings, get_settings
from src.domain.entities import Document
from src.infrastructure.vector_db.pgvector_repo import PgVectorRepo

log = structlog.get_logger()

router = APIRouter(prefix="/documents", tags=["documents"])


@router.post("", status_code=status.HTTP_202_ACCEPTED, response_model=Document)
async def upload_document(
    background_tasks: BackgroundTasks,
    file: UploadFile = File(...),
    user_id: UUID = Depends(get_current_user_id),
    settings: Settings = Depends(get_settings),
    repo: PgVectorRepo = Depends(get_repo),
    doc_service: DocumentService = Depends(get_document_service),
) -> Document:
    """
    Sube un PDF y encola su procesamiento.

    Devuelve 202 con el registro del documento en estado 'pending'.
    El cliente debe hacer polling a GET /documents/{id} para saber
    cuándo cambia a 'ready'.
    """
    if not file.filename or not file.filename.lower().endswith(".pdf"):
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Solo se aceptan archivos .pdf",
        )

    # Validar tamaño
    contents = await file.read()
    max_bytes = settings.max_upload_size_mb * 1024 * 1024
    if len(contents) > max_bytes:
        raise HTTPException(
            status_code=status.HTTP_413_REQUEST_ENTITY_TOO_LARGE,
            detail=f"El archivo excede {settings.max_upload_size_mb} MB",
        )

    # Guardar archivo temporal
    upload_dir = Path(settings.upload_dir)
    upload_dir.mkdir(parents=True, exist_ok=True)
    tmp_path = upload_dir / f"{uuid4()}.pdf"
    tmp_path.write_bytes(contents)

    # Crear registro en DB
    doc = await repo.create_document(user_id=user_id, filename=file.filename)

    # Encolar procesamiento en background
    background_tasks.add_task(doc_service.process_document, doc.id, tmp_path)

    log.info("doc.upload_accepted", doc_id=str(doc.id), user_id=str(user_id))
    return doc


@router.get("/{document_id}", response_model=Document)
async def get_document(
    document_id: UUID,
    user_id: UUID = Depends(get_current_user_id),
    repo: PgVectorRepo = Depends(get_repo),
) -> Document:
    """Consulta el estado y metadata de un documento propio."""
    doc = await repo.get_document(document_id)
    if doc is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Documento no encontrado",
        )

    # Autorización: solo el dueño puede consultar (o si es contenido curado)
    if doc.user_id is not None and doc.user_id != user_id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="No tienes acceso a este documento",
        )

    return doc
