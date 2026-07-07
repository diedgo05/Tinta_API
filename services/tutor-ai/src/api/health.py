"""
Health check endpoint.

Railway usa este endpoint para saber si el servicio está vivo. También
sirve para debugging: si algo falla, indica cuál componente.
"""
from fastapi import APIRouter, Depends

from src.api.deps import get_embeddings, get_llm, get_repo
from src.infrastructure.embeddings.minilm_model import EmbeddingsModel
from src.infrastructure.llm.llama_runner import LlamaRunner
from src.infrastructure.vector_db.pgvector_repo import PgVectorRepo

router = APIRouter()


@router.get("/health")
async def health(
    llm: LlamaRunner = Depends(get_llm),
    embeddings: EmbeddingsModel = Depends(get_embeddings),
    repo: PgVectorRepo = Depends(get_repo),
):
    db_ok = await repo.health_check()
    llm_ok = llm.is_loaded
    emb_ok = embeddings.is_loaded

    all_ok = db_ok and llm_ok and emb_ok
    return {
        "status": "ok" if all_ok else "degraded",
        "llm_loaded": llm_ok,
        "embeddings_loaded": emb_ok,
        "db_connected": db_ok,
    }
