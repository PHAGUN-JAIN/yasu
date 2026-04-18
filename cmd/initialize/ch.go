package initialize

import (
	"context"
	"errors"
	"log"
	"urlshort/migrations"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse" // ClickHouse driver for golang-migrate
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// reference:https://clickhouse.com/docs/integrations/go

func GetChClient() driver.Conn {
	ctx := context.Background()
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"clickhouse:9000"},
		Auth: clickhouse.Auth{
			Database: "yasu_analytics",
			Username: "default",
			Password: "default",
		},
		// Detailed client info for ClickHouse system logs
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
			log.Printf("ClickHouse Exception [%d]: %s", exception.Code, exception.Message)
		} else {
			log.Printf("ClickHouse Ping failed: %v", err)
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
		log.Printf("found migration file: %s", e.Name())
	}

	src, err := iofs.New(migrations.ClickHouse, "clickhouse")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, "clickhouse://clickhouse:9000?database=yasu_analytics")
	if err != nil {
		return err
	}

	m.Log = &MigrateLogger{}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
