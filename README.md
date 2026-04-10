# urlshort — URL shortener with ClickHouse analytics

## Build order (suggested)

Phase 1 — Core flow (no analytics yet)
  1. store/pg.go        → Save and FindByCode
  2. handler/handler.go → ShortenHandler + RedirectHandler (skip cache for now)
  3. cmd/server/main.go → wire up, test with curl

Phase 2 — Caching
  4. store/redis.go     → Set and Get
  5. handler/handler.go → add Redis cache check to RedirectHandler

Phase 3 — Analytics (the ClickHouse part)
  6. Run migrations/001_clicks.sql against ClickHouse
  7. store/ch.go        → BatchInsert
  8. buffer/buffer.go   → Buffer with flush goroutine
  9. handler/handler.go → enqueue ClickEvent on redirect
  10. store/ch.go       → GetStats query
  11. handler/handler.go → StatsHandler

Phase 4 — Privacy & compliance
  12. Extract real client IP from X-Forwarded-For header (not the load balancer IP)
  13. Map IP → country via GeoIP (github.com/oschwald/geoip2-golang)
  14. Discard the raw IP — store only the country in ClickHouse (GDPR/CCPA)

## Run dependencies
  cd docker && docker compose up -d

## Deliberate design decisions (be ready to explain these in interviews)

### Why 302 and not 301?
301 = permanent redirect. Browser caches it — future clicks never hit your server.
302 = temporary redirect. Every click hits your server, so you can track analytics.
For this project, always use 302. Add a comment in RedirectHandler explaining this.

### Why Postgres for URLs and not NoSQL?
A URL shortener is a simple key-value lookup on a short code with a unique constraint.
Postgres with an index on `code` handles millions of lookups/sec easily.
DynamoDB/Cassandra only make sense at Bit.ly scale (billions of rows, multi-region ops).
Don't add complexity you don't need — this is the right call for 99% of real companies.

### Why no Kafka?
Single producer (this Go app) → single consumer (ClickHouse).
Kafka shines when multiple services produce events or you need replay.
The in-memory buffer gives you batching without the ops overhead.
Correct answer in interviews: "I'd add Kafka when I have multiple producers or need replay."

### Why not generate keys with Base62(MD5(url))?
Leads to collisions + "check-then-insert" race conditions under concurrent load.
Better: generate a random short code and use INSERT ... ON CONFLICT DO NOTHING
with a unique constraint in Postgres. Simple, correct, no KGS service needed.

## Key things to figure out yourself (intentional)
  - Why does buffer.flush() swap the slice instead of clearing it?
  - What happens to buffered events if the process crashes mid-flush?
  - Why LowCardinality(String) for country and device in ClickHouse?
  - What ORDER BY makes GetStats fast?
  - What's the difference between ClickHouse native port (9000) vs HTTP (8123)?


# notes

The pattern underneath most DB switches is the same:

- Why does this DB exist? (what problem it solves that others don't)

- Data model (relational, document, columnar, vector, key-value)

- Read/write tradeoffs (OLTP vs OLAP, latency vs throughput)

- When it breaks down (at what scale or use case it fails)

 