"""
Servicio de aplicación para el pipeline de ingesta de documentos.

Orquesta:
    1. Guardar el archivo temporalmente
    2. Extraer texto y chunkear
    3. Generar embeddings
    4. Persistir en Postgres
    5. Actualizar el estado del documento

Este servicio se ejecuta en background para no bloquear el request HTTP.
"""
from __future__ import annotations

from pathlib import Path
from uuid import UUID

import structlog

from src.domain.entities import DocumentStatus
from src.infrastructure.embeddings.minilm_model import EmbeddingsModel
from src.infrastructure.pdf.extractor import PdfExtractor
from src.infrastructure.vector_db.pgvector_repo import PgVectorRepo

log = structlog.get_logger()


class DocumentService:
    """
    Coordina la ingesta de un PDF de punta a punta.
    """

    def __init__(
        self,
        pdf_extractor: PdfExtractor,
        embeddings: EmbeddingsModel,
        repo: PgVectorRepo,
    ):
        self._extractor = pdf_extractor
        self._embeddings = embeddings
        self._repo = repo

    async def process_document(
        self,
        document_id: UUID,
        pdf_path: Path,
    ) -> None:
        """
        Ejecuta el pipeline completo.

        Este método SIEMPRE es llamado en background (BackgroundTasks).
        No lanza excepciones al caller: si algo falla, marca el documento
        como 'failed' con el mensaje de error.
        """
        try:
            log.info("doc.processing_started", doc_id=str(document_id))
            await self._repo.update_document_status(
                document_id, DocumentStatus.PROCESSING,
            )

            # 1. Extraer y chunkear
            total_pages, extracted_chunks = self._extractor.extract_and_chunk(pdf_path)

            if not extracted_chunks:
                await self._repo.update_document_status(
                    document_id,
                    DocumentStatus.FAILED,
                    total_pages=total_pages,
                    error_message="No se pudo extraer texto del PDF.",
                )
                return

            # 2. Generar embeddings en batch (mucho más rápido que uno a uno)
            texts = [c.text for c in extracted_chunks]
            embeddings = await self._embeddings.encode_batch(texts)

            # 3. Preparar tuplas y persistir
            chunks_data = [
                (
                    extracted.text,
                    extracted.position,
                    extracted.page_number,
                    embedding,
                )
                for extracted, embedding in zip(extracted_chunks, embeddings)
            ]

            inserted = await self._repo.insert_chunks_with_embeddings(
                document_id, chunks_data,
            )

            # 4. Marcar como listo
            await self._repo.update_document_status(
                document_id,
                DocumentStatus.READY,
                total_pages=total_pages,
            )

            log.info(
                "doc.processing_done",
                doc_id=str(document_id),
                pages=total_pages,
                chunks=inserted,
            )

        except Exception as e:
            log.exception("doc.processing_failed", doc_id=str(document_id))
            await self._repo.update_document_status(
                document_id,
                DocumentStatus.FAILED,
                error_message=str(e)[:500],
            )
        finally:
            # Limpiar archivo temporal
            try:
                if pdf_path.exists():
                    pdf_path.unlink()
            except Exception:
                pass
