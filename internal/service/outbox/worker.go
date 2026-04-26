package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// ClickEvent represents a single redirect event
type ClickEvent struct {
	ID        int64
	Code      string
	Country   string
	Device    string
	Referer   string
	IP        string
	Source    string // "rest" | "mcp" | "a2a"
	City      string // from City()
	ASN       string // from ASN() — useful to know if traffic is from cloud/bots
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
		flushEvery: 30 * time.Second,
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

	if err := bulkInsert(ctx, w.pg, batch); err != nil {
		log.Printf("worker: ERROR: failed to flush %d events to Postgres: %v", len(batch), err)
		return
	}
	log.Printf("worker: flushed %d events to Postgres", len(batch))
}

func bulkInsert(ctx context.Context, db *sql.DB, events []ClickEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Ensure rollback if any step fails
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("rollback error: %v", err)
		}
	}()

	// build: INSERT INTO click_events(code, device, referer, clicked_at)
	// VALUES ($1,$2,$3,$4), ($5,$6,$7,$8), ...
	var query strings.Builder
	query.WriteString("INSERT INTO click_events(code, ip, device, referer, country, city, asn, source, clicked_at) VALUES ")
	args := make([]any, 0, len(events)*9)

	for i, e := range events {
		if i > 0 {
			query.WriteString(",")
		}
		base := i * 9
		fmt.Fprintf(&query, "($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9)
		args = append(args, e.Code, e.IP, e.Device, e.Referer, e.Country, e.City, e.ASN, e.Source, e.ClickedAt)
	}

	if _, err := tx.ExecContext(ctx, query.String(), args...); err != nil {
		return err
	}
	return tx.Commit()
}
