"""
Repositorio para operaciones sobre el esquema `knowledge`.

Usa psycopg 3 con connection pooling y pgvector para operaciones vectoriales.
Todas las operaciones son async.
"""
from __future__ import annotations

from typing import Optional
from uuid import UUID

import psycopg
import structlog
from pgvector.psycopg import register_vector_async
from psycopg_pool import AsyncConnectionPool

from src.config import Settings
from src.domain.entities import (
    Chunk,
    ChunkWithScore,
    Document,
    DocumentStatus,
)

log = structlog.get_logger()


class PgVectorRepo:
    """
    Repositorio único para documents, chunks y embeddings.

    Uso:
        repo = PgVectorRepo(settings)
        await repo.connect()
        doc = await repo.create_document(user_id, filename)
        ...
        await repo.close()
    """

    def __init__(self, settings: Settings):
        self._settings = settings
        self._pool: Optional[AsyncConnectionPool] = None

    async def connect(self) -> None:
        """Crea el pool de conexiones y registra pgvector."""
        if self._pool is not None:
            return

        log.info("pgvector.connecting")

        # Callback para registrar el adaptador de vector en cada conexión nueva
        async def _configure(conn: psycopg.AsyncConnection) -> None:
            await register_vector_async(conn)

        self._pool = AsyncConnectionPool(
            conninfo=self._settings.database_url,
            min_size=1,
            max_size=5,
            configure=_configure,
            open=False,
        )
        await self._pool.open()
        log.info("pgvector.connected")

    async def close(self) -> None:
        if self._pool is not None:
            await self._pool.close()
            self._pool = None

    async def health_check(self) -> bool:
        """Ping simple para el endpoint /health."""
        if self._pool is None:
            return False
        try:
            async with self._pool.connection() as conn:
                async with conn.cursor() as cur:
                    await cur.execute("SELECT 1")
                    await cur.fetchone()
            return True
        except Exception as e:
            log.error("pgvector.health_failed", error=str(e))
            return False

    # ══════════════════════════════════════════════════════════════════
    # DOCUMENTS
    # ══════════════════════════════════════════════════════════════════

    async def create_document(
        self,
        user_id: Optional[UUID],
        filename: str,
    ) -> Document:
        assert self._pool is not None
        async with self._pool.connection() as conn:
            async with conn.cursor() as cur:
                await cur.execute(
                    """
                    INSERT INTO knowledge.documents (user_id, filename, status)
                    VALUES (%s, %s, 'pending')
                    RETURNING id, user_id, filename, total_pages, status,
                              error_message, uploaded_at, processed_at
                    """,
                    (user_id, filename),
                )
                row = await cur.fetchone()
        return self._row_to_document(row)

    async def get_document(self, doc_id: UUID) -> Optional[Document]:
        assert self._pool is not None
        async with self._pool.connection() as conn:
            async with conn.cursor() as cur:
                await cur.execute(
                    """
                    SELECT d.id, d.user_id, d.filename, d.total_pages, d.status,
                           d.error_message, d.uploaded_at, d.processed_at,
                           COALESCE((SELECT COUNT(*) FROM knowledge.chunks
                                     WHERE document_id = d.id), 0) as chunks_count
                    FROM knowledge.documents d
                    WHERE d.id = %s
                    """,
                    (doc_id,),
                )
                row = await cur.fetchone()
        return self._row_to_document(row) if row else None

    async def update_document_status(
        self,
        doc_id: UUID,
        status: DocumentStatus,
        total_pages: Optional[int] = None,
        error_message: Optional[str] = None,
    ) -> None:
        assert self._pool is not None
        async with self._pool.connection() as conn:
            async with conn.cursor() as cur:
                await cur.execute(
                    """
                    UPDATE knowledge.documents
                    SET status = %s,
                        total_pages = COALESCE(%s, total_pages),
                        error_message = %s,
                        processed_at = CASE
                            WHEN %s IN ('ready', 'failed') THEN NOW()
                            ELSE processed_at
                        END
                    WHERE id = %s
                    """,
                    (status.value, total_pages, error_message, status.value, doc_id),
                )

    # ══════════════════════════════════════════════════════════════════
    # CHUNKS + EMBEDDINGS
    # ══════════════════════════════════════════════════════════════════

    async def insert_chunks_with_embeddings(
        self,
        document_id: UUID,
        chunks_data: list[tuple[str, int, Optional[int], list[float]]],
    ) -> int:
        """
        Inserta chunks y sus embeddings en la misma transacción.

        Cada tupla es: (chunk_text, chunk_position, page_number, embedding)
        Devuelve el número de chunks insertados.
        """
        assert self._pool is not None
        if not chunks_data:
            return 0

        async with self._pool.connection() as conn:
            async with conn.cursor() as cur:
                # Insertar chunks y capturar sus IDs
                # Usamos executemany para eficiencia
                chunk_ids: list[UUID] = []
                for chunk_text, position, page_number, _ in chunks_data:
                    await cur.execute(
                        """
                        INSERT INTO knowledge.chunks
                            (document_id, chunk_text, chunk_position, page_number)
                        VALUES (%s, %s, %s, %s)
                        RETURNING id
                        """,
                        (document_id, chunk_text, position, page_number),
                    )
                    row = await cur.fetchone()
                    chunk_ids.append(row[0])

                # Insertar embeddings
                for chunk_id, (_, _, _, embedding) in zip(chunk_ids, chunks_data):
                    await cur.execute(
                        """
                        INSERT INTO knowledge.chunk_embeddings (chunk_id, embedding)
                        VALUES (%s, %s)
                        """,
                        (chunk_id, embedding),
                    )

            await conn.commit()

        return len(chunk_ids)

    async def search_similar_chunks(
        self,
        query_embedding: list[float],
        document_id: UUID,
        top_k: int,
        min_similarity: float,
    ) -> list[ChunkWithScore]:
        """
        Búsqueda semántica: recupera los top_k chunks más similares a
        `query_embedding` dentro del documento indicado.

        Usa distancia coseno (`<=>`) que es 1 - cosine_similarity.
        Convertimos: similarity = 1 - distance.
        """
        assert self._pool is not None
        async with self._pool.connection() as conn:
            async with conn.cursor() as cur:
                await cur.execute(
                    """
                    SELECT
                        c.id, c.document_id, c.chunk_text,
                        c.chunk_position, c.page_number,
                        1 - (ce.embedding <=> %s::vector) AS similarity
                    FROM knowledge.chunk_embeddings ce
                    JOIN knowledge.chunks c ON c.id = ce.chunk_id
                    WHERE c.document_id = %s
                      AND 1 - (ce.embedding <=> %s::vector) >= %s
                    ORDER BY ce.embedding <=> %s::vector ASC
                    LIMIT %s
                    """,
                    (
                        query_embedding, document_id, query_embedding,
                        min_similarity, query_embedding, top_k,
                    ),
                )
                rows = await cur.fetchall()

        results: list[ChunkWithScore] = []
        for row in rows:
            chunk = Chunk(
                id=row[0],
                document_id=row[1],
                chunk_text=row[2],
                chunk_position=row[3],
                page_number=row[4],
            )
            results.append(ChunkWithScore(chunk=chunk, similarity=float(row[5])))
        return results

    # ══════════════════════════════════════════════════════════════════
    # HELPERS
    # ══════════════════════════════════════════════════════════════════

    def _row_to_document(self, row) -> Document:
        chunks_count = row[8] if len(row) > 8 else 0
        return Document(
            id=row[0],
            user_id=row[1],
            filename=row[2],
            total_pages=row[3],
            status=DocumentStatus(row[4]),
            error_message=row[5],
            uploaded_at=row[6],
            processed_at=row[7],
            chunks_count=chunks_count,
        )
