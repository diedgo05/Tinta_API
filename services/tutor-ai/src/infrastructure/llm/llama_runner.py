"""
Wrapper sobre llama-cpp-python para ejecutar Gemma 3 4B GGUF.

El modelo se carga UNA vez al arrancar el servicio (lazy singleton) y se
reutiliza para todas las peticiones. La generación es streaming: cada
token se emite tan pronto como está disponible.

Como Python es single-threaded para el intérprete pero llama.cpp corre
en C bajo el hood, la inferencia NO bloquea el event loop de asyncio
cuando se usa con `run_in_executor`.
"""
from __future__ import annotations

import asyncio
from typing import AsyncIterator, Optional

import structlog
from llama_cpp import Llama

from src.config import Settings

log = structlog.get_logger()


class LlamaRunner:
    """
    Singleton que encapsula la instancia de llama_cpp.Llama.

    Uso:
        runner = LlamaRunner(settings)
        await runner.load()                # una vez al arrancar
        async for token in runner.generate(prompt):
            ...
    """

    def __init__(self, settings: Settings):
        self._settings = settings
        self._llm: Optional[Llama] = None
        # Lock para serializar generaciones. llama.cpp NO es thread-safe
        # a nivel de contexto: dos generaciones simultáneas corromperían
        # el estado interno.
        self._lock = asyncio.Lock()

    @property
    def is_loaded(self) -> bool:
        return self._llm is not None

    async def load(self) -> None:
        """
        Carga el modelo GGUF a memoria. Bloquea el proceso ~10-15 seg.
        Se llama desde el lifespan de FastAPI al arrancar.
        """
        if self._llm is not None:
            log.info("llm.already_loaded")
            return

        log.info(
            "llm.loading",
            path=self._settings.llm_model_path,
            n_ctx=self._settings.llm_context_size,
            n_threads=self._settings.llm_n_threads,
        )

        loop = asyncio.get_running_loop()

        def _load_sync() -> Llama:
            return Llama(
                model_path=self._settings.llm_model_path,
                n_ctx=self._settings.llm_context_size,
                n_threads=self._settings.llm_n_threads,
                n_gpu_layers=self._settings.llm_n_gpu_layers,
                verbose=False,
            )

        # Bloquea el thread ejecutor, no el event loop
        self._llm = await loop.run_in_executor(None, _load_sync)
        log.info("llm.loaded")

    async def generate(self, prompt: str) -> AsyncIterator[str]:
        """
        Genera tokens en streaming.

        Emite cada token tan pronto llegue. Filtra tokens especiales
        (<end_of_turn>) y corta la generación cuando aparecen.
        """
        if self._llm is None:
            raise RuntimeError("LLM no cargado. Llama load() primero.")

        # Serializar: si otra generación está en curso, esperar.
        async with self._lock:
            loop = asyncio.get_running_loop()

            # llama_cpp.Llama.__call__ con stream=True devuelve un
            # generador sincrónico. Lo consumimos en un executor para
            # no bloquear el event loop.
            def _create_stream():
                return self._llm(
                    prompt=prompt,
                    max_tokens=self._settings.llm_max_tokens,
                    temperature=self._settings.llm_temperature,
                    top_p=self._settings.llm_top_p,
                    stream=True,
                    stop=["<end_of_turn>", "<start_of_turn>"],
                )

            stream = await loop.run_in_executor(None, _create_stream)

            # Consumir el generador sync en el mismo executor.
            # Cada iteración devuelve un dict con {choices: [{text: "..."}]}
            def _next_chunk(it):
                try:
                    return next(it)
                except StopIteration:
                    return None

            it = iter(stream)
            while True:
                chunk = await loop.run_in_executor(None, _next_chunk, it)
                if chunk is None:
                    break

                token = chunk["choices"][0]["text"]
                if token:
                    yield token

    async def unload(self) -> None:
        """Libera memoria. Se llama al shutdown del servicio."""
        if self._llm is not None:
            del self._llm
            self._llm = None
            log.info("llm.unloaded")
