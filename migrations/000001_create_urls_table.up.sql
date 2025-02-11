CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    original_url TEXT NOT NULL UNIQUE,
    short_url VARCHAR(10) NOT NULL UNIQUE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_short_url ON urls (short_url);
CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON urls (original_url);
