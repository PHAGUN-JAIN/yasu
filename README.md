# YASU — Yet Another Short URL

A production-grade URL shortener with real-time analytics, built with Go. Designed to explore the intersection of operational databases (Postgres, Redis) and analytical databases (ClickHouse), with both a REST API and an MCP interface for AI agent integration.

## What it does

- Shortens URLs with 7-char Base62 codes and optional expiry
- Tracks every redirect — country, city, ASN, device, referrer, source
- Streams click events asynchronously to ClickHouse via a Postgres outbox
- Exposes a REST API and an MCP server so AI agents can use it directly

## Stack

| Component | Technology | Why |
|---|---|---|
| URL storage | Postgres | Simple key-value with unique constraint, no joins needed |
| Cache | Redis | Hot URLs never touch the DB |
| Analytics | ClickHouse | Columnar storage, fast aggregations on click data |
| Event pipeline | Postgres outbox + SKIP LOCKED | Crash-safe, no Kafka needed |
| GeoIP | MaxMind GeoLite2 | IP → country/city/ASN, raw IP discarded (GDPR) |
| AI interface | MCP (official Go SDK) | Claude and other agents can shorten URLs autonomously |

## Architecture

```
                        ┌─────────────────────────────────┐
                        │           Go Service             │
                        │                                  │
  REST  ──────────────▶ │  Handler → Service → Postgres    │
  MCP   ──────────────▶ │                    → Redis       │
  A2A   ──────────────▶ │                                  │
                        │  Worker (channel buffer)         │
                        │    └─▶ Postgres outbox           │
                        │         └─▶ ClickHouse           │
                        └─────────────────────────────────┘
```

Every redirect enqueues a `ClickEvent` into a buffered Go channel — the redirect itself never waits for analytics writes. A background worker drains the channel into Postgres in batches. A separate drainer polls the outbox with `SKIP LOCKED` and inserts into ClickHouse.

## Design decisions

**302 not 301** — 301 is a permanent redirect that browsers cache. Future clicks never reach the server, so analytics break. 302 ensures every click is tracked.

**Postgres over NoSQL** — a URL shortener is a key-value lookup on a short code with a unique constraint. Postgres with an index handles millions of lookups/sec easily. DynamoDB or Cassandra only make sense at Bit.ly scale.

**Outbox pattern over Kafka** — single producer, single consumer. The Postgres outbox with `SKIP LOCKED` gives crash durability and horizontal worker scaling without the operational overhead of running a Kafka cluster. Kafka would be the right call if multiple services were producing events or replay was needed.

**Random Base62 over hash-based generation** — `Base62(MD5(url))` creates check-then-insert races under concurrency and hash collisions across different URLs. Random Base62 with `ON CONFLICT DO NOTHING` and a retry loop is simpler and correct.

**GeoIP at the edge, IP discarded** — the raw IP is resolved to country/city/ASN in the handler and never written to disk. Only the derived data reaches ClickHouse.

## Local development

**Prerequisites:** Docker, Go 1.22+, Make

```sh
# first time setup
make setup       # configure git hooks
make geoip       # download MaxMind databases (needs MAXMIND_KEY env var)

# run everything
make up

# verify
curl 127.0.0.1:8000/api/v1/health -ki
```

## API reference

### Shorten a URL
```sh
curl -X POST http://localhost:8000/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

### Redirect
```sh
curl -v http://localhost:8000/api/v1/{code}
# HTTP/1.1 302 Found
# Location: https://example.com
```

### MCP (AI agent interface)

Start a session:
```sh
curl -X POST http://localhost:8001/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0", "id": 1, "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "clientInfo": {"name": "curl-test", "version": "1.0.0"},
      "capabilities": {}
    }
  }'
```

Call a tool (use session ID from initialize response):
```sh
curl -X POST http://localhost:8001/mcp \
  -H "Content-Type: application/json" \
  -H "Mcp-Session-Id: {session_id}" \
  -d '{
    "jsonrpc": "2.0", "id": 2, "method": "tools/call",
    "params": {"name": "shorten_url", "arguments": {"url": "https://example.com"}}
  }'
```

Available tools: `shorten_url`, `resolve_url`, `get_stats`

## Database access

```sh
# Postgres
docker exec -it compose-postgres-1 psql -U postgres -d yasu

# ClickHouse
docker exec -it compose-clickhouse-1 clickhouse-client
```

## Short code generation

Seven random characters from `[a-zA-Z0-9]` using `crypto/rand`. This gives a keyspace of 62^7 ≈ 3.5 trillion codes. Collisions are handled with `INSERT ... ON CONFLICT DO NOTHING` and a retry loop — at this keyspace size, retries should never occur in practice.

A bias exists because 256 (byte range) doesn't divide evenly by 62 — the first 8 characters of the alphabet appear with 1.95% probability vs the fair 1.56%. The difference is 0.39% and has no practical consequence for a URL shortener.