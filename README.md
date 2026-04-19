# YASU - Yet Another Short URL
urlshortner service — URL shortener with ClickHouse analytics

## Build order

Phase 1 — Core flow (no analytics yet) ✅

  1. store/pg.go        → Save and FindByCode
  2. handler/handler.go → ShortenHandler + RedirectHandler (skip cache for now)
  3. cmd/server/main.go → wire up, test with curl

Phase 2 — Caching ✅

  4. store/redis.go     → Set and Get
  5. handler/handler.go → add Redis cache check to RedirectHandler

Phase 3 — Analytics (the ClickHouse part) ✅

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

## Short Code Generation — Approaches & Trade-offs

There are several ways to generate short codes. Here's why we chose what we chose.

Approaches considered

### Random Base62 (what this project uses)
Generate 7 random characters from [a-zA-Z0-9] using crypto/rand. Insert with
ON CONFLICT DO NOTHING and retry on collision. Simple, unpredictable, and safe
under concurrency. Collision probability at 7 chars is astronomically low until
you reach Bit.ly scale (billions of URLs).

### Base62(MD5(url))
Hash the original URL and take the first 7 chars. Sounds clever — same URL always
gets the same code. In practice, fragile. Different URLs can share the same first
7 chars of MD5 (collision), and concurrent inserts create check-then-insert race
conditions. Avoid.

### KGS — Key Generation Service
A separate service pre-generates millions of codes and stores them in an
unused_keys table. App servers pull batches of 1000 keys into memory and consume
one per request — zero DB collision checks. Correct at Twitter/Bit.ly scale.
Overkill for everything else — you're running an extra service to solve a problem
you don't have yet.

### Auto-increment + Base62
Encode a Postgres SERIAL integer into Base62. Zero collisions, very short codes.
Fatal flaw: sequential codes let anyone enumerate all URLs in your system. Never use for a public URL shortener.

### ULID/UUID as the code
Globally unique and collision-free but 26-36 chars long — defeats the purpose of a short URL. ULIDs are used in this project as internal primary keys (sortable,opaque, never exposed), not as the user-facing short code.

### What this project does
Random Base62 for the short code (user-facing, 7 chars).
ULID as the internal primary key (never exposed, sortable by time).
Two different jobs, two different columns.


Now with 62 characters and 0-255:
```
256 / 62 = 4 remainder 8
```

So 0-255 divides into 62 cleanly 4 times (248 numbers), with 8 left over:
```
numbers 0-247   → map fairly to all 62 chars (each gets 4 chances)

numbers 248-255 → map to chars 'a','b','c','d','e','f','g','h' only
                  these 8 chars get a 5th extra chance
```

So a through h are very slightly more likely than the rest.
How much does it actually matter?
```
fair probability:        4/256 = 1.5625%
biased probability:      5/256 = 1.9531%
difference:              0.39%
```
Across 7 characters in a URL shortener — completely negligible. Nobody can exploit this. It's a theoretical impurity, not a real bug.

The fix (rejection sampling) just means:
```
"If you drew a number in the unfair zone (248-255), throw it away and draw again."
That's it. Discard the biased numbers, only use the fair ones.
```

### When to upgrade

Switch to KGS when: you have multiple app servers burning through codes so fast that retry collisions become measurable. At 7-char Base62 that's roughly 3.5 trillion possible codes — you won't hit this problem for a very long time.

# notes

The pattern underneath most DB switches is the same:

- Why does this DB exist? (what problem it solves that others don't)

- Data model (relational, document, columnar, vector, key-value)

- Read/write tradeoffs (OLTP vs OLAP, latency vs throughput)

- When it breaks down (at what scale or use case it fails)


# Local Dev Guide

1. build image

 ```shell
 make build
 ```

2. run service with dependencies

  ```shell
  make up     /OR

  make dev-up
  ```

3. 

 ```shell
 curl 127.0.0.1:8000/api/v1/health -ki
 ```

4. How to test

  #### step 1 — shorten

  ```
  curl -X POST http://localhost:8000/api/v1/shorten \
    -H "Content-Type: application/json" \
    -d '{"url": "https://google.com"}'
  ```

  #### step 2 — copy the code from response, test redirect
  ```
  curl -v http://localhost:8000/api/v1/<code>
  ```

  ```
  should see:
   < HTTP/1.1 302 Found
   < Location: https://google.com
  ```

  #### step 3 — follow the redirect
  ```
  curl -L http://localhost:8000/api/v1/<code>
  ```

5. access postgres

```
docker exec -it compose-postgres-1 psql -U postgres
```

6. access clickhouse

```
docker exec -it compose-clickhouse-1 clickhouse-client
```
