package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/lib/pq"
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
		pollEvery: 30 * time.Second,
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
	    SELECT id, code, ip, country, device, referer, clicked_at
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
	var ids []int64
	for rows.Next() {
		var event ClickEvent
		if err := rows.Scan(&event.ID, &event.Code, &event.IP, &event.Country, &event.Device, &event.Referer, &event.ClickedAt); err != nil {
			log.Printf("drainer: failed to scan row: %v", err)
			return
		}
		batch = append(batch, event)
		ids = append(ids, event.ID)
	}

	if len(batch) == 0 {
		log.Printf("drainer: no new events found")
		return
	}

	log.Printf("drainer: fetched %d events", len(batch))

	if err := ingestClickHouse(ctx, d.ch, batch); err != nil {
		log.Printf("drainer: failed to insert batch into ClickHouse: %v", err)
		return
	}

	_, err = tx.Exec(`UPDATE click_events SET processed = true, updated_at = NOW() WHERE id = ANY($1::bigint[])`, pq.Array(ids))
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

func ingestClickHouse(ctx context.Context, ch driver.Conn, b []ClickEvent) error {
	batch, err := ch.PrepareBatch(ctx, "INSERT INTO clicks (short_code, ip, country, device, referer, clicked_at)")
	if err != nil {
		return err
	}

	for _, event := range b {
		if err := batch.Append(event.Code, event.IP, event.Country, event.Device, event.Referer, event.ClickedAt); err != nil {
			return fmt.Errorf("clickhouse: append event for code %s: %w", event.Code, err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("clickhouse: send batch: %w", err)
	}

	return nil
}
