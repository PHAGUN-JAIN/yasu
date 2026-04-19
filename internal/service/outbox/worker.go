package outbox

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// ClickEvent represents a single redirect event
type ClickEvent struct {
	ID        int
	Code      string
	Country   string
	Device    string
	Referer   string
	IP        string
	ClickedAt time.Time
}

// Worker drains the channel into Postgres, then Postgres into ClickHouse
type Worker struct {
	eventChan  chan ClickEvent
	pg         *sql.DB
	batchSize  int           // flush when channel hits this size
	flushEvery time.Duration // flush on this interval even if not full
}

func NewWorker(pg *sql.DB) *Worker {
	return &Worker{
		eventChan:  make(chan ClickEvent, 1000), // buffered — redirects never block
		pg:         pg,
		batchSize:  1000,
		flushEvery: 5 * time.Second,
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
	log.Printf("Worker: started with batch size %d and flush interval %s", w.batchSize, w.flushEvery)
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
				log.Printf("worker: shutting down — flushing %d remaining events", len(batch))
				w.flush(ctx, batch)
			}
			return
		}
	}
}

// flush writes batch to Postgres outbox
func (w *Worker) flush(ctx context.Context, batch []ClickEvent) {
	log.Printf("worker: flushing %d events to Postgres outbox", len(batch))
	if len(batch) == 0 {
		return
	}

	if err := bulkInsert(w.pg, batch); err != nil {
		log.Printf("worker: ERROR: failed to flush %d events to Postgres: %v", len(batch), err)
		return
	}
	log.Printf("worker: flushed %d events to Postgres", len(batch))
}

func bulkInsert(db *sql.DB, click_events []ClickEvent) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	// Ensure rollback if any step fails
	defer tx.Rollback()

	// Prepare the statement once
	// PostgreSQL uses $1, $2, $3 placeholders, not ?
	stmt, err := tx.Prepare("INSERT INTO click_events(code, device, referer, clicked_at) VALUES($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the prepared statement for each event
	for _, u := range click_events {
		log.Printf("worker: Inserting click event: code=%s, device=%s, referer=%s, clicked_at=%s", u.Code, u.Device, u.Referer, u.ClickedAt)
		if _, err := stmt.Exec(u.Code, u.Device, u.Referer, u.ClickedAt); err != nil {
			return err
		}
	}

	return tx.Commit()
}
