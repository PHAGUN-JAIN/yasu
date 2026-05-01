-- Step 1  the target table that stores pre-aggregated data
-- AggregatingMergeTree is the correct engine for materialized views
CREATE TABLE IF NOT EXISTS clicks_daily
(
    short_code  String,
    date        Date,
    country     LowCardinality(String),
    city        LowCardinality(String),
    device      LowCardinality(String),
    source      LowCardinality(String),
    clicks      AggregateFunction(count)   -- stores partial aggregation state
)
ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (short_code, date, country, city, device, source);

-- Step 2  the materialized view that feeds clicks_daily
-- fires on every INSERT into clicks
CREATE MATERIALIZED VIEW IF NOT EXISTS clicks_daily_mv
TO clicks_daily
AS
SELECT
    short_code,
    toDate(clicked_at)          AS date,
    country,
    city,
    device,
    source,
    countState()                AS clicks   -- countState() not count()
FROM clicks
GROUP BY short_code, date, country, city, device, source;