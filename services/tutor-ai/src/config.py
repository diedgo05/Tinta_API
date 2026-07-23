"""
Configuración central del servicio tutor-ai.

Todas las variables de entorno se declaran aquí como campos de Pydantic
Settings. Esto da validación de tipos, defaults sensatos y errores tempranos
si falta configuración crítica.

En Railway, estas variables se definen en el panel Settings → Variables.
En desarrollo local, se toman del archivo .env
"""
from functools import lru_cache
from pathlib import Path

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    # ── Servicio ──────────────────────────────────────────────
    service_name: str = "tutor-ai"
    environment: str = "development"      # development | production
    log_level: str = "INFO"

    # ── Base de datos ─────────────────────────────────────────
    # Formato esperado:
    #   postgresql://user:pass@host:5432/dbname?search_path=knowledge,public
    database_url: str

    # ── LLM (Gemma 3 4B GGUF en servidor) ─────────────────────
    # Path absoluto al archivo .gguf dentro del container.
    # El Dockerfile lo descarga durante el build en /models/
    llm_model_path: str = "/models/gemma-2-2b-it-q4_k_m.gguf"

    llm_context_size: int = 4096          # tokens de contexto
    llm_max_tokens: int = 512             # tokens máximos por respuesta
    llm_temperature: float = 0.7
    llm_top_p: float = 0.9
    llm_n_threads: int = 4                # cores CPU dedicados al LLM
    llm_n_gpu_layers: int = 0             # 0 = CPU only en Railway

    # ── Embeddings (MiniLM multilingüe) ───────────────────────
    embedding_model_name: str = "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
    embedding_dimension: int = 384

    # ── RAG ───────────────────────────────────────────────────
    chunk_size: int = 500                 # palabras por chunk
    chunk_overlap: int = 50               # palabras de solapamiento
    top_k_chunks: int = 3                 # cuántos chunks recuperar por query
    min_similarity: float = 0.38           # umbral de similitud coseno

    # ── Autenticación (JWT compartido con identity service) ───
    # Se comparte la clave pública RSA del identity para validar tokens.
    # En Railway: pegar el contenido PEM en la variable JWT_PUBLIC_KEY.
    jwt_public_key: str
    jwt_algorithm: str = "RS256"
    jwt_issuer: str = "tinta-identity"

    # ── Uploads ───────────────────────────────────────────────
    upload_dir: str = "/tmp/tinta-uploads"
    max_upload_size_mb: int = 20          # límite por archivo


@lru_cache
def get_settings() -> Settings:
    """
    Cached factory. `lru_cache` garantiza una sola instancia por proceso.
    Se usa como dependencia de FastAPI: `settings: Settings = Depends(get_settings)`.
    """
    return Settings()
