package outbox

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// Drainer polls the Postgres outbox and inserts into ClickHouse
type Drainer struct {
	pg        *sql.DB
	ch        driver.Conn
	pollEvery time.Duration
}

func NewDrainer(pg *sql.DB, ch driver.Conn) *Drainer {
	return &Drainer{
		pg:        pg,
		ch:        ch,
		pollEvery: 10 * time.Second,
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
	log.Printf("drainer: polling for new events...")

	tx, err := d.pg.Begin()
	if err != nil {
		log.Printf("drainer: failed to begin tx: %v", err)
		return
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
	    SELECT code, country, device, referer, clicked_at
	    FROM click_events
	    WHERE processed = false
	    ORDER BY id
	    LIMIT 1000
	    FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		log.Printf("drainer: failed to query click_events: %v", err)
		return
	}
	defer rows.Close()

	var batch []ClickEvent
	for rows.Next() {
		var event ClickEvent
		if err := rows.Scan(&event.Code, &event.Country, &event.Device, &event.Referer, &event.ClickedAt); err != nil {
			log.Printf("drainer: failed to scan row: %v", err)
			return
		}
		batch = append(batch, event)
	}

	if len(batch) == 0 {
		log.Printf("drainer: no new events found")
		return
	}

	log.Printf("drainer: fetched %d events", len(batch))

	if err := IngestClickHouse(d.ch, batch); err != nil {
		log.Printf("drainer: failed to insert batch into ClickHouse: %v", err)
		return
	}

	var codes []string
	for _, event := range batch {
		codes = append(codes, event.Code) // assuming code is unique identifier, replace with actual ID if available
	}

	_, err = tx.Exec(`UPDATE click_events SET processed = true WHERE code = ANY($1)`, codes)
	if err != nil {
		log.Printf("drainer: failed to mark events as processed: %v", err)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("drainer: failed to commit transaction: %v", err)
		return
	}

	log.Printf("drainer: successfully processed %d events", len(batch))

}

func IngestClickHouse(conn driver.Conn, batch []ClickEvent) error {
	// TODO: implement me

	return nil
}
