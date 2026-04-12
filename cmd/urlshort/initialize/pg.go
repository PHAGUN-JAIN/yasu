package initialize

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq" // Import PostgreSQL driver
)

func init() {
	// TODO: add tablest creation and migration logic here, or create a separate function for it and call it from app.go
	log.Println("Initializing PostgreSQL client...")
}

func GetPostgressClient() *sql.DB {

	log.Println("Attempting to connect with: host=postgres user=postgres")

	//connect to database
	db, err := sql.Open("postgres", "host=postgres port=5432 user=postgres password=postgres dbname=yasu sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()

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
