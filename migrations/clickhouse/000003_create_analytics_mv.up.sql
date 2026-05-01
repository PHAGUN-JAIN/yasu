-- Step 1: the target table that stores pre-aggregated data
-- AggregatingMergeTree is the correct engine for the target aggregate table used by the materialized view
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

-- Step 2: Backfill historical data into clicks_daily BEFORE creating the materialized view.
-- The materialized view (Step 3) only captures rows inserted into clicks AFTER it is created.
-- Any rows that already existed in clicks before this migration are invisible to the MV, so
-- analytics queries reading from clicks_daily would return 0 or incomplete counts for historical
-- periods. By running this backfill before the MV is created, we ensure there is no overlap
-- between the rows the MV will capture and the rows we insert here, preventing double-counting.
INSERT INTO clicks_daily
SELECT
    short_code,
    toDate(clicked_at)          AS date,
    country,
    city,
    device,
    source,
    countState()                AS clicks
FROM clicks
GROUP BY short_code, date, country, city, device, source;

-- Step 3: the materialized view that feeds clicks_daily
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