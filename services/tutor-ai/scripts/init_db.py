"""
Aplica todas las migraciones SQL contra la base de datos configurada
en DATABASE_URL.

Se puede correr manualmente después de deploy:
    railway run --service tutor-ai python -m scripts.init_db

O ejecutarlo automáticamente en el startup añadiéndolo al lifespan de main.py
(NO recomendado en producción; mejor manual y controlado).
"""
import os
import sys
from pathlib import Path

import psycopg


def apply_migrations(database_url: str, migrations_dir: Path) -> None:
    if not migrations_dir.exists():
        print(f"[init_db] Directorio no existe: {migrations_dir}", file=sys.stderr)
        sys.exit(1)

    sql_files = sorted(migrations_dir.glob("*.sql"))
    if not sql_files:
        print("[init_db] No hay migraciones para aplicar.")
        return

    print(f"[init_db] Aplicando {len(sql_files)} migraciones...")

    with psycopg.connect(database_url, autocommit=True) as conn:
        for sql_file in sql_files:
            print(f"[init_db] → {sql_file.name}")
            sql = sql_file.read_text(encoding="utf-8")
            with conn.cursor() as cur:
                cur.execute(sql)
            print(f"[init_db] ✓ {sql_file.name}")

    print("[init_db] Migraciones aplicadas.")


if __name__ == "__main__":
    database_url = os.environ.get("DATABASE_URL")
    if not database_url:
        print("[init_db] ERROR: DATABASE_URL no definida", file=sys.stderr)
        sys.exit(1)

    migrations_dir = Path(__file__).resolve().parent.parent / "migrations"
    apply_migrations(database_url, migrations_dir)
