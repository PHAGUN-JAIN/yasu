package repository

import (
	"context"
	"errors"
	"log"
	"time"
	"urlshort/migrations"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse" // ClickHouse driver for golang-migrate
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// reference:https://clickhouse.com/docs/integrations/go

func GetChClient() driver.Conn {
	log.Printf("Attempting to connect with: clickhouse:9000")

	ctx := context.Background()
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"clickhouse:9000"},
		Auth: clickhouse.Auth{
			Database: "yasu_analytics",
			Username: "yasu",
			Password: "yasu",
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct{ Name, Version string }{
				{Name: "yasu-go-client", Version: "0.1"},
			},
		},
	})

	if err != nil {
		log.Printf("Failed to initialize ClickHouse client: %v", err)
		return nil
	}

	if err := conn.Ping(ctx); err != nil {
		// Handle specific ClickHouse server exceptions
		if exception, ok := err.(*clickhouse.Exception); ok {
			log.Fatalf("ClickHouse Exception [%d]: %s", exception.Code, exception.Message)
		} else {
			log.Fatalf("ClickHouse Ping failed: %v", err)
		}
		return nil
	}

	return conn
}

func RunClickHouseMigrations() error {
	conn := GetChClient()
	if conn == nil {
		return errors.New("failed to connect to ClickHouse for migrations")
	}
	defer (conn).Close()

	entries, _ := migrations.ClickHouse.ReadDir("clickhouse")
	for _, e := range entries {
		log.Printf("found migration file CH: %s", e.Name())
	}

	src, err := iofs.New(migrations.ClickHouse, "clickhouse")
	if err != nil {
		return err
	}

	dsn := "clickhouse://yasu:yasu@clickhouse:9000/yasu_analytics?x-multi-statement=true"
	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return err
	}

	m.Log = &MigrateLogger{}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// CHStore handles analytics writes and reads
// Use: github.com/ClickHouse/clickhouse-go/v2

// TODO: implement BatchInsert(ctx, []ClickEvent) error
// Key learning: use clickhouse.PrepareBatch for efficient bulk inserts
// Read the clickhouse-go docs on batch inserts - it's different from pgx

// TODO: implement GetStats(ctx, shortCode string) (*Stats, error)
// Hint: think about what aggregations you need
// SELECT count(), uniq(ip), any(country) ... GROUP BY toDate(clicked_at)

type Stats struct {
	Code        string            `json:"code"`
	TotalClicks uint64            `json:"total_clicks"`
	ByCountry   map[string]uint64 `json:"by_country"`
	ByDevice    map[string]uint64 `json:"by_device"`
	BySource    map[string]uint64 `json:"by_source"`
	ByDay       []DayStat         `json:"by_day"`
}

type DayStat struct {
	Date   time.Time `json:"date"`
	Clicks uint64    `json:"clicks"`
}
