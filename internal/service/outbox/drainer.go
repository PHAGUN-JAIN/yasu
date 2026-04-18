package outbox

import (
	"context"
	"database/sql"
	"log"
	"time"
	"urlshort/internal/store"
)

// Drainer polls the Postgres outbox and inserts into ClickHouse
type Drainer struct {
	pg *sql.DB
	ch *store.CHStore

	pollEvery time.Duration
}

func NewDrainer(pg *sql.DB, ch *store.CHStore) *Drainer {
	return &Drainer{
		pg:        pg,
		ch:        ch,
		pollEvery: 5 * time.Second,
	}
}

// Start runs the poll loop — call this in main.go as a goroutine
func (d *Drainer) Start(ctx context.Context) {
	log.Printf("starting the events drain to clickhouse...")
	ticker := time.NewTicker(d.pollEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.drain(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (d *Drainer) drain(ctx context.Context) {
	// tx, err := d.pg.Begin(ctx)
	// if err != nil {
	// 	log.Printf("drainer: failed to begin tx: %v", err)
	// 	return
	// }
	// defer tx.Rollback(ctx)

	// TODO: SELECT with SKIP LOCKED
	// rows, err := tx.Query(ctx, `
	//     SELECT id, code, country, device, referer, clicked_at
	//     FROM click_events
	//     WHERE processed = false
	//     ORDER BY id
	//     LIMIT 1000
	//     FOR UPDATE SKIP LOCKED
	// `)

	// TODO: scan rows into []ClickEvent

	// TODO: batch insert into ClickHouse
	// if insert fails — return, don't commit, Postgres retries next tick

	// TODO: mark rows as processed
	// UPDATE click_events SET processed = true WHERE id = ANY($1)

	// TODO: tx.Commit(ctx)
}
