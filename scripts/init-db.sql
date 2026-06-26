-- ============================================================================
-- Tinta · Inicialización de la base de datos
-- Crea los esquemas separados por microservicio (SOA: aislamiento de datos)
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE SCHEMA IF NOT EXISTS identity;
CREATE SCHEMA IF NOT EXISTS community;
CREATE SCHEMA IF NOT EXISTS recommendations;
CREATE SCHEMA IF NOT EXISTS catalog;
CREATE SCHEMA IF NOT EXISTS reading;
CREATE SCHEMA IF NOT EXISTS knowledge;
CREATE SCHEMA IF NOT EXISTS notifications;

GRANT ALL PRIVILEGES ON SCHEMA identity        TO tinta;
GRANT ALL PRIVILEGES ON SCHEMA community       TO tinta;
GRANT ALL PRIVILEGES ON SCHEMA recommendations TO tinta;
GRANT ALL PRIVILEGES ON SCHEMA catalog         TO tinta;
GRANT ALL PRIVILEGES ON SCHEMA reading         TO tinta;
GRANT ALL PRIVILEGES ON SCHEMA knowledge       TO tinta;
GRANT ALL PRIVILEGES ON SCHEMA notifications   TO tinta;
