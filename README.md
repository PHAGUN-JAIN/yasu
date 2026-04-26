# YASU — Yet Another Short URL

A production-grade URL shortener with real-time analytics, built with Go. Explores the intersection of operational databases (Postgres, Redis) and analytical databases (ClickHouse), with both a REST API and an MCP server for AI agent integration.

## Features

- Shortens URLs with 7-char Base62 codes and configurable expiry
- Tracks every redirect — country, city, ASN, device, referrer, and request source
- Streams click events asynchronously to ClickHouse via a Postgres outbox (no Kafka)
- GeoIP resolution at the edge — raw IP discarded, only derived data stored (GDPR)
- REST API for traditional clients and MCP server for AI agents

## Stack

| Component | Technology | Why |
|---|---|---|
| URL storage | Postgres | Key-value lookup with unique constraint, no joins needed |
| Cache | Redis | Hot URLs never touch the DB on redirect |
| Analytics | ClickHouse | Columnar storage built for fast aggregations on event data |
| Event pipeline | Postgres outbox + SKIP LOCKED | Crash-safe batching without Kafka |
| GeoIP | MaxMind GeoLite2 | IP → country/city/ASN lookup |
| AI interface | MCP (official Go SDK) | AI agents can shorten and resolve URLs autonomously |

## Architecture

```
  REST ─────────────────┐
  MCP ──────────────────┤
                        ▼
                 ┌─────────────┐
                 │  Go Service  │
                 │             │──▶ Postgres (URL store)
                 │             │──▶ Redis (cache)
                 │             │
                 │  Worker     │──▶ Postgres outbox
                 └─────────────┘         │
                                         ▼
                                    ClickHouse
```

Every redirect enqueues a `ClickEvent` into a buffered Go channel — the redirect path never waits for analytics writes. A background worker drains the channel into a Postgres outbox in batches. A separate drainer polls the outbox with `SKIP LOCKED` and inserts into ClickHouse.

## Design decisions

**302 not 301** — 301 is cached by browsers, meaning future clicks never reach the server and analytics break. 302 ensures every redirect is tracked.

**Postgres over NoSQL** — a URL shortener is a key-value lookup on a short code with a unique constraint. Postgres with an index on `code` handles millions of lookups/sec easily. DynamoDB or Cassandra only make sense at Bit.ly scale (billions of rows, multi-region).

**Outbox pattern over Kafka** — single producer, single consumer. The Postgres outbox with `SKIP LOCKED` gives crash durability and supports horizontal worker scaling without the operational overhead of running a Kafka cluster. The right time to add Kafka is when multiple services are producing events or event replay is needed — neither applies here.

**Random Base62 over hash-based generation** — `Base62(MD5(url))` causes check-then-insert race conditions under concurrency and hash collisions across different URLs. Random Base62 with `ON CONFLICT DO NOTHING` and a retry loop handles collisions correctly with no coordination overhead. At 62^7 ≈ 3.5 trillion possible codes, collisions are negligible.

**Channel buffer before outbox** — the redirect path only does a non-blocking channel send (nanoseconds). The worker goroutine drains the channel into Postgres in batches, decoupling write latency from redirect latency entirely.

**GeoIP at the edge** — IP is resolved to country/city/ASN in the handler before the event is enqueued. The raw IP never reaches the outbox or ClickHouse — only the derived geographic data is stored.

## Local development

**Prerequisites:** Docker, Go 1.22+, Make

```sh
# first time setup
make setup                          # configure git hooks
make geoip MAXMIND_KEY=your_key     # download MaxMind GeoLite2 databases

# start everything
make up

# verify
curl http://localhost:8000/api/v1/health -ki
```

## API

### Shorten a URL
```sh
curl -X POST http://localhost:8000/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'

# response
{"code": "zEuZJsq", "short_url": "http://localhost:8000/zEuZJsq"}
```

### Redirect
```sh
curl -v http://localhost:8000/api/v1/{code}
# HTTP/1.1 302 Found
# Location: https://example.com
```

### Analytics
```sh
curl http://localhost:8000/api/v1/stats/{code}

# response
{
  "code": "zEuZJsq",
  "total_clicks": 42,
  "by_country": {"US": 20, "IN": 15, "GB": 7},
  "by_device":  {"desktop": 30, "mobile": 10, "cli": 2},
  "by_source":  {"rest": 38, "mcp": 4},
  "by_day": [
    {"date": "2026-04-25", "clicks": 28},
    {"date": "2026-04-26", "clicks": 14}
  ]
}
```

### MCP (AI agent interface)

Initialize a session:
```sh
curl -X POST http://localhost:8001/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "clientInfo": {"name": "curl-test", "version": "1.0.0"},
      "capabilities": {}
    }
  }'
```

Call a tool (pass session ID from initialize response):
```sh
curl -X POST http://localhost:8001/mcp \
  -H "Content-Type: application/json" \
  -H "Mcp-Session-Id: {session_id}" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "shorten_url",
      "arguments": {"url": "https://example.com"}
    }
  }'
```

Available MCP tools: `shorten_url`, `resolve_url`, `get_stats`

## Database access

```sh
# Postgres
docker exec -it compose-postgres-1 psql -U postgres -d yasu

# ClickHouse
docker exec -it compose-clickhouse-1 clickhouse-client
```

## Contributing

Issues and PRs are welcome. Run `make lint` and `make fmt-fix` before opening a PR — the CI pipeline enforces both.

## License

MIT