CREATE TABLE IF NOT EXISTS click_events (
    id          BIGSERIAL PRIMARY KEY,
    code        VARCHAR(12) NOT NULL,
    ip          INET DEFAULT '0.0.0.0',
    country     VARCHAR(2) DEFAULT 'XX',
    device      VARCHAR(50) DEFAULT 'unknown',
    referer     TEXT DEFAULT 'none',
    clicked_at  TIMESTAMPTZ DEFAULT now(),
    
    processed   BOOLEAN DEFAULT FALSE,
    
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

-- index for the SKIP LOCKED query
CREATE INDEX idx_click_events_unprocessed 
ON click_events (id) 
WHERE processed = false;