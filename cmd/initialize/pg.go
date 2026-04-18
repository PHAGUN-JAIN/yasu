package initialize

import (
	"database/sql"
	"urlshort/migrations"

	_ "github.com/lib/pq" // Import PostgreSQL driver
	// ↑ blank import — just runs the package's init() to register the driver
	"log"
	"time"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func GetPostgressClient() *sql.DB {

	log.Println("Attempting to connect with: host=postgres user=postgres")

	//connect to database
	// db, err := sql.Open("postgres", "host=postgres port=5432 user=postgres password=postgres dbname=yasu sslmode=disable")
	db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=postgres password=postgres dbname=yasu sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	err = db.Ping()
	if err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	return db
}

// handle migrations here as well, or create a separate function for it and call it from app.go

func RunMigrations(db *sql.DB) error {
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	entries, _ := migrations.Postgres.ReadDir("postgres")
	for _, e := range entries {
		log.Printf("found migration file: %s", e.Name())
	}

	src, err := iofs.New(migrations.Postgres, "postgres")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return err
	}

	m.Log = &MigrateLogger{}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

type MigrateLogger struct{}

func (l *MigrateLogger) Printf(format string, v ...interface{}) {
	log.Printf("[migrate] "+format, v...)
}

func (l *MigrateLogger) Verbose() bool {
	return true // set false to suppress debug-level logs
}
