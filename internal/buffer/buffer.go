package buffer

import (
	"sync"
	"time"
)

// ClickEvent represents a single URL click to be stored in ClickHouse
type ClickEvent struct {
	// TODO: add fields
	// ShortCode string
	// IP        string
	// UserAgent string
	// Referer   string
	// Country   string  (hint: use a GeoIP lib like oschwald/geoip2-golang)
	// Device    string  (hint: parse UserAgent)
	// ClickedAt time.Time
}

// Buffer holds events in memory before flushing to ClickHouse
type Buffer struct {
	mu     sync.Mutex
	events []ClickEvent

	maxSize    int           // flush when buffer reaches this size
	flushEvery time.Duration // flush on this interval even if not full

	// TODO: add a channel or reference to ClickHouse store for flushing
}

// TODO: implement Add(event ClickEvent)
// - lock, append, unlock
// - if len > maxSize, trigger flush

// TODO: implement Start(ctx context.Context)
// - runs in a goroutine
// - ticker fires every flushEvery duration
// - calls flush()

// TODO: implement flush()
// - lock, swap events slice with empty one, unlock  <-- why swap instead of clear?
// - batch insert swapped slice to ClickHouse
// - handle insert error (log? retry? dead letter?)
