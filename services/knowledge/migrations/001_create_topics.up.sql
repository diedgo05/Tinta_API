CREATE TABLE IF NOT EXISTS topics (
    id          UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
    name        VARCHAR(120) NOT NULL UNIQUE,
    slug        VARCHAR(120) NOT NULL UNIQUE,
    description TEXT,
    icon        VARCHAR(120),
    size_mb     INT NOT NULL DEFAULT 0,
    version     VARCHAR(20) NOT NULL DEFAULT 'v1',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_topics_slug ON topics(slug);

-- Seed iniciales (los 9 temas del proyecto Tinta)
INSERT INTO topics (name, slug, description, icon, size_mb) VALUES
    ('Física',                     'fisica',        'Mecánica, termodinámica, electromagnetismo y física moderna.', '🔭', 180),
    ('Matemáticas',                'matematicas',   'Álgebra, cálculo, estadística y geometría.',                    '📐', 220),
    ('Química',                    'quimica',       'Química general, orgánica e inorgánica.',                       '⚗️', 150),
    ('Biología',                   'biologia',      'Biología celular, genética, ecología.',                         '🧬', 170),
    ('Medicina y anatomía',        'medicina',      'Anatomía, fisiología, farmacología.',                           '⚕️', 250),
    ('Historia',                   'historia',      'Historia universal y de México.',                               '📜', 140),
    ('Literatura y humanidades',   'literatura',    'Clásicos, análisis literario y filosofía.',                     '📚', 160),
    ('Programación e informática', 'programacion',  'Algoritmos, estructuras de datos, lenguajes de programación.',  '💻', 130),
    ('Idiomas',                    'idiomas',       'Inglés y francés.',                                             '🌍', 110)
ON CONFLICT (slug) DO NOTHING;
