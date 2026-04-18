CREATE TABLE IF NOT EXISTS click_events (
    id          BIGSERIAL PRIMARY KEY,
    code        VARCHAR(12) NOT NULL,
    country     VARCHAR(2),
    device      VARCHAR(50),
    referer     TEXT,
    clicked_at  TIMESTAMPTZ DEFAULT now(),
    processed   BOOLEAN DEFAULT FALSE
);

-- index for the SKIP LOCKED query
CREATE INDEX idx_click_events_unprocessed 
ON click_events (id) 
WHERE processed = false;