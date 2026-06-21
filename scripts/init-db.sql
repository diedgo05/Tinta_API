-- ============================================================================
-- Tinta · Inicialización de la base de datos
-- Crea los esquemas separados por microservicio (SOA: aislamiento de datos)
-- ============================================================================

-- Extensiones
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Esquemas por microservicio
CREATE SCHEMA IF NOT EXISTS identity;
CREATE SCHEMA IF NOT EXISTS community;
CREATE SCHEMA IF NOT EXISTS recommendations;

-- Nota: en producción cada servicio tendría su propio usuario de BD con
-- permisos restringidos a su esquema. Para desarrollo usamos un solo
-- usuario "tinta" con acceso a todos los esquemas.

GRANT ALL PRIVILEGES ON SCHEMA identity        TO tinta;
GRANT ALL PRIVILEGES ON SCHEMA community       TO tinta;
GRANT ALL PRIVILEGES ON SCHEMA recommendations TO tinta;
