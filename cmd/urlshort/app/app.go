package app

import (
	"log"
	// "urlshort/cmd/urlshort/initialize"
	router "urlshort/internal/http"
)

// TODO: wire everything together here
// 1. Load config from env
// 2. Connect to Postgres (store/pg.go)
// 3. Connect to ClickHouse (store/ch.go)
// 4. Connect to Redis (store/redis.go)
// 5. Start buffer flush goroutine (buffer/buffer.go)
// 6. Register routes (handler/routes.go)
// 7. Start HTTP server

func RunApplication() {
	log.Println("Starting URL Shortener Application...")
	// clients := initialize.GetClients()

	//create the table if it doesn't exist
	// _, err := clients.PgClient.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	err := router.RunYasuServer("8000")
	if err != nil {
		log.Fatal(err)
	}

}
