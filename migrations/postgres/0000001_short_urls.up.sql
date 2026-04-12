CREATE TABLE IF NOT EXISTS short_urls (
    code        VARCHAR(12) PRIMARY KEY,
    original_url TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT now(),
    expires_at  TIMESTAMPTZ,
    is_custom    BOOLEAN DEFAULT FALSE
);