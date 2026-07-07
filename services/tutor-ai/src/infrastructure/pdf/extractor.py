"""
Extracción de texto de PDFs y chunking.

Usa pypdf para extraer texto página por página. Después parte el texto
en chunks de N palabras con overlap, conservando el número de página
para poder citar.
"""
from __future__ import annotations

import re
from dataclasses import dataclass
from pathlib import Path
from typing import Optional

import structlog
from pypdf import PdfReader

log = structlog.get_logger()


@dataclass
class ExtractedChunk:
    """Chunk generado por el pipeline: texto + posición + página estimada."""
    text: str
    position: int
    page_number: Optional[int]


class PdfExtractor:
    """
    Extractor de texto con chunking basado en palabras.

    La estrategia:
        1. Extraer texto de cada página con pypdf, guardando el nro de página
        2. Normalizar espacios y saltos de línea
        3. Partir el texto acumulado en chunks de `chunk_size` palabras
           con `chunk_overlap` palabras de solapamiento
        4. Cada chunk conoce la página donde INICIA (aproximado)
    """

    def __init__(self, chunk_size: int = 500, chunk_overlap: int = 50):
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap

    def extract_and_chunk(self, pdf_path: Path) -> tuple[int, list[ExtractedChunk]]:
        """
        Devuelve (total_pages, chunks).

        Levanta Exception si el PDF no se puede leer.
        """
        log.info("pdf.extracting", path=str(pdf_path))

        reader = PdfReader(str(pdf_path))
        total_pages = len(reader.pages)

        # Lista de tuplas: (page_num, palabra)
        # Guardamos la palabra con su página de origen para chunking preciso.
        words_with_pages: list[tuple[int, str]] = []

        for page_num, page in enumerate(reader.pages, start=1):
            try:
                text = page.extract_text() or ""
            except Exception as e:
                log.warning("pdf.page_extract_failed", page=page_num, error=str(e))
                continue

            text = self._normalize(text)
            for word in text.split():
                words_with_pages.append((page_num, word))

        if not words_with_pages:
            log.warning("pdf.no_text_extracted", path=str(pdf_path))
            return total_pages, []

        # Chunking con overlap
        chunks = self._chunk_words(words_with_pages)

        log.info(
            "pdf.extracted",
            pages=total_pages,
            words=len(words_with_pages),
            chunks=len(chunks),
        )
        return total_pages, chunks

    def _normalize(self, text: str) -> str:
        """Colapsa espacios múltiples, quita caracteres de control."""
        # Normalizar saltos de línea y espacios
        text = re.sub(r"\s+", " ", text)
        # Quitar caracteres de control (excepto los normales)
        text = "".join(ch for ch in text if ch.isprintable() or ch.isspace())
        return text.strip()

    def _chunk_words(
        self,
        words_with_pages: list[tuple[int, str]],
    ) -> list[ExtractedChunk]:
        """
        Ventana deslizante sobre palabras.
        Cada chunk toma `chunk_size` palabras y avanza `chunk_size - overlap`.
        """
        chunks: list[ExtractedChunk] = []
        step = self.chunk_size - self.chunk_overlap
        if step <= 0:
            raise ValueError("chunk_size debe ser mayor que chunk_overlap")

        total = len(words_with_pages)
        position = 0

        for start in range(0, total, step):
            end = min(start + self.chunk_size, total)
            window = words_with_pages[start:end]
            if not window:
                break

            # Página del chunk: la de su primera palabra
            page_number = window[0][0]
            text = " ".join(word for _, word in window)

            chunks.append(ExtractedChunk(
                text=text,
                position=position,
                page_number=page_number,
            ))
            position += 1

            if end == total:
                break

        return chunks
