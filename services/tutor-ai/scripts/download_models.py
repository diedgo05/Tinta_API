"""
Descarga los modelos durante el build de Docker.

- Gemma 3 4B Instruct GGUF Q4_K_M (~2.5 GB) de bartowski en HuggingFace
- MiniLM multilingüe (~470 MB) via sentence-transformers cache

Se ejecuta UNA vez durante el docker build, no en cada startup.
"""
import os
import sys
import urllib.request
from pathlib import Path


GEMMA_URL = (
    "https://huggingface.co/bartowski/gemma-2-2b-it-GGUF/resolve/main/"
    "gemma-2-2b-it-Q4_K_M.gguf"
)
GEMMA_TARGET = Path("/models/gemma-2-2b-it-q4_k_m.gguf")
EMBED_MODEL = "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"


def download_llm() -> None:
    print(f"[download] Gemma 3 4B → {GEMMA_TARGET}")
    GEMMA_TARGET.parent.mkdir(parents=True, exist_ok=True)

    if GEMMA_TARGET.exists() and GEMMA_TARGET.stat().st_size > 2 * 1024 * 1024 * 1024:
        print("[download] Gemma ya existe, se omite.")
        return

    def _hook(block_num, block_size, total_size):
        downloaded = block_num * block_size
        pct = min(100, downloaded * 100 // total_size) if total_size > 0 else 0
        print(f"\r[download] Gemma: {pct}%", end="", flush=True)

    urllib.request.urlretrieve(GEMMA_URL, GEMMA_TARGET, _hook)
    print()
    print(f"[download] Gemma listo ({GEMMA_TARGET.stat().st_size // 1024 // 1024} MB)")


def download_embeddings() -> None:
    print(f"[download] Embeddings model → {EMBED_MODEL}")
    from sentence_transformers import SentenceTransformer
    # Al instanciarlo, sentence-transformers lo baja al cache de HuggingFace
    # que Docker capturará en la layer.
    SentenceTransformer(EMBED_MODEL)
    print("[download] Embeddings listo.")


if __name__ == "__main__":
    try:
        download_llm()
        download_embeddings()
    except Exception as e:
        print(f"[download] ERROR: {e}", file=sys.stderr)
        sys.exit(1)
