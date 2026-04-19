-- clicks table
-- Run this manually or wire into your startup

CREATE TABLE IF NOT EXISTS clicks
(
    short_code  String,
    ip          String,
    country     LowCardinality(String),  -- why LowCardinality?
    device      LowCardinality(String),
    referer     String,
    clicked_at  DateTime
)
ENGINE = MergeTree()
ORDER BY (short_code, clicked_at);