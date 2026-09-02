CREATE TABLE IF NOT EXISTS students (
    id         SERIAL PRIMARY KEY,
    nim        INTEGER NOT NULL,
    name       VARCHAR(100) NOT NULL,
    grade      DOUBLE PRECISION NOT NULL CHECK (grade >= 0 AND grade <= 100),
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT students_nim_key UNIQUE (nim),
    CONSTRAINT students_name_not_blank CHECK (length(btrim(name)) > 0)
);

CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));
