package initialize

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // Import PostgreSQL driver
)

func GetPostgressClient() *sql.DB {

	//connect to database
	db, err := sql.Open("postgres", "host=postgres port=5432 user=user password=pass dbname=urlshort sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()
	return db
}

// handle migrations here as well, or create a separate function for it and call it from app.go
