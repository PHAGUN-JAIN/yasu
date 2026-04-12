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
-- TODO: think carefully about ORDER BY
-- This determines how ClickHouse sorts data on disk
-- and directly affects query performance
-- Hint: what column do you filter/group by most?
ORDER BY (short_code, clicked_at);

-- TODO: try adding a materialized view later
-- that pre-aggregates clicks per day per short_code
-- this is a key ClickHouse exam concept