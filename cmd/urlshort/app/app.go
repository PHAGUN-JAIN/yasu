package app

import (
	"log"
	"urlshort/cmd/initialize"
	router "urlshort/internal/http"
)

// TODO: wire everything together here
// 1. Load config from env
// 2. Connect to Postgres (store/pg.go) ✅
// 3. Connect to ClickHouse (store/ch.go)
// 4. Connect to Redis (store/redis.go)
// 5. Start buffer flush goroutine (buffer/buffer.go)
// 6. Register routes (handler/routes.go)  ✅
// 7. Start HTTP server (router/router.go) ✅

func RunApplication() {
	log.Println("Running database migrations...")
	err := initialize.RunPostgresMigrations()
	if err != nil {
		log.Fatal(err)
	}

	err = initialize.RunClickHouseMigrations()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting URL Shortener Application...")
	err = router.RunYasuServer("8000")
	if err != nil {
		log.Fatal(err)
	}
}
