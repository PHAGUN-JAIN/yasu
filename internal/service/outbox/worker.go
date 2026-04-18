package outbox

import (
	"context"
	"database/sql"
	"log"
	"time"
	"urlshort/internal/store"
)

// ClickEvent represents a single redirect event
type ClickEvent struct {
	Code      string
	Country   string
	Device    string
	Referer   string
	ClickedAt time.Time
}

// Worker drains the channel into Postgres, then Postgres into ClickHouse
type Worker struct {
	eventChan chan ClickEvent
	pg        *sql.DB
	ch        *store.CHStore

	batchSize  int           // flush when channel hits this size
	flushEvery time.Duration // flush on this interval even if not full
}

func NewWorker(pg *sql.DB, ch *store.CHStore) *Worker {
	return &Worker{
		eventChan:  make(chan ClickEvent, 10000), // buffered — redirects never block
		pg:         pg,
		ch:         ch,
		batchSize:  1000,
		flushEvery: 2 * time.Second,
	}
}

// Enqueue is called by the redirect handler — non-blocking
func (w *Worker) Enqueue(e ClickEvent) {
	select {
	case w.eventChan <- e:
		// queued successfully
	default:
		// channel full — drop and log
		// this should never happen at normal traffic
		log.Printf("WARN: event channel full, dropping click event for code %s", e.Code)
	}
}

// Start runs the drain loop — call this in main.go as a goroutine
func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.flushEvery)
	defer ticker.Stop()

	var batch []ClickEvent

	for {
		select {
		case event := <-w.eventChan:
			batch = append(batch, event)
			if len(batch) >= w.batchSize {
				w.flush(ctx, batch)
				batch = nil // reset
			}

		case <-ticker.C:
			if len(batch) > 0 {
				w.flush(ctx, batch)
				batch = nil // reset
			}

		case <-ctx.Done():
			// drain remaining events before shutdown
			if len(batch) > 0 {
				log.Printf("shutting down — flushing %d remaining events", len(batch))
				w.flush(ctx, batch)
			}
			return
		}
	}
}

// flush writes batch to Postgres outbox
func (w *Worker) flush(ctx context.Context, batch []ClickEvent) {
	if len(batch) == 0 {
		return
	}

	// TODO: batch insert into click_events Postgres table
	// hint: use pgx CopyFrom for fastest bulk insert into Postgres
	// pgx CopyFrom is faster than a multi-row INSERT for large batches
	log.Printf("flushing %d events to Postgres outbox", len(batch))
}
