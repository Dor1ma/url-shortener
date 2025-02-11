CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    original_url TEXT NOT NULL UNIQUE,
    short_url VARCHAR(10) NOT NULL UNIQUE
);

CREATE INDEX idx_short_url ON urls (short_url);
CREATE INDEX idx_original_url ON urls (original_url);
