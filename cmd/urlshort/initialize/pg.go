package initialize

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // Import PostgreSQL driver
)

func init() {
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

	err = db.Ping()
	if err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	return db
}

// handle migrations here as well, or create a separate function for it and call it from app.go
