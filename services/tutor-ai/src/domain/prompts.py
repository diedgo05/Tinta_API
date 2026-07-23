"""
Prompts y templates para el LLM.

Aislar los prompts en un solo archivo permite iterar sobre ellos sin tocar
la lógica de negocio. Cambiar el tono del tutor o añadir instrucciones es
una edición aquí, no una migración.
"""
from src.domain.entities import ChatMessage, ChunkWithScore

# ── System prompt base ────────────────────────────────────────────
BASE_SYSTEM_PROMPT = """Eres Tinta AI, un tutor académico que SIEMPRE responde en español neutro y claro.
Ayudas a estudiantes universitarios mexicanos a entender los temas de sus lecturas.

Reglas:
- Sé conciso pero pedagógico. Usa ejemplos cuando ayuden.
- Si no sabes algo, dilo honestamente. NO inventes datos.
- Estructura tus respuestas con párrafos cortos.
- No uses inglés salvo para términos técnicos que no tienen traducción.
- No saludes en cada mensaje; responde directo a la pregunta."""


# ── Instrucción cuando hay contexto RAG disponible ────────────────
RAG_INSTRUCTION_TEMPLATE = """

Tienes acceso a los siguientes fragmentos del documento que el estudiante está leyendo.
Estos fragmentos son tu ÚNICA fuente de información para esta respuesta. Si citas
información específica, indica la página entre paréntesis: (p. 15).

═══════════════════════════════════════════════════════════════
FRAGMENTOS DEL DOCUMENTO:
{context}
═══════════════════════════════════════════════════════════════

IMPORTANTE: Responde ÚNICAMENTE con base en estos fragmentos. Si la pregunta del
estudiante no puede responderse con esta información, dilo explícitamente:
"No encuentro esa información en el documento." NO uses tu conocimiento general
para completar la respuesta, aunque lo conozcas."""


# ── Mensaje cuando la pregunta no tiene relación con el documento ──
# Se usa ANTES de llamar al LLM, cuando ningún chunk recuperado superó
# el umbral de similitud mínima. Evita que el modelo alucine una
# respuesta sobre un tema completamente ajeno al PDF.
OUT_OF_SCOPE_MESSAGE = (
    "Esa pregunta no parece estar relacionada con el documento que estás "
    "leyendo. Intenta preguntarme algo sobre su contenido, o si quieres "
    "ayuda con otro tema, dímelo explícitamente."
)


def build_system_prompt(chunks: list[ChunkWithScore] | None) -> str:
    """
    Construye el system prompt final.

    Si hay chunks recuperados por RAG, los inyecta al prompt.
    Si no, devuelve solo el prompt base.
    """
    if not chunks:
        return BASE_SYSTEM_PROMPT

    context_parts = []
    for i, item in enumerate(chunks, 1):
        page = item.chunk.page_number
        page_str = f" (p. {page})" if page else ""
        context_parts.append(f"[Fragmento {i}{page_str}]\n{item.chunk.chunk_text}")

    context = "\n\n".join(context_parts)
    rag_addition = RAG_INSTRUCTION_TEMPLATE.format(context=context)
    return BASE_SYSTEM_PROMPT + rag_addition

def build_gemma_chat_prompt(
    system_prompt: str,
    history: list[ChatMessage],
    user_question: str,
) -> str:
    """
    Construye el prompt en formato chat de Gemma 3.

    Gemma usa tokens especiales:
        <start_of_turn>user
        {contenido}<end_of_turn>
        <start_of_turn>model

    Gemma NO tiene rol "system" nativo; el system prompt va como prefijo
    del primer mensaje del usuario.
    """
    buf: list[str] = []
    system_injected = False

    # Historial previo
    for msg in history:
        if msg.role.value == "user":
            buf.append("<start_of_turn>user\n")
            if not system_injected:
                buf.append(f"{system_prompt}\n\n")
                system_injected = True
            buf.append(f"{msg.content}<end_of_turn>\n")
        elif msg.role.value == "assistant":
            buf.append(f"<start_of_turn>model\n{msg.content}<end_of_turn>\n")

    # Pregunta actual
    buf.append("<start_of_turn>user\n")
    if not system_injected:
        buf.append(f"{system_prompt}\n\n")
    buf.append(f"{user_question}<end_of_turn>\n")

    # Abrir turno del modelo para que empiece a generar
    buf.append("<start_of_turn>model\n")

    return "".join(buf)
