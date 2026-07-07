"""
Wrapper sobre sentence-transformers para generar embeddings.

Usamos MiniLM multilingüe de 384 dimensiones. Es un buen balance:
    - 471 MB de peso (aceptable en el container)
    - Soporte decente para español
    - Rápido en CPU (~50-100 embeddings/seg)
    - Vectores de 384-dim (más chicos = búsqueda más rápida en pgvector)
"""
from __future__ import annotations

import asyncio
from typing import Optional

import structlog
from sentence_transformers import SentenceTransformer

from src.config import Settings

log = structlog.get_logger()


class EmbeddingsModel:
    """
    Singleton del modelo de embeddings.

    Uso:
        model = EmbeddingsModel(settings)
        await model.load()
        vec = await model.encode("¿Qué es la recursión?")
        vecs = await model.encode_batch(["texto 1", "texto 2", ...])
    """

    def __init__(self, settings: Settings):
        self._settings = settings
        self._model: Optional[SentenceTransformer] = None

    @property
    def is_loaded(self) -> bool:
        return self._model is not None

    async def load(self) -> None:
        """
        Descarga (si no está en cache) y carga el modelo.
        En Railway, si el Dockerfile pre-descargó el modelo, es instantáneo.
        """
        if self._model is not None:
            return

        log.info("embeddings.loading", model=self._settings.embedding_model_name)

        loop = asyncio.get_running_loop()

        def _load_sync() -> SentenceTransformer:
            return SentenceTransformer(self._settings.embedding_model_name)

        self._model = await loop.run_in_executor(None, _load_sync)
        log.info("embeddings.loaded")

    async def encode(self, text: str) -> list[float]:
        """Genera un embedding para un solo texto."""
        if self._model is None:
            raise RuntimeError("Modelo de embeddings no cargado.")

        loop = asyncio.get_running_loop()

        def _encode_sync() -> list[float]:
            vec = self._model.encode(
                text,
                convert_to_numpy=True,
                normalize_embeddings=True,   # normaliza para similitud coseno
            )
            return vec.tolist()

        return await loop.run_in_executor(None, _encode_sync)

    async def encode_batch(self, texts: list[str]) -> list[list[float]]:
        """
        Genera embeddings de un batch. Mucho más eficiente que llamar
        `encode()` en loop porque sentence-transformers batchea internamente.
        """
        if self._model is None:
            raise RuntimeError("Modelo de embeddings no cargado.")

        loop = asyncio.get_running_loop()

        def _encode_sync() -> list[list[float]]:
            vecs = self._model.encode(
                texts,
                convert_to_numpy=True,
                normalize_embeddings=True,
                batch_size=32,
                show_progress_bar=False,
            )
            return vecs.tolist()

        return await loop.run_in_executor(None, _encode_sync)
