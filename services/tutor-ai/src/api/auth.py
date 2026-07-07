"""
Dependency de FastAPI para validar JWT emitidos por el servicio identity.

El servicio identity firma tokens con RSA privada. Este servicio valida
con la clave pública correspondiente (compartida vía env var).
"""
from __future__ import annotations

from uuid import UUID

import jwt
import structlog
from fastapi import Depends, Header, HTTPException, status

from src.config import Settings, get_settings

log = structlog.get_logger()


async def get_current_user_id(
    authorization: str | None = Header(default=None),
    settings: Settings = Depends(get_settings),
) -> UUID:
    """
    Extrae y valida el JWT del header Authorization.
    Devuelve el UUID del usuario (claim `sub`).

    Uso en endpoints:
        async def my_endpoint(user_id: UUID = Depends(get_current_user_id)):
            ...
    """
    if not authorization or not authorization.lower().startswith("bearer "):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Header Authorization Bearer requerido",
            headers={"WWW-Authenticate": "Bearer"},
        )

    token = authorization.split(" ", 1)[1].strip()

    try:
        payload = jwt.decode(
            token,
            settings.jwt_public_key,
            algorithms=[settings.jwt_algorithm],
            issuer=settings.jwt_issuer,
            options={"require": ["sub", "exp"]},
        )
    except jwt.ExpiredSignatureError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Token expirado",
        )
    except jwt.InvalidTokenError as e:
        log.warning("auth.invalid_token", error=str(e))
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Token inválido",
        )

    try:
        return UUID(payload["sub"])
    except (KeyError, ValueError):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Claim 'sub' inválido",
        )
